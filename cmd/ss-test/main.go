package main

import (
	. "../../relay"
	"context"
	"github.com/pkg/profile"
	"log"
	"math"
	"runtime"
	"time"
)

func memstats() runtime.MemStats {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return *stats
}
func bToMb(b uint64) float64 {
	value := float64(b) / 1024 / 1024
	return math.Trunc(value*1e2+0.5) * 1e-2
}
func Monitor(ctx context.Context) {
	for {
		log.Printf("goroutine Num[%d]", runtime.NumGoroutine())

		stats := memstats()
		log.Printf("Process TotalAlloc:%f", bToMb(stats.TotalAlloc))
		log.Printf("Process Sys:%f", bToMb(stats.Sys))
		log.Printf("Process HeapSys:%f", bToMb(stats.HeapSys))
		log.Printf("Process HeapIdle:%f", bToMb(stats.HeapIdle))

		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Second * 5)
		}
	}
}
func main() {
	stoper := profile.Start(profile.MemProfile, profile.MemProfileRate(100), profile.ProfilePath("."))
	defer stoper.Stop()
	//go Monitor(context.Background())
	ch := make(chan int, 1)
	pi, _ := NewProxy(29999, "AES-128-cfb", "test", 300)
	pr, _ := NewProxyRelay(*pi)
	time.AfterFunc(time.Second*10, func() {
		pr.Stop()
		time.AfterFunc(time.Second*5, func() {
			tu, td, uu, ud := pr.GetTraffic()
			log.Printf("总量 [%d] [%d] [%d] [%d]", tu, td, uu, ud)
		})
	})
	time.AfterFunc(time.Second*20, func() {
		pr.Start()
	})
	time.AfterFunc(time.Second*30, func() {
		pr.Stop()
	})
	time.AfterFunc(time.Second*35, func() {
		pr.Start()
	})
	time.AfterFunc(time.Second*50, func() {
		pr.Stop()
		pr.Close()
		_ = pr
	})
	time.AfterFunc(time.Second*100, func() {
		ch <- 1
	})
	pr.Start()
	<-ch
}
