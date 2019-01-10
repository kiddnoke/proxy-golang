package main

import (
	"Vpn-golang/manager"
	"flag"
	"runtime"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	var flags struct {
		BeginPort int
		EndPort   int
		Limit     int
		Password  string
		Method    string
	}
	var core int
	flag.IntVar(&flags.BeginPort, "beginport", 20000, "beginport 起始端口")
	flag.IntVar(&flags.EndPort, "endport", 30000, "endport 结束端口")
	flag.IntVar(&flags.Limit, "limit", 0, "限制速度")
	flag.StringVar(&flags.Password, "k", "test", "Password")
	flag.StringVar(&flags.Method, "m", "aes-128-cfb", "encryption method, default: aes-128-cfb")
	flag.IntVar(&core, "Core", 0, "maximum number of CPU cores to use, default is determinied by Go runtime")
	flag.Parse()
	if core > 0 {
		runtime.GOMAXPROCS(core)
	}
	Manager := manager.New()
	for freeport := flags.BeginPort; freeport <= flags.EndPort; freeport++ {
		pi := manager.Proxy{
			ServerPort: freeport,
			Method:     flags.Method,
			Password:   flags.Password,
			Limit:      flags.Limit,
		}
		Manager.Add(pi)
	}
	wg.Wait()
}
