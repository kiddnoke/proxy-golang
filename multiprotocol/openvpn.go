package multiprotocol

import (
	"time"
)

type OpenVpn struct {
	Config
}

func (o *OpenVpn) Start() {
	panic("implement me")
}

func (o *OpenVpn) Stop() {
	panic("implement me")
}

func (o *OpenVpn) Close() {
	panic("implement me")
}

func (o *OpenVpn) IsTimeout() bool {
	panic("implement me")
}

func (o *OpenVpn) IsExpire() bool {
	panic("implement me")
}

func (o *OpenVpn) IsOverflow() bool {
	panic("implement me")
}

func (o *OpenVpn) IsNotify() bool {
	panic("implement me")
}

func (o *OpenVpn) IsStairCase() (limit int, flag bool) {
	panic("implement me")
}

func (o *OpenVpn) GetTraffic() (tu, td, uu, ud int64) {
	panic("implement me")
}

func (o *OpenVpn) GetStartTimeStamp() time.Time {
	panic("implement me")
}

func (o *OpenVpn) GetLastTimeStamp() time.Time {
	panic("implement me")
}

func (o *OpenVpn) Clear() {
	panic("implement me")
}

func (o *OpenVpn) SetLimit(bytesPerSec int) {
	panic("implement me")
}

func (o *OpenVpn) Burst() int {
	panic("implement me")
}

func (o *OpenVpn) GetConfig() *Config {
	return &o.Config
}
