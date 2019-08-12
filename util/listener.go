package util

import (
	"net"
)

const PortBegin = 20000
const PortEnd = 65535

var maxPort = PortBegin

func FreeListenerRange(start, end int) (tl *net.TCPListener, ul *net.UDPConn) {
	var err error
	var freeport int
	for freeport = start; freeport <= end && freeport <= 65535; freeport++ {
		tl, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if err != nil {
			continue
		}
		ul, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if err != nil {
			tl.Close()
			continue
		}
		break
	}
	if freeport > maxPort {
		maxPort = freeport
	}
	return
}
func MaxListener() (tl *net.TCPListener, ul *net.UDPConn) {
	for {
		maxPort++
		maxPort = maxPort % PortEnd
		if maxPort < PortBegin {
			maxPort = PortBegin
		}
		var err error
		tl, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: maxPort})
		if err != nil {
			continue
		}
		ul, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: maxPort})
		if err != nil {
			tl.Close()
			continue
		}
		break
	}
	return
}

func IsFreePort(port int) (err error) {
	tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	defer tl.Close()
	if t_err != nil {
		return t_err
	}
	ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	defer ul.Close()
	if u_err != nil {
		return u_err
	}
	return
}
