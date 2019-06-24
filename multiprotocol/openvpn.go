package multiprotocol

import (
	"time"
)

func NewOpen() (r Relayer, err error) {
	return
}

type OpenVpn struct {
	running bool
}

func (OpenVpn) Start() {
	panic("implement me")
}

func (OpenVpn) Stop() {
	panic("implement me")
}

func (OpenVpn) Close() {
	panic("implement me")
}

func (OpenVpn) IsTimeout() bool {
	panic("implement me")
}

func (OpenVpn) IsExpire() bool {
	panic("implement me")
}

func (OpenVpn) IsOverflow() bool {
	panic("implement me")
}

func (OpenVpn) IsNotify() bool {
	panic("implement me")
}

func (OpenVpn) IsStairCase() (limit int, flag bool) {
	panic("implement me")
}

func (OpenVpn) GetTraffic() (tu, td, uu, ud int64) {
	panic("implement me")
}

func (OpenVpn) AddTraffic(tu, td, uu, ud int) {
	panic("implement me")
}

func (OpenVpn) Clear() {
	panic("implement me")
}

func (OpenVpn) GetStartTimeStamp() time.Time {
	panic("implement me")
}

func (OpenVpn) GetLastTimeStamp() time.Time {
	panic("implement me")
}

func (OpenVpn) WaitN(n int) (err error) {
	panic("implement me")
}

func (OpenVpn) SetLimit(bytesPerSec int) {
	panic("implement me")
}
