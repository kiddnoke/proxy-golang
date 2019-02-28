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
	proxyTable sync.Map
	*eventemitter.EventEmitter
}

func New() (m *Manager) {
	return &Manager{proxyTable: sync.Map{}, EventEmitter: eventemitter.New()}
}
func (m *Manager) Add(proxy *Proxy) (err error) {
	var key string
	proxy.ServerPort = GetFreePort(BeginPort, EndPort)
	key = strconv.FormatInt(int64(proxy.Uid), 10)
	key += strconv.FormatInt(int64(proxy.Sid), 10)
	key += strconv.FormatInt(int64(proxy.ServerPort), 10)
	if _, found := m.proxyTable.Load(key); found {
		return KeyExist
	} else {
		if err = proxy.Init(); err != nil {
			return err
		}
		m.proxyTable.Store(key, proxy)
		proxy.Start()
		return err
	}
}
func (m *Manager) Delete(keys Proxy) error {
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if p, found := m.proxyTable.Load(key); found {
		p.(*Proxy).Close()
		m.proxyTable.Delete(key)
		ClearPort(keys.ServerPort)
	} else {
		return KeyNotExist
	}

	return nil
}
func (m *Manager) Update(keys Proxy) error {
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if p, found := m.proxyTable.Load(key); found {
		if keys.CurrLimitDown != 0 {
			p.(*Proxy).CurrLimitDown = keys.CurrLimitDown
		}
		if keys.Timeout != 0 {
			p.(*Proxy).Timeout = keys.Timeout
		}
		if keys.Expire != 0 {
			p.(*Proxy).Expire = keys.Expire
		}
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Size() (size int) {
	length := 0
	m.proxyTable.Range(func(key, p interface{}) bool {
		length++
		return true
	})
	return length
}
func (m *Manager) Health() (h int) {
	health := 0
	m.proxyTable.Range(func(key, p interface{}) bool {
		if b := p.(*Proxy).Burst(); b == 0 {
			health += 1024 * 1024 * 5
		} else {
			health += b
		}
		return true
	})
	return health
}
func (m *Manager) Get(keys Proxy) (proxy *Proxy, err error) {
	var key string
	key = strconv.FormatInt(int64(keys.Uid), 10)
	key += strconv.FormatInt(int64(keys.Sid), 10)
	key += strconv.FormatInt(int64(keys.ServerPort), 10)
	if p, found := m.proxyTable.Load(key); found {
		return p.(*Proxy), nil
	} else {
		return nil, KeyNotExist
	}
}
func (m *Manager) CheckLoop() {
	// 10 second timer
	setInterval(time.Second*10, func(when time.Time) {
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(*Proxy)
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
			return true
		})
	})
	// 1 min timer
	setInterval(time.Minute, func(when time.Time) {
		//m.Lock()
		//defer m.Unlock()
		<-m.Emit("health", m.Health())
		var transferLists []interface{}
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(*Proxy)
			if p.GetLastTimeStamp().Add(time.Minute * 2).Before(time.Now().UTC()) {
				return true
			}
			tu, td, uu, ud := p.GetTraffic()
			item := make(map[string]interface{})
			item["sid"] = p.Sid
			item["transfer"] = []int64{tu, td, uu, ud}
			transferLists = append(transferLists, item)
			return true
		})
		if len(transferLists) > 0 {
			<-m.Emit("transfer", transferLists)
		}
	})
	// Benchmark Case
	setInterval(time.Second*30, func(when time.Time) {
		for _, p := range m.proxyTable {
			<-m.Emit("benchmark", p.Uid, p.Sid)
		}
	})
}
