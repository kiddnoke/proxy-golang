package main

import (
	"context"
	"flag"
	"log"
	"math"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"sync"
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

	go HttpSrv(flags.ServerPort % 10000)

	pi, _ := NewProxyInfo(flags.ServerPort, flags.Method, flags.Password, flags.Speed)
	pr, _ := NewProxyRelay(pi)

	pr.SetFlags(log.LstdFlags | log.Lmicroseconds)
	pr.Start()
	wg.Wait()
}

func HttpSrv(port int) {

	// Create a new router
	router := http.NewServeMux()

	// Register pprof handlers
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	// Register wsServer handlers

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:" + strconv.Itoa(port),
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
