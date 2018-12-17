package shadowsocks

import (
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	idType  = 0 // address type index
	idIP0   = 1 // ip address start index
	idDmLen = 1 // domain address length index
	idDm0   = 2 // domain address start index

	typeIPv4 = 1 // type is ipv4 address
	typeDm   = 3 // type is domain address
	typeIPv6 = 4 // type is ipv6 address

	lenIPv4   = 1 + net.IPv4len + 2 // 1addrType + ipv4 + 2port
	lenIPv6   = 1 + net.IPv6len + 2 // 1addrType + ipv6 + 2port
	lenDmBase = 1 + 1 + 2           // 1addrType + 1addrLen + 2port, plus addrLen
	// lenHmacSha1 = 10
)

type UdpRelay struct {
	*net.UDPConn
	config     SSconfig
	cipher     *Cipher // 加密子
	limiter    *speedlimiter
	running    bool
	ctx        context.Context
	stopFunc   context.CancelFunc
	reqList    *requestHeaderList
	natlist    *natTable
	addTraffic func(tu, td, uu, ud int)
}

func newUdpRelay(l *net.UDPConn, config SSconfig, addTraffic func(tu, td, uu, ud int)) *UdpRelay {
	ctx := context.Background()
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}
	return &UdpRelay{UDPConn: l, limiter: util.NewSpeedLimiterWithContext(ctx, config.Limit*1024), config: config, cipher: cipher, ctx: ctx, reqList: newReqList(), natlist: newNatTable(), addTraffic: addTraffic}
}
func makeUdpRelay(l *net.UDPConn, config SSconfig, addTraffic func(tu, td, uu, ud int)) UdpRelay {
	return *newUdpRelay(l, config, addTraffic)
}
func (l *UdpRelay) Listening() {
	defer l.Close()
	log.Printf("SS listening at udp port[%d]", l.config.ServerPort)
	SecurePacketConn := NewSecurePacketConn(l, l.cipher.Copy())
	for l.running {
		if err := l.handleConnection(SecurePacketConn); err != nil {
			log.Printf("udp read error:[%s]", err.Error())
			break
		}
	}
	log.Printf("UdpRelayer port:[%d] Uid:[%d] Sid:[%d] Close", l.config.ServerPort, l.config.Uid, l.config.Sid)
}
func (l *UdpRelay) Start() {
	l.running = true
	go l.Listening()
}
func (l *UdpRelay) Stop() {
	l.running = false
	l.reqList.running = false
	l.Close()
}

