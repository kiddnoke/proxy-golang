package manager

import (
	"sync"
	"testing"
	"time"
)

func TestSetInterval(t *testing.T) {
	var lock sync.Mutex
	var flag int
	flag = 0
	timer := setInterval(time.Second, func(when time.Time) {
		lock.Lock()
		defer lock.Unlock()
		flag++
	})
	time.Sleep(time.Millisecond * 50)
	if flag > 0 {
		t.FailNow()
	}
	time.Sleep(time.Second)
	if flag != 1 {
		t.FailNow()
	}
	clearInterval(timer)
	time.Sleep(time.Second)
	if flag > 1 {
		t.FailNow()
	}
}
func TestSetTimeout(t *testing.T) {
	var lock sync.Mutex
	var flag int
	flag = 0
	setTimeout(time.Second, func(when time.Time) {
		lock.Lock()
		defer lock.Unlock()
		flag++
	})
	time.Sleep(time.Millisecond * 50)
	if flag > 0 {
		t.FailNow()
	}
	time.Sleep(time.Second)
	if flag != 1 {
		t.FailNow()
	}
	timer2 := setTimeout(time.Millisecond*500, func(when time.Time) {
		lock.Lock()
		defer lock.Unlock()
		flag++
	})
	time.AfterFunc(time.Millisecond*400, func() {
		clearTimeout(timer2)
	})
	time.Sleep(time.Millisecond * 200)
	if flag > 1 {
		t.FailNow()
	}
	time.Sleep(time.Millisecond * 300)
	if flag > 1 {
		t.FailNow()
	}
}
