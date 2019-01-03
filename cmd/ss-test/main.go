package main

import (
	. "../../relay"
	"context"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	// go Monitor(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	pi, _ := NewProxy(8488, "AES-128-cfb", "test", 100)
	pr, _ := NewProxyRelay(*pi)
	time.AfterFunc(time.Second*15, func() {
		pr.Stop()
		tu, td, uu, ud := pr.GetTraffic()
		log.Printf("总量 [%d] [%d] [%d] [%d]", tu, td, uu, ud)
	})
	pr.Start()
	<-sigCh
}
