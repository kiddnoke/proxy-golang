package manager

import (
	"errors"
	"net"
)

func GetFreePort(start, end int) (freeport int) {
	for freeport = start; freeport <= end; freeport++ {
		tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		defer func() {
			tl.Close()
			ul.Close()
			_ = tl
			_ = ul
		}()
		if t_err != nil || u_err != nil {
			continue
		} else {
			return
		}
	}
	return freeport
}
func IsFreePort(port int) (err error) {
	tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	defer func() {
		tl.Close()
		ul.Close()
		_ = tl
		_ = ul
	}()
	if t_err != nil || u_err != nil {
		return errors.New(t_err.Error() + u_err.Error())
	} else {
		return nil
	}
}
