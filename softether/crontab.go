package softether

import (
	"reflect"
	"time"
)

const duration = time.Hour * 6

func Cron() {
	timer := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-timer.C:
				callback(time.Now())
			}
		}
	}()
	return
}
func callback(timestamp time.Time) {
	// 只要检查 hub 的最后通信时间，
	// 如果hub的最后通信时间非常久远，就可以把hub删除了。
	hubs, err := API.ListHub()
	if err != nil {
		return
	}
	_, ok := hubs["lastCommTime"]
	if ok == false {
		return
	}

	var clear_hubname_list []string
	if reflect.TypeOf(hubs["LastCommTime"]).Kind() == reflect.Slice {
		i_lastCommTime := hubs["LastCommTime"].([]interface{})
		i_hubName := hubs["HubName"].([]interface{})

		for index, value := range i_lastCommTime {
			lastcommtime := time.Unix(value.(int64)/1e3, value.(int64)%1e3*1e6)
			now := time.Now()
			if now.Sub(lastcommtime) >= duration*2 {
				clear_hubname := i_hubName[index].(string)
				clear_hubname_list = append(clear_hubname_list, clear_hubname)
			}
		}
	} else if reflect.TypeOf(hubs["LastCommTime"]).Kind() == reflect.Int64 {
		lastCommTime := hubs["LastCommTime"].(int64)
		lastcommtime := time.Unix(lastCommTime/1e3, 0)
		now := time.Now()
		if now.Sub(lastcommtime) >= duration*2 {
			clear_hubname_list = append(clear_hubname_list, hubs["HubName"].(string))
		}
	}

	for _, hubname := range clear_hubname_list {
		API.DeleteHub(hubname)
	}

}
