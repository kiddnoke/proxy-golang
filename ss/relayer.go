package ss

import (
	"net"
	"strconv"

	"github.com/pkg/errors"

	"proxy-golang/common"
	"proxy-golang/util"
)

type ProxyRelayer interface {
	Stop()
	Start()
	Close()
}
type ProxyRelay struct {
	*TcpRelay
	*UdpRelay
	*proxyinfo
}

var leakyBuf = common.BuffPoll

func NewProxyRelay(p *proxyinfo) (r *ProxyRelay, err error) {
	t, err_t := NewTcpRelayByProxyInfo(p)
	if err_t != nil {
		return nil, errors.Wrapf(err_t, "new tcp relay proxyinfo[%v]", p)
	}
	u, err_u := NewUdpRelayByProxyInfo(p)
	if err_u != nil {
		return nil, errors.Wrapf(err_u, "new udp relay proxyinfo[%v]", p)
	}
	return &ProxyRelay{TcpRelay: t, UdpRelay: u, proxyinfo: p}, nil
}
func (r *ProxyRelay) Start() {
	if r.running == false {
		r.Warn("ProxyRelay Start")
		r.TcpRelay.Start()
		r.UdpRelay.Start()
	}
}
func (r *ProxyRelay) Stop() {
	r.Warn("ProxyRelay Stop")
	r.TcpRelay.Stop()
	r.UdpRelay.Stop()
}
func (r *ProxyRelay) Close() {
	r.Warn("ProxyRelay Close")
	r.Stop()
	r.TcpRelay.Close()
	r.UdpRelay.Close()
}
func NewRelay(c *proxyinfo) (r *ProxyRelay, err error) {
	r = new(ProxyRelay)
	var tl *net.TCPListener
	var ul *net.UDPConn
	if c.ServerPort == 0 {
		tl, ul = util.MaxListener()

		_, port, _ := net.SplitHostPort(tl.Addr().String())
		c.ServerPort, _ = strconv.Atoi(port)
	} else {
		tl, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: c.ServerPort})
		ul, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: c.ServerPort})
	}

	tr := &TcpRelay{
		l:         listener{tl, c.Cipher},
		proxyinfo: c,
		handlerId: 0,
	}
	r.TcpRelay = tr
	ur := &UdpRelay{
		l:         c.Cipher.PacketConn(ul),
		proxyinfo: c,
	}
	r.UdpRelay = ur

	r.proxyinfo = c
	return
}
