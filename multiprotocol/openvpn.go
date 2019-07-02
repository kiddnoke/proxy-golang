package multiprotocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/kiddnoke/SoftetherGo"

	"proxy-golang/common"
	"proxy-golang/softether"
)

type OpenVpn struct {
	Config
	HubName  string
	UserName string
	Password string
	common.Traffic
	timer interval
	common.Logger
}

func NewOpenVpn(c *Config) (*OpenVpn, error) {
	r := new(OpenVpn)
	_, level := common.GetDefaultLevel()
	prefix := fmt.Sprintf("Uid[%d] Sid[%d] Port[%d] AppId[%d] Protocol[%s]", c.Uid, c.Sid, c.ServerPort, c.AppId, c.Protocol)
	r.Logger = *common.NewLogger(level, prefix)
	searchLimit, err := searchLimit(int64(c.CurrLimitDown), c.LimitArray, c.FlowArray, c.UsedTotalTraffic)
	c.CurrLimitUp = int(searchLimit)
	c.CurrLimitUp = int(searchLimit)

	r.HubName = fmt.Sprintf("%d", c.Uid)
	r.UserName = fmt.Sprintf("%d", c.Sid)
	r.Password = c.Password
	r.Traffic = common.MakeTraffic()

	c.Method = r.UserName + "@" + r.HubName

	_, err = softether.API.CreateHub(r.HubName, true, softetherApi.HUB_TYPE_STANDALONE)
	if err != nil {
		if e, ok := err.(*softetherApi.ApiError); ok && e.Code() == softetherApi.ERR_HUB_ALREADY_EXISTS {
			r.Logger.Debug("Hub[%s] %s ,then goto CreateUser[%s]", r.HubName, err.Error(), r.UserName)
			goto CreateUser
		}
		r.Error("%s", err.Error())
		return nil, err
	}

	_, err = softether.API.EnableSecureNat(r.HubName)
	if err != nil {
		r.Error("%s", err.Error())
		return nil, err
	}
CreateUser:
	_, err = softether.API.CreateUser(r.HubName, r.UserName, r.Password)
	if err != nil {
		if e, ok := err.(*softetherApi.ApiError); ok && e.Code() == softetherApi.ERR_USER_ALREADY_EXISTS {
			r.Logger.Debug("User[%s] %s ,then goto SetPassword[%s]", r.HubName, err.Error(), r.UserName)
			goto SetPassword
		}
		r.Error("%s", err.Error())
		return nil, err
	} else {
		goto End
	}
SetPassword:
	softether.API.SetUserPassword(r.HubName, r.UserName, r.Password)
	_, err = softether.API.SetUserPolicy(r.HubName, r.UserName, int(searchLimit)*8, int(searchLimit)*8)
	if err != nil {
		return nil, err
	}
End:
	c.ServerCert = softether.ServerCert
	c.RemoteAccess = strings.Replace(softether.RemoteAccess, "1194", fmt.Sprintf("%d", c.ServerPort), -1)
	c.Ipv4Address = softether.Ipv4Address
	r.Config = *c
	return r, nil
}

func (o *OpenVpn) Start() {
	o.Info("Start")
	o.timer = *setInterval(time.Second*30, func(when time.Time) {
		o.syncUserTraffic()
	})
}

func (o *OpenVpn) Stop() {
	o.Info("Stop")
	o.timer.Stop()
}

func (o *OpenVpn) Close() {
	o.Stop()
	o.Info("Close")
	hubname := o.HubName
	username := o.UserName
	out, err := softether.API.ListUser(hubname)
	if err != nil {
		o.Error("%s", err.Error())
		softether.API.DeleteUser(hubname, username)
		return
	}
	names, ok := out["Name"].([]interface{})
	if ok && len(names) > 1 {
		o.Info("Delete User[%s]", username)
		softether.API.DeleteUser(hubname, username)
	} else {
		o.Info("Delete Hub[%s]", hubname)
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
	o.Debug("Clear")
	o.Traffic.GetTrafficWithClear()
}

func (o *OpenVpn) SetLimit(bytesPerSec int) {
	o.Debug("SetLimit")
	if bytesPerSec == 0 {
		return
	}
	bytesPerSec = bytesPerSec * 8
	softether.API.SetUserPolicy(o.HubName, o.UserName, bytesPerSec, bytesPerSec)
}

func (o *OpenVpn) Burst() int {
	brust := o.Config.CurrLimitUp
	o.Debug("Burst is %d", brust)
	return brust
}

func (o *OpenVpn) GetConfig() *Config {
	return &o.Config
}

func (o *OpenVpn) syncUserTraffic() {
	o.Debug("syncUserTraffic")
	out, err := softether.API.GetUser(o.HubName, o.UserName)
	if err != nil {
		o.Error("%s", err.Error())
		return
	}
	new_tu := out["Send.UnicastBytes"].(int64)
	new_td := out["Recv.UnicastBytes"].(int64)
	new_uu := out["Send.BroadcastBytes"].(int64)
	new_ud := out["Recv.BroadcastBytes"].(int64)

	tu, td, uu, ud := o.Traffic.GetTraffic()
	if new_tu != tu || new_td != td || new_uu != uu || new_ud != ud {
		o.Debug("traffic update: tu[%d], td[%d], uu[%d], ud[%d]", new_tu, new_td, new_uu, new_ud)
		o.Traffic.SetTraffic(new_tu, new_td, new_uu, new_ud)
	}
	//
	maxupload, ok := out["policy:MaxUpload"]
	if ok {
		MaxUpload := maxupload.(int)
		MaxUpload = MaxUpload / (1024 * 8)
		o.GetConfig().CurrLimitUp = MaxUpload
		o.GetConfig().CurrLimitDown = MaxUpload
	} else {
		o.Debug("do not get MaxUpload ")
		o.GetConfig().CurrLimitDown = 0
	}
}
func (o *OpenVpn) GetMaxRate() (float64, float64) {
	return 0, 0
}
