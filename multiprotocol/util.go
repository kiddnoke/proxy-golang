package multiprotocol

import (
	"errors"
	"math"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

var opLock sync.Mutex
var lockPortTable sync.Map
var BeginPort int
var EndPort int

func getFreePort(start, end int) (freeport int) {
	opLock.Lock()
	defer opLock.Unlock()
	for freeport = start; freeport <= end; freeport++ {
		if _, ok := lockPortTable.Load(freeport); ok == true {
			continue
		}
		tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if t_err != nil {
			lockPortTable.Store(freeport, true)
			continue
		}
		ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: freeport})
		if u_err != nil {
			lockPortTable.Store(freeport, true)
			continue
		}
		tl.Close()
		ul.Close()
		lockPortTable.Store(freeport, true)
		return
	}
	return freeport
}
func clearPort(port int) (flag bool) {
	opLock.Lock()
	defer opLock.Unlock()
	if _, flag = lockPortTable.Load(port); flag == true {
		lockPortTable.Delete(port)
	}
	return
}
func IsFreePort(port int) (err error) {
	tl, t_err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})
	ul, u_err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port})

	if t_err != nil || u_err != nil {
		if t_err != nil && u_err == nil {
			err = t_err
		} else if u_err != nil && t_err == nil {
			err = u_err
		} else {
			err = errors.New(t_err.Error() + u_err.Error())
		}

	} else {
		tl.Close()
		ul.Close()
	}
	return
}
func searchLimit(CurrLimit int64, limitArray []int64, flowArray []int64, TotalFlow int64) (limit int64, err error) {
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
		return flowArray[i] > TotalFlow
	})
	if index != 0 {
		limit = limitArray[index-1]
	}
	//
	if limit > 0 && CurrLimit > 0 {
		return int64(math.Min(float64(limit), float64(CurrLimit))), nil
	} else if limit > 0 && CurrLimit == 0 {
		return limit, nil
	} else if limit == 0 && CurrLimit > 0 {
		return CurrLimit, nil
	} else if limit == 0 && CurrLimit == 0 {
		return 0, nil
	}
	return
}

const ManagerBeginPort = 10000

func InstanceIdGen(curr int) (next int) {
	curr = curr + ManagerBeginPort
	for {
		if err := IsFreePort(curr); err != nil {
			curr += 1
			continue
		}
		break
	}
	next = curr - ManagerBeginPort
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
