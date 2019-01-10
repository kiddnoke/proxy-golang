package manager

import (
	"net"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port1 := GetFreePort(20000, 30000)
	if port1 != 20000 {
		t.FailNow()
	} else {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
	}

	port2 := GetFreePort(20000, 30000)
	if port2 != 20001 {
		t.FailNow()
	} else {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port2})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port2})
	}
}
func TestIsFreePort(t *testing.T) {
	port1 := GetFreePort(20000, 30000)
	if err := IsFreePort(port1); err == nil {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
	} else {
		t.FailNow()
	}
	if err := IsFreePort(port1); err == nil {
		t.FailNow()
	}
}
