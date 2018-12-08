package shadowsocks

import (
	"log"
	"net"
	"strings"
	"syscall"
	"time"
)

type TcpListener struct {
	*net.TCPListener
	config       SSconfig
	speedlimiter *Bucket
	cipher       *Cipher
	connCnt      int
	channel      chan interface{}
	running      bool
}

const readtimeout = 180

func newTcpListener(tcp *net.TCPListener, config SSconfig) *TcpListener {
	speedlimiter := NewBucket(time.Second, config.Limit*1024)
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}
	return &TcpListener{TCPListener: tcp, speedlimiter: speedlimiter, config: config, cipher: cipher}
}
func makeTcpListener(tcp *net.TCPListener, config SSconfig) TcpListener {
	return *newTcpListener(tcp, config)
}
func (l *TcpListener) Listening() {
	log.Printf("SS listening at tcp port[%d]", l.config.ServerPort)
	for l.running {
		conn, err := l.Accept()
		if err != nil {
			// listener maybe closed to update password
			//debug.Printf("accept error: %v\n", err)
			return
		}
		// Creating cipher upon first connection.
		go l.handleConnection(NewConn(conn, l.cipher.Copy()))
	}
}
func (l *TcpListener) handleConnection(conn *SsConn) {
	closed := false
	l.connCnt++
	defer func() {
		l.connCnt--
		if !closed {
			_ = conn.Close()
		}
	}()
	host, err := util.getRequestbySsConn(conn)
	if err != nil {
		log.Println("error getting request", conn.RemoteAddr(), conn.LocalAddr(), err)
		closed = true
		return
	}
	if strings.ContainsRune(host, 0x00) {
		log.Println("invalid domain name.")
		closed = true
		return
	}
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Println("dial error:", err)
		} else {
			log.Println("error connecting to:", host, err)
		}
		return
	}
	defer func() {
		if !closed {
			_ = remote.Close()
		}
	}()
	go func() {
		PipeThenClose(conn, remote, func(Traffic int) {
			// 把消耗的流量推出去
			// 限制速度
		})
	}()

	PipeThenClose(remote, conn, func(Traffic int) {
		// 如上
	})
	closed = true
	return
}
func (l *TcpListener) Start() {
	l.running = true
	go l.Listening()
}
func (l *TcpListener) Stop() {
	l.running = false
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn, addTraffic func(int)) {
	defer func() {
		_ = dst.Close()
	}()
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		SetReadTimeout(src, readtimeout)
		n, err := src.Read(buf)
		if addTraffic != nil {
			addTraffic(n)
		}
		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			// Note: avoid overwrite err returned by Read.
			if _, err := dst.Write(buf[0:n]); err != nil {
				log.Println("write:", err)
				break
			}
		}
		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			/*
				if bool(Debug) && err != io.EOF {
					Debug.Println("read:", err)
				}
			*/
			break
		}
	}
	return
}
