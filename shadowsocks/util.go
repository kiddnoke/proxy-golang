package shadowsocks

import (
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type Util struct{}

var util Util

func IsOccupiedPort(port int) (tcpconn *net.TCPListener, udpconn *net.UDPConn, ret error) {
	udpconn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	if err != nil {
		log.Printf("IsOccupiedPort Error:%s", err.Error())
		return nil, nil, err
	}
	tcpl, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	if err != nil {
		log.Printf("IsOccupiedPort Error:%s", err.Error())
		return nil, nil, err
	}
	return tcpl, udpconn, nil
}
func (u Util) sanitizeAddr(addr net.Addr) string {
	return addr.String()
}
func (u Util) getRequestbySsConn(conn *SsConn) (host string, err error) {
	const (
		idType  = 0 // address type index
		idIP0   = 1 // ip address start index
		idDmLen = 1 // domain address length index
		idDm0   = 2 // domain address start index

		typeIPv4 = 1 // type is ipv4 address
		typeDm   = 3 // type is domain address
		typeIPv6 = 4 // type is ipv6 address

		lenIPv4   = net.IPv4len + 2 // ipv4 + 2port
		lenIPv6   = net.IPv6len + 2 // ipv6 + 2port
		lenDmBase = 2               // 1addrLen + 2port, plus addrLen
		// lenHmacSha1 = 10
	)
	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 255(max length address) + 2(port) + 10(hmac-sha1)
	buf := make([]byte, 269)
	// read till we get possible domain length field
	if _, err = io.ReadFull(conn, buf[:idType+1]); err != nil {
		return
	}

	var reqStart, reqEnd int
	addrType := buf[idType]
	switch addrType & AddrMask {
	case typeIPv4:
		reqStart, reqEnd = idIP0, idIP0+lenIPv4
	case typeIPv6:
		reqStart, reqEnd = idIP0, idIP0+lenIPv6
	case typeDm:
		if _, err = io.ReadFull(conn, buf[idType+1:idDmLen+1]); err != nil {
			return
		}
		reqStart, reqEnd = idDm0, idDm0+int(buf[idDmLen])+lenDmBase
	default:
		err = fmt.Errorf("addr type %d not supported", addrType&AddrMask)
		return
	}

	if _, err = io.ReadFull(conn, buf[reqStart:reqEnd]); err != nil {
		return
	}

	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch addrType & AddrMask {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+int(buf[idDmLen])])
	}
	// parse port
	port := binary.BigEndian.Uint16(buf[reqEnd-2 : reqEnd])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return
}

func (u Util) IsOccupiedPort(port int) (tcpconn *net.TCPListener, udpconn *net.UDPConn, ret error) {
	return IsOccupiedPort(port)
}

type speedlimiter struct {
	*rate.Limiter
	ctx context.Context
}

func (u Util) NewSpeedLimiterWithContext(ctx context.Context, bytesPerSec int) *speedlimiter {
	burstsize := bytesPerSec * 3
	limiter := rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
	limiter.AllowN(time.Now(), burstsize)
	ctx = context.Background()
	return &speedlimiter{Limiter: limiter, ctx: ctx}
}
func (u Util) MakeSpeedLimiterWithContext(ctx context.Context, bytesPerSec int) speedlimiter {
	return *u.NewSpeedLimiterWithContext(ctx, bytesPerSec)
}
func (u Util) NewSpeedLimiter(bytesPerSec int) *speedlimiter {
	burstsize := bytesPerSec * 3
	limiter := rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
	limiter.AllowN(time.Now(), burstsize)
	ctx := context.Background()
	return &speedlimiter{Limiter: limiter, ctx: ctx}
}
func (u Util) MakeSpeedLimiter(bytesPerSec int) speedlimiter {
	return *u.NewSpeedLimiterWithContext(context.Background(), bytesPerSec)
}
func (s *speedlimiter) WaitN(n int) error {
	return s.Limiter.WaitN(s.ctx, n)
}

type Traffic struct {
	tcpup   uint64
	tcpdown uint64
	udpup   uint64
	udpdown uint64
}
