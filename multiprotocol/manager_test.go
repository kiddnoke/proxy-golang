package multiprotocol

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"
	"sync"
	"testing"
	"time"
)

/*

{\"Localip\":\"10.15.121.17\",\"StateAbb\":\"HK\",\"AreaCode\":1,\"DeviceOS\":\"1\",\"DeviceId\":\"c0aeeafba07a36a44000000000000000\",\"TimeSpanUsable\":81143,\"StartTimeStamp\":1561364663,\"ServerType\":1,\"NetworkType\":\"wifi\",\"Protocol\":\"ss\",\"SnId\":100547,\"Uid\":100547,\"AppVersion\":\"10000041\",\"CarrierOperators\":\"\",\"UserType\":\"vip0\",\"Limit\":{\"CurrLimitUp\":0,\"CurrLimitDown\":0,\"UsedTotalTraffic\":30930},\"EventId\":1300772}
{\"app_id\":0,\"app_version\":\"10000041\",\"balancenotifytime\":0,\"carrier_operators\":\"\",\"currlimitdown\":0,\"currlimitup\":0,\"device_id\":\"c0aeeafba07a36a44000000000000000\",\"expire\":1561445705,\"flow_array\":[10240,20480,30720],\"ip\":\"10.0.2.71\",\"limit_array\":[100,50,20],\"method\":\"aes-256-cfb\",\"network_type\":\"wifi\",\"os\":\"android\",\"password\":\"9TpDEzI9O4ap\",\"sid\":1300771,\"sn_id\":100547,\"timeout\":300,\"uid\":100547,\"used_total_traffic\":30930,\"user_type\":\"vip0\"}
*/
var Json []string = []string{
	"{\"app_id\":0,\"app_version\":\"10000041\",\"balancenotifytime\":0,\"carrier_operators\":\"\",\"currlimitdown\":0,\"currlimitup\":0,\"device_id\":\"c0aeeafba07a36a44000000000000000\",\"expire\":1561445705,\"flow_array\":[10240,20480,30720],\"ip\":\"10.0.2.71\",\"limit_array\":[100,50,20],\"method\":\"aes-256-cfb\",\"network_type\":\"wifi\",\"os\":\"android\",\"password\":\"9TpDEzI9O4ap\",\"sid\":1300771,\"sn_id\":100547,\"timeout\":300,\"uid\":100547,\"used_total_traffic\":30930,\"user_type\":\"vip0\"}",
}
var m *Manager

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

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:" + strconv.Itoa(port),
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func init() {
	m = New()
	BeginPort = 2000
	EndPort = 3000
	go HttpSrv(2999)
}

func TestManager_Add(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var proxyinfo Config
	json_str := []byte(Json[0])
	if err := json.Unmarshal(json_str, &proxyinfo); err != nil {
		log.Printf(err.Error())
		t.FailNow()
	}
	if err := m.Add(&proxyinfo); err != nil {
		log.Printf(err.Error())
		t.FailNow()
	}
	m.CheckLoop()
	wg.Wait()
}
