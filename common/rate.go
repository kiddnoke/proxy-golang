package common

import (
	"sync/atomic"
	"time"
)

type RateTraffic struct {
	MaxUploadRate   float64
	MaxDownloadRate float64
	UploadTraffic   int64
	DownloadTraffic int64
	Samplingtime    time.Duration
}

func NewRateTraffic() *RateTraffic {
	r := &RateTraffic{
		MaxUploadRate:   0,
		MaxDownloadRate: 0,
		UploadTraffic:   0,
		DownloadTraffic: 0,
		Samplingtime:    time.Second,
	}
	r.sampling()
	return r
}

func (r *RateTraffic) GetMaxRate() (float64, float64) {
	return r.MaxUploadRate, r.MaxDownloadRate
}

func (r *RateTraffic) AddTraffic(u int, d int) {
	r.AddTrafficInt64(int64(u), int64(d))
}
func (r *RateTraffic) AddTrafficInt64(u, d int64) {
	if u > 0 {
		atomic.AddInt64(&r.UploadTraffic, u)
	}

	if d > 0 {
		atomic.AddInt64(&r.DownloadTraffic, d)
	}
}
func (r *RateTraffic) sampling() {
	callback := func(time2 time.Time) {
		if uptotal := atomic.LoadInt64(&r.UploadTraffic); uptotal > 0 {
			rate := float64(uptotal) / r.Samplingtime.Seconds() / 1024
			if rate > r.MaxUploadRate {
				r.MaxUploadRate = rate
				atomic.CompareAndSwapInt64(&r.UploadTraffic, uptotal, 0)
			}
		}
		if downtotal := atomic.LoadInt64(&r.DownloadTraffic); downtotal > 0 {
			rate := float64(downtotal) / r.Samplingtime.Seconds() / 1024
			if rate > r.MaxDownloadRate {
				r.MaxDownloadRate = rate
				atomic.CompareAndSwapInt64(&r.DownloadTraffic, downtotal, 0)
			}
		}
	}
	timer := time.NewTicker(r.Samplingtime)
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
