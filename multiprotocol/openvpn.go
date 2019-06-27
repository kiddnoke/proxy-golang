package multiprotocol

import (
	"fmt"
	"github.com/kiddnoke/SoftetherGo"
	"proxy-golang/common"
	"proxy-golang/softether"
	"time"
)

type OpenVpn struct {
	Config
	HubName  string
	UserName string
	Password string
	common.Traffic
	timer interval
}

func NewOpenVpn(c *Config) (*OpenVpn, error) {
	r := new(OpenVpn)
	searchLimit, err := searchLimit(int64(c.CurrLimitDown), c.LimitArray, c.FlowArray, c.UsedTotalTraffic)
	c.CurrLimitUp = int(searchLimit)
	c.CurrLimitUp = int(searchLimit)

	r.HubName = fmt.Sprintf("%d", c.Uid)
	r.UserName = fmt.Sprintf("%d", c.Sid)
	r.Password = c.Password
	r.Traffic = common.MakeTraffic()
	c.ServerPort = 1194
	c.Method = r.UserName + "@" + r.HubName

	_, err = softether.API.CreateHub(r.HubName, true, softetherApi.HUB_TYPE_STANDALONE)
	if err != nil {
		if e, ok := err.(softetherApi.ApiError); ok && e.Code() != softetherApi.ERR_HUB_ALREADY_EXISTS {
			goto CreateUser
		}
		return nil, err
	}

	_, err = softether.API.EnableSecureNat(r.HubName)
	if err != nil {
		return nil, err
	}
CreateUser:
	_, err = softether.API.CreateUser(r.HubName, r.UserName, r.Password)
	if err != nil {
		if e, ok := err.(softetherApi.ApiError); ok && e.Code() != softetherApi.ERR_USER_ALREADY_EXISTS {
			goto SetPassword
		}
		return nil, err
	}
SetPassword:
	_, err = softether.API.SetUserPolicy(r.HubName, r.UserName, int(searchLimit)*8, int(searchLimit)*8)
	if err != nil {
		return nil, err
	}

	c.ServerCert = softether.ServerCert
	c.RemoteAccess = softether.RemoteAccess
	c.Ipv4Address = softether.Ipv4Address
	r.Config = *c
	return r, nil
}

func (o *OpenVpn) Start() {
	o.timer = *setInterval(time.Second*30, func(when time.Time) {
		o.syncUserTraffic()
	})
}

func (o *OpenVpn) Stop() {
	o.timer.Stop()
}

func (o *OpenVpn) Close() {
	o.Stop()
	hubname := o.HubName
	username := o.UserName
	out, err := softether.API.ListUser(hubname)
	if err != nil {
		softether.API.DeleteUser(hubname, username)
		return
	}
	names, ok := out["Name"].([]string)
	if ok && len(names) > 1 {
		softether.API.DeleteUser(hubname, username)
	} else {
		softether.API.DeleteHub(hubname)
	}
}

func (o *OpenVpn) IsTimeout() bool {
	if o.Timeout == 0 {
		return false
	}
	if o.GetLastTimeStamp().Unix()+int64(o.Timeout) < time.Now().UTC().Unix() {
		return true
	} else {
		return false
	}
}

func (o *OpenVpn) IsExpire() bool {
	if o.Expire == 0 {
		return false
	}
	if time.Now().UTC().Unix() > int64(o.Expire) {
		return true
	} else {
		return false
	}
}

func (o *OpenVpn) IsOverflow() bool {
	tu, td, uu, ud := o.GetTraffic()
	if tu+td+uu+ud > int64(o.UsedTotalTraffic*1024) {
		return true
	} else {
		return false
	}
}

func (o *OpenVpn) IsNotify() bool {
	if o.BalanceNotifyDuration == 0 {
		return false
	} else {
		if int64(o.Expire)-time.Now().UTC().Unix() < int64(o.BalanceNotifyDuration) {
			return true
		} else {
			return false
		}
	}
}

func (o *OpenVpn) IsStairCase() (limit int, flag bool) {
	tu, td, uu, ud := o.GetTraffic()
	totalFlow := o.UsedTotalTraffic + (tu+td+uu+ud)/1024
	preLimit := int64(o.CurrLimitDown)
	nextLimit, err := searchLimit(preLimit, o.LimitArray, o.FlowArray, totalFlow)
	if preLimit != nextLimit && err == nil {
		return int(nextLimit), true
	} else {
		return 0, false
	}
}

func (o *OpenVpn) GetTraffic() (tu, td, uu, ud int64) {
	return o.Traffic.GetTraffic()
}

func (o *OpenVpn) GetStartTimeStamp() time.Time {
	return o.Traffic.GetStartTimeStamp()
}

func (o *OpenVpn) GetLastTimeStamp() time.Time {
	return o.Traffic.GetLastTimeStamp()
}

func (o *OpenVpn) Clear() {
	o.Traffic.GetTrafficWithClear()
}

func (o *OpenVpn) SetLimit(bytesPerSec int) {
	bytesPerSec = bytesPerSec * 8
	softether.API.SetUserPolicy(o.HubName, o.UserName, bytesPerSec, bytesPerSec)
}

func (o *OpenVpn) Burst() int {
	if out, err := softether.API.GetUser(o.HubName, o.UserName); err == nil {
		maxupload, ok := out["policy:MaxUpload"]
		if ok {
			MaxUpload := maxupload.(int)
			MaxUpload = MaxUpload / (1024 * 8)
			return MaxUpload
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (o *OpenVpn) GetConfig() *Config {
	return &o.Config
}

func (o *OpenVpn) syncUserTraffic() {
	out, err := softether.API.GetUser(o.HubName, o.UserName)
	if err != nil {
		return
	}
	new_tu := out["Send.UnicastBytes"].(int64)
	new_td := out["Recv.UnicastBytes"].(int64)
	new_uu := out["Send.BroadcastBytes"].(int64)
	new_ud := out["Recv.BroadcastBytes"].(int64)

	tu, td, uu, ud := o.Traffic.GetTraffic()
	if new_tu != tu || new_td != td || new_uu != uu || new_ud != ud {
		o.Traffic.SetTraffic(new_tu, new_td, new_uu, new_ud)
	}
}
