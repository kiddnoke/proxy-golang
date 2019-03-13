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
const MaxAcceptConnection = 400

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
}

func NewTcpRelayByProxyInfo(c *proxyinfo) (tp *TcpRelay, err error) {
	l, err := Listen("tcp", c.ServerPort, c.Cipher)
	return &TcpRelay{l: l, proxyinfo: c}, err
}
func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(time.Minute)
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
			c.Close()
			continue
		}
		go func() {
			defer func() {
				t.conns.Delete(c.RemoteAddr().String())
				c.Close()
				ConnCount--
			}()
			t.conns.Store(c.RemoteAddr().String(), c)
			ConnCount++
			tgt, err := socks.ReadAddr(c)

			if err != nil {
				return
			}

			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				return
			}
			defer func() {
				t.conns.Delete(rc.RemoteAddr().String())
				rc.Close()
			}()
			t.conns.Store(rc.RemoteAddr().String(), rc)
			tcpKeepAlive(rc)
			t.Active()

			var flow int
			currstamp := time.Now()
			go func() {
				defer func() {
					duration := time.Since(currstamp)
					time_stamp := time.Now().UnixNano() / 1000000
					if rate := float64(flow) / duration.Seconds() / 1024; rate > 1.0 {
						t.proxyinfo.Printf("proxy %s <-> %s\trate[%f kb/s]\tflow[%d kb]\tDuration[%f sec]",
							c.RemoteAddr(), tgt, rate, flow/1024, duration.Seconds())
						ip := fmt.Sprintf("%v", c.RemoteAddr())
						website := fmt.Sprintf("%v", tgt)
						rate := int64(rate * 100)
						if t.ConnectInfoCallback != nil {
							t.ConnectInfoCallback(time_stamp, int64(rate*100), ip, website, int64(flow/1024)*100, duration)
						}
					}
				}()
				PipeThenClose(rc, c, func(n int) {
					flow += n
					if err := t.Limiter.WaitN(n); err != nil {
						t.Printf("[%v] -> [%v] speedlimiter err:%v", tgt, c.RemoteAddr(), err)
					}
					go t.AddTraffic(0, n, 0, 0)
				})
			}()
			PipeThenClose(c, rc, func(n int) {
				flow += n
				if err := t.Limiter.WaitN(n); err != nil {
					t.Printf("[%v] -> [%v] speedlimiter err:%v", c.RemoteAddr(), tgt, err)
				}
				go t.AddTraffic(n, 0, 0, 0)
			})
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
func PipeThenClose(left, right net.Conn, addTraffic func(n int)) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		left.SetReadDeadline(time.Now().Add(time.Second * 10))
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
}
