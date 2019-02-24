package manager

import (
	"encoding/json"
	"testing"
	"time"
)

var m *Manager

const json_str = "{\"app_version\":\"benchmark\",\"balancenotifytime\":0,\"carrier_operators\":null,\"currlimitdown\":0,\"currlimitup\":0,\"device_id\":\"20000\",\"expire\":0,\"flow_array\":[10240,20480,30720],\"ip\":\"10.15.121.114\",\"limit_array\":[20,40,60],\"method\":\"aes-128-cfb\",\"os\":\"pc\",\"password\":\"d9giTTB1EMGI\",\"sid\":20096,\"sn_id\":20000,\"timeout\":3600,\"uid\":20000,\"used_total_traffic\":0,\"user_type\":\"free\"}"

func init() {
	m = New()
}
func TestManager_Add(t *testing.T) {
	var proxy Proxy
	if err := json.Unmarshal([]byte(json_str), &proxy); err != nil {
		t.FailNow()
	}
	if err := m.Add(proxy); err != nil {
		t.FailNow()
	}
}
func TestManager_Get(t *testing.T) {
	var proxy Proxy
	if err := json.Unmarshal([]byte(json_str), &proxy); err != nil {
		t.FailNow()
	}
	if p, err := m.Get(proxy); err != nil {
		t.FailNow()
	} else {
		if p.AppVersion != "benchmark" {
			t.FailNow()
		}
	}
}
func TestManager_Delete(t *testing.T) {
	var proxy Proxy
	proxy = Proxy{
		Uid:                   1,
		Sid:                   2,
		ServerPort:            23233,
		Method:                "aes-128-cfb",
		Password:              "11111",
		CurrLimitUp:           10,
		CurrLimitDown:         80,
		Timeout:               180,
		Expire:                time.Now().Add(time.Minute * 3).Unix(),
		BalanceNotifyDuration: 800,
		// v1.1.1
		SnId:             1,
		AppVersion:       "test",
		UserType:         "free",
		CarrierOperators: "中国移动",
		Os:               "ios",
		DeviceId:         "12121212",
		UsedTotalTraffic: 1024 * 1024 * 1024,
		LimitArray:       []int64{20, 40, 60},
		FlowArray:        []int64{10240, 20480, 30720},
	}
	if err := m.Add(proxy); err != nil {
		t.FailNow()
	}
	if p, err := m.Get(proxy); err != nil {
		t.FailNow()
	} else {
		if p.Burst() != 60*1024 {
			t.FailNow()
		}
	}
	if err := m.Delete(proxy); err != nil {
		t.FailNow()
	}
}
func TestManager_Update(t *testing.T) {

}

func TestManager_Size(t *testing.T) {
	if m.Size() != 1 {
		t.FailNow()
	}
}
