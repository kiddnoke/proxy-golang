package relay

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

const AcceptTimeout = 1000
const MaxAcceptConnection = 300

type listener struct {
	*net.TCPListener
	core.StreamConnCipher
}

func Listen(network string, Port int, ciph core.StreamConnCipher) (listener, error) {
	l, err := net.ListenTCP(network, &net.TCPAddr{Port: Port})
	return listener{l, ciph}, err
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.TCPListener.Accept()
	return l.StreamConn(c), err
}

type TcpRelay struct {
	l listener
	*proxyinfo
	conns               sync.Map
	ConnectInfoCallback func(time_stamp int64, rate int64, localAddress, RemoteAddress string, traffic int64, duration time.Duration)
	handlerId           int
}

func NewTcpRelayByProxyInfo(c *proxyinfo) (tp *TcpRelay, err error) {
	l, err := Listen("tcp", c.ServerPort, c.Cipher)
	return &TcpRelay{l: l, proxyinfo: c, handlerId: 0}, err
}
func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(time.Second * 30)
	}
}
func (t *TcpRelay) Loop() {
	ConnCount := 0
	for t.running {
		_ = t.l.SetDeadline(time.Now().Add(time.Millisecond * AcceptTimeout))
		c, err := t.l.Accept()
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				continue
			}
			continue
		}
		if ConnCount > MaxAcceptConnection {
			t.proxyinfo.Println("MaxAcceptConnection")
			c.Close()
			continue
		}
		go func() {
			t.handlerId++
			handlerId := t.handlerId
			defer func() {
				t.conns.Delete(c.RemoteAddr().String())
				c.Close()
				ConnCount--
			}()
			t.conns.Store(c.RemoteAddr().String(), c)
			ConnCount++
			tgt, err := socks.ReadAddr(c)

			if err != nil {
				t.proxyinfo.Printf("socks.ReadAddr [%s],handlerId[%d], local[%s]", err.Error(), handlerId, c.RemoteAddr())
				return
			}

			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				t.proxyinfo.Printf("net.Dial Error [%s], handlerId[%d], tgt[%s]", err.Error(), handlerId, tgt.String())
				return
			}
			defer func() {
				t.conns.Delete(rc.RemoteAddr().String())
				rc.Close()
			}()
			t.conns.Store(rc.RemoteAddr().String(), rc)
			t.Active()

			var flow int
			currstamp := time.Now()
			errorChannel := make(chan error, 1)

			go func() {
				tcpKeepAlive(c)
				t.proxyinfo.Printf("handlerId[%d] TransferStart [%s] => [%s]", handlerId, c.RemoteAddr(), tgt.String())
				errC2Rc := PipeThenClose(c, rc, func(up int) {
					flow += up
					if err := t.Limiter.WaitN(up); err != nil {
						t.proxyinfo.Printf("[%v] -> [%v] speedlimiter err:%v", tgt, c.RemoteAddr(), err)
					}
					t.AddTraffic(up, 0, 0, 0)
				})
				if errC2Rc == nil {
					t.proxyinfo.Printf("handlerId[%d] TransferFinish [%s] => [%s]", handlerId, c.RemoteAddr(), tgt.String())
				} else {
					t.proxyinfo.Printf("handlerId[%d] TransferFinish [%s] => [%s] with error:%s", handlerId, c.RemoteAddr(), tgt.String(), errC2Rc.Error())
				}
				errorChannel <- errC2Rc
			}()

			go func() {
				t.proxyinfo.Printf("handlerId[%d] TransferStart [%s] => [%s]", handlerId, tgt.String(), c.RemoteAddr())
				errRc2C := PipeThenClose(rc, c, func(down int) {
					flow += down
					if err := t.Limiter.WaitN(down); err != nil {
						t.proxyinfo.Printf("[%v] -> [%v] speedlimiter err:%v", tgt, c.RemoteAddr(), err)
					}
					t.AddTraffic(0, down, 0, 0)
				})
				if errRc2C == nil {
					t.proxyinfo.Printf("handlerId[%d] TransferFinish [%s] => [%s]", handlerId, tgt.String(), c.RemoteAddr())

				} else {
					t.proxyinfo.Printf("handlerId[%d] TransferFinish [%s] => [%s] with error:%s", handlerId, tgt.String(), c.RemoteAddr(), errRc2C.Error())
				}
				errorChannel <- errRc2C
			}()

			err = <-errorChannel
			defer func() {
				duration := time.Since(currstamp)
				time_stamp := time.Now().UnixNano() / 1000000
				rate := float64(flow) / duration.Seconds() / 1024
				t.proxyinfo.Printf("handlerId[%d] TransferStatisc [%s] <=> [%s]\trate[%f kb/s]\tflow[%d kb]\tDuration[%f sec]", handlerId, c.RemoteAddr(), tgt, rate, flow/1024, duration.Seconds())
				ip := fmt.Sprintf("%v", c.RemoteAddr())
				website := fmt.Sprintf("%v", tgt)
				if t.ConnectInfoCallback != nil && rate > 1.0 {
					t.ConnectInfoCallback(time_stamp, int64(rate), ip, website, int64(flow/1024), duration)
				}
			}()
		}()
	}
}
func (t *TcpRelay) Start() {
	t.running = true
	go t.Loop()
}
func (t *TcpRelay) Stop() {
	t.running = false
	t.conns.Range(func(key, value interface{}) bool {
		value.(net.Conn).Close()
		return true
	})
}
func (t *TcpRelay) Close() {
	if t.running == false {
		t.l.Close()
	}
}
func PipeThenClose(left, right net.Conn, addTraffic func(n int)) (err error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		n, err := left.Read(buf)
		if addTraffic != nil && n > 0 {
			addTraffic(n)
		}
		if n > 0 {
			if _, err := right.Write(buf[0:n]); err != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func PipeWithError(tpcid int, left, right net.Conn, addTraffic func(n, m int)) (tcpId int, err error) {
	tcpId = tpcid
	errc := make(chan error, 1)
	pipe := func(front, back net.Conn, callback func(n int)) {
		buf := leakyBuf.Get()
		defer leakyBuf.Put(buf)
		for {
			n, err := front.Read(buf)
			if addTraffic != nil && n > 0 {
				callback(n)
			}
			if n > 0 {
				if _, err := back.Write(buf[0:n]); err != nil {
					errc <- err
					break
				}
			}
			if err != nil {
				errc <- err
				break
			}
		}
	}
	go pipe(left, right, func(n int) {
		addTraffic(n, 0)
	})
	go pipe(right, left, func(n int) {
		addTraffic(0, n)
	})
	err = <-errc

	return
}
