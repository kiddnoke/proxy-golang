package manager

import (
	"errors"
	"net"
)

func GetFreePort(start, end int) (freeport int) {
	for freeport = start; freeport <= end; freeport++ {
		tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if t_err != nil || u_err != nil {
			continue
		} else {
			tl.Close()
			ul.Close()
			return
		}
	}
	return freeport
}
func IsFreePort(port int) (err error) {
	tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})

	if t_err != nil || u_err != nil {
		return errors.New(t_err.Error() + u_err.Error())
	} else {
		tl.Close()
		ul.Close()
		return nil
	}
}
