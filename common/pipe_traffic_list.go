package common

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type PipeTrafficSet struct {
	list sync.Map
}

func NewPipTrafficSet() *PipeTrafficSet {
	return &PipeTrafficSet{sync.Map{}}
}
func (p *PipeTrafficSet) AddTraffic(key string, n int64) {
	v, ok := p.list.Load(key)
	if ok {
		t := v.(*int64)
		atomic.AddInt64(t, n)
	} else {
		p.list.Store(key, &n)
	}
}
func (p *PipeTrafficSet) GetTraffic(Key string) int64 {
	v, ok := p.list.Load(Key)
	if ok {
		t := v.(*int64)
		return atomic.LoadInt64(t)
	} else {
		return 0
	}
}
func ratter(n int64, duration time.Duration) float64 {
	if n > 0 {
		return float64(n) / duration.Seconds() / 1024
	}
	return 0
}
func (p *PipeTrafficSet) SamplingAndString(d time.Duration) string {
	var str string
	p.list.Range(func(key, value interface{}) bool {
		tr := atomic.LoadInt64(value.(*int64))
		str += fmt.Sprintf(`{%s}:%f kb/s +`, key, ratter(tr, d))
		p.list.Delete(key)
		return true
	})
	return str[:len(str)-1]
}
func (p *PipeTrafficSet) SamplingAndPrint(d time.Duration) {
	p.list.Range(func(key, value interface{}) bool {
		tr := atomic.LoadInt64(value.(*int64))
		fmt.Printf(`{%s}:%f kb/s +`, key, ratter(tr, d))
		p.list.Delete(key)
		return true
	})
	fmt.Printf("\b\b\n")
}
