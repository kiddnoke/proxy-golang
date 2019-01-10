package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"strconv"
	"sync"

	"proxy-golang/comm/websocket"
	"proxy-golang/manager"
)

var Manager *manager.Manager

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	var flags struct {
		BeginPort      int
		EndPort        int
		ManagerPort    int
		ControllerPort int
		CenterUrl      string
		State          string
		Area           string
	}
	flag.IntVar(&flags.ManagerPort, "manager-port", 8000, "管理端口(作废)")
	flag.IntVar(&flags.ControllerPort, "controller-port", 9000, "控制端口(作废)")
	flag.IntVar(&flags.BeginPort, "beginport", 20000, "beginport 起始端口")
	flag.IntVar(&flags.EndPort, "endport", 30000, "endport 结束端口")
	flag.StringVar(&flags.CenterUrl, "url", "localhost:7001", "中心的url地址")
	flag.StringVar(&flags.State, "state", "SG", "本实例所要注册的国家")
	flag.StringVar(&flags.Area, "area", "1", "本实例所要注册的地区")
	flag.Parse()

	client := wswarpper.New()

	host, port_str, err := net.SplitHostPort(flags.CenterUrl)
	if err != nil {
		log.Println(err.Error())
	}
	port, err := strconv.Atoi(port_str)
	if err != nil {
		log.Println(err.Error())
	}

	_ = client.Connect(host, port)
	Manager = manager.New()
	go Manager.CheckLoop()
	Manager.On("timeout", func(uid, sid, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: int64(sid), Uid: int64(uid), ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}

		timestamp := pr.GetLastTimeStamp()
		client.Timeout(sid, uid, transfer, timestamp.Unix())
	})
	Manager.On("expire", func(uid, sid, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: int64(sid), Uid: int64(uid), ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}

		client.Expire(sid, uid, transfer)
	})
	Manager.On("overflow", func(uid, sid, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: int64(sid), Uid: int64(uid), ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		client.Overflow(sid, uid, pr.Limit)
	})
	Manager.On("balance", func(uid, sid, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: int64(sid), Uid: int64(uid), ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		client.Balance(sid, uid, pr.BalanceNotifyDuration)
	})
	client.OnConnect(func(c wswarpper.Channel) {
		client.OnOpened(func(msg []byte) {
			var proxyinfo manager.Proxy
			port := manager.GetFreePort(flags.BeginPort, flags.EndPort)
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				log.Printf(err.Error())
			}
			proxyinfo.ServerPort = port
			err := Manager.Add(proxyinfo)
			if err != nil {
				log.Printf(err.Error())
			}
			client.Notify("open", proxyinfo)
		})
		client.OnClosed(func(msg []byte) {
			var proxyinfo manager.Proxy
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				log.Printf(err.Error())
			}
			if p, err := Manager.Get(proxyinfo); err != nil {
				log.Printf(err.Error())
			} else {
				tu, td, uu, ud := p.GetTraffic()
				transfer := []int64{tu, td, uu, ud}
				p.Close()
				Manager.Delete(proxyinfo)
				CloseRetMsg := make(map[string]interface{})
				CloseRetMsg["server_port"] = proxyinfo.ServerPort
				CloseRetMsg["transfer"] = transfer
				CloseRetMsg["sid"] = proxyinfo.Sid
				CloseRetMsg["uid"] = proxyinfo.Uid
				client.Notify("close", CloseRetMsg)
			}
		})
		client.Login(flags.ManagerPort, flags.BeginPort, flags.EndPort, flags.ControllerPort, flags.State, flags.Area)
	})
	client.OnDisconnect(func(c wswarpper.Channel) {
		client.Connect(host, port)
		client.Login(flags.ManagerPort, flags.BeginPort, flags.EndPort, flags.ControllerPort, flags.State, flags.Area)
	})
	wg.Wait()
}
