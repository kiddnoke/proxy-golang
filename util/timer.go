package util

import (
	"math/rand"
	"time"
)

type Timer interface {
	Stop() bool
}

type interval struct {
	*time.Ticker
	done chan<- bool
}

func Interval(duration time.Duration, callback func(when time.Time)) (t *interval) {
	t = new(interval)
	done := make(chan bool)
	t.done = done
	t.Ticker = time.NewTicker(duration)
	go func() {
		defer t.Ticker.Stop()
		for {
			select {
			case <-t.C:
				callback(time.Now())
			case <-done:
				return
			}
		}
	}()
	return
}

func (i *interval) Stop() bool {
	i.done <- true
	return true
}

type timeout struct {
	*time.Timer
}

func Timeout(duration time.Duration, callback func(when time.Time)) (t *timeout) {
	t = new(timeout)
	t.Timer = time.AfterFunc(duration, func() {
		callback(time.Now())
	})
	return
}

type intervalrandom struct {
	*time.Timer
}

func (i *intervalrandom) Stop() bool {
	i.Timer.Stop()
	return true
}
func IntervalRange(first, end time.Duration, callback func(when time.Time)) (t *intervalrandom, err error) {
	t = new(intervalrandom)
	if first > end {
		first, end = end, first
	}
	durationrange := float64((end - first).Nanoseconds()) * rand.Float64()
	durationRange := int64(durationrange)
	duration := first + time.Duration(durationRange)
	t.Timer = time.AfterFunc(duration, func() {
		callback(time.Now())
		IntervalRange(first, end, callback)
	})
	return
}
func IntervalRandom(jichu, fudong time.Duration, callback func(when time.Time)) (t *intervalrandom, err error) {
	return IntervalRange(jichu-fudong, jichu+fudong, callback)
}
