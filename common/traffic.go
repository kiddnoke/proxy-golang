package common

import "time"

type Traffic struct {
	tu              int64
	td              int64
	uu              int64
	ud              int64
	startstamp      time.Time
	lastactivestamp time.Time
}

func MakeTraffic() Traffic {
	return Traffic{
		0, 0, 0, 0,
		time.Now().UTC(),
		time.Now().UTC(),
	}
}
func (t *Traffic) GetTraffic() (tu, td, uu, ud int64) {
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) GetTrafficWithClear() (tu, td, uu, ud int64) {
	defer func() {
		t.tu = 0
		t.td = 0
		t.uu = 0
		t.ud = 0
	}()
	return t.tu, t.td, t.uu, t.ud
}
func (t *Traffic) AddTraffic(tu, td, uu, ud int) {
	if tu+td+uu+ud == 0 {
		return
	}
	t.tu += int64(tu)
	t.td += int64(td)
	t.uu += int64(uu)
	t.ud += int64(ud)
	t.Active()
}

func (t *Traffic) SetTraffic(tu, td, uu, ud int64) {
	if tu+td+uu+ud == 0 {
		return
	}
	t.tu = tu
	t.td = td
	t.uu = uu
	t.ud = ud
	t.Active()
}
func (t *Traffic) Active() {
	t.lastactivestamp = time.Now().UTC()
}
func (t *Traffic) GetLastTimeStamp() time.Time {
	return t.lastactivestamp
}
func (t *Traffic) GetStartTimeStamp() time.Time {
	return t.startstamp
}
