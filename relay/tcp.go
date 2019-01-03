package relay

import (
	"../manager/speedlimit"
	"github.com/riobard/go-shadowsocks2/core"
	"github.com/riobard/go-shadowsocks2/socks"
	"io"
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

// WrapperConnWithTraffic wraps a stream-oriented net.Conn with stream cipher encryption/decryption.
func WrapperConnWithTraffic(c net.Conn, add func(tu, td int)) *Conn { return &Conn{Conn: c, add: add} }

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.add(n, 0)
	return n, err
}

func (c *Conn) WriteTo(w io.Writer) (int64, error) {
	n, err := c.Conn.(io.WriterTo).WriteTo(w)
	c.add(0, int(n))
	return n, err
}

func (c *Conn) ReadFrom(r io.Reader) (int64, error) {
	n, err := c.Conn.(io.ReaderFrom).ReadFrom(r)
	c.add(int(n), 0)
	return n, err
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
			c = WrapperConnWithTraffic(c, func(tu, td int) {
				if err = t.WaitN(tu + td); err != nil {
					log.Printf("[%s] [%s]", c.RemoteAddr().String(), err.Error())
				} else {
					log.Printf("[%s] [%d] [%d]", c.RemoteAddr().String(), tu, td)
				}
				t.AddTraffic(td, tu, 0, 0)
			})
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
			if err = relay(c, rc); err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					return // ignore i/o timeout
				}
				//logf("relay error: %v", err)
			}
		}()
	}
}
func relay(left, right net.Conn) error {
	var err, err1 error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err1 = io.Copy(right, left)
		right.SetReadDeadline(time.Now()) // unblock read on right
	}()

	_, err = io.Copy(left, right)
	left.SetReadDeadline(time.Now()) // unblock read on left
	wg.Wait()

	if err1 != nil {
		err = err1
	}
	return err
}
func (t *TcpRelay) Stop() {
	t.running = false
	t.l.Close()
	t.conns.Range(func(key, value interface{}) bool {
		value.(net.Conn).Close()
		return true
	})
}
