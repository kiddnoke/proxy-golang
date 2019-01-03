package relay

import (
	"../manager/speedlimit"
	"github.com/riobard/go-shadowsocks2/core"
	"log"
)

type Config struct {
	Uid        int    `json:"uid"`
	Sid        int    `json:"sid"`
	Timeout    int    `json:"timeout"`
	Method     string `json:"method"`
	Password   string `json:"password"`
	Expire     int64  `json:"expire"`
	ServerPort int    `json:"server_port"`
	Limit      int    `json:"limit"`
	Overflow   int64  `json:"overflow"`
}

type ProxyInfo struct {
	Config
	core.Cipher
	*speedlimit.Limiter
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

func NewProxyInfo(c Config) (pt *ProxyInfo) {
	ciph, err := core.PickCipher(c.Method, nil, c.Password)
	if err != nil {
		log.Fatal(err)
	}
	limiter := speedlimit.New(c.Limit * 1024)
	return &ProxyInfo{
		Cipher:  ciph,
		Config:  c,
		Limiter: limiter,
		Traffic: Traffic{0, 0, 0, 0},
		running: false,
	}
}
func MakeProxyInfo(c Config) (pi ProxyInfo) {
	return *NewProxyInfo(c)
}
