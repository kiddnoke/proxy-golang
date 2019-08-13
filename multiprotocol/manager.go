package multiprotocol

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"proxy-golang/ss"
	"proxy-golang/util"
	"sync"
	"time"

	"github.com/kiddnoke/eventemitter"

	"proxy-golang/softether"
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
	for _, value := range args[:len(args)] {
		keystr += fmt.Sprintf("%v-", value)
	}
	keystr += fmt.Sprintf("\b")
	return keystr
}

func genPassword(n int) string {
	return util.RandStringBytesMaskImprSrcSB(n)
}

type Manager struct {
	manager
	proxyTable sync.Map
	*eventemitter.EventEmitter
}

func (m *Manager) Add(proxy *Config) (err error) {
	var key string
	var relay Relayer
	key = generatorKey(proxy.Sid)
	if _, found := m.proxyTable.Load(key); found {
		err = errors.Errorf("key[%s] config.Sid[%d],config.Uid[%d]", key, proxy.Sid, proxy.Uid)
		log.Printf("%+v", err)
		return
	} else {

		if proxy.Password == "" {
			proxy.Password = genPassword(5)
		}

		if proxy.Protocol == "open" {
			proxy.ServerPort = softether.OpenVpnServicePort
			relay, err = NewOpenVpn(proxy)
		} else {
			proxy.Protocol = "ss"
			if proxy.Method == "" {
				proxy.Method = ss.GenCipherMethod(2)
			}
			relay, err = NewSS(proxy)
		}
		if err != nil {
			return err
		}
		relay.Start()
		time.AfterFunc(time.Minute/2, func() {
			c := relay.GetConfig()
			tu, td, uu, ud := relay.GetTraffic()
			if tu+td+uu+ud == 0 {
				<-m.Emit("fast_release", c.Uid, c.Sid, c.ServerPort, c.AppId, c.Protocol)
			}
		})
		m.proxyTable.Store(key, relay)
		return nil
	}
}

func (m *Manager) Delete(config Config) error {
	var key string
	key = generatorKey(config.Sid)
	if p, found := m.proxyTable.Load(key); found {
		p.(Relayer).Close()
		m.proxyTable.Delete(key)
	} else {
		err := errors.Wrapf(KeyNotExist, "key[%s],config.Sid[%d],config.Uid[%d]", key, config.Sid, config.Uid)
		return err
	}
	return nil
}

func (m *Manager) Update(config Config) error {
	var key string
	key = generatorKey(config.Sid)
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
		err := errors.Errorf("key[%s],config.Sid[%d],config.Uid[%d]", key, config.Sid, config.Uid)
		log.Printf("%+v", err)
		return err
	}
	return nil
}

func (m *Manager) Get(config Config) (Re Relayer, err error) {
	var key string
	key = generatorKey(config.Sid)
	if p, found := m.proxyTable.Load(key); found {
		return p.(Relayer), nil
	} else {
		err = errors.Wrapf(KeyNotExist, "key[%s],config.Sid[%d],config.Uid[%d]", key, config.Sid, config.Uid)
		return nil, err
	}
}

func (m *Manager) CheckLoop() {
	// 10 second timer
	util.Interval(time.Second*30, func(when time.Time) {
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(Relayer)
			c := p.GetConfig()
			if p.IsTimeout() {
				<-m.Emit("timeout", c.Uid, c.Sid, c.ServerPort, c.AppId, c.Protocol)
			}
			if p.IsExpire() {
				<-m.Emit("expire", c.Uid, c.Sid, c.ServerPort, c.AppId, c.Protocol)
			}
			if p.IsNotify() {
				<-m.Emit("balance", c.Uid, c.Sid, c.ServerPort, c.AppId, c.Protocol)
			}
			if limit, flag := p.IsStairCase(); flag == true {
				<-m.Emit("overflow", c.Uid, c.Sid, c.ServerPort, c.AppId, limit, c.Protocol)
			}
			return true
		})
	})
	// 1 min timer
	util.Interval(time.Minute, func(when time.Time) {
		<-m.Emit("health")
		var transferLists []interface{}
		m.proxyTable.Range(func(key, proxy interface{}) bool {
			p := proxy.(Relayer)
			c := p.GetConfig()
			if p.GetLastTimeStamp().Add(time.Minute * 5).Before(time.Now()) {
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
			avgrate, maxrate := p.GetRate()
			item["maxrate"] = []float64{avgrate, maxrate}
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
