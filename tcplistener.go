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
	config       Config
	speedlimiter *Bucket
	cipher       *Cipher
	connCnt      int
	channel      chan interface{}
}

var util Util

func NewTcpListener(tcp *net.TCPListener, config Config) *TcpListener {
	speedlimiter := NewBucket(time.Second, config.Limit*1024)
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}
	return &TcpListener{speedlimiter: speedlimiter, config: config, cipher: cipher, TCPListener: tcp}
}
func (l *TcpListener) Listening() {
	//if l.config.Expiration > 0 {
	//	time.AfterFunc(time.Until(time.Unix(l.config.Expiration,0)) , func() {
	//
	//	})
	//}
	for {
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
		})
	}()

	PipeThenClose(remote, conn, func(Traffic int) {
	})
	closed = true
	return
}
func SetReadTimeout(c net.Conn, timeout int /*sec*/) {
	if timeout != 0 {
		_ = c.SetReadDeadline(time.Now().Add(time.Duration(timeout)))
	}
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn, addTraffic func(int)) {
	defer func() {
		_ = dst.Close()
	}()
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		n, _ := src.Read(buf)
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
	}
	return
}
