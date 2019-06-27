package softether

import "time"

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
	// TODO
	// 只要检查 hub 的最后通信时间，
	// 如果hub的最后通信时间非常久远，就可以把hub删除了。
	hubs, err := API.ListHub()
	if err != nil {
		return
	}
	i_lastCommTime := hubs["LastCommTime"].([]interface{})
	i_hubName := hubs["HubName"].([]interface{})
	for index, value := range i_lastCommTime {
		lastcommtime := time.Unix(value.(int64)/1e3, 0)
		now := time.Now()
		if now.Sub(lastcommtime) >= time.Hour*18 {
			clear_hubname := i_hubName[index].(string)
			API.DeleteHub(clear_hubname)
		}
	}
}
