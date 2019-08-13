package common

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

const duration = time.Second

type Traffic struct {
	Tu              int64 `json:"tcp_up"`
	Td              int64 `json:"tcp_down"`
	Uu              int64 `json:"udp_up"`
	Ud              int64 `json:"udp_down"`
	StartStamp      int64 `json:"start_stamp"`
	LastActiveStamp int64 `json:"last_active_stamp"`

	PreU         int64   `json:"pre_u"`
	PreD         int64   `json:"pre_d"`
	PreTimeStamp int64   `json:"pre_time_stamp"`
	AvgRate      float64 `json:"avg_rate"`
	MaxRate      float64 `json:"max_rate"`

	SamplingTimer time.Ticker
}

func MakeTraffic() Traffic {
	return Traffic{
		0, 0, 0, 0,
		time.Now().UnixNano(),
		time.Now().UnixNano(),
		0, 0, time.Now().UnixNano(), 0, 0,
		time.Ticker{},
	}
}
func NewTraffic() *Traffic {
	return &Traffic{
		0, 0, 0, 0,
		time.Now().UnixNano(),
		time.Now().UnixNano(),
		0, 0, time.Now().UnixNano(), 0, 0,
		time.Ticker{},
	}
}
func (t *Traffic) GetTraffic() (tu, td, uu, ud int64) {
	return t.Tu, t.Td, t.Uu, t.Ud
}
func (t *Traffic) GetTrafficWithClear() (tu, td, uu, ud int64) {
	defer func() {
		atomic.StoreInt64(&t.Tu, 0)
		atomic.StoreInt64(&t.Td, 0)
		atomic.StoreInt64(&t.Uu, 0)
		atomic.StoreInt64(&t.Ud, 0)
	}()
	return t.Tu, t.Td, t.Uu, t.Ud
}
func (t *Traffic) AddTraffic(tu, td, uu, ud int64) {
	if tu+td+uu+ud == 0 {
		return
	}
	atomic.AddInt64(&t.Tu, tu)
	atomic.AddInt64(&t.Td, td)
	atomic.AddInt64(&t.Uu, uu)
	atomic.AddInt64(&t.Ud, ud)
	t.Active()
}

func (t *Traffic) SetTraffic(tu, td, uu, ud int64) {
	if tu+td+uu+ud == 0 {
		return
	}
	atomic.StoreInt64(&t.Tu, tu)
	atomic.StoreInt64(&t.Td, td)
	atomic.StoreInt64(&t.Uu, uu)
	atomic.StoreInt64(&t.Ud, ud)
	t.Active()
}
func (t *Traffic) Active() {
	timeNowToUint64(&t.LastActiveStamp)
}
func (t *Traffic) GetLastTimeStamp() time.Time {
	return UInt64ToTime(&t.LastActiveStamp)
}
func (t *Traffic) GetStartTimeStamp() time.Time {
	return UInt64ToTime(&t.StartStamp)
}
func (t *Traffic) StartSampling() {
	t.SamplingTimer = *time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.SamplingTimer.C:
				{
					if t != nil {
						t.OnceSampling()
					} else {
						break
					}
				}
			}
		}
	}()
	return
}
func (t *Traffic) StopSampling() {
	t.SamplingTimer.Stop()
}
func (t *Traffic) OnceSampling() float64 {
	var ratter = func(n int64, duration time.Duration) float64 {
		if duration.Seconds() > 0 {
			return float64(n) / duration.Seconds() / 1024
		}
		return 0
	}

	curr := atomic.LoadInt64(&t.Td) + atomic.LoadInt64(&t.Ud)
	d := time.Since(UInt64ToTime(&t.PreTimeStamp))

	rate := ratter(curr-atomic.LoadInt64(&t.PreD), d)
	if rate > t.MaxRate {
		t.MaxRate = rate
	}
	atomic.CompareAndSwapInt64(&t.PreD, t.PreD, curr)
	timeNowToUint64(&t.PreTimeStamp)

	dall := time.Since(UInt64ToTime(&t.StartStamp))
	t.AvgRate = ratter(curr, dall)

	return rate
}

func (t *Traffic) GetRate() (float64, float64) {
	d := UInt64ToTime(&t.LastActiveStamp).Sub(UInt64ToTime(&t.StartStamp))
	if d.Seconds() > 0 {
		t.AvgRate = float64(t.Td+t.Ud) / d.Seconds() / 1024
	}

	t.AvgRate, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", t.AvgRate), 64)
	t.MaxRate, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", t.MaxRate), 64)

	return t.AvgRate, t.MaxRate
}

func UInt64ToTime(u *int64) time.Time {
	value := atomic.LoadInt64(u)
	return time.Unix(value/1e9, value%1e9)
}
func timeNowToUint64(u *int64) {
	t := time.Now().UnixNano()
	atomic.StoreInt64(u, t)
}
