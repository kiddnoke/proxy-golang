package relay

import (
	"github.com/shadowsocks/go-shadowsocks2/core"
	"log"
	"os"
	"proxy-golang/common"
)

type proxyinfo struct {
	ServerPort int `json:"server_port"`
	core.Cipher
	*common.Limiter
	common.Traffic
	*log.Logger
	running bool
}

func NewProxyInfo(ServerPort int, Method string, Password string, Speed int) (pi *proxyinfo, err error) {
	ciph, err := core.PickCipher(Method, nil, Password)
	if err != nil {
		log.Fatal(err)
	}
	limiter := common.NewSpeedLimiter(Speed * 1024)
	return &proxyinfo{
		Cipher:     ciph,
		ServerPort: ServerPort,
		Limiter:    limiter,
		Traffic:    common.MakeTraffic(),
		running:    false,
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
	}, err
}
