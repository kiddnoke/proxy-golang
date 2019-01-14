package main

import (
	"flag"
	"time"

	"proxy-golang/manager"
)

func main() {
	var flags struct {
		Uid             int64
		Sid             int64
		Timeout         int64
		Remain          uint
		Expire          uint
		NotifyId        uint
		NotifyTimestamp uint
		ServerPort      int
		Method          string
		Password        string
		Limit           int
	}
	flag.Int64Var(&flags.Uid, "uid", 0, "uid")
	flag.Int64Var(&flags.Sid, "sid", 0, "sid")
	flag.Int64Var(&flags.Timeout, "timeout", 0, "timeout")
	flag.UintVar(&flags.Remain, "remain", 0, "remain")
	flag.UintVar(&flags.Expire, "expire", 0, "expire")
	flag.UintVar(&flags.NotifyId, "notifyid", 0, "notifyid")
	flag.UintVar(&flags.NotifyTimestamp, "notifytimestamp", 0, "notifytimestamp")
	flag.StringVar(&flags.Password, "password", "test", "Password")
	flag.StringVar(&flags.Method, "method", "AES-128-cfb", "Method")
	flag.IntVar(&flags.Limit, "limit", 500, "Limit")
	flag.IntVar(&flags.ServerPort, "port", 0, "ServerPort")
	flag.Parse()
	ch := make(chan int, 1)
	m := manager.New()
	p := &manager.Proxy{
		Uid:        flags.Uid,
		Sid:        flags.Sid,
		Timeout:    flags.Timeout,
		Limit:      flags.Limit,
		ServerPort: flags.ServerPort,
		Method:     flags.Method,
		Password:   flags.Password,
	}
	m.Add(*p)
	timer2 := setInterval(time.Second*5, func() {
		pr, _ := m.Get(*p)
		tu, td, uu, ud := pr.GetTraffic()
		pr.Printf("[%d] [%d] [%d] [%d]", tu, td, uu, ud)
	})
	<-ch
	clearInterval(timer2)
}

type interval struct {
	*time.Ticker
	duration time.Duration
}

func setInterval(duration time.Duration, callback func()) (t *interval) {
	t = new(interval)
	t.duration = duration
	t.Ticker = time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.C:
				callback()
			}
		}
	}()
	return
}
func clearInterval(t *interval) {
	t.Stop()
}

type timeout struct {
	*time.Timer
	time time.Time
}

func setTimeout(duration time.Duration, callback func()) (t *timeout) {
	t = new(timeout)
	t.time = time.Now().Add(duration)
	t.Timer = time.AfterFunc(duration, callback)
	return
}
func clearTimeout(t *timeout) bool {
	t.Stop()
	return true
}
