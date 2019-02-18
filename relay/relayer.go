package relay

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

func NewProxyRelay(p *proxyinfo) (r *ProxyRelay, err error) {
	t, err1 := NewTcpRelayByProxyInfo(p)
	u, _ := NewUdpRelayByProxyInfo(p)
	return &ProxyRelay{TcpRelay: t, UdpRelay: u, proxyinfo: p}, err1
}
func (r *ProxyRelay) Start() {
	if r.running == false {
		r.proxyinfo.Logger.Printf("ProxyRelay Start")
		r.TcpRelay.Start()
		r.UdpRelay.Start()
	}
}
func (r *ProxyRelay) Stop() {
	r.Printf("ProxyRelay Stop")
	r.TcpRelay.Stop()
	r.UdpRelay.Stop()
}
func (r *ProxyRelay) Close() {
	r.Printf("ProxyRelay Close")
	r.Stop()
	r.TcpRelay.Close()
	r.UdpRelay.Close()
}
