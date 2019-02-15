package manager

import "testing"

func TestProxy_Init(t *testing.T) {
	//{"balancenotifytime":0,"currlimitdown":0,"currlimitup":0,"expire":0,"ip":"10.0.2.71","method":"aes-128-cfb","nextlimitdown":0,"nextlimitup":0,"password":"fNnWVk4zcxsD","remain":0,"sid":3135,"timeout":180,"uid":100248}
	p := &Proxy{
		Uid:                   2222,
		Sid:                   2222,
		ServerPort:            10000,
		Method:                "aes-128-cfb",
		Password:              "test",
		Timeout:               180,
		CurrLimitDown:         0,
		NextLimitDown:         0,
		Remain:                0,
		Expire:                0,
		BalanceNotifyDuration: 0,
	}
	if err := p.Init(); err != nil {
		t.FailNow()
	}

}
