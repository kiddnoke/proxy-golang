package common

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestPipeTrafficSet_AddTraffic(t *testing.T) {
	var wg sync.WaitGroup
	p := NewPipTrafficSet()
	wg.Add(4)
	currtimestamp := time.Now()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 1 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 2 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 3 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 4 finish")
	}()
	wg.Wait()
	log.Println(p.GetTraffic("t2"))
	p.SamplingAndPrint(time.Since(currtimestamp))

	wg.Add(4)
	currtimestamp = time.Now()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t3", 1)
		}
		log.Println("task 1 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t3", 1)
		}
		log.Println("task 2 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t4", 1)
		}
		log.Println("task 3 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t4", 1)
		}
		log.Println("task 4 finish")
	}()
	wg.Wait()
	log.Println(p.GetTraffic("t3"))
	p.SamplingAndPrint(time.Since(currtimestamp))
}

func TestPipeTrafficSet_SamplingAndString(t *testing.T) {
	var wg sync.WaitGroup
	p := NewPipTrafficSet()
	wg.Add(4)
	currtimestamp := time.Now()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 1 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 2 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 3 finish")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			p.AddTraffic("t2", 1)
		}
		log.Println("task 4 finish")
	}()
	wg.Wait()
	log.Println(p.GetTraffic("t2"))
	log.Println(p.SamplingAndString(time.Since(currtimestamp)))
}
