package multiprotocol

import (
	"fmt"
	"github.com/CHH/eventemitter"
	"sync"
	"time"
)

type manager interface {
	Add(Config *Config) error
	Delete(Config Config) error
	Update(Config Config) error
	Get(Config Config) (Re Relayer, err error)
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

type Manager struct {
	manager
	proxyTable sync.Map
	*eventemitter.EventEmitter
}

func (m *Manager) Add(proxy *Config) (err error) {
	var key string
	if proxy.Protocol == "open" {
		return nil
	} else {
		proxy.ServerPort = getFreePort(BeginPort, EndPort)
		key = generatorKey(proxy.Uid, proxy.Sid, proxy.ServerPort, proxy.AppId)
		if _, found := m.proxyTable.Load(key); found {
			return KeyExist
		} else {
			proxy, err := NewSS(*proxy)
			if err != nil {
				return err
			}
			m.proxyTable.Store(key, proxy)
			proxy.Start()
			return nil
		}
	}
}

func (m *Manager) Delete(config Config) error {
	var key string
	key = generatorKey(config.Uid, config.Sid, config.ServerPort, config.AppId)
	if p, found := m.proxyTable.Load(key); found {
		p.(Relayer).Close()
		m.proxyTable.Delete(key)
		clearPort(config.ServerPort)
	} else {
		return KeyNotExist
	}
	return nil
}

func (m *Manager) Update(config Config) error {
	var key string
	key = generatorKey(config.Uid, config.Sid, config.ServerPort, config.AppId)
	if p, found := m.proxyTable.Load(key); found {
		cp := p.(Config)
		if config.CurrLimitDown != 0 {
			cp.CurrLimitDown = config.CurrLimitDown
		}
		if config.Timeout != 0 {
			cp.Timeout = config.Timeout
		}
		if config.Expire != 0 {
			cp.Expire = config.Expire
		}
	} else {
		return KeyNotExist
	}
	return nil
}

func (m *Manager) Get(keys Config) (Re Relayer, err error) {
	var key string
	key = generatorKey(keys.Uid, keys.Sid, keys.ServerPort, keys.AppId)
	if p, found := m.proxyTable.Load(key); found {
		return p.(Relayer), nil
	} else {
		return nil, KeyNotExist
	}
}

func (m *Manager) CheckLoop() {
	// 10 second timer
	setInterval(time.Second*30, func(when time.Time) {
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(Relayer)
			c := proxy.(Config)
			if p.IsTimeout() {
				<-m.Emit("timeout", c.Uid, c.Sid, c.ServerPort, c.AppId)
			}
			if p.IsExpire() {
				<-m.Emit("expire", c.Uid, c.Sid, c.ServerPort, c.AppId)
			}
			if p.IsNotify() {
				<-m.Emit("balance", c.Uid, c.Sid, c.ServerPort, c.AppId)
			}
			if limit, flag := p.IsStairCase(); flag == true {
				<-m.Emit("overflow", c.Uid, c.Sid, c.ServerPort, c.AppId, limit)
			}
			return true
		})
	})
	// 1 min timer
	setInterval(time.Minute, func(when time.Time) {
		<-m.Emit("health")
		var transferLists []interface{}
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(Relayer)
			c := proxy.(Config)
			if p.GetLastTimeStamp().Add(time.Minute * 5).Before(time.Now().UTC()) {
				return true
			}
			tu, td, uu, ud := p.GetTraffic()
			if tu+td+uu+ud == 0 {
				return true
			}
			item := make(map[string]interface{})
			item["app_id"] = c.AppId
			item["sid"] = c.Sid
			item["transfer"] = []int64{tu, td, uu, ud}
			transferLists = append(transferLists, item)
			return true
		})
		if len(transferLists) > 0 {
			<-m.Emit("transferlist", transferLists)
		}
	})
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
		if b := p.(Relayer).Burst(); b == 0 {
			health += 1024 * 5
		} else {
			health += b / 1024
		}
		return true
	})
	return health
}
