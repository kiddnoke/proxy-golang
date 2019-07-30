package multiprotocol

import (
	"fmt"
	"time"

	"github.com/kiddnoke/SoftetherGo"

	"proxy-golang/common"
	"proxy-golang/softether"
)

const defaulthub = "DEFAULT"

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

	r.HubName = defaulthub
	r.UserName = fmt.Sprintf("%d", c.Sid)
	r.Password = c.Password
	r.Traffic = common.MakeTraffic()

	c.Method = r.UserName + "@" + r.HubName

	API, _ := softether.PoolGetConn()
	_, err = API.CreateHub(r.HubName, true, softetherApi.HUB_TYPE_STANDALONE)
	if err != nil {
		if e, ok := err.(*softetherApi.ApiError); ok && e.Code() == softetherApi.ERR_HUB_ALREADY_EXISTS {
			r.Logger.Debug("Hub[%s] %s ,then goto CreateUser[%s]", r.HubName, err.Error(), r.UserName)
			goto CreateUser
		}
		r.Error("%s", err.Error())
		return nil, err
	}

	_, err = API.EnableSecureNat(r.HubName)
	if err != nil {
		r.Error("%s", err.Error())
		return nil, err
	}
CreateUser:
	_, err = API.CreateUser(r.HubName, r.UserName, fmt.Sprintf("%d", c.Uid), fmt.Sprintf("%d-%d", c.Sid, c.Uid), r.Password)
	if err != nil {
		if e, ok := err.(*softetherApi.ApiError); ok && e.Code() == softetherApi.ERR_USER_ALREADY_EXISTS {
			r.Logger.Debug("User[%s] %s ,then goto SetPassword[%s]", r.HubName, err.Error(), r.UserName)
			goto SetUser
		}
		r.Error("%s", err.Error())
		return nil, err
	} else {
		API.SetUserExpireTime(r.HubName, r.UserName, time.Now().Add(time.Second*time.Duration(c.Timeout)))
		goto End
	}
SetUser:
	API.SetUserPassword(r.HubName, r.UserName, r.Password)
	_, err = API.SetUserPolicy(r.HubName, r.UserName, int(searchLimit)*8, int(searchLimit)*8)
	if err != nil {
		return nil, err
	}
	API.SetUserExpireTime(r.HubName, r.UserName, time.Now().Add(time.Second*time.Duration(c.Timeout)))
End:
	c.ServerCert = softether.ServerCert
	c.RemoteAccess = softether.RemoteAccess
	c.Ipv4Address = softether.Ipv4Address
	r.Config = *c
	return r, nil
}

func (o *OpenVpn) Start() {
	o.Info("Start")
	o.timer = *setInterval(time.Second*10, func(when time.Time) {
		o.syncUserTraffic()
	})
}

func (o *OpenVpn) Stop() {
	defer o.Info("Stop")
	o.timer.Stop()
}

func (o *OpenVpn) Close() {
	API, _ := softether.PoolGetConn()

	o.Stop()
	defer o.Info("Close")
	hubname := o.HubName
	username := o.UserName

	sess, err := softether.NewHubSessions(hubname)
	if err != nil {
		o.Error("%s", err.Error())
		return
	}
	sess.DeleteSessionBySid(username)
	_, err = API.DeleteUser(hubname, username)
	if err != nil {
		o.Error("%v", err)
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
	API, _ := softether.PoolGetConn()
	o.Debug("SetLimit")
	if bytesPerSec == 0 {
		return
	}
	bytesPerSec = bytesPerSec * 8
	API.SetUserPolicy(o.HubName, o.UserName, bytesPerSec, bytesPerSec)
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
	API, _ := softether.PoolGetConn()

	o.Debug("syncUserTraffic")
	out, err := API.GetUser(o.HubName, o.UserName)
	if err != nil {
		o.Error("%s", err.Error())
		if e, ok := err.(softetherApi.ApiError); ok && e.Code() == softetherApi.ERR_OBJECT_NOT_FOUND {
			o.Close()
		} else {
			o.Close()
		}
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
		API.SetUserExpireTime(o.HubName, o.UserName, time.Now().Add(time.Second*time.Duration(o.Timeout)))
	}
	o.Traffic.OnceSampling(time.Second * 10)
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
func (o *OpenVpn) GetRate() (float64, float64) {
	return o.Traffic.GetRate()
}
