package manager

import (
	"fmt"
	"log"
	"time"

	"proxy-golang/relay"

	"proxy-golang/udpposter"
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
	Timeout               int64  `json:"timeout"`
	Expire                int64  `json:"expire"`
	BalanceNotifyDuration int    `json:"balancenotifytime"`
	// v1.1.1
	SnId             int64   `json:"sn_id"`
	AppVersion       string  `json:"app_version"`
	UserType         string  `json:"user_type"`
	CarrierOperators string  `json:"carrier_operators"`
	Os               string  `json:"os"`
	DeviceId         string  `json:"device_id"`
	UsedTotalTraffic int64   `json:"used_total_traffic" unit:"kb"`
	LimitArray       []int64 `json:"limit_array" unit:"kb"`
	FlowArray        []int64 `json:"flow_array" unit:"kb"`

	relay.ProxyRelay
}

func (p *Proxy) Init() (err error) {
	searchLimit, err := SearchLimit(int64(p.CurrLimitDown), p.LimitArray, p.FlowArray, p.UsedTotalTraffic)

	pi, e := relay.NewProxyInfo(p.ServerPort, p.Method, p.Password, int(searchLimit))
	if e != nil {
		return NewError("Proxy Init", e, relay.NewProxyRelay, p.ServerPort, p.Method, p.Password, p.CurrLimitDown)
	}
	log.Printf("Proxy.Init UsedTotalTraffic[%v] DefaultLimi[%v] CurrLimit[%v]", p.UsedTotalTraffic, p.CurrLimitDown, searchLimit)
	p.CurrLimitDown = int(searchLimit)
	p.CurrLimitUp = int(searchLimit)

	pr, e := relay.NewProxyRelay(pi)
	if e != nil {
		return NewError("Proxy Init", e, relay.NewProxyRelay, pi)
	}

	pr.ConnectInfoCallback = func(time_stamp int64, rate int64, localAddress, RemoteAddress string, traffic int64, duration time.Duration) {
		user_id := p.Uid
		sn_id := p.SnId
		device_id := p.DeviceId
		app_version := p.AppVersion
		os := p.Os
		user_type := p.UserType
		carrier_operator := p.CarrierOperators
		connect_time := int64(duration.Seconds() * 100)
		_ = udpposter.PostParams(user_id, sn_id,
			device_id, app_version, os, user_type, carrier_operator,
			localAddress, RemoteAddress, time_stamp,
			rate, connect_time, traffic)
	}
	pr.Start()
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
func (p *Proxy) IsStairCase() (limit int, flag bool) {
	tu, td, uu, ud := p.GetTraffic()
	totalFlow := p.UsedTotalTraffic + (tu+td+uu+ud)/1024
	preLimit := int64(p.CurrLimitDown)
	nextLimit, err := SearchLimit(preLimit, p.LimitArray, p.FlowArray, totalFlow)
	if preLimit != nextLimit && err == nil {
		log.Printf("Proxy.IsStairCase totalFlow[%v] CurrLimit[%v] NextLimit[%v]", totalFlow, p.CurrLimitDown, nextLimit)
		return int(nextLimit), true
	} else {
		return 0, false
	}
}
