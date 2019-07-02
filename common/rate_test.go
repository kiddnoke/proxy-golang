package common

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestMakeRateTraffic(t *testing.T) {
	NewRateTraffic()
}
func TestMakeRateTraffic2(t *testing.T) {
	r := NewRateTraffic()
	log.Printf("%.2f", r.Samplingtime.Seconds())
}

func TestRateTraffic_GetMaxRate(t *testing.T) {
	r := NewRateTraffic()
	var sw sync.WaitGroup
	sw.Add(2)
	go func() {
		for i := 0; i < 100; i++ {
			r.AddTraffic(i, i)
		}
		sw.Add(-1)
	}()
	go func() {
		for i := 0; i < 100; i++ {
			r.AddTraffic(i, i)
		}
		sw.Add(-1)
	}()
	sw.Wait()
	<-time.After(time.Second * 1)
	log.Println(r.GetMaxRate())

}
