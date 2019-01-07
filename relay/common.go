package relay

import (
	"github.com/riobard/go-shadowsocks2/core"
	"log"
	"time"
)

type proxyinfo struct {
	ServerPort int `json:"server_port"`
	core.Cipher
	*Limiter
	Traffic
	*log.Logger
	running bool
}

type Traffic struct {
	tu              int64
	td              int64
	uu              int64
	ud              int64
	lastactivestamp int64
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
	t.lastactivestamp = time.Now().UTC().Unix()
}
func (t *Traffic) GetLastTimeStamp() time.Time {
	return time.Unix(t.lastactivestamp, 0)
}
func NewProxyInfo(ServerPort int, Method string, Password string, Speed int) (pi *proxyinfo, err error) {
	ciph, err := core.PickCipher(Method, nil, Password)
	if err != nil {
		log.Fatal(err)
	}
	limiter := NewSpeedLimiter(Speed * 1024)
	return &proxyinfo{
		Cipher:     ciph,
		ServerPort: ServerPort,
		Limiter:    limiter,
		Traffic:    Traffic{0, 0, 0, 0, time.Now().UTC().Unix()},
		running:    false,
		Logger:     &log.Logger{},
	}, err
}
