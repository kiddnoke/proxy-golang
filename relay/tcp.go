package relay

import (
	"github.com/riobard/go-shadowsocks2/core"
	"github.com/riobard/go-shadowsocks2/socks"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type TcpRelay struct {
	l net.Listener
	ProxyInfo
	conns sync.Map
}

func NewTcpRelayByProxyInfo(c ProxyInfo) (tp *TcpRelay, err error) {
	addr := strconv.Itoa(c.ServerPort)
	addr = ":" + addr
	l, err := core.Listen("tcp", addr, c.Cipher)
	return &TcpRelay{l: l, ProxyInfo: c}, err
}
func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(3 * time.Minute)
	}
}
func (t *TcpRelay) Loop() {
	for t.running {
		c, err := t.l.Accept()
		if err != nil {
			//logf("failed to accept: %v", err)
			continue
		}

		go func() {
			defer func() {
				t.conns.Delete(c.RemoteAddr().String())
				c.Close()
			}()
			t.conns.Store(c.RemoteAddr().String(), c)
			tcpKeepAlive(c)
			tgt, err := socks.ReadAddr(c)

			if err != nil {
				//logf("failed to get target address: %v", err)
				return
			}

			rc, err := net.Dial("tcp", tgt.String())
			if err != nil {
				//logf("failed to connect to target: %v", err)
				return
			}
			defer func() {
				t.conns.Delete(rc.RemoteAddr().String())
				rc.Close()
			}()
			t.conns.Store(rc.RemoteAddr().String(), rc)
			tcpKeepAlive(rc)

			//logf("proxy %s <-> %s", c.RemoteAddr(), tgt)
			go func() {
				PipeThenClose(rc, c, func(n int) {
					t.Limiter.WaitN(n)
					t.AddTraffic(0, n, 0, 0)
				})
			}()
			PipeThenClose(c, rc, func(n int) {
				t.Limiter.WaitN(n)
				t.AddTraffic(n, 0, 0, 0)
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
	t.l.Close()
	t.conns.Range(func(key, value interface{}) bool {
		value.(net.Conn).Close()
		return true
	})
}

func PipeThenClose(left, right net.Conn, addTraffic func(n int)) {
	defer func() {
		right.Close()
	}()
	buf := make([]byte, 1024*16)
	for {
		left.SetReadDeadline(time.Now().Add(time.Second * 15))
		n, err := left.Read(buf)
		if addTraffic != nil && n > 0 {
			addTraffic(n)
		}
		if n > 0 {
			if _, err := right.Write(buf[0:n]); err != nil {
				log.Println("write:", err)
				break
			}
		}
		if err != nil {
			break
		}
	}
}
