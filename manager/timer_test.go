package manager

import (
	"log"
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
	timer.Stop()
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
		timer2.Stop()
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
func TestSetIntervalRange(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)
	var lock sync.Mutex
	var flag int
	flag = 0
	log.Printf("%v", time.Now().Unix())
	timer, _ := setIntervalRange(10*time.Second, 2*time.Second, func(when time.Time) {
		lock.Lock()
		defer lock.Unlock()
		log.Printf("%v", when.Unix())
		flag++
		wg.Done()
	})
	wg.Wait()
	timer.Stop()
}
func TestSetIntervalRandom(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	var lock sync.Mutex
	var flag int
	flag = 0
	log.Printf("%v", time.Now().Unix())
	timer, _ := setIntervalRandom(10*time.Second, 2*time.Second, func(when time.Time) {
		lock.Lock()
		defer lock.Unlock()
		log.Printf("%v", when.Unix())
		flag++
		wg.Done()
	})
	wg.Wait()
	timer.Stop()
}
