package util

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	tick := Interval(time.Second, func(when time.Time) {
		log.Println("ls")
	})
	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
	time.AfterFunc(time.Second*5, func() {
		tick.Stop()
		wg.Done()
	})
	wg.Wait()
}
