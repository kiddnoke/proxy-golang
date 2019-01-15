package manager

import (
	"fmt"
	"time"

	"proxy-golang/relay"
)

type Proxy struct {
	Uid                   int64  `json:"uid"`
	Sid                   int64  `json:"sid"`
	ServerPort            int    `json:"server_port"`
	Method                string `json:"method"`
	Password              string `json:"password"`
	CurrLimitUp           int    `json:"currlimitup"`
	CurrLimitDown         int    `json:"currlimitdown"`
	NextLimitUp           int    `json:"nextlimitup"`
	NextLimitDown         int    `json:"nextlimitdown"`
	Timeout               int64  `json:"timeout"`
	Remain                int64  `json:"remain"`
	Expire                int64  `json:"expire"`
	BalanceNotifyDuration int    `json:"balancenotifytime"`
	relay.ProxyRelay
}

func (p *Proxy) Init() (err error) {
	pi, e := relay.NewProxyInfo(p.ServerPort, p.Method, p.Password, p.CurrLimitDown)
	if e != nil {
		return e
	}
	pr, e := relay.NewProxyRelay(*pi)
	if e != nil {
		return e
	}
	pr.SetPrefix(fmt.Sprintf("Uid[%d] Sid[%d] Port[%d] ", p.Uid, p.Sid, p.ServerPort))
	p.ProxyRelay = *pr
	return
}
func (p *Proxy) IsTimeout() bool {
	if p.Timeout == 0 {
		return false
	}
	if p.GetLastTimeStamp().Unix()+int64(p.Timeout) < time.Now().UTC().Unix() {
		return true
	} else {
		return false
	}
}
func (p *Proxy) IsExpire() bool {
	if p.Expire == 0 {
		return false
	}
	if time.Now().UTC().Unix() > int64(p.Expire) {
		return true
	} else {
		return false
	}
}
func (p *Proxy) IsOverflow() bool {
	if p.Remain == 0 {
		return false
	}
	tu, td, uu, ud := p.GetTraffic()
	if tu+td+uu+ud > int64(p.Remain) {
		return true
	} else {
		return false
	}
}
func (p *Proxy) IsNotify() bool {
	if p.BalanceNotifyDuration == 0 {
		return false
	} else {
		if int64(p.Expire)-time.Now().UTC().Unix() < int64(p.BalanceNotifyDuration) {
			return true
		} else {
			return false
		}
	}
}
