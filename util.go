package shadowsocks

import (
	"net"
	"os"
)

type Util struct {}
func (u Util) IsOccupiedPort(port int) (ret bool){
	udpconn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	defer udpconn.Close()
	if err != nil {
		os.Exit(1)
		return true
	}
	tcpl , err := net.ListenTCP("tcp" ,&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	defer tcpl.Close()
	if err != nil {
		os.Exit(1)
		return true
	}
	return false
}