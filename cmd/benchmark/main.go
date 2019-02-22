package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"proxy-golang/comm/websocket"
	"proxy-golang/manager"
)

var Manager *manager.Manager

func init() {
	Manager = manager.New()
	Manager.CheckLoop()
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	var generate bool
	var LinkMode string
	var flags struct {
		BeginPort      int
		EndPort        int
		ManagerPort    int
		ControllerPort int
		CenterUrl      string
		State          string
		Area           string
	}
	flag.BoolVar(&generate, "pm2", false, "生成pm2可识别的版本文件")
	flag.StringVar(&LinkMode, "link-mode", "1", "通信模式")
	flag.IntVar(&flags.ManagerPort, "manager-port", 8000, "管理端口(作废)")
	flag.IntVar(&flags.BeginPort, "beginport", 20000, "beginport 起始端口")
	flag.IntVar(&flags.EndPort, "endport", 30000, "endport 结束端口")
	flag.StringVar(&flags.CenterUrl, "url", "localhost:7001", "中心的url地址")
	flag.StringVar(&flags.State, "state", "NULL", "本实例所要注册的国家")
	flag.StringVar(&flags.Area, "area", "0", "本实例所要注册的地区")
	flag.Parse()

	if generate {
		log.Printf("生成pm2版本文件")
		var writeString = fmt.Sprintf("{\"version\":\"%s\"}", BuildDate)
		filename := "./package.json"
		var d1 = []byte(writeString)
		ioutil.WriteFile(filename, d1, 0666)
		return
	}

	client := wswarpper.New()

	host, portStr, err := net.SplitHostPort(flags.CenterUrl)
	if err != nil {
		log.Println(err.Error())
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err.Error())
	}

	_ = client.Connect(host, port)

	Manager.On("timeout", func(uid, sid int64, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: sid, Uid: uid, ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}

		timestamp := pr.GetLastTimeStamp()
		pr.Timeout = 0
		time.AfterFunc(time.Minute*2, func() {
			// 关闭实例
			pr.Close()
			// 回收
			Manager.Delete(proxyinfo)
		})
		client.Timeout(sid, uid, transfer, timestamp.Unix())
		log.Printf("timeout: sid[%d] uid[%d] ,transfer[%d,%d,%d,%d] ,timestamp[%d]", sid, uid, tu, td, uu, ud, timestamp.Unix())
		client.Health(Manager.Size())
	})
	Manager.On("expire", func(uid, sid int64, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: sid, Uid: uid, ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		tu, td, uu, ud := pr.GetTraffic()
		transfer := []int64{tu, td, uu, ud}
		pr.Expire = 0
		time.AfterFunc(time.Minute, func() {
			// 关闭实例
			pr.Close()
			// 回收
			Manager.Delete(proxyinfo)
		})

		client.Expire(sid, uid, transfer)
		log.Printf("expire: sid[%d] uid[%d] ,transfer[%d,%d,%d,%d]", sid, uid, tu, td, uu, ud)
		client.Health(Manager.Size())
	})
	Manager.On("overflow", func(uid, sid int64, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: sid, Uid: uid, ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		tu, td, uu, ud := pr.GetTraffic()
		nextLimit, err := manager.SearchLimit(pr.LimitArray, pr.FlowArray, tu+td+uu+ud/1024+pr.UsedTotalTraffic)
		if err != nil {
			log.Printf("Manager.On overflow event err:%v", err.Error())
		}
		pr.CurrLimitDown = int(math.Min(float64(nextLimit), float64(pr.CurrLimitDown)))
		pr.CurrLimitUp = pr.CurrLimitDown
		log.Printf("overflow: sid[%d] uid[%d] ,Frome CurrLimit[%d]->NextLimit[%d]", sid, uid, pr.CurrLimitDown, nextLimit)
		client.Overflow(sid, uid, int(pr.CurrLimitDown))
		pr.SetLimit(int(pr.CurrLimitDown) * 1024)
	})
	Manager.On("balance", func(uid, sid int64, port int) {
		var proxyinfo manager.Proxy
		proxyinfo = manager.Proxy{Sid: sid, Uid: uid, ServerPort: port}
		pr, err := Manager.Get(proxyinfo)
		if err != nil {
			log.Println(err)
		}
		log.Printf("balance: sid[%d] uid[%d] ,BalanceNotifyDuration[%d]", sid, uid, pr.BalanceNotifyDuration)
		client.Balance(sid, uid, pr.BalanceNotifyDuration)
		pr.BalanceNotifyDuration = 0
	})
	Manager.On("health", func(n int) {
		client.Health(n)
	})
	Manager.On("transfer", func(transferList []interface{}) {
		client.TransferList(transferList)
	})
	Manager.On("benchmark", func(uid, sid int64) {
		client.BenchMark(uid, sid)
	})
	client.OnConnect(func(c wswarpper.Channel) {
		client.OnOpened(func(msg []byte) {
			log.Printf("OnOpend %s", msg)
			var proxyinfo manager.Proxy
			port := manager.GetFreePort(flags.BeginPort, flags.EndPort)
			if err := json.Unmarshal(msg, &proxyinfo); err != nil {
				log.Printf(err.Error())
			}
			proxyinfo.ServerPort = port
			err := Manager.Add(proxyinfo)
			if err != nil {
				log.Printf(err.Error())
				return
			}
			client.Notify("open", proxyinfo)
			client.Health(Manager.Size())
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
				p.Close()
				CloseRetMsg := make(map[string]interface{})
				CloseRetMsg["server_port"] = proxyinfo.ServerPort
				CloseRetMsg["transfer"] = []int64{tu, td, uu, ud}
				CloseRetMsg["sid"] = proxyinfo.Sid
				CloseRetMsg["uid"] = proxyinfo.Uid
				Manager.Delete(proxyinfo)
				client.Notify("close", CloseRetMsg)
				client.Health(Manager.Size())
			}
		})
		client.Login(flags.ManagerPort, flags.BeginPort, flags.EndPort, flags.ManagerPort+1000, flags.State, flags.Area)
	})
	client.OnDisconnect(func(c wswarpper.Channel) {
		client.Connect(host, port)
	})

	wg.Wait()
}

func Profile(port int) {

	// Create a new router
	router := mux.NewRouter()

	// Register pprof handlers
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))

	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:" + strconv.Itoa(port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
