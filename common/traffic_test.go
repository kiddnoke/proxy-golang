package common

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestMakeTraffic(t *testing.T) {
	m := MakeTraffic()
	m.AddTraffic(2, 2, 2, 2)
	tu, td, uu, ud := m.GetTraffic()
	log.Println(tu, td, uu, ud)
}
func TestNewTraffic(t *testing.T) {
	m := NewTraffic()
	m.AddTraffic(1, 1, 1, 1)
	tu, td, uu, ud := m.GetTraffic()
	log.Println(tu, td, uu, ud)
}
func TestTraffic_Sampling(t *testing.T) {
	r := NewTraffic()
	r.Sampling()
	var sw sync.WaitGroup
	sw.Add(2)
	go func() {
		for i := 0; i < 1100; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	go func() {
		for i := 0; i < 100; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	sw.Wait()
	<-time.After(time.Second * 2)
	log.Println(r.GetTraffic())
	log.Println(r.GetMaxRate())
	sw.Add(2)
	go func() {
		for i := 0; i < 11000; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	go func() {
		for i := 0; i < 11000; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	<-time.After(time.Second * 2)
	log.Println(r.GetTraffic())
	log.Println(r.GetMaxRate())

	sw.Add(2)
	go func() {
		for i := 0; i < 110; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	go func() {
		for i := 0; i < 110; i++ {
			r.AddTraffic(i, i, i, i)
		}
		sw.Add(-1)
	}()
	<-time.After(time.Second * 2)
	log.Println(r.GetTraffic())
	log.Println(r.GetMaxRate())
}
