package ss

import (
	"fmt"
	"net"
	"proxy-golang/common"
	"proxy-golang/util"
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
	ConnectInfoCallback func(time_stamp int64, rate float64, localAddress, RemoteAddress string, traffic float64, duration time.Duration, max_rate float64)
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
	totalRateTicker := util.Interval(time.Second, func(when time.Time) {
		if t == nil || t.running == false {
			t.Warn("Sampling Ticker Stop , this goroutine will be released")
			return
		}
		if rate := t.OnceSampling(); rate > 50.0 {
			t.Info("total[%f kb/s] = %v", rate, pipe_set.SamplingAndString(time.Second))
		}
	})
	defer totalRateTicker.Stop()

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
			{
				t.conns.Store(shadowconn.RemoteAddr().String(), shadowconn)
				ConnCount++
				setTcpConnKeepAlive(shadowconn)
				defer func() {
					t.conns.Delete(shadowconn.RemoteAddr().String())
					shadowconn.Close()
					ConnCount--
				}()
			}
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
			var preDownFlow int64 = down_flow
			var maxRate float64
			maxRateTicker := util.Interval(time.Millisecond*500, func(when time.Time) {
				currDownFlow := atomic.LoadInt64(&down_flow)
				if rate := common.Ratter(currDownFlow-preDownFlow, time.Second/2); rate > 0 && rate > maxRate {
					maxRate = rate
				}
				atomic.StoreInt64(&preDownFlow, currDownFlow)
			})
			defer maxRateTicker.Stop()
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
				key := shadowconn.RemoteAddr().String() + "=>" + tgt.String()
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
			pipe_err := <-ErrC
			defer func() {
				if down_flow < 30*1024 {
					return
				}
				duration := time.Since(currstamp)
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					duration = duration - ReadDeadlineDuration
				}
				time_stamp := time.Now().UnixNano() / 1e6
				_flow := atomic.LoadInt64(&down_flow)
				avgRate := common.Ratter(_flow, duration)
				ip := fmt.Sprintf("%v", shadowconn.RemoteAddr())
				website := fmt.Sprintf("%v", tgt)
				if avgRate > maxRate {
					maxRate = avgRate
				}
				t.Info("handler[%d] flow[%f k] duration[%f sec] avg_rate[%f kb/s] max_rate[%f kb/s] domain[%v] remote_addr[%v] Error[%v]", handlerId, float64(_flow)/1024.0, duration.Seconds(), avgRate, maxRate, tgt, remoteconn.RemoteAddr(), pipe_err)
				if t.ConnectInfoCallback != nil {
					t.ConnectInfoCallback(time_stamp, avgRate, ip, website, float64(_flow)/1024.0, duration, maxRate)
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
