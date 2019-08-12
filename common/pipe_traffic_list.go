package common

import (
	"fmt"
	"log"
	"sort"
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
	var list PipeTrafficList
	p.list.Range(func(key, value interface{}) bool {
		tr := atomic.LoadInt64(value.(*int64))
		list = append(list, struct {
			pipename string
			traffic  int64
		}{pipename: key.(string), traffic: tr})
		p.list.Delete(key.(string))
		return true
	})

	sort.Sort(sort.Reverse(list))
	var str string
	for _, item := range list {
		pipename := item.pipename
		traffic := item.traffic
		str += fmt.Sprintf(`{%s}:%f kb/s +`, pipename, ratter(traffic, d))
	}
	if len(str) > 0 {
		return str[:len(str)-1]
	} else {
		return ""
	}
}
func (p *PipeTrafficSet) SamplingAndPrint(d time.Duration) {
	log.Println(p.SamplingAndString(d))
}

type PipeTrafficList []struct {
	pipename string
	traffic  int64
}

func (p PipeTrafficList) Len() int { return len(p) }
func (p PipeTrafficList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PipeTrafficList) Less(i, j int) bool {
	return p[i].traffic < p[j].traffic
}
