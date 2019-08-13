package common

import (
	"fmt"
	"log"
	"proxy-golang/common"
	"strconv"
	"sync"
	"testing"
)

func TestTraffic_AddTraffic(t *testing.T) {
	var wg sync.WaitGroup

	t1 := common.NewTraffic()
	wg.Add(4)

	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 1 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 2 finish`)

	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 3 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 4 finish`)
	}()
	wg.Wait()
	log.Println(t1.GetTraffic())
}

func TestTraffic_AddTrafficNew(t *testing.T) {
	var wg sync.WaitGroup

	t1 := NewTraffic()
	wg.Add(4)

	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 1 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 2 finish`)

	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 3 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 4 finish`)
	}()
	wg.Wait()
	log.Println(t1.GetTraffic())
	log.Println(t1.OnceSampling())

	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 1 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 2 finish`)

	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 3 finish`)
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 102400; i++ {
			t1.AddTraffic(0, 1, 0, 0)
		}
		log.Printf(`task 4 finish`)
	}()
	wg.Wait()
	log.Println(t1.GetTraffic())
	log.Println(t1.OnceSampling())
}

func TestTraffic_timeNowToUint64(t *testing.T) {
	var i int64
	timeNowToUint64(&i)
	log.Println(UInt64ToTime(&i))
}
func TestTraffic_UInt64ToTime(t *testing.T) {
	var i int64 = 1565245987729728700
	log.Println(UInt64ToTime(&i))
}
func TestFloat(t *testing.T) {
	s, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", 1.05), 64)
	log.Println(s)
}
