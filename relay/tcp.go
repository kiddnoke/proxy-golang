package relay

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

const ReadDeadlineDuration = time.Second * 5
const KeepAlivePeriod = time.Second * 3
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
	ConnectInfoCallback func(time_stamp int64, rate float64, localAddress, RemoteAddress string, traffic float64, duration time.Duration)
	handlerId           int
}

func NewTcpRelayByProxyInfo(c *proxyinfo) (tp *TcpRelay, err error) {
	l, err := Listen("tcp", c.ServerPort, c.Cipher)
	return &TcpRelay{l: l, proxyinfo: c, handlerId: 0}, err
}
func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(KeepAlivePeriod)
	}
}
func (t *TcpRelay) Loop() {
	ConnCount := 0
	for t.running {
		_ = t.l.SetDeadline(time.Now().Add(time.Millisecond * AcceptTimeout))
		c, err := t.l.Accept()
		localConnectionStartStatmp := time.Now()
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

		go func(shadowConn net.Conn, localtimestamp time.Time) {
			t.handlerId++
			handlerId := t.handlerId
			defer func() {
				t.conns.Delete(shadowConn.RemoteAddr().String())
				shadowConn.Close()
				ConnCount--
			}()
			t.conns.Store(shadowConn.RemoteAddr().String(), shadowConn)
			ConnCount++
			tgt, err := socks.ReadAddr(shadowConn)

			if err != nil {
				t.proxyinfo.Printf("socks.ReadAddr Error [%s],handlerId[%d], local[%s]", err.Error(), handlerId, shadowConn.RemoteAddr())
				return
			}
			remoteConnectionStartStamp := time.Now()
			remoteConn, err := net.Dial("tcp", tgt.String())
			if err != nil {
				t.proxyinfo.Printf("net.Dial Error [%s], handlerId[%d], tgt[%s]", err.Error(), handlerId, tgt.String())
				return
			}
			defer func() {
				t.conns.Delete(remoteConn.RemoteAddr().String())
				remoteConn.Close()
			}()
			t.conns.Store(remoteConn.RemoteAddr().String(), remoteConn)
			t.Active()

			flow, PipeError, _duation := PipeThenCloseWithError(shadowConn, localConnectionStartStatmp, remoteConn, remoteConnectionStartStamp, func(up, down int) {
				if err := t.Limiter.WaitN(up + down); err != nil {
					t.proxyinfo.Printf("[%v] -> [%v] speedlimiter err:%v", tgt, shadowConn.RemoteAddr(), err)
				}
				t.AddTraffic(up, down, 0, 0)
			})

			defer func() {
				time_stamp := time.Now().UnixNano() / 1000000
				rate := float64(flow[1]) / 1024 / _duation[1].Seconds()
				ip := fmt.Sprintf("%v", shadowConn.RemoteAddr())
				website := fmt.Sprintf("%v", tgt)
				if flow[1] > 1024 && rate > 1.0 {
					t.proxyinfo.Printf("handler[%d] flow[%d kb] duration[%f] rate[%f k/s]  domain[%s] remoteaddr[%v] shadowErr[%v] remoteErr[%v]", handlerId, flow[1]/1024, _duation[1].Seconds(), rate, tgt, remoteConn.RemoteAddr(), PipeError[0].Error(), PipeError[1].Error())
					if t.ConnectInfoCallback != nil {
						t.ConnectInfoCallback(time_stamp, rate, ip, website, float64(flow[1]/1024), _duation[1])
					}
				}
			}()
		}(c, localConnectionStartStatmp)
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

func PipeWithError(tpcid int, left, right net.Conn, addTraffic func(n, m int)) (tcpId int, Error [2]error) {
	tcpId = tpcid
	errc0 := make(chan error, 1)
	errc1 := make(chan error, 1)
	pipe := func(front, back net.Conn, errc chan error, callback func(n int)) {
		buf := leakyBuf.Get()
		defer leakyBuf.Put(buf)
		for {
			front.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
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
	go pipe(left, right, errc0, func(n int) {
		addTraffic(n, 0)
	})
	go pipe(right, left, errc1, func(n int) {
		addTraffic(0, n)
	})
	Error[0] = <-errc0
	Error[1] = <-errc1
	return
}

func PipeThenCloseWithError(left net.Conn, lefttimestamp time.Time, right net.Conn, righttimestamp time.Time, addTraffic func(n, m int)) (flow [2]int64, Error [2]error, duration [2]time.Duration) {
	eleft := make(chan error, 1)
	eright := make(chan error, 1)
	getduration := func(timstamp time.Time) time.Duration {
		d := time.Now().Sub(timstamp)
		return d
	}

	pipe := func(left net.Conn, right net.Conn,
		ls, rs time.Time,
		errleft, erright chan error,
		dl, dr *time.Duration,
		flow *int64) {
		buf := leakyBuf.Get()
		defer leakyBuf.Put(buf)
		for {
			left.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
			n, err := left.Read(buf)
			if addTraffic != nil && n > 0 {
				addTraffic(n, 0)
			}
			if err != nil {
				errleft <- err
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					*dl = getduration(lefttimestamp) - ReadDeadlineDuration
				} else {
					*dl = getduration(lefttimestamp)
				}
				break
			}
			if n > 0 {
				*flow += int64(n)
				if _, err := right.Write(buf[0:n]); err != nil {
					erright <- err
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						*dr = getduration(righttimestamp) - ReadDeadlineDuration
					} else {
						*dr = getduration(righttimestamp)
					}
					break
				}
			}
		}

	}
	go pipe(left, right, lefttimestamp, righttimestamp, eleft, eright, &duration[0], &duration[1], &flow[0])
	go pipe(right, left, righttimestamp, lefttimestamp, eright, eleft, &duration[1], &duration[0], &flow[1])
	Error[0] = <-eleft
	Error[1] = <-eright
	return
}
