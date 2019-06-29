package ss

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

var ReadDeadlineDuration = time.Second * 15
var WriteDeadlineDuration = ReadDeadlineDuration

const DialTimeoutDuration = time.Second * 20
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
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				continue
			}
			continue
		}
		if ConnCount > MaxAcceptConnection {
			t.Warn("MaxAcceptConnection")
			c.Close()
			continue
		}
		go func(shadowconn net.Conn) {
			t.handlerId++
			handlerId := t.handlerId
			defer func() {
				t.conns.Delete(shadowconn.RemoteAddr().String())
				shadowconn.Close()
				ConnCount--
			}()
			t.conns.Store(shadowconn.RemoteAddr().String(), shadowconn)
			ConnCount++
			tgt, err := socks.ReadAddr(shadowconn)

			if err != nil {
				t.Info("socks.ReadAddr Error [%s],handlerId[%d], local[%s]", err.Error(), handlerId, shadowconn.RemoteAddr())
				return
			}
			currstamp := time.Now()
			remoteconn, err := net.DialTimeout("tcp", tgt.String(), DialTimeoutDuration)
			if err != nil {
				t.Info("net.Dial Error [%s], handlerId[%d], tgt[%s]", err.Error(), handlerId, tgt.String())
				return
			}
			defer func() {
				t.conns.Delete(remoteconn.RemoteAddr().String())
				remoteconn.Close()
			}()
			t.conns.Store(remoteconn.RemoteAddr().String(), remoteconn)
			tcpKeepAlive(remoteconn)
			t.Active()

			var flow int
			s2rErrC := make(chan error, 1)
			go func() {
				err := PipeThenClose(shadowconn, remoteconn, func(n int) {
					if err := t.Limiter.WaitN(n); err != nil {
						t.Error("[%v] -> [%v] speedlimiter err:%v", shadowconn.RemoteAddr(), tgt, err)
					}
					t.AddTraffic(n, 0, 0, 0)
				})
				s2rErrC <- err
			}()
			r2sErr := PipeThenClose(remoteconn, shadowconn, func(n int) {
				flow += n
				if err := t.Limiter.WaitN(n); err != nil {
					t.Error("[%v] -> [%v] speedlimiter err:%v", tgt, shadowconn.RemoteAddr(), err)
				}
				t.AddTraffic(0, n, 0, 0)
			})

			defer func() {
				duration := time.Since(currstamp)
				if ne, ok := r2sErr.(net.Error); ok && ne.Timeout() {
					duration = duration - ReadDeadlineDuration
				}
				time_stamp := time.Now().UnixNano() / 1000000
				rate := float64(flow) / duration.Seconds() / 1024
				ip := fmt.Sprintf("%v", shadowconn.RemoteAddr())
				website := fmt.Sprintf("%v", tgt)
				if flow > 5*1024 {
					s2rErr := <-s2rErrC
					t.Info("handler[%d] flow[%f k] duration[%f sec] rate[%f kb/s] domain[%v] remoteaddr[%v] s2rErr[%v] r2sErr[%v]", handlerId, float64(flow)/1024.0, duration.Seconds(), rate, tgt, remoteconn.RemoteAddr(), s2rErr.Error(), r2sErr.Error())
					if t.ConnectInfoCallback != nil && flow > 10*1024 {
						t.ConnectInfoCallback(time_stamp, rate, ip, website, float64(flow)/1024.0, duration)
					}
				}
			}()
		}(c)
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

func PipeThenClose(left, right net.Conn, addTraffic func(n int)) (netErr error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		left.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
		n, err := left.Read(buf)
		if addTraffic != nil && n > 0 {
			addTraffic(n)
		}
		if n > 0 {
			if _, err := right.Write(buf[0:n]); err != nil {
				netErr = err
				break
			}
		}
		if err != nil {
			netErr = err
			break
		}
	}
	return
}

func PipeWithError(left, right net.Conn, addTraffic func(n, m int)) (err error) {
	errc := make(chan error, 1)
	pipe := func(front, back net.Conn, callback func(n int)) {
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
	go pipe(left, right, func(n int) {
		addTraffic(n, 0)
	})
	go pipe(right, left, func(n int) {
		addTraffic(0, n)
	})
	err = <-errc
	return
}
