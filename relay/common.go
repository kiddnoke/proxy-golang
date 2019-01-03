package relay

import (
	"github.com/riobard/go-shadowsocks2/core"
	"log"
)

type ProxyInfo struct {
	ServerPort int `json:"server_port"`
	core.Cipher
	*Limiter
	Traffic
	running bool
}

type Traffic struct {
	tu int64
	td int64
	uu int64
	ud int64
}

func (t *Traffic) GetTraffic() (tu, td, uu, ud int64) {
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) GetTrafficWithClear() (tu, td, uu, ud int64) {
	defer func() {
		t.tu = 0
		t.td = 0
		t.uu = 0
		t.ud = 0
	}()
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) AddTraffic(tu, td, uu, ud int) {
	t.tu += int64(tu)
	t.td += int64(td)
	t.uu += int64(uu)
	t.ud += int64(ud)
}
func NewProxy(ServerPort int, Method string, Password string, Speed int) (pi *ProxyInfo, err error) {
	ciph, err := core.PickCipher(Method, nil, Password)
	if err != nil {
		log.Fatal(err)
	}
	limiter := NewSpeedLimiter(Speed * 1024)
	return &ProxyInfo{
		Cipher:     ciph,
		ServerPort: ServerPort,
		Limiter:    limiter,
		Traffic:    Traffic{0, 0, 0, 0},
		running:    false,
	}, err
}
