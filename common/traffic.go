package common

import (
	"time"
)

const duration = time.Second

type Traffic struct {
	tu              int64
	td              int64
	uu              int64
	ud              int64
	startstamp      time.Time
	lastactivestamp time.Time

	pre_u   int64
	pre_d   int64
	minRate float64
	maxRate float64
}

func MakeTraffic() Traffic {
	return Traffic{
		0, 0, 0, 0,
		time.Now().UTC(),
		time.Now().UTC(),
		0, 0, 0, 0,
	}
}
func NewTraffic() *Traffic {
	return &Traffic{
		0, 0, 0, 0,
		time.Now().UTC(),
		time.Now().UTC(),
		0, 0, 0, 0,
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
func (t *Traffic) Sampling() {
	timer := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-timer.C:
				{
					t.OnceSampling(duration)
				}
			}
		}
	}()
	return
}
func (t *Traffic) OnceSampling(duration2 time.Duration) {
	var ratter = func(n int64, duration time.Duration) float64 {
		if n > 0 {
			return float64(n) / duration.Seconds() / 1024
		}
		return 0
	}
	curr := t.tu + t.uu
	t.pre_u = curr

	curr = t.td + t.ud
	if rate := ratter(curr-t.pre_d, duration2); rate > t.maxRate {
		t.maxRate = rate
	} else if rate > 0 && rate < t.minRate {
		t.minRate = rate
	}
	t.pre_d = curr
}
func (t *Traffic) GetRate() (float64, float64) {
	return t.minRate, t.maxRate
}
