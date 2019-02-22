package manager

import (
	"errors"
	"net"
	"sort"
	"sync"
)

var lock sync.Mutex
var lockTable sync.Map

func GetFreePort(start, end int) (freeport int) {
	lock.Lock()
	defer lock.Unlock()
	for freeport = start; freeport <= end; freeport++ {
		if _, ok := lockTable.Load(freeport); ok == true {
			continue
		}
		tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if t_err != nil {
			continue
		}
		ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if u_err != nil {
			continue
		}
		lockTable.Store(freeport, true)
		tl.Close()
		ul.Close()
		return
	}
	return freeport
}
func IsFreePort(port int) (err error) {
	tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})

	if t_err != nil || u_err != nil {
		return errors.New(t_err.Error() + u_err.Error())
	} else {
		tl.Close()
		ul.Close()
		return nil
	}
}
func SearchLimit(limitArray []int64, flowArray []int64, Total int64) (limit int64, err error) {
	limit = 0
	if len(limitArray) == 0 {
		err = errors.New("limitArray size is 0")
		return
	}
	if len(flowArray) == 0 {
		err = errors.New("flowArray size is 0")
		return
	}
	if len(limitArray) != len(flowArray) {
		err = errors.New("limitArray != flowArray")
		return
	}
	index := sort.Search(len(flowArray), func(i int) bool {
		return flowArray[i] > Total
	})
	if index == 0 {
		return
	}
	limit = limitArray[index-1]
	return limit, err
}
