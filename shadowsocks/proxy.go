package shadowsocks

type Proxy struct {
	TcpInstance     TcpRelayer
	UdpInstance     UdpListener
	Conf            SSconfig
	TransferChannel chan interface{}
	master          *Manager
}

func NewProxy(config SSconfig) (p *Proxy, e error) {
	t, u, err := util.IsOccupiedPort(config.ServerPort)
	if err != nil {
		return nil, err
	}
	tl := makeTcpListener(t, config)
	tl.Start()
	ul := makeUdpListener(u, config)
	ul.Start()
	return &Proxy{TcpInstance: tl, UdpInstance: ul, Conf: config}, nil
}
func MakeProxy(config SSconfig) (p Proxy, err error) {
	ptr, err := NewProxy(config)
	return *ptr, err
}
func (p *Proxy) Start() {
	p.TcpInstance.Start()
	p.UdpInstance.Start()
}
func (p *Proxy) Stop() {
	p.TcpInstance.Stop()
	p.UdpInstance.Stop()
}
