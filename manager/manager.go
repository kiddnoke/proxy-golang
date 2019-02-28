package manager

import (
	"strconv"
	"sync"
	"time"

	"github.com/CHH/eventemitter"
)

type manager interface {
	Add(proxy Proxy) error
	Delete(key Proxy) error
	Update(key Proxy) error
	Get(key Proxy) (proxy *Proxy, err error)
}
type Manager struct {
	manager
	proxyTable map[string]*Proxy
	sync.Mutex
	*eventemitter.EventEmitter
}

func New() (m *Manager) {
	return &Manager{proxyTable: make(map[string]*Proxy), EventEmitter: eventemitter.New()}
}
func (m *Manager) Add(proxy *Proxy) (err error) {
	m.Lock()
	defer m.Unlock()
	var key string
	proxy.ServerPort = GetFreePort(BeginPort, EndPort)
	key = strconv.FormatInt(int64(proxy.Uid), 10)
	key += strconv.FormatInt(int64(proxy.Sid), 10)
	key += strconv.FormatInt(int64(proxy.ServerPort), 10)
	if _, found := m.proxyTable[key]; found {
		return KeyExist
	} else {
		if err = proxy.Init(); err != nil {
			return err
		}
		m.proxyTable[key] = proxy
		proxy.Start()
		return err
	}
}
func (m *Manager) Delete(keys Proxy) error {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if _, found := m.proxyTable[key]; found {
		p := m.proxyTable[key]
		p.Close()
		delete(m.proxyTable, key)
		ClearPort(keys.ServerPort)
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Update(keys Proxy) error {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if p, found := m.proxyTable[key]; found {
		if keys.CurrLimitDown != 0 {
			p.CurrLimitDown = keys.CurrLimitDown
		}
		if keys.Timeout != 0 {
			p.Timeout = keys.Timeout
		}
		if keys.Expire != 0 {
			p.Expire = keys.Expire
		}
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Size() (size int) {
	m.Lock()
	defer m.Unlock()
	return len(m.proxyTable)
}
func (m *Manager) Health() (h int) {
	m.Lock()
	defer m.Unlock()
	h = 0
	for _, p := range m.proxyTable {
		if b := p.Burst(); b == 0 {
			h += 1024 * 1024 * 5
		} else {
			h += b
		}
	}
	return h
}
func (m *Manager) Get(keys Proxy) (proxy *Proxy, err error) {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if p, found := m.proxyTable[key]; found {
		return p, nil
	} else {
		return nil, KeyNotExist
	}
}
func (m *Manager) CheckLoop() {
	// 10 second timer
	setInterval(time.Second*10, func(when time.Time) {
		for _, p := range m.proxyTable {
			if p.IsTimeout() {
				<-m.Emit("timeout", p.Uid, p.Sid, p.ServerPort)
			}
			if p.IsExpire() {
				<-m.Emit("expire", p.Uid, p.Sid, p.ServerPort)
			}
			if p.IsNotify() {
				<-m.Emit("balance", p.Uid, p.Sid, p.ServerPort)
			}
			if limit, flag := p.IsStairCase(); flag == true {
				<-m.Emit("overflow", p.Uid, p.Sid, p.ServerPort, limit)
			}
		}
	})
	// 1 min timer
	setInterval(time.Minute, func(when time.Time) {
		//m.Lock()
		//defer m.Unlock()
		<-m.Emit("health", m.Health())
		var transferLists []interface{}
		for _, p := range m.proxyTable {
			if p.GetLastTimeStamp().Add(time.Minute * 2).Before(time.Now().UTC()) {
				continue
			}
			tu, td, uu, ud := p.GetTraffic()
			item := make(map[string]interface{})
			item["sid"] = p.Sid
			item["transfer"] = []int64{tu, td, uu, ud}
			transferLists = append(transferLists, item)
		}
		if len(transferLists) > 0 {
			<-m.Emit("transfer", transferLists)
		}
	})
}
