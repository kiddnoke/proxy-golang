package ss

import (
	"github.com/pkg/errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"

	"proxy-golang/common"
	"proxy-golang/util"
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
func setTcpConnNoDelay(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
	}
}
func setTcpConnBuffer(c net.Conn, size int) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetReadBuffer(size)
		tcp.SetWriteBuffer(size)
	}
}
func setTcpDefault(c net.Conn) {
	setTcpConnKeepAlive(c)
	setTcpConnNoDelay(c)
	setTcpConnBuffer(c, 1024*32)
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
			t.Debug("total[%f kb/s] = %v", rate, pipe_set.SamplingAndString(time.Second))
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
		go func(shadowconn net.Conn, connectionComming time.Time) {
			t.handlerId++
			handlerId := t.handlerId

			t.conns.Store(shadowconn.RemoteAddr().String(), shadowconn)
			ConnCount++
			defer func() {
				t.conns.Delete(shadowconn.RemoteAddr().String())
				ConnCount--
			}()

			/**
			 * tcp perfermance params
			 */
			var remoteConnectionStart, remoteConnectionFinish time.Time
			var ConnectionStalled time.Duration

			var localFFTB_dur time.Duration
			var remoteFFTB_dur time.Duration

			tgt, err := socks.ReadAddr(shadowconn)
			if err != nil {
				if err == io.EOF { // 仅仅只是握手就端开了
					t.Debug("shadowconn[%s] just tcp handshake", shadowconn.RemoteAddr().String())
					return
				}
				t.Warn("socks.ReadAddr Error [%s],handlerId[%d], local[%s]", err.Error(), handlerId, shadowconn.RemoteAddr())
				return
			}
			localFFTB_dur = time.Since(connectionComming)
			t.Debug("localFFTB_dur[%v]", localFFTB_dur)
			remoteConnectionStart = time.Now()
			remoteconn, err := net.DialTimeout("tcp", tgt.String(), DialTimeoutDuration)
			if err != nil {
				t.Warn("net.Dial Error [%s], handlerId[%d], tgt[%s]", err.Error(), handlerId, tgt.String())
				return
			}
			remoteConnectionFinish = time.Now()
			ConnectionStalled = remoteConnectionFinish.Sub(remoteConnectionStart)
			t.Debug("handlerId[%d], dialing tgt[%s] duration[%f sec]", handlerId, tgt.String(), ConnectionStalled.Seconds())
			setTcpDefault(remoteconn)
			defer func() {
				t.conns.Delete(remoteconn.RemoteAddr().String())
				remoteconn.Close()
			}()
			t.conns.Store(remoteconn.RemoteAddr().String(), remoteconn)
			t.Active()

			//remote -> downloadPipeW downloadPipeR -> shadows_conn

			//downloadPipeR, downloadPipeW := common.Pipe(1024)
			downloadPipeR, downloadPipeW := io.Pipe()
			//downloadPipeR, downloadPipeW := net.Pipe()

			//downloadPipeRW := common.NewBuffPipe(1024)

			ErrChan := make(chan error, 3)
			var wg sync.WaitGroup
			wg.Add(3)
			var request_sent time.Time
			go func() {
				defer wg.Done()
				var once sync.Once
				err := PipeThenClose(shadowconn, remoteconn, func(n int) {
					shadowconn.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
					once.Do(func() {
						request_sent = time.Now()
					})

					t.AddTraffic(int64(n), 0, 0, 0)
				})
				ErrChan <- err
			}()
			var RecvBytes int64
			var preRecvBytes int64
			var RecvBytesAvgRate float64
			var RecvDuration time.Duration
			var RecvBytesMaxRate float64
			go func() {
				defer wg.Done()
				var once sync.Once
				err := PipeThenClose(remoteconn, downloadPipeW, func(n int) {
					once.Do(func() {
						remoteFFTB_dur = time.Since(remoteConnectionFinish)
					})
					remoteconn.SetReadDeadline(time.Now().Add(ReadDeadlineDuration))
					atomic.AddInt64(&RecvBytes, int64(n))
				})
				RecvDuration = time.Since(remoteConnectionFinish)

				ne, ok := errors.Cause(err).(*net.OpError)
				if ok && ne.Timeout() {
					RecvDuration -= ReadDeadlineDuration
				}
				RecvBytesAvgRate = common.Ratter(RecvBytes, RecvDuration)
				t.Debug("remote[%v]=>pipe Error[%v]", remoteconn.RemoteAddr(), err.Error())
				t.Info("handle[%d] remote[%s]=>proxy Bytes[%.3f Kb] AvgRate[%.3f kb/s] MaxRate[%.3f kb/s] Duration[%.3f sec] Error[%v]", handlerId, tgt.String(), float64(RecvBytes)/1024.0, RecvBytesAvgRate, RecvBytesMaxRate, RecvDuration.Seconds(), err)
				ErrChan <- err
			}()

			var SendBytes int64
			var preSendBytes int64
			var SendBytesAvgRate float64
			var SendDuration time.Duration
			var SendBytesMaxRate float64
			go func() {
				defer wg.Done()
				key := tgt.String()
				err := PipeThenClose(common.TimeOutReaderWrap(downloadPipeR, ReadDeadlineDuration), shadowconn, func(n int) {
					if err := t.Limiter.WaitN(n); err != nil {
						t.Error("[%v] -> [%v] speedlimiter err:%v", shadowconn.RemoteAddr(), tgt, err)
					}
					atomic.AddInt64(&SendBytes, int64(n))
					t.AddTraffic(0, int64(n), 0, 0)
					pipe_set.AddTraffic(key, int64(n))
				})
				SendDuration = time.Since(connectionComming)
				type timeouter interface {
					Timeout() bool
				}
				ne, ok := errors.Cause(err).(timeouter)
				if ok && ne.Timeout() {
					SendDuration -= ReadDeadlineDuration
				}
				SendBytesAvgRate = common.Ratter(SendBytes, SendDuration)
				t.Debug("pipe=>shadows_conn[%v] Error[%v]", shadowconn.RemoteAddr(), err)
				t.Info("handle[%d] proxy=>shadows_conn[%v] Bytes[%.3f Kb] AvgRate[%.3f kb/s] MaxRate[%.3f kb/s] Duration[%.3f sec] Error[%v]", handlerId, shadowconn.RemoteAddr(), float64(SendBytes)/1024.0, SendBytesAvgRate, SendBytesMaxRate, SendDuration.Seconds(), err)
				ErrChan <- err
			}()
			//
			fourPipeRaterTick := util.Interval(time.Second/2, func(when time.Time) {
				var ratter = func(pre *int64, curr *int64, tick time.Duration) float64 {
					defer atomic.StoreInt64(pre, *curr)
					diff := atomic.LoadInt64(curr) - atomic.LoadInt64(pre)
					return common.Ratter(diff, tick)
				}
				if r := ratter(&preSendBytes, &SendBytes, time.Second/2); r > SendBytesMaxRate {
					SendBytesMaxRate = r
				}
				if r := ratter(&preRecvBytes, &RecvBytes, time.Second/2); r > RecvBytesMaxRate {
					RecvBytesMaxRate = r
				}
			})
			defer fourPipeRaterTick.Stop()

			wg.Wait()
			<-ErrChan
			defer func() {

			}()
		}(c, time.Now())

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

func PipeThenClose(left io.Reader, right io.Writer, addTraffic func(n int)) (netErr error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	for {
		n, err := left.Read(buf)
		if err != nil || n == 0 {
			netErr = errors.Wrap(err, "pipethenclose read")
			break
		}
		if addTraffic != nil && n > 0 {
			addTraffic(n)
		}
		if n > 0 {
			if _, err := right.Write(buf[0:n]); err != nil {
				netErr = errors.Wrap(err, "pipethenclose write")
				break
			}
		}
	}
	return
}
