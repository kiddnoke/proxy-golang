package multiprotocol

import (
	"math/rand"
	"time"
)

type Stopper interface {
	Stop() bool
}
type Timer struct {
	Stopper
}

func clearTimer(t *Timer) bool {
	return t.Stop()
}

type interval struct {
	*time.Ticker
}

func setInterval(duration time.Duration, callback func(when time.Time)) (t *interval) {
	t = new(interval)
	t.Ticker = time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.C:
				callback(time.Now())
			}
		}
	}()
	return
}

type timeout struct {
	*time.Timer
}

func setTimeout(duration time.Duration, callback func(when time.Time)) (t *timeout) {
	t = new(timeout)
	t.Timer = time.AfterFunc(duration, func() {
		callback(time.Now())
	})
	return
}

type intervalrandom struct {
	*time.Timer
}

func setIntervalRange(first, end time.Duration, callback func(when time.Time)) (t *intervalrandom, err error) {
	t = new(intervalrandom)
	if first > end {
		first, end = end, first
	}
	durationrange := float64((end - first).Nanoseconds()) * rand.Float64()
	durationRange := int64(durationrange)
	duration := first + time.Duration(durationRange)
	t.Timer = time.AfterFunc(duration, func() {
		callback(time.Now())
		setIntervalRange(first, end, callback)
	})
	return
}
func setIntervalRandom(jichu, fudong time.Duration, callback func(when time.Time)) (t *intervalrandom, err error) {
	return setIntervalRange(jichu-fudong, jichu+fudong, callback)
}
