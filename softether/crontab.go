package softether

import (
	"reflect"
	"time"

	"proxy-golang/common"
)

const duration = time.Minute

var selflogger *common.Logger

func CronInit() {
	selflogger = common.NewLogger(common.LOG_DEFAULT, "CronTask")
	//Cron()
	clearExpireUser(time.Now())
	clearExpireHub(time.Now())
}

func Cron() {
	timer := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-timer.C:
				clearExpireHub(time.Now())
				clearExpireUser(time.Now())
			}
		}
	}()
	return
}
func clearExpireHub(now time.Time) {
	// 只要检查 hub 的最后通信时间，
	// 如果hub的最后通信时间非常久远，就可以把hub删除了。
	selflogger.Debug("Check Hub Begin")
	defer selflogger.Debug("Check Hub End")
	API, _ := PoolGetConn()
	hubs, err := API.ListHub()
	if err != nil {
		return
	}
	_, ok := hubs["LastCommTime"]
	if ok == false {
		selflogger.Debug("there are not hubs")
		return
	}

	var clear_hubname_list []string
	if reflect.TypeOf(hubs["LastCommTime"]).Kind() == reflect.Slice {
		i_lastCommTime := hubs["LastCommTime"].([]interface{})
		i_hubName := hubs["HubName"].([]interface{})

		for index, value := range i_lastCommTime {
			lastcommtime := time.Unix(value.(int64)/1e3, value.(int64)%1e3*1e6)
			if now.Sub(lastcommtime) >= time.Hour*24 {
				selflogger.Warn("LastCommTime[%v] now[%v]", lastcommtime, now)
				clear_hubname := i_hubName[index].(string)
				clear_hubname_list = append(clear_hubname_list, clear_hubname)
			}
		}
	} else if reflect.TypeOf(hubs["LastCommTime"]).Kind() == reflect.Int64 {
		lastCommTime := hubs["LastCommTime"].(int64)
		lastcommtime := time.Unix(lastCommTime/1e3, lastCommTime%1e3*1e6)
		if now.Sub(lastcommtime) >= time.Hour*24 {
			selflogger.Warn("LastCommTime[%v] now[%v]", lastcommtime, now)
			clear_hubname_list = append(clear_hubname_list, hubs["HubName"].(string))
		}
	}
	if len(clear_hubname_list) > 0 {
		selflogger.Debug("Delete Hub:%v", clear_hubname_list)
		for _, hubname := range clear_hubname_list {
			API.DeleteHub(hubname)
			selflogger.Warn("Delete Hub[%s]", hubname)
		}
		selflogger.Info("Delete Hubs has been finish", clear_hubname_list)
	}
}
func clearExpireUser(now time.Time) {
	selflogger.Debug("ClearExpireUser Begin")
	defer selflogger.Debug("ClearExpireUser End")
	//遍历所有hub
	API, _ := PoolGetConn()
	hubs, err := API.ListHub()
	if err != nil {
		return
	}
	_, ok := hubs["HubName"]
	if ok == false {
		selflogger.Debug("there are not hubs ")
		return
	}
	var hubs_ []string
	if reflect.TypeOf(hubs["HubName"]).Kind() == reflect.Slice {
		i_hubs := hubs["HubName"].([]interface{})
		for _, ii_hub := range i_hubs {
			hubs_ = append(hubs_, ii_hub.(string))
		}
	} else if reflect.TypeOf(hubs["HubName"]).Kind() == reflect.String {
		hubs_ = append(hubs_, hubs["HubName"].(string))
	}
	for _, hubname := range hubs_ {
		user_out, err := API.ListUser(hubname)
		if err != nil {
			continue
		}
		if user_out["Expires"] == nil {
			continue
		}
		if reflect.TypeOf(user_out["Expires"]).Kind() == reflect.Slice {
			i_Expires := user_out["Expires"].([]interface{})
			for index, expire := range i_Expires {
				_hubname := hubname
				_username := user_out["Name"].([]interface{})[index].(string)
				_expire := time.Unix(expire.(int64)/1e3, expire.(int64)%1e3*1e6)
				selflogger.Debug("expire[%v] now[%v]", _expire, now)
				if now.Sub(_expire) >= time.Second {
					API.DeleteUser(_hubname, _username)
					selflogger.Warn("DeleteUser [%s] in Hub[%s]", _username, _hubname)
				}
			}
		} else if reflect.TypeOf(user_out["Expires"]).Kind() == reflect.Int64 {
			i_expire := user_out["Expires"]
			_expire := time.Unix(i_expire.(int64)/1e3, i_expire.(int64)%1e3*1e6)
			_hubname := hubname
			_username := user_out["Name"].(string)
			selflogger.Debug("expire[%v] now[%v]", _expire, now)
			if now.Sub(_expire) > time.Second {
				API.DeleteUser(_hubname, _username)
				selflogger.Warn("DeleteUser [%s] in Hub[%s]", _username, _hubname)
			}
		}
	}
}
