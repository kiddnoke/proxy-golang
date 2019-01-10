package manager

import (
	"github.com/CHH/eventemitter"
	"strconv"
	"sync"
	"time"
)

type manager interface {
	Add(proxy interface{}) error
	Delete(key interface{}) error
	Update(key interface{}) error
	Get(key interface{}) (proxy *Proxy, err error)
}
type Manager struct {
	manager
	proxyTable map[string]*Proxy
	sync.Mutex
	checktimer *interval
	*eventemitter.EventEmitter
}

func New() (m *Manager) {
	return &Manager{proxyTable: make(map[string]*Proxy), EventEmitter: eventemitter.New()}
}
func (m *Manager) Add(proxy interface{}) (err error) {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(proxy.(Proxy).Uid), 10)
	key += strconv.FormatInt(int64(proxy.(Proxy).Sid), 10)
	key += strconv.FormatInt(int64(proxy.(Proxy).ServerPort), 10)
	if _, found := m.proxyTable[key]; found {
		return KeyExist
	} else {
		p := &Proxy{
			Uid:        proxy.(Proxy).Uid,
			Sid:        proxy.(Proxy).Sid,
			Timeout:    proxy.(Proxy).Timeout,
			Limit:      proxy.(Proxy).Limit,
			ServerPort: proxy.(Proxy).ServerPort,
			Method:     proxy.(Proxy).Method,
			Password:   proxy.(Proxy).Password,
		}
		err = p.Init()
		m.proxyTable[key] = p
		p.Start()
		return err
	}
}
func (m *Manager) Delete(keys interface{}) error {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(keys.(Proxy).Uid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).Sid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).ServerPort), 10)
	if _, found := m.proxyTable[key]; found {
		p := m.proxyTable[key]
		p.Close()
		delete(m.proxyTable, key)
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Update(keys interface{}) error {
	m.Lock()
	defer m.Unlock()
	var key string
	key = strconv.FormatInt(int64(keys.(Proxy).Uid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).Sid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).ServerPort), 10)
	if p, found := m.proxyTable[key]; found {
		if keys.(Proxy).Limit != 0 {
			p.Limit = keys.(Proxy).Limit
			p.SetLimit(p.Limit * 1024)
		}
		if keys.(Proxy).Timeout != 0 {
			p.Timeout = keys.(Proxy).Timeout
		}
		if keys.(Proxy).Remain != 0 {
			p.Remain = keys.(Proxy).Remain
		}
		if keys.(Proxy).Expire != 0 {
			p.Expire = keys.(Proxy).Expire
		}
	} else {
		return KeyNotExist
	}
	return nil
}
func (m *Manager) Get(keys interface{}) (proxy *Proxy, err error) {
	var key string
	key = strconv.FormatInt(int64(keys.(Proxy).Uid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).Sid), 10)
	key += strconv.FormatInt(int64(keys.(Proxy).ServerPort), 10)
	if p, found := m.proxyTable[key]; found {
		return p, nil
	} else {
		return nil, KeyNotExist
	}
}
func (m *Manager) CheckLoop() {
	m.checktimer = setInterval(time.Minute, func(when time.Time) {
		for _, p := range m.proxyTable {
			if p.IsTimeout() {
				m.Emit("timeout", p.Uid, p.Sid, p.ServerPort)
			}
			if p.IsExpire() {
				m.Emit("expire", p.Uid, p.Sid, p.ServerPort)
			}
			if p.IsOverflow() {
				m.Emit("overflow", p.Uid, p.Sid, p.ServerPort)
			}
			if p.IsNotify() {
				m.Emit("balance", p.Uid, p.Sid, p.ServerPort)
			}
		}
	})
}
