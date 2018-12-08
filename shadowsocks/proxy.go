package shadowsocks

import "errors"

type Proxy struct {
	TcpInstance     TcpRelayer
	UdpInstance     UdpListener
	Conf            SSconfig
	TransferChannel chan interface{}
	master          *Manager
}

func NewProxy(config SSconfig) (p *Proxy, e error) {
	e = nil
	t, u, ok := util.IsOccupiedPort(config.ServerPort)
	if ok == true {
		e = errors.New("端口被占用")
		return nil, e
	}
	tl := makeTcpListener(t, config)
	tl.Start()
	ul := makeUdpListener(u, config)
	ul.Start()
	return &Proxy{TcpInstance: tl, UdpInstance: ul, Conf: config}, nil
}
func MakeProxy(config SSconfig) (p Proxy) {
	ptr, _ := NewProxy(config)
	return *ptr
}
func (p *Proxy) Start() {
	p.TcpInstance.Start()
	p.UdpInstance.Start()
}
func (p *Proxy) Stop() {
	p.TcpInstance.Stop()
	p.UdpInstance.Stop()
}
