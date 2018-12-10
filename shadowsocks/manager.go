package shadowsocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

type Manager struct {
	sync.Mutex
	proxyTable map[int]Proxy
	conn       *net.UDPConn
	running    bool
}

func NewManager(port int) (m *Manager) {
	// 端口是否在用
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("localhost"), Port: port})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening:", err)
		os.Exit(1)
	}
	return &Manager{proxyTable: make(map[int]Proxy), conn: conn, running: false}
}
func (m *Manager) Add(config SSconfig) (e error) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 没有就加
	// 有就报错
	p, ok := m.proxyTable[config.ServerPort]
	if ok == true {
		e = errors.New(fmt.Sprintf("这个实例已经存在了 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Conf.Uid, config.Sid, p.Conf.Sid))
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
		e = errors.New(fmt.Sprintf("没有这个实例 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Conf.Uid, config.Sid, p.Conf.Sid))
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
	if ok == true && p.Conf.Uid == config.Uid && p.Conf.Sid == config.Sid {
		p.Stop()
		delete(m.proxyTable, config.ServerPort)
	} else {
		e = errors.New(fmt.Sprintf("没有这个实例 params.Uid[%d] Proxy.Uid[%d] params.Sid[%d] Proxy.Sid[%d] ", config.Uid, p.Conf.Uid, config.Sid, p.Conf.Sid))
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
	Cmd    string   `json:"cmd"`
	Config SSconfig `json:"config"`
}

func (m *Manager) Loop() {
	go func() {
		for {
			time.Sleep(time.Second * 5)
			log.Printf("gorontine Num[%d]", runtime.NumGoroutine())
		}
	}()
	var params Params
	//var config SSconfig
	conn := m.conn
	for m.running {
		data := make([]byte, 300)
		_, remote, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("Failed to read UDP manage msg, error: %s", err.Error())
			continue
		}
		if err := json.Unmarshal(bytes.Trim(data, "\x00\r\n "), &params); err != nil {
			log.Printf("Failed to Unmarshal json, error: %s", err.Error())
		}
		var res []byte
		switch params.Cmd {
		case "open":
			if err := m.Add(params.Config); err == nil {
				log.Printf("open SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.Config.Uid, params.Config.Sid, params.Config.ServerPort)
			} else {
				log.Printf("open SS proxy error:[%s]", err.Error())
			}
		case "close":
			if err := m.Remove(params.Config); err == nil {
				log.Printf("close SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.Config.Uid, params.Config.Sid, params.Config.ServerPort)
			} else {
				log.Printf("close SS proxy error:[%s]", err.Error())
			}
		case "remove":
			if err := m.Remove(params.Config); err == nil {
				log.Printf("remove SS proxy success")
			} else {
				log.Printf("remove SS proxy error:[%s]", err.Error())
			}
		case "update":

		case "query":
			if proxy, err := m.Get(params.Config); err == nil {
				log.Printf("query SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", proxy.Conf.Uid, proxy.Conf.Sid, proxy.Conf.ServerPort)
			} else {
				log.Printf("query SS proxy error:[%s]", err.Error())
			}
		default:
			res = []byte("error , command not found")
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
