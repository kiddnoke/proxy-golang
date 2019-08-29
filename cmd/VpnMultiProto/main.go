package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"sync"
	"time"

	"proxy-golang/comm/websocket"
	"proxy-golang/common"
	"proxy-golang/multiprotocol"
	"proxy-golang/pushService"
	"proxy-golang/softether"
	"proxy-golang/udpposter"
	"proxy-golang/util"
)

const ManagerBeginPort = 10000

var Manager *multiprotocol.Manager
var pushSrv *pushService.PushService
var mainlog *common.Logger

func init() {
	mainlog = common.NewLogger(common.LOG_INFO, "MainLogic")
	Manager = multiprotocol.New()
	Manager.CheckLoop()
	pushSrv, _ = pushService.NewPushService()
	softether.SoftHost = "10.0.2.70"
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

	// Init InstanceID
	{
		tl, ul := util.FreeListenerRange(10001, 11000)
		ul.Close()
		_, port, _ := net.SplitHostPort(tl.Addr().String())
		flags.InstanceID, _ = strconv.Atoi(port)
		flags.InstanceID = flags.InstanceID - ManagerBeginPort
		go HttpSrv(tl)
	}
	// udp echo service
	{
		go delayer()
	}

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
		proxyinfo = *pr.GetConfig()

		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}

		timestamp := pr.GetLastTimeStamp()
		duration := int64(pr.GetLastTimeStamp().Sub(pr.GetStartTimeStamp()).Seconds())
		pr.GetConfig().Timeout = 0
		time.AfterFunc(time.Second*30, func() {
			// 回收
			err := Manager.Delete(proxyinfo)
			if err != nil && errors.Cause(err) != multiprotocol.KeyNotExist {
				mainlog.Error("%+v", err)
			}
		})
		avgrate, maxrate := pr.GetRate()
		client.Timeout(appid, sid, uid, transfer, timestamp.Unix(), duration, [2]float64{avgrate, maxrate})
		pr.Warn("timeout: transfer[%d,%d,%d,%d] ,timestamp[%d] duration[%d] ,rate[%v]", appid, sid, uid, tu, td, uu, ud, timestamp.Unix(), duration, []float64{avgrate, maxrate})
		client.Health(Manager.Health())
		client.Size(Manager.Size())
		udpposter.PostMaxRate(appid, uid, sid, proxyinfo.DeviceId, proxyinfo.AppVersion, proxyinfo.Os, proxyinfo.UserType, proxyinfo.CarrierOperators, proxyinfo.NetworkType, int64(maxrate*100), duration*100, int64(tu+td+uu+ud), proxyinfo.Ip, proxyinfo.State, proxyinfo.UserType)
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "timeout", time.Now().Unix())
	})
	Manager.On("fast_release", func(uid, sid int64, port int, appid int64, protocol string) {
		var proxyinfo multiprotocol.Config
		proxyinfo = multiprotocol.Config{Sid: sid, Uid: uid, ServerPort: port, AppId: appid, Protocol: protocol}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			if errors.Cause(err) != multiprotocol.KeyNotExist {
				log.Printf("%+v", err)
			}
			return
		}
		proxyinfo = *pr.GetConfig()
		tu, td, uu, ud := pr.GetTraffic()
		timestamp := pr.GetLastTimeStamp()
		duration := int64(pr.GetLastTimeStamp().Sub(pr.GetStartTimeStamp()).Seconds())
		pr.GetConfig().Timeout = 0
		err = Manager.Delete(proxyinfo)
		if err != nil {
			mainlog.Error("%+v", err)
		}
		pr.Warn("fast_release: transfer[%d,%d,%d,%d] ,timestamp[%d] duration[%d]", appid, sid, uid, tu, td, uu, ud, timestamp.Unix(), duration)
		client.Health(Manager.Health())
		client.Size(Manager.Size())
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
		proxyinfo = *pr.GetConfig()

		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}
		pr.GetConfig().Expire = 0
		time.AfterFunc(time.Second*30, func() {
			// 回收
			Manager.Delete(proxyinfo)
		})

		duration := int64(pr.GetLastTimeStamp().Sub(pr.GetStartTimeStamp()).Seconds())
		avgrate, maxrate := pr.GetRate()
		client.Expire(appid, sid, uid, transfer, duration, [2]float64{avgrate, maxrate})
		pr.Warn("expire: transfer[%d,%d,%d,%d] duration[%d] rate[%v]", appid, sid, uid, tu, td, uu, ud, duration, []float64{avgrate, maxrate})
		client.Health(Manager.Health())
		client.Size(Manager.Size())
		udpposter.PostMaxRate(appid, uid, sid, proxyinfo.DeviceId, proxyinfo.AppVersion, proxyinfo.Os, proxyinfo.UserType, proxyinfo.CarrierOperators, proxyinfo.NetworkType, int64(maxrate*100), duration*100, int64(tu+td+uu+ud), proxyinfo.Ip, proxyinfo.State, proxyinfo.UserType)
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "expire", time.Now().Unix())
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
		pr.Warn("balance: BalanceNotifyDuration[%d]", appid, sid, uid, pr.GetConfig().BalanceNotifyDuration)
		client.Balance(appid, sid, uid, pr.GetConfig().BalanceNotifyDuration)
		pr.GetConfig().BalanceNotifyDuration = 0
		key := pushService.GeneratorKey(uid, sid, port, appid)
		_ = pushSrv.Push(key, "balance", time.Now().Unix())
	})
	// health Handle
	Manager.On("health", func() {
		client.Health(Manager.Health())
		client.Size(Manager.Size())
	})
	// transfer Handle
	Manager.On("transfer", func(appid, sid int64, transfer []int64, maxrate [2]float64) {
		client.Transfer(appid, sid, transfer, maxrate)
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
			mainlog.Info("OnOpen %s", msg)
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				mainlog.Error(err.Error())
			}
			mainlog.Debug("OnOpen Struct :%v", proxyinfo)

			OpenRetMsg := make(map[string]interface{})
			OpenRetMsg["sid"] = proxyinfo.Sid
			OpenRetMsg["uid"] = proxyinfo.Uid
			OpenRetMsg["app_id"] = proxyinfo.AppId
			OpenRetMsg["protocol"] = proxyinfo.Protocol
			err := Manager.Add(&proxyinfo)
			if err != nil {
				mainlog.Error(err.Error())
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
			mainlog.Warn("OnClose %s", msg)
			var proxyinfo multiprotocol.Config
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				mainlog.Error(err.Error())
			}
			mainlog.Debug("OnClose Struct :%v", proxyinfo)

			CloseRetMsg := make(map[string]interface{})
			CloseRetMsg["sid"] = proxyinfo.Sid
			CloseRetMsg["uid"] = proxyinfo.Uid
			CloseRetMsg["app_id"] = proxyinfo.AppId
			CloseRetMsg["protocol"] = proxyinfo.Protocol

			if p, err := Manager.Get(proxyinfo); err != nil {
				mainlog.Error(err.Error())
				CloseRetMsg["error"] = fmt.Sprintf("%s", err.Error())
				client.Notify("close", CloseRetMsg)
				client.Health(Manager.Health())
			} else {
				proxyinfo = *p.GetConfig()
				tu, td, uu, ud := p.GetTraffic()
				CloseRetMsg["server_port"] = proxyinfo.ServerPort
				CloseRetMsg["port"] = proxyinfo.ServerPort
				CloseRetMsg["transfer"] = []int64{tu, td, uu, ud}
				CloseRetMsg["duration"] = int64(p.GetLastTimeStamp().Sub(p.GetStartTimeStamp()).Seconds())
				ava, max := p.GetRate()
				CloseRetMsg["maxrate"] = []float64{ava, max}
				Manager.Delete(proxyinfo)
				client.Notify("close", CloseRetMsg)
				client.Health(Manager.Health())
				udpposter.PostMaxRate(proxyinfo.AppId, proxyinfo.Uid, proxyinfo.Sid, proxyinfo.DeviceId, proxyinfo.AppVersion, proxyinfo.Os, proxyinfo.UserType, proxyinfo.CarrierOperators, proxyinfo.NetworkType, int64(max*100), int64(p.GetLastTimeStamp().Sub(p.GetStartTimeStamp()).Seconds())*100, int64(tu+td+uu+ud), proxyinfo.Ip, proxyinfo.State, proxyinfo.UserType)
				return
			}
		})
		client.Login(flags.InstanceID, util.PortBegin, util.PortEnd, flags.InstanceID+10000, flags.State, flags.Area)
	})
	// OnDisConnect Handle
	client.OnDisconnect(func(c wswrapper.Channel) {
		_ = client.Connect(host, port)
		return
	})

	wg.Wait()
}

func HttpListenPort(port int) {
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
func HttpSrv(tl *net.TCPListener) {
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
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	log.Fatal(srv.Serve(tl))
}