//func (l *UdpRelay) ReadAndHandleUDPReq(c *SecurePacketConn, addTraffic func(int)) error {
//	buf := leakyBuf.Get()
//	n, src, err := c.ReadFrom(buf[0:])
//	if err != nil {
//		return err
//	}
//	return nil
//}
func (l *UdpRelay) handleConnection(handle *SecurePacketConn) error {
	receive := leakyBuf.Get()
	n, src, err := handle.ReadFrom(receive[0:])
	if err != nil {
		return err
	}
	defer leakyBuf.Put(receive)
	dst, reqLen := parseDstIp(receive)
	if _, ok := l.reqList.Get(dst.String()); !ok {
		req := make([]byte, reqLen)
		copy(req, receive)
		l.reqList.Put(dst.String(), req)
	}

	remote, exist, err := l.natlist.Get(src.String())
	if err != nil {
		return err
	}
	if !exist {
		log.Printf("[udp]new client %s->%s via %s\n", src, dst, remote.LocalAddr())
		go func() {
			l.Pipeloop(handle, src, remote, func(traffic int) {

			})
			l.natlist.Delete(src.String())
		}()
	} else {
		log.Printf("[udp]using cached client %s->%s via %s\n", src, dst, remote.LocalAddr())
	}
	if remote == nil {
		log.Printf("WTF")
	}
	remote.SetDeadline(time.Now().Add(udpTimeout))
	n, err = remote.WriteTo(receive[reqLen:n], dst)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Println("[udp]write error:", err)
		} else {
			log.Println("[udp]error connecting to:", dst, err)
		}
		if conn := l.natlist.Delete(src.String()); conn != nil {
			conn.Close()
		}
	}
	// Pipeloop
	return nil
}
func (l *UdpRelay) Pipeloop(write net.PacketConn, writeAddr net.Addr, readClose net.PacketConn, addTraffic func(int)) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	defer readClose.Close()
	for {
		readClose.SetDeadline(time.Now().Add(udpTimeout))
		n, raddr, err := readClose.ReadFrom(buf)
		if err != nil {
			if ne, ok := err.(*net.OpError); ok {
				if ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE {
					// log too many open file error
					// EMFILE is process reaches open file limits, ENFILE is system limit
					log.Println("[udp]read error:", err)
				}
			}
			log.Printf("[udp]closed pipe %s<-%s\n", writeAddr, readClose.LocalAddr())
			return
		}
		// need improvement here
		if req, ok := l.reqList.Get(raddr.String()); ok {
			n, _ := write.WriteTo(append(req, buf[:n]...), writeAddr)
			addTraffic(n)
		} else {
			header, hlen := parseHeaderFromAddr(raddr)
			n, _ := write.WriteTo(append(header[:hlen], buf[:n]...), writeAddr)
			addTraffic(n)
		}
	}
}
func (l *UdpRelay) SetLimit(bytesPerSec int) {
	l.limiter.SetLimit(rate.Limit(bytesPerSec))
}
func parseHeaderFromAddr(addr net.Addr) ([]byte, int) {
	// if the request address type is domain, it cannot be reverselookuped
	ip, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return nil, 0
	}
	buf := make([]byte, 20)
	IP := net.ParseIP(ip)
	b1 := IP.To4()
	iplen := 0
	if b1 == nil { //ipv6
		b1 = IP.To16()
		buf[0] = typeIPv6
		iplen = net.IPv6len
	} else { //ipv4
		buf[0] = typeIPv4
		iplen = net.IPv4len
	}
	copy(buf[1:], b1)
	port_i, _ := strconv.Atoi(port)
	binary.BigEndian.PutUint16(buf[1+iplen:], uint16(port_i))
	return buf[:1+iplen+2], 1 + iplen + 2
}
func parseDstIp(receive []byte) (dst *net.UDPAddr, reqLen int) {
	var dstIP net.IP
	addrType := receive[idType]

	switch addrType & AddrMask {
	case typeIPv4:
		reqLen = lenIPv4
		if len(receive) < reqLen {
			log.Println("[udp]invalid received message.")
		}
		dstIP = net.IP(receive[idIP0 : idIP0+net.IPv4len])
	case typeIPv6:
		reqLen = lenIPv6
		if len(receive) < reqLen {
			log.Println("[udp]invalid received message.")
		}
		dstIP = net.IP(receive[idIP0 : idIP0+net.IPv6len])
	case typeDm:
		reqLen = int(receive[idDmLen]) + lenDmBase
		if len(receive) < reqLen {
			log.Println("[udp]invalid received message.")
		}
		name := string(receive[idDm0 : idDm0+int(receive[idDmLen])])
		// avoid panic: syscall: string with NUL passed to StringToUTF16 on windows.
		if strings.ContainsRune(name, 0x00) {
			fmt.Println("[udp]invalid domain name.")
			return
		}
		dIP, err := net.ResolveIPAddr("ip", name) // carefully with const type
		if err != nil {
			log.Printf("[udp]failed to resolve domain name: %s\n", string(receive[idDm0:idDm0+receive[idDmLen]]))
			return
		}
		dstIP = dIP.IP
	default:
		log.Printf("[udp]addrType %d not supported", addrType)
		return
	}
	dst = &net.UDPAddr{
		IP:   dstIP,
		Port: int(binary.BigEndian.Uint16(receive[reqLen-2 : reqLen])),
	}
	return
}
