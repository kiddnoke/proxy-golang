package manager

import (
	"fmt"
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
func generatorKey(args ...interface{}) (keystr string) {
	keystr = ""
	for _, value := range args[:len(args)-1] {
		keystr += fmt.Sprintf("%v-", value)
	}
	keystr += fmt.Sprintf("%v", args[len(args)-1])
	return keystr
}
func (m *Manager) Add(proxy *Proxy) (err error) {
	var key string
	proxy.ServerPort = getFreePort(BeginPort, EndPort)
	key = generatorKey(proxy.Uid, proxy.Sid, proxy.ServerPort, proxy.AppId)
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
	key = generatorKey(keys.Uid, keys.Sid, keys.ServerPort, keys.AppId)
	if p, found := m.proxyTable.Load(key); found {
		p.(*Proxy).Close()
		m.proxyTable.Delete(key)
		clearPort(keys.ServerPort)
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Update(keys Proxy) error {
	var key string
	key = generatorKey(keys.Uid, keys.Sid, keys.ServerPort, keys.AppId)
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
			health += 1024 * 5
		} else {
			health += b / 1024
		}
		return true
	})
	return health
}
func (m *Manager) Get(keys Proxy) (proxy *Proxy, err error) {
	var key string
	key = generatorKey(keys.Uid, keys.Sid, keys.ServerPort, keys.AppId)
	if p, found := m.proxyTable.Load(key); found {
		return p.(*Proxy), nil
	} else {
		return nil, KeyNotExist
	}
}
func (m *Manager) CheckLoop() {
	// 10 second timer
	setInterval(time.Second*30, func(when time.Time) {
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
		<-m.Emit("health")
		var transferLists []interface{}
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(*Proxy)
			if p.GetLastTimeStamp().Add(time.Minute * 5).Before(time.Now().UTC()) {
				return true
			}
			tu, td, uu, ud := p.GetTraffic()
			if tu+td+uu+ud == 0 {
				return true
			}
			item := make(map[string]interface{})
			item["sid"] = p.Sid
			item["transfer"] = []int64{tu, td, uu, ud}
			transferLists = append(transferLists, item)
			return true
		})
		if len(transferLists) > 0 {
			<-m.Emit("transferlist", transferLists)
		}
	})
}
