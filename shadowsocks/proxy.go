package shadowsocks

type Proxy struct {
	TcpInstance     TcpRelay
	UdpInstance     UdpRelay
	Config          SSconfig
	TransferChannel chan interface{}
	traffic         Traffic
	master          *Manager
}

func NewProxy(config SSconfig) (p *Proxy, e error) {
	t, u, err := util.IsOccupiedPort(config.ServerPort)
	if err != nil {
		return nil, err
	}
	p = &Proxy{
		Config:  config,
		traffic: Traffic{0, 0, 0, 0},
	}
	tl := makeTcpRelay(t, config, p.AddTraffic)
	p.TcpInstance = tl
	tl.Start()
	ul := makeUdpRelay(u, config, p.AddTraffic)
	p.UdpInstance = ul
	ul.Start()
	return
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
func (p *Proxy) SetLimit(bytesPerSec int) {
	p.TcpInstance.SetLimit(bytesPerSec)
	p.UdpInstance.SetLimit(bytesPerSec)
}
func (p *Proxy) AddTraffic(tu, td, uu, ud int) {
	p.traffic.tcpup += uint64(tu)
	p.traffic.tcpdown += uint64(td)
	p.traffic.udpup += uint64(uu)
	p.traffic.udpdown += uint64(ud)
}
