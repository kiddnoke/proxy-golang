package ss

import (
	"fmt"
	"log"
	"net"
	"proxy-golang/common"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

var ReadDeadlineDuration = time.Second * 15
var WriteDeadlineDuration = ReadDeadlineDuration

const DialTimeoutDuration = time.Second * 5
const KeepAlivePeriod = time.Second * 15
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
func setTcpConnKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(KeepAlivePeriod)
	}
}
func (t *TcpRelay) Loop() {
	pipe_set := common.NewPipTrafficSet()
	ConnCount := 0
	go func() {
		tick := time.NewTicker(time.Second)
		for {
			select {
			case <-tick.C:
				if rate := t.OnceSampling(); rate > 0 {
					log.Printf("total[%f kb/s] = %v", rate, pipe_set.SamplingAndString(time.Second))
				}
			}
		}
	}()
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
			setTcpConnKeepAlive(shadowconn)

			tgt, err := socks.ReadAddr(shadowconn)
			if err != nil {
				t.Warn("socks.ReadAddr Error [%s],handlerId[%d], local[%s]", err.Error(), handlerId, shadowconn.RemoteAddr())
				return
			}

			pre_connectstamp := time.Now()
			remoteconn, err := net.DialTimeout("tcp", tgt.String(), DialTimeoutDuration)
			if err != nil {
				t.Warn("net.Dial Error [%s], handlerId[%d], tgt[%s]", err.Error(), handlerId, tgt.String())
				return
			}
			currstamp := time.Now()
			t.Debug("handlerId[%d], dialing tgt[%s] duration[%f sec]", handlerId, tgt.String(), currstamp.Sub(pre_connectstamp).Seconds())
			setTcpConnKeepAlive(remoteconn)
			defer func() {
				t.conns.Delete(remoteconn.RemoteAddr().String())
				remoteconn.Close()
			}()
			t.conns.Store(remoteconn.RemoteAddr().String(), remoteconn)
			t.Active()

			var down_flow int64
			ErrC := make(chan error, 1)
			go func() {
				err := PipeThenClose(shadowconn, remoteconn, func(n int) {
					if err := t.Limiter.WaitN(n); err != nil {
						t.Error("[%v] -> [%v] speedlimiter err:%v", shadowconn.RemoteAddr(), tgt, err)
					}
					t.AddTraffic(int64(n), 0, 0, 0)
				})
				ErrC <- err
			}()
			go func() {
				key := shadowconn.RemoteAddr().String() + "=>" + remoteconn.RemoteAddr().String()
				err := PipeThenClose(remoteconn, shadowconn, func(n int) {
					remoteconn.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
					if err := t.Limiter.WaitN(n); err != nil {
						t.Error("[%v] -> [%v] speedlimiter err:%v", tgt, shadowconn.RemoteAddr(), err)
					}
					atomic.AddInt64(&down_flow, int64(n))
					t.AddTraffic(0, int64(n), 0, 0)
					pipe_set.AddTraffic(key, int64(n))
				})
				ErrC <- err
			}()
			err = <-ErrC
			defer func() {
				duration := time.Since(currstamp)
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					duration = duration - ReadDeadlineDuration
				}
				time_stamp := time.Now().UnixNano() / 1e6
				_flow := atomic.LoadInt64(&down_flow)
				rate := float64(_flow) / duration.Seconds() / 1024
				ip := fmt.Sprintf("%v", shadowconn.RemoteAddr())
				website := fmt.Sprintf("%v", tgt)

				t.Debug("handler[%d] flow[%f k] duration[%f sec] rate[%f kb/s] domain[%v] remoteaddr[%v] Error[%s]", handlerId, float64(_flow)/1024.0, duration.Seconds(), rate, tgt, remoteconn.RemoteAddr(), err.Error())
				if t.ConnectInfoCallback != nil && down_flow > 10*1024 {
					t.ConnectInfoCallback(time_stamp, rate, ip, website, float64(_flow)/1024.0, duration)
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
