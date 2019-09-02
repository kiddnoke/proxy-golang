package softether

import (
	"errors"
	"fmt"
	"time"

	"github.com/kiddnoke/SoftetherGo"

	"proxy-golang/util"
)

type Pool struct {
	Conns map[string]softetherApi.API
}

func NewPool() *Pool {
	p := &Pool{Conns: make(map[string]softetherApi.API)}
	return p
}
func (p *Pool) NewConn(key string) (*softetherApi.API, error) {
	a := softetherApi.NewAPI(SoftHost, SoftPort, SoftPassword)
	if err := a.HandShake(); err != nil {
		selflogger.Error(err.Error())
		return nil, err
	}
	p.Conns[key] = *a
	return a, nil
}
func (p *Pool) GetConn(key string) (*softetherApi.API, error) {
	a, ok := p.Conns[key]
	if ok == false {
		return nil, errors.New(fmt.Sprintf("conn[%s] is not existed", key))
	}
	return &a, nil
}

var DefaultConnPool *Pool

func init() {
	DefaultConnPool = NewPool()
}
func PoolConnect() error {
	_, err := DefaultConnPool.NewConn("DEFAULT")
	if err != nil {
		panic(err)
	}
	return err
}
func PoolReConnect() error {
	return PoolConnect()
}
func PoolDiscConnect() error {
	c, _ := DefaultConnPool.GetConn("DEFAULT")
	c.Conn.Close()
	return nil
}
func PoolGetConn() (*softetherApi.API, error) {
	return DefaultConnPool.GetConn("DEFAULT")
}
func PoolHeartBeatLoop() {
	c, err := PoolGetConn()
	if err != nil {
		panic(err)
	}

	util.Interval(30*time.Second, func(when time.Time) {
		if _, err := c.Test(); err != nil {
			panic(err)
		}
	})

	return
}
