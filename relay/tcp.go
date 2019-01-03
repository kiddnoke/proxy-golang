package relay

import (
	"../manager/speedlimit"
	"github.com/riobard/go-shadowsocks2/core"
	"github.com/riobard/go-shadowsocks2/socks"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Conn struct {
	net.Conn
	add func(tu, td int)
}
type Config struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	Timeout    int    `json:"timeout"`
	Method     string `json:"method"`
	Password   string `json:"password"`
	Expire     int64  `json:"expire"`
	ServerPort int    `json:"server_port"`
	Limit      int    `json:"limit"`
	Overflow   int64  `json:"overflow"`
}
type Traffic struct {
	tu int64
	td int64
	uu int64
	ud int64
}

func (t *Traffic) GetTraffic() (tu, td, uu, ud int64) {
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) GetTrafficWithClear() (tu, td, uu, ud int64) {
	defer func() {
		t.tu = 0
		t.td = 0
		t.uu = 0
		t.ud = 0
	}()
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) AddTraffic(tu, td, uu, ud int) {
	t.tu += int64(tu)
	t.td += int64(td)
	t.uu += int64(uu)
	t.ud += int64(ud)
}

type ProxyInfo struct {
	Config
	core.Cipher
	*speedlimit.Limiter
	Traffic
	running bool
}

func NewProxyInfo(c Config) (pt *ProxyInfo) {
	ciph, err := core.PickCipher(c.Method, nil, c.Password)
	if err != nil {
		log.Fatal(err)
	}
	limiter := speedlimit.New(c.Limit * 1024)
	return &ProxyInfo{
		Cipher:  ciph,
		Config:  c,
		Limiter: limiter,
		Traffic: Traffic{0, 0, 0, 0},
		running: false,
	}
}
func MakeProxyInfo(c Config) (pi ProxyInfo) {
	return *NewProxyInfo(c)
}

type TcpRelay struct {
	l net.Listener
	ProxyInfo
	conns sync.Map
}

func NewTcpRelay(c ProxyInfo) (tp *TcpRelay, err error) {
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
func (t *TcpRelay) Start() {
	t.running = true
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
				rc.Close()
				t.conns.Delete(rc.RemoteAddr().String())
			}()
			t.conns.Store(rc.RemoteAddr().String(), rc)
			tcpKeepAlive(rc)

			//logf("proxy %s <-> %s", c.RemoteAddr(), tgt)
			go func() {
				PipeThenClose(rc, c, func(n int) {
					log.Printf("[0] [%d]", n)
					t.Limiter.WaitN(n)
					t.AddTraffic(0, n, 0, 0)
				})
			}()
			PipeThenClose(c, rc, func(n int) {
				log.Printf("[%d] [0]", n)
				t.Limiter.WaitN(n)
				t.AddTraffic(n, 0, 0, 0)
			})
		}()
	}
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
