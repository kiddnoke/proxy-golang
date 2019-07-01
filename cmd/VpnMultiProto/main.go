package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"proxy-golang/softether"
	"strconv"
	"sync"
	"time"

	"proxy-golang/comm/websocket"
	"proxy-golang/multiprotocol"
	"proxy-golang/pushService"
)

const ssBeginPort = 20000

var Manager *multiprotocol.Manager
var pushSrv *pushService.PushService

func init() {
	Manager = multiprotocol.New()
	Manager.CheckLoop()
	pushSrv, _ = pushService.NewPushService()
	softether.SoftHost = "localhost"
	softether.SoftPort = 443
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	var generate bool
	var LinkMode string
	var flags struct {
		BeginPort      int
		EndPort        int
		InstanceID     int
		ControllerPort int
		CenterUrl      string
		State          string
		Area           string
	}
	flag.BoolVar(&generate, "pm2", false, "生成pm2可识别的版本文件")
	flag.StringVar(&LinkMode, "link-mode", "1", "通信模式")
	flag.IntVar(&flags.InstanceID, "Id", 0, "实例id")
	flag.IntVar(&flags.BeginPort, "beginport", 0, "beginport 起始端口")
	flag.IntVar(&flags.EndPort, "endport", 0, "endport 结束端口")
	flag.StringVar(&flags.CenterUrl, "url", "localhost:7001", "中心的url地址")
	flag.StringVar(&flags.State, "state", "NULL", "本实例所要注册的国家")
	flag.StringVar(&flags.Area, "area", "0", "本实例所要注册的地区")
	flag.Parse()

	if generate {
		GeneratePm2ConfigFile()
		return
	}
	//
	softether.SoftPassword = flags.CenterUrl
	go softether.Init()

	flags.InstanceID = multiprotocol.InstanceIdGen(flags.InstanceID)

	if flags.BeginPort == 0 && flags.EndPort == 0 {
		flags.BeginPort = ssBeginPort + flags.InstanceID*1000
		flags.EndPort = flags.BeginPort + 999
	}

	multiprotocol.BeginPort = flags.BeginPort
	multiprotocol.EndPort = flags.EndPort

	go HttpSrv(flags.InstanceID)

	client := wswrapper.New()

	host, portStr, err := net.SplitHostPort(flags.CenterUrl)
	if err != nil {
		log.Println(err.Error())
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err.Error())
	}

	_ = client.Connect(host, port)
	// timeout Handle
	Manager.On("timeout", func(uid, sid int64, port int, appid int64, protocol string) {
		var proxyinfo multiprotocol.Config
		proxyinfo = multiprotocol.Config{Sid: sid, Uid: uid, ServerPort: port, AppId: appid, Protocol: protocol}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
			return
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}

		timestamp := pr.GetLastTimeStamp()
		duration := int64(pr.GetLastTimeStamp().Sub(pr.GetStartTimeStamp()).Seconds())
		pr.GetConfig().Timeout = 0
		time.AfterFunc(time.Minute*2, func() {
			// 回收
			Manager.Delete(proxyinfo)
		})
		client.Timeout(appid, sid, uid, transfer, timestamp.Unix(), duration)
		log.Printf("timeout: appid[%d] sid[%d] uid[%d] ,transfer[%d,%d,%d,%d] ,timestamp[%d] duration[%d]", appid, sid, uid, tu, td, uu, ud, timestamp.Unix(), duration)
		client.Health(Manager.Health())
		client.Size(Manager.Size())
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "timeout", time.Now().UTC().Unix())
	})
	//expire Handle
	Manager.On("expire", func(uid, sid int64, port int, appid int64, protocol string) {
		var proxyinfo multiprotocol.Config
		proxyinfo = multiprotocol.Config{Sid: sid, Uid: uid, ServerPort: port, AppId: appid, Protocol: protocol}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
			return
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}
		pr.GetConfig().Expire = 0
		time.AfterFunc(time.Minute, func() {
			// 回收
			Manager.Delete(proxyinfo)
		})

		duration := int64(pr.GetLastTimeStamp().Sub(pr.GetStartTimeStamp()).Seconds())
		client.Expire(appid, sid, uid, transfer, duration)
		log.Printf("expire: appid[%d] sid[%d] uid[%d] ,transfer[%d,%d,%d,%d] duration[%d]", appid, sid, uid, tu, td, uu, ud, duration)
		client.Health(Manager.Health())
		client.Size(Manager.Size())
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "expire", time.Now().UTC().Unix())
	})
	// overflow Handle
	Manager.On("overflow", func(uid, sid int64, port int, appid int64, limit int, protocol string) {
		var proxyinfo multiprotocol.Config
		proxyinfo = multiprotocol.Config{Sid: sid, Uid: uid, ServerPort: port, AppId: appid, Protocol: protocol}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
			return
		}

		pr.GetConfig().CurrLimitDown = limit
		pr.GetConfig().CurrLimitUp = limit
		client.Overflow(appid, sid, uid, limit)
		client.Health(Manager.Health())
		pr.SetLimit(limit * 1024)
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "overflow", limit)

	})
	// balance Handle
	Manager.On("balance", func(uid, sid int64, port int, appid int64, protocol string) {
		var proxyinfo multiprotocol.Config
		proxyinfo = multiprotocol.Config{Sid: sid, Uid: uid, ServerPort: port, AppId: appid, Protocol: protocol}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("balance: appid[%d] sid[%d] uid[%d] ,BalanceNotifyDuration[%d]", appid, sid, uid, pr.GetConfig().BalanceNotifyDuration)
		client.Balance(appid, sid, uid, pr.GetConfig().BalanceNotifyDuration)
		pr.GetConfig().BalanceNotifyDuration = 0
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "balance", time.Now().UTC().Unix())
	})
	// health Handle
	Manager.On("health", func() {
		client.Health(Manager.Health())
		client.Size(Manager.Size())
	})
	// transfer Handle
	Manager.On("transfer", func(appid, sid int64, transfer []int64) {
		client.Transfer(appid, sid, transfer)
	})
	// trnasferlist Handle
	Manager.On("transferlist", func(transferlist []interface{}) {
		client.TransferList(transferlist)
	})
	// OnConnect Handle
	client.OnConnect(func(c wswrapper.Channel) {
		// OnOpened Handle
		client.OnOpened(func(msg []byte) {
			var proxyinfo multiprotocol.Config
			log.Printf("OnOpen %s", msg)
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				log.Printf(err.Error())
			}
			log.Printf("OnOpen Struct :%v", proxyinfo)

			OpenRetMsg := make(map[string]interface{})
			OpenRetMsg["sid"] = proxyinfo.Sid
			OpenRetMsg["uid"] = proxyinfo.Uid
			OpenRetMsg["app_id"] = proxyinfo.AppId
			OpenRetMsg["protocol"] = proxyinfo.Protocol
			err := Manager.Add(&proxyinfo)
			if err != nil {
				log.Printf(err.Error())
				OpenRetMsg["error"] = fmt.Sprintf("%s", err.Error())
			} else {
				OpenRetMsg["server_port"] = proxyinfo.ServerPort
				OpenRetMsg["port"] = proxyinfo.ServerPort
				OpenRetMsg["limit"] = proxyinfo.CurrLimitDown
				OpenRetMsg["method"] = proxyinfo.Method
				OpenRetMsg["password"] = proxyinfo.Password
				if proxyinfo.Protocol == "open" {
					OpenRetMsg["server_cert"] = proxyinfo.ServerCert
					OpenRetMsg["remote_access"] = proxyinfo.RemoteAccess
					OpenRetMsg["ip"] = proxyinfo.Ipv4Address
				}
			}
			client.Notify("open", OpenRetMsg)
			client.Health(Manager.Health())
			client.Size(Manager.Size())
			return
		})
		// OnClosed Handle
		client.OnClosed(func(msg []byte) {
			log.Printf("OnClose %s", msg)
			var proxyinfo multiprotocol.Config
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				log.Printf(err.Error())
			}
			log.Printf("OnClose Struct :%v", proxyinfo)

			CloseRetMsg := make(map[string]interface{})
			CloseRetMsg["sid"] = proxyinfo.Sid
			CloseRetMsg["uid"] = proxyinfo.Uid
			CloseRetMsg["app_id"] = proxyinfo.AppId
			CloseRetMsg["protocol"] = proxyinfo.Protocol

			if p, err := Manager.Get(proxyinfo); err != nil {
				log.Printf(err.Error())
				CloseRetMsg["error"] = fmt.Sprintf("%s", err.Error())
				client.Notify("close", CloseRetMsg)
				client.Health(Manager.Health())
			} else {
				tu, td, uu, ud := p.GetTraffic()
				CloseRetMsg["server_port"] = proxyinfo.ServerPort
				CloseRetMsg["port"] = proxyinfo.ServerPort
				CloseRetMsg["transfer"] = []int64{tu, td, uu, ud}
				CloseRetMsg["duration"] = int64(p.GetLastTimeStamp().Sub(p.GetStartTimeStamp()).Seconds())
				Manager.Delete(proxyinfo)
				client.Notify("close", CloseRetMsg)
				client.Health(Manager.Health())
				return
			}
		})
		client.Login(flags.InstanceID, flags.BeginPort, flags.EndPort, flags.InstanceID+10000, flags.State, flags.Area)
	})
	// OnDisConnect Handle
	client.OnDisconnect(func(c wswrapper.Channel) {
		client.Connect(host, port)
	})

	wg.Wait()
}

func HttpSrv(port int) {
	port = port + multiprotocol.ManagerBeginPort
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
	router.Handle("/socket.io/", pushSrv)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:" + strconv.Itoa(port),
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
