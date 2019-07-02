package multiprotocol

import (
	"fmt"
	"strconv"
	"time"

	"proxy-golang/ss"
	"proxy-golang/udpposter"
)

type SS struct {
	Config
	ss.ProxyRelay
}

func NewSS(p *Config) (r *SS, err error) {
	r = new(SS)
	searchLimit, err := searchLimit(int64(p.CurrLimitDown), p.LimitArray, p.FlowArray, p.UsedTotalTraffic)

	pi, e := ss.NewProxyInfo(p.ServerPort, p.Method, p.Password, int(searchLimit))
	if e != nil {
		return nil, err
	}
	pi.Info("Proxy Init:Uid[%d] Sid[%d] Port[%d] AppId[%d] Proxy.Init UsedTotalTraffic[%v] DefaultLimi[%v] CurrLimit[%v]", p.Uid, p.Sid, p.ServerPort, p.AppId, p.UsedTotalTraffic, p.CurrLimitDown, searchLimit)
	p.CurrLimitDown = int(searchLimit)
	p.CurrLimitUp = int(searchLimit)

	pr, e := ss.NewProxyRelay(pi)
	if e != nil {
		pi.Error("%s", err.Error())
		return nil, err
	}

	pr.ConnectInfoCallback = func(time_stamp int64, rate float64, localAddress, RemoteAddress string, traffic float64, duration time.Duration) {
		_ = udpposter.PostParams(p.AppId, p.Uid, p.SnId,
			p.DeviceId, p.AppVersion, p.Os, p.UserType, p.CarrierOperators, p.NetworkType,
			localAddress, RemoteAddress, time_stamp,
			int64(rate*100), int64(duration.Seconds()*100), int64(traffic*100), p.Ip+":"+strconv.Itoa(p.ServerPort), p.State, p.UserType)
	}
	pr.SetPrefix(fmt.Sprintf("Uid[%d] Sid[%d] Port[%d] AppId[%d] Protocol[%s]", p.Uid, p.Sid, p.ServerPort, p.AppId, p.Protocol))
	r.Config = *p
	r.ProxyRelay = *pr
	return
}

func (s *SS) Start() {
	s.ProxyRelay.Start()
}

func (s *SS) Stop() {
	s.ProxyRelay.Stop()
}

func (s *SS) Close() {
	s.ProxyRelay.Close()
}

func (s *SS) IsTimeout() bool {
	if s.Timeout == 0 {
		return false
	}
	if s.GetLastTimeStamp().Unix()+int64(s.Timeout) < time.Now().UTC().Unix() {
		return true
	} else {
		return false
	}
}

func (s *SS) IsExpire() bool {
	if s.Expire == 0 {
		return false
	}
	if time.Now().UTC().Unix() > int64(s.Expire) {
		return true
	} else {
		return false
	}
}

func (s *SS) IsOverflow() bool {
	tu, td, uu, ud := s.GetTraffic()
	if tu+td+uu+ud > int64(s.UsedTotalTraffic*1024) {
		return true
	} else {
		return false
	}
}

func (s *SS) IsNotify() bool {
	if s.BalanceNotifyDuration == 0 {
		return false
	} else {
		if int64(s.Expire)-time.Now().UTC().Unix() < int64(s.BalanceNotifyDuration) {
			return true
		} else {
			return false
		}
	}
}

func (s *SS) IsStairCase() (limit int, flag bool) {
	tu, td, uu, ud := s.GetTraffic()
	totalFlow := s.UsedTotalTraffic + (tu+td+uu+ud)/1024
	preLimit := int64(s.CurrLimitDown)
	nextLimit, err := searchLimit(preLimit, s.LimitArray, s.FlowArray, totalFlow)
	if preLimit != nextLimit && err == nil {
		s.Warn("Proxy.IsStairCase totalFlow[%v] CurrLimit[%v] NextLimit[%v]", totalFlow, s.CurrLimitDown, nextLimit)
		return int(nextLimit), true
	} else {
		return 0, false
	}
}

func (s *SS) GetTraffic() (tu, td, uu, ud int64) {
	return s.ProxyRelay.GetTraffic()
}

func (s *SS) Clear() {
	s.ProxyRelay.GetTrafficWithClear()
}

func (s *SS) GetStartTimeStamp() time.Time {
	return s.ProxyRelay.GetStartTimeStamp()
}

func (s *SS) GetLastTimeStamp() time.Time {
	return s.ProxyRelay.GetLastTimeStamp()
}

func (s *SS) SetLimit(bytesPerSec int) {
	s.ProxyRelay.SetLimit(bytesPerSec)
}

func (s *SS) Burst() int {
	return s.ProxyRelay.Burst()
}

func (s *SS) GetConfig() *Config {
	return &s.Config
}

func (s *SS) GetMaxRate() (float64, float64) {
	return s.Traffic.GetMaxRate()
}
