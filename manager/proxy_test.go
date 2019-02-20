package manager

import (
	"encoding/json"
	"log"
	"testing"
)

func TestProxy_Init(t *testing.T) {
	p := &Proxy{
		Uid:                   2222,
		Sid:                   2222,
		ServerPort:            10000,
		Method:                "aes-128-cfb",
		Password:              "test",
		Timeout:               180,
		CurrLimitDown:         0,
		Expire:                0,
		BalanceNotifyDuration: 0,
		// v1.1.1
		SnId:             1,
		AppVersion:       "v1.1.1",
		UserType:         "free",
		CarrierOperators: "SamMobile",
		Os:               "ios",
		DeviceId:         "sdfasdfasdf",
		UsedTotalTraffic: 0,
		LimitArray:       []int64{0},
		FlowArray:        []int64{0},
	}
	if err := p.Init(); err != nil {
		t.FailNow()
	}
}

func TestProxy_Init2(t *testing.T) {
	msg := "{\"app_version\":\"10000012\",\"balancenotifytime\":0,\"carrier_operators\":\"\",\"currlimitdown\":0,\"currlimitup\":0,\"expire\":0,\"flow_array\":[0],\"ip\":\"10.0.2.71\",\"limit_array\":[0],\"method\":\"aes-128-cfb\",\"os\":\"os\",\"password\":\"11111111\",\"sid\":3255,\"sn_id\":100013,\"timeout\":180,\"uid\":100013,\"used_total_traffic\":0}"
	var proxyinfo Proxy
	if err := json.Unmarshal([]byte(msg), &proxyinfo); err != nil {
		log.Printf(err.Error())
	}
	if err := proxyinfo.Init(); err != nil {
		t.FailNow()
	}
}
