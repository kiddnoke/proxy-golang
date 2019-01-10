package main

import (
	"context"
	"flag"
	"log"
	"math"
	"runtime"
	"time"

	. "proxy-golang/relay"
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

	var flags struct {
		ServerPort int
		Method     string
		Password   string
		Speed      int
	}
	flag.StringVar(&flags.Password, "k", "test", "Password")
	flag.StringVar(&flags.Method, "m", "AES-128-cfb", "Method")
	flag.IntVar(&flags.Speed, "s", 500, "Limit")
	flag.IntVar(&flags.ServerPort, "port", 0, "ServerPort")
	flag.Parse()
	ch := make(chan int, 1)
	pi, _ := NewProxyInfo(flags.ServerPort, flags.Method, flags.Password, flags.Speed)
	pr, _ := NewProxyRelay(*pi)
	pr.Start()
	<-ch
}
