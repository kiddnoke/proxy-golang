package common

import (
	"time"
)

const duration = time.Second

type Traffic struct {
	Tu              int64     `json:"tcp_up"`
	Td              int64     `json:"tcp_down"`
	Uu              int64     `json:"udp_up"`
	Ud              int64     `json:"udp_down"`
	StartStamp      time.Time `json:"start_stamp"`
	LastactiveStamp time.Time `json:"lastactive_stamp"`

	Pre_u   int64   `json:"pre_u"`
	Pre_d   int64   `json:"pre_d"`
	AvaRate float64 `json:"ava_rate"`
	MaxRate float64 `json:"max_rate"`

	SamplingTimer time.Ticker
}

func MakeTraffic() Traffic {
	return Traffic{
		0, 0, 0, 0,
		time.Now().UTC(),
		time.Now().UTC(),
		0, 0, 0, 0,
		time.Ticker{},
	}
}
func NewTraffic() *Traffic {
	return &Traffic{
		0, 0, 0, 0,
		time.Now().UTC(),
		time.Now().UTC(),
		0, 0, 0, 0,
		time.Ticker{},
	}
}
func (t *Traffic) GetTraffic() (tu, td, uu, ud int64) {
	return t.Tu, t.Td, t.Uu, t.Ud
}
func (t *Traffic) GetTrafficWithClear() (tu, td, uu, ud int64) {
	defer func() {
		t.Tu = 0
		t.Td = 0
		t.Uu = 0
		t.Ud = 0
	}()
	return t.Tu, t.Td, t.Uu, t.Ud
}
func (t *Traffic) AddTraffic(tu, td, uu, ud int) {
	if tu+td+uu+ud == 0 {
		return
	}
	t.Tu += int64(tu)
	t.Td += int64(td)
	t.Uu += int64(uu)
	t.Ud += int64(ud)
	t.Active()
}

func (t *Traffic) SetTraffic(tu, td, uu, ud int64) {
	if tu+td+uu+ud == 0 {
		return
	}
	t.Tu = tu
	t.Td = td
	t.Uu = uu
	t.Ud = ud
	t.Active()
}
func (t *Traffic) Active() {
	t.LastactiveStamp = time.Now().UTC()
}
func (t *Traffic) GetLastTimeStamp() time.Time {
	return t.LastactiveStamp
}
func (t *Traffic) GetStartTimeStamp() time.Time {
	return t.StartStamp
}
func (t *Traffic) Sampling() {
	t.SamplingTimer = *time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.SamplingTimer.C:
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
	curr := t.Tu + t.Uu
	t.Pre_u = curr

	curr = t.Td + t.Ud
	if rate := ratter(curr-t.Pre_d, duration2); rate > t.MaxRate {
		t.MaxRate = rate
	}
	t.Pre_d = curr
}

func (t *Traffic) GetRate() (float64, float64) {
	d := t.LastactiveStamp.Sub(t.StartStamp)
	t.AvaRate = float64(t.Td+t.Ud) / d.Seconds() / 1024
	return t.AvaRate, t.MaxRate
}
