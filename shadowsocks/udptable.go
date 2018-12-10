package shadowsocks

import (
	"log"
	"net"
	"sync"
	"time"
)

var (
	udpTimeout         = 30 * time.Second
	reqListRefreshTime = 1 * time.Minute
)

type natTable struct {
	sync.Mutex
	conns map[string]net.PacketConn
}

func newNatTable() *natTable {
	return &natTable{conns: map[string]net.PacketConn{}}
}

func (table *natTable) Delete(index string) net.PacketConn {
	table.Lock()
	defer table.Unlock()
	c, ok := table.conns[index]
	if ok {
		delete(table.conns, index)
		return c
	}
	return nil
}

func (table *natTable) Get(index string) (c net.PacketConn, ok bool, err error) {
	table.Lock()
	defer table.Unlock()
	c, ok = table.conns[index]
	if !ok {
		c, err = net.ListenPacket("udp", "")
		if err != nil {
			return nil, false, err
		}
		table.conns[index] = c
	}
	return
}

//noinspection GoRedundantParens
type requestHeaderList struct {
	sync.Mutex
	List    map[string]([]byte)
	running bool
}

//noinspection GoRedundantParens
func newReqList() *requestHeaderList {
	ret := &requestHeaderList{List: map[string]([]byte){}, running: true}
	var refreshHandler func()
	refreshHandler = func() {
		time.Sleep(reqListRefreshTime)
		ret.Refresh()
		if ret.running {
			log.Printf("refresh reqlist")
			time.AfterFunc(reqListRefreshTime, refreshHandler)
		}
	}
	time.AfterFunc(reqListRefreshTime, refreshHandler)
	return ret
}

func (r *requestHeaderList) Refresh() {
	r.Lock()
	defer r.Unlock()
	for k := range r.List {
		delete(r.List, k)
	}
}

func (r *requestHeaderList) Get(dstaddr string) (req []byte, ok bool) {
	r.Lock()
	defer r.Unlock()
	req, ok = r.List[dstaddr]
	return
}

func (r *requestHeaderList) Put(dstaddr string, req []byte) {
	r.Lock()
	defer r.Unlock()
	r.List[dstaddr] = req
	return
}
