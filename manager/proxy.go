package manager

import (
	"Vpn-golang/relay"
	"fmt"
	"time"
)

type Proxy struct {
	Uid             uint   `json:"uid"`
	Sid             uint   `json:"sid"`
	ServerPort      int    `json:"server_port"`
	Method          string `json:"method"`
	Password        string `json:"password"`
	Limit           int    `json:"limit"`
	Timeout         uint   `json:"timeout"`
	Remain          uint   `json:"remain"`
	Expire          uint   `json:"expire"`
	NotifyId        uint   `json:"notifyid"`
	NotifyTimestamp uint   `json:"notifytimestamp"`
	relay.ProxyRelay
}

func (p *Proxy) Init() (err error) {
	pi, e := relay.NewProxyInfo(p.ServerPort, p.Method, p.Password, p.Limit)
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
	if p.NotifyTimestamp == 0 {
		return false
	} else {
		if int64(p.Expire)-time.Now().UTC().Unix() < int64(p.NotifyTimestamp) {
			return true
		} else {
			return false
		}
	}
}
