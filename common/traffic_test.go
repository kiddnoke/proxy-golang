package common

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestTraffic_AddTraffic(t *testing.T) {
	var wg sync.WaitGroup
	tra := NewTraffic()
	wg.Add(4)
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		t.Logf("task 1 finish")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		t.Logf("task 2 finish")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		t.Logf("task 3 finish")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		t.Logf("task 4 finish")
	}()
	wg.Wait()
	tu, td, _, _ := tra.GetTraffic()
	if tu == 40000 || td == 40000 {
		t.Skipf("AddTraffic Sucess")
	} else {
		t.FailNow()
	}
}

func TestTraffic_AddUsedDuration(t *testing.T) {
	var wg sync.WaitGroup
	tra := NewTraffic()
	wg.Add(4)
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		log.Println("task 1 finish")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		log.Println("task 2 finish")
	}()
	<-time.After(time.Second * 2)
	tra.OnceSampling()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		log.Println("task 3 finish")
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			tra.AddTraffic(1, 1, 0, 0)
		}
		wg.Done()
		log.Println("task 4 finish")
	}()
	wg.Wait()
	tra.OnceSampling()
	d := tra.GetUsedDuration()
	t.Skipf("tra is usedDuration [%v]", d)
}
