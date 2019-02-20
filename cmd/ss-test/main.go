package main

import (
	"context"
	"flag"
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	. "proxy-golang/relay"

	"proxy-golang/udpposter"
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
	var wg sync.WaitGroup
	wg.Add(1)
	var flags struct {
		ServerPort int
		Method     string
		Password   string
		Speed      int
	}
	flag.StringVar(&flags.Password, "k", "test", "Password")
	flag.StringVar(&flags.Method, "m", "AES-128-cfb", "Method")
	flag.IntVar(&flags.Speed, "limit", 0, "Limit")
	flag.IntVar(&flags.ServerPort, "port", 29999, "ServerPort")
	flag.Parse()
	pi, _ := NewProxyInfo(flags.ServerPort, flags.Method, flags.Password, flags.Speed)
	pr, _ := NewProxyRelay(pi)
	pr.ConnectInfoCallback = func(time_stamp int64, rate int64, localAddress, RemoteAddress string, traffic int64, duration time.Duration) {
		user_id := int64(10203040)
		sn_id := int64(10203040)
		device_id := "11111"
		app_version := "11111"
		os := "zhangsen"
		user_type := "zhangsen"
		carrier_operator := "zhangsen"
		connect_time := int64(duration.Seconds() * 100)
		_ = udpposter.PostParams(user_id, sn_id,
			device_id, app_version, os, user_type, carrier_operator,
			localAddress, RemoteAddress, time_stamp,
			rate, connect_time, traffic)
	}
	pr.SetFlags(log.LstdFlags | log.Lmicroseconds)
	pr.Start()
	wg.Wait()
}
