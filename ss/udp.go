package ss

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
)

const udpBufSize = 4 * 1024
const UDPTimeout = time.Minute * 3

var bufPool = sync.Pool{New: func() interface{} { return make([]byte, udpBufSize) }}

type UdpRelay struct {
	l net.PacketConn
	*proxyinfo
	conns sync.Map
}

func NewUdpRelayByProxyInfo(p *proxyinfo) (up *UdpRelay, err error) {
	addr := strconv.Itoa(p.ServerPort)
	addr = ":" + addr
	l, err := core.ListenPacket("udp", addr, p.Cipher)
	return &UdpRelay{l: l, proxyinfo: p}, err
}
func (u *UdpRelay) Start() {
	u.running = true
	go u.Loop()
}
func (u *UdpRelay) Stop() {
	u.running = false
	u.conns.Range(func(key, value interface{}) bool {
		value.(net.PacketConn).Close()
		return true
	})
}
func (u *UdpRelay) Close() {
	if u.running == false {
		u.l.Close()
	}
}
func (u *UdpRelay) Loop() {
	m := make(map[string]chan []byte)
	var lock sync.Mutex
	c := u.l
	for u.running {
		buf := bufPool.Get().([]byte)
		c.SetReadDeadline(time.Now().Add(time.Millisecond * AcceptTimeout))
		n, raddr, err := c.ReadFrom(buf)
		u.Limiter.WaitN(n)
		u.AddTraffic(0, 0, int64(n), 0)
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				continue
			}
			// log.Printf("UDP remote read error: %s", err.Error())
			continue
		}

		lock.Lock()
		k := raddr.String()
		ch := m[k]
		if ch == nil {
			pc, err := net.ListenPacket("udp", "")
			if err != nil {
				//logf("failed to create UDP socket: %v", err)
				goto Unlock
			}
			u.conns.Store(pc.LocalAddr().String(), pc)
			ch = make(chan []byte, 1) // must use buffered chan
			m[k] = ch

			go func() { // receive from udpLocal and send to target
				var tgtUDPAddr *net.UDPAddr
				var err error

				for buf := range ch {
					tgtAddr := socks.SplitAddr(buf)
					if tgtAddr == nil {
						//logf("failed to split target address from packet: %q", buf)
						goto End
					}
					tgtUDPAddr, err = net.ResolveUDPAddr("udp", tgtAddr.String())
					if err != nil {
						//logf("failed to resolve target UDP address: %v", err)
						goto End
					}
					u.Info("localAddr[%v] => tgtAddr[%v]", k, tgtAddr.String())
					pc.SetReadDeadline(time.Now().Add(UDPTimeout))
					if _, err = pc.WriteTo(buf[len(tgtAddr):], tgtUDPAddr); err != nil {
						//logf("UDP remote write error: %v", err)
						goto End
					}
				End:
					bufPool.Put(buf[:cap(buf)])
				}
			}()

			go func() { // receive from udpLocal and send to client
				handler := func(n int) {
					u.Limiter.WaitN(n)
					u.AddTraffic(0, 0, 0, int64(n))
				}
				if err := timedCopy(raddr, c, pc, UDPTimeout, true, handler); err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						// ignore i/o timeout
					} else {
						//logf("timedCopy error: %v", err)
					}
				}
				u.conns.Delete(pc.LocalAddr().String())
				pc.Close()
				lock.Lock()
				if ch := m[k]; ch != nil {
					close(ch)
				}
				delete(m, k)
				lock.Unlock()
			}()
		}
	Unlock:
		lock.Unlock()

		select {
		case ch <- buf[:n]: // sent
		default: // drop
			bufPool.Put(buf)
		}
	}
}

// copy from src to dst at target with read timeout
func timedCopy(target net.Addr, dst, src net.PacketConn, timeout time.Duration, prependSrcAddr bool, addTraffic func(n int)) error {
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	for {
		src.SetReadDeadline(time.Now().Add(timeout))
		n, raddr, err := src.ReadFrom(buf)
		if err != nil {
			return err
		}
		addTraffic(n)
		if prependSrcAddr { // server -> client: prepend original packet source address
			srcAddr := socks.ParseAddr(raddr.String())
			copy(buf[len(srcAddr):], buf[:n])
			copy(buf, srcAddr)
			if _, err = dst.WriteTo(buf[:len(srcAddr)+n], target); err != nil {
				return err
			}
		} else { // client -> user: strip original packet source address
			srcAddr := socks.SplitAddr(buf[:n])
			if _, err = dst.WriteTo(buf[len(srcAddr):n], target); err != nil {
				return err
			}
		}
	}
}
