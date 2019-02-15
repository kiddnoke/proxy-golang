package manager

import (
	"fmt"
	"log"
	"time"

	"proxy-golang/relay"
)

type Proxy struct {
	// v1.1
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
	// v1.1.1
	SnId             int64   `json:"sn_id"`
	AppVersion       string  `json:"app_version"`
	UserType         string  `json:"user_type"`
	CarrierOperators string  `json:"carrier_operators"`
	Os               int     `json:"os"`
	UsedTotalTraffic int64   `json:"used_total_traffic" unit:"kb"`
	LimitArray       []int64 `json:"limit_array" unit:"kb"`
	FlowArray        []int64 `json:"flow_array" unit:"kb"`

	relay.ProxyRelay
}

func (p *Proxy) Init() (err error) {
	if currLimit, err := SearchLimit(p.LimitArray, p.FlowArray, p.UsedTotalTraffic); currLimit < int64(p.CurrLimitDown) && err == nil {
		p.CurrLimitDown = int(currLimit)
		p.CurrLimitUp = int(currLimit)
	}
	pi, e := relay.NewProxyInfo(p.ServerPort, p.Method, p.Password, p.CurrLimitDown)
	if e != nil {
		return e
	}
	pr, e := relay.NewProxyRelay(pi)
	if e != nil {
		return e
	}
	pr.SetFlags(log.LstdFlags | log.Lmicroseconds)
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
	tu, td, uu, ud := p.GetTraffic()
	if tu+td+uu+ud > int64(p.UsedTotalTraffic*1024) {
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
