package relay

import (
	"fmt"
	"testing"
)

func TestNewProxyInfo(t *testing.T) {
	pi, e := NewProxyInfo(10000, "aes-128-cfb", "test", 100)
	if e != nil {
		t.FailNow()
	}
	if pi.ServerPort != 10000 {
		t.FailNow()
	}
	if pi.Limit() != 100*1024 {
		t.FailNow()
	}
}
func TestNewProxyRelay(t *testing.T) {
	pi, e := NewProxyInfo(10000, "aes-128-cfb", "test", 100)
	if e != nil {
		t.FailNow()
	}
	pr, e := NewProxyRelay(pi)
	if e != nil {
		t.FailNow()
	}
	fmt.Printf("pi          [%p]=%v\n", pi, pi)
	fmt.Printf("pr.proxyinfo[%p]=%v\n", pr.proxyinfo, pr.proxyinfo)
	if pr.proxyinfo != pi {
		t.FailNow()
	}
	fmt.Printf("pi            [%p]=%v\n", pi, pi)
	fmt.Printf("pr.t.proxyinfo[%p]=%v\n", pr.TcpRelay.proxyinfo, pr.TcpRelay.proxyinfo)
	if pr.TcpRelay.proxyinfo != pi {
		t.FailNow()
	}

	fmt.Printf("pi            [%p]=%v\n", pi, pi)
	fmt.Printf("pr.u.proxyinfo[%p]=%v\n", pr.UdpRelay.proxyinfo, pr.UdpRelay.proxyinfo)
	if pr.UdpRelay.proxyinfo != pi {
		t.FailNow()
	}
	pr.Close()
}
func TestProxyRelay_Close(t *testing.T) {

}
func TestProxyRelay_Start(t *testing.T) {

}
func TestProxyRelay_Stop(t *testing.T) {

}
