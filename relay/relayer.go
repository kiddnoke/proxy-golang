package relay

type ProxyRelayer interface {
	Stop()
	Start()
	Close()
}
type ProxyRelay struct {
	t *TcpRelay
	u *UdpRelay
	*proxyinfo
	ProxyRelayer
}

func NewProxyRelay(p proxyinfo) (r *ProxyRelay, err error) {
	t, err1 := NewTcpRelayByProxyInfo(&p)
	u, _ := NewUdpRelayByProxyInfo(&p)
	return &ProxyRelay{t: t, u: u, proxyinfo: &p}, err1
}
func (r *ProxyRelay) Start() {
	if r.running == false {
		r.t.Start()
		r.u.Start()
	}
}
func (r *ProxyRelay) Stop() {
	r.t.Stop()
	r.u.Stop()
}
func (r *ProxyRelay) Close() {
	r.Stop()
	r.t.Close()
	r.u.Close()
}
