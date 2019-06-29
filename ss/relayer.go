package ss

import (
	"errors"
	"fmt"
	"proxy-golang/common"
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
	u, err_u := NewUdpRelayByProxyInfo(p)
	if err_t == nil && err_u == nil {
		return &ProxyRelay{TcpRelay: t, UdpRelay: u, proxyinfo: p}, err_t
	} else {
		return nil, errors.New(fmt.Sprintf("NewProxyRelay Error:%v, %v", err_t, err_u))
	}
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
