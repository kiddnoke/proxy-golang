package shadowsocks

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"net"
	"strings"
	"syscall"
	"time"
)

type TcpRelay struct {
	*net.TCPListener
	config     SSconfig
	limiter    *speedlimiter
	cipher     *Cipher
	connCnt    int
	channel    chan interface{}
	ctx        context.Context
	stopFunc   context.CancelFunc
	addTraffic func(tu, td, uu, ud int)
	running    bool
}

const readtimeout = 180

func newTcpRelay(tcp *net.TCPListener, config SSconfig, addTraffic func(tu, td, uu, ud int)) *TcpRelay {
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &TcpRelay{
		TCPListener: tcp,
		limiter:     util.NewSpeedLimiterWithContext(ctx, config.Currlimitdown*1024),
		config:      config,
		cipher:      cipher,
		ctx:         ctx,
		stopFunc:    cancel,
		addTraffic:  addTraffic}
}
func makeTcpRelay(tcp *net.TCPListener, config SSconfig, addTraffic func(tu, td, uu, ud int)) TcpRelay {
	return *newTcpRelay(tcp, config, addTraffic)
}
func (l *TcpRelay) Listening() {
	log.Printf("SS listening at tcp port[%d]", l.config.ServerPort)
	defer log.Printf("TcpRelay port:[%d] Uid:[%d] Sid:[%d] Close", l.config.ServerPort, l.config.Uid, l.config.Sid)
	for l.running {
		conn, err := l.Accept()
		if err != nil {
			// listener maybe closed to update password
			//debug.Printf("accept error: %v\n", err)
			log.Printf("accept error:[%s]", err.Error())
			return
		}
		// Creating cipher upon first connection.
		go l.handleConnection(NewConn(conn, l.cipher.Copy()))
		select {
		case <-l.ctx.Done():
			log.Printf("TcpRelay port:[%d] Uid:[%d] Sid:[%d]  by Error:[%s]", l.config.ServerPort, l.config.Uid, l.config.Sid, l.ctx.Err())
			return
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}
func (l *TcpRelay) handleConnection(conn *SsConn) {
	closed := false
	l.connCnt++
	log.Printf("new client %s->%s\n", util.sanitizeAddr(conn.RemoteAddr()), conn.LocalAddr())
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
	} else {
		log.Printf("connecting %s", host)
	}
	defer func() {
		log.Printf("closed pipe %s<->%s\n", util.sanitizeAddr(conn.RemoteAddr()), host)
		l.connCnt--
		if !closed {
			_ = conn.Close()
			_ = remote.Close()
		}
	}()
	go func() {
		PipeThenClose(l.ctx, conn, remote, func(Traffic int) {
			// 把消耗的流量推出去
			l.addTraffic(Traffic, 0, 0, 0)
			// 限制速度
			l.limiter.WaitN(Traffic)
		})
	}()

	PipeThenClose(l.ctx, remote, conn, func(Traffic int) {

		// 如上
		l.addTraffic(0, Traffic, 0, 0)
		l.limiter.WaitN(Traffic)
	})
	closed = true
	return
}
func (l *TcpRelay) Start() {
	l.running = true
	go l.Listening()
}
func (l *TcpRelay) Stop() {
	l.running = false
	l.stopFunc()
	l.Close()
}
func (l *TcpRelay) SetLimit(bytesPerSec int) {
	l.limiter.SetLimit(rate.Limit(bytesPerSec))
}

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(ctx context.Context, src, dst net.Conn, addTraffic func(int)) {
	defer func() {
		dst.Close()
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
		select {
		case <-ctx.Done():
			log.Printf("TcpRelayerPipe is closed by Manager")
			return
		default:
			continue
		}
	}
	return
}
