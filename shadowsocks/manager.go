package shadowsocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Manager struct {
	sync.Mutex
	proxyTable map[int]Proxy
	conn       *net.UDPConn
	clients    map[*net.UDPAddr]bool
	running    bool
}

func NewManager(port int) (m *Manager) {
	// 端口是否在用
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("localhost"), Port: port})
	if err != nil {
		log.Printf("Error listening: %s", err.Error())
		os.Exit(1)
	}
	return &Manager{proxyTable: make(map[int]Proxy), conn: conn, clients: make(map[*net.UDPAddr]bool), running: false}
}
func MakeManger(port int) (m Manager) {
	return *NewManager(port)
}
func (m *Manager) Add(config SSconfig) (e error) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 没有就加
	// 有就报错
	p, ok := m.proxyTable[config.ServerPort]
	if ok == true {
		e = errors.New(fmt.Sprintf("这个实例已经存在了 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Config.Uid, config.Sid, p.Config.Sid))
		return
	} else {
		if proxy, err := MakeProxy(config); err == nil {
			m.proxyTable[config.ServerPort] = proxy
		} else {
			log.Printf("Add Proxy Error:%s", err.Error())
		}
	}
	return
}
func (m *Manager) Get(config SSconfig) (p Proxy, e error) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有就返回
	// 没有就报错
	p, ok := m.proxyTable[config.ServerPort]
	if ok == false {
		e = errors.New(fmt.Sprintf("没有这个实例 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Config.Uid, config.Sid, p.Config.Sid))
	}
	return
}
func (m *Manager) Remove(config SSconfig) (e error) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有就删除
	// 没有就说没这个key
	p, ok := m.proxyTable[config.ServerPort]
	if ok == true /*&& p.Config.Uid == config.Uid && p.Config.Sid == config.Sid */ {
		p.Stop()
		delete(m.proxyTable, config.ServerPort)
	} else {
		e = errors.New(fmt.Sprintf("没有这个实例 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Config.Uid, config.Sid, p.Config.Sid))
	}
	return
}
func (m *Manager) Limit(config SSconfig) (e error) {
	m.Lock()
	defer m.Unlock()
	p, ok := m.proxyTable[config.ServerPort]
	if ok == true /* && p.Config.Uid == config.Uid && p.Config.Sid == config.Sid */ {
		p.SetLimit(config.Limitdown * 1024)
	} else {
		e = errors.New(fmt.Sprintf("没有这个实例 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Config.Uid, config.Sid, p.Config.Sid))
	}
	return
}
func (m *Manager) Update(port int, config SSconfig) (e error) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有的话，先删除， 在添加
	// 没有的话， 啥都不干 报个错
	return
}

type Params struct {
	Open   *SSconfig `json:"open"`
	Close  *SSconfig `json:"close"`
	Remove *SSconfig `json:"remove"`
	Limit  *SSconfig `json:"limit"`
	Ping   int64     `json:"ping"`
}

func (m *Manager) Loop() {
	conn := m.conn
	clients := m.clients
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for range ticker.C {
			for clientAddr, _ := range clients {
				type Transfer struct {
					transfer map[int][]int
				}
				tt := make(map[int][]uint64)
				ut := make(map[int][]uint64)
				for port, p := range m.proxyTable {
					t := p.GetTraffic()
					tt[port] = []uint64{t.tcpup, t.tcpdown}
					ut[port] = []uint64{t.udpup, t.udpdown}
				}
				tcptransfer, _ := json.Marshal(struct {
					Transfer map[int][]uint64 `json:"tcptransfer"`
				}{Transfer: tt})
				_, _ = conn.WriteToUDP(tcptransfer, clientAddr)
				udptransfer, _ := json.Marshal(struct {
					Transfer map[int][]uint64 `json:"udptransfer"`
				}{Transfer: ut})
				_, _ = conn.WriteToUDP(udptransfer, clientAddr)
			}
		}
	}()
	for m.running {
		var params Params
		data := make([]byte, 300)
		_, remote, err := conn.ReadFromUDP(data)
		clients[remote] = true
		if err != nil {
			log.Printf("Failed to read UDP manage msg, error: %s", err.Error())
			continue
		}
		if err := json.Unmarshal(bytes.Trim(data, "\x00\r\n "), &params); err != nil {
			log.Printf("Failed to Unmarshal json, error: %s", err.Error())
		}
		var res []byte
		if params.Open != nil {
			if err := m.Add(*params.Open); err == nil {
				log.Printf("open SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.Open.Uid, params.Open.Sid, params.Open.ServerPort)
				ret := struct {
					Open int `json:"open"`
				}{Open: params.Open.ServerPort}
				res, _ = json.Marshal(ret)
			} else {
				log.Printf("open SS proxy error:[%s]", err.Error())
				ret := struct {
					Open int `json:"opened"`
				}{Open: params.Open.ServerPort}
				res, _ = json.Marshal(ret)
			}
		} else if params.Close != nil {
			if err := m.Remove(*params.Close); err == nil {
				log.Printf("close SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.Close.Uid, params.Close.Sid, params.Close.ServerPort)
				ret := struct {
					Close int `json:"close"`
				}{Close: params.Close.ServerPort}
				res, _ = json.Marshal(ret)
			} else {
				log.Printf("close SS proxy error:[%s]", err.Error())
				ret := struct {
					Close int `json:"closed"`
				}{Close: params.Close.ServerPort}
				res, _ = json.Marshal(ret)
			}
		} else if params.Remove != nil {
			if err := m.Remove(*params.Remove); err == nil {
				log.Printf("remove SS proxy success")
				ret := struct {
					Remove int `json:"remove"`
				}{Remove: params.Remove.ServerPort}
				res, _ = json.Marshal(ret)
			} else {
				log.Printf("remove SS proxy error:[%s]", err.Error())
				ret := struct {
					Remove int `json:"removed"`
				}{Remove: params.Remove.ServerPort}
				res, _ = json.Marshal(ret)
			}
		} else if params.Limit != nil {
			if err := m.Limit(*params.Limit); err == nil {
				log.Printf("Limit SS proxy success")
				ret := struct {
					Limit int `json:"limit"`
				}{Limit: params.Limit.ServerPort}
				res, _ = json.Marshal(ret)
			} else {
				log.Printf("Limit SS proxy error:[%s]", err.Error())
				ret := struct {
					Limit int `json:"limited"`
				}{Limit: params.Limit.ServerPort}
				res, _ = json.Marshal(ret)
			}
		} else if params.Ping > 0 {
			ret := struct {
				Pong int64 `json:"pong"`
			}{Pong: params.Ping}
			res, _ = json.Marshal(ret)
		} else {
			res = []byte("error")
		}
		if len(res) == 0 {
			continue
		}
		_, err = conn.WriteToUDP(res, remote)
		if err != nil {
			log.Printf("Failed to write UDP manage msg, error:[%s]", err.Error())
			continue
		}
	}
}

func ManagerDaemon(m *Manager) {
	m.Start()
}
func (m *Manager) Stop() {
	m.running = false
	m.conn.Close()
}
func (m *Manager) Start() {
	m.running = true
	m.Loop()
}
