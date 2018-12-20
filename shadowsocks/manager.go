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
		log.Printf("Error listening: %s", err.Error())
		os.Exit(1)
	}
	return &Manager{proxyTable: make(map[int]Proxy), conn: conn, running: false}
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
	if ok == true && p.Config.Uid == config.Uid && p.Config.Sid == config.Sid {
		p.Stop()
		delete(m.proxyTable, config.ServerPort)
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
	open   interface{} `json:"open"`
	close  interface{} `json:"close"`
	remove interface{} `json:"remove"`
}

func (m *Manager) Loop() {
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
		if params.open != nil {
			if err := m.Add(params.open.(SSconfig)); err == nil {
				log.Printf("open SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.open.(SSconfig).Uid, params.open.(SSconfig).Sid, params.open.(SSconfig).ServerPort)
			} else {
				log.Printf("open SS proxy error:[%s]", err.Error())
			}
		} else if params.close != nil {
			if err := m.Remove(params.close.(SSconfig)); err == nil {
				log.Printf("close SS proxy success : proxy.Uid[%d] ,proxy.Sid[%d] ,proxy.ServerProt[%d]", params.close.(SSconfig).Uid, params.close.(SSconfig).Sid, params.close.(SSconfig).ServerPort)
			} else {
				log.Printf("close SS proxy error:[%s]", err.Error())
			}
		} else if params.remove != nil {
			if err := m.Remove(params.remove.(SSconfig)); err == nil {
				log.Printf("remove SS proxy success")
			} else {
				log.Printf("remove SS proxy error:[%s]", err.Error())
			}
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
