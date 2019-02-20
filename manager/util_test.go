package manager

import (
	"net"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port1 := GetFreePort(20000, 30000)
	if port1 != 20000 {
		t.FailNow()
	} else {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
	}

	port2 := GetFreePort(20000, 30000)
	if port2 != 20001 {
		t.FailNow()
	} else {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port2})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port2})
	}
}
func TestIsFreePort(t *testing.T) {
	port1 := GetFreePort(20000, 30000)
	if err := IsFreePort(port1); err == nil {
		net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
		net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port1})
	} else {
		t.FailNow()
	}
	if err := IsFreePort(port1); err == nil {
		t.FailNow()
	}
}
func TestSearchLimit(t *testing.T) {
	limitArray := []int64{0, 1000, 500, 300, 200}
	flowArray := []int64{0, 5 * 1024 * 1024, 10 * 1024 * 1024, 15 * 1024 * 1024, 20 * 1024 * 1024}
	var totalflow int64
	totalflow = 3 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 0 || err != nil {
		t.FailNow()
	}
	totalflow = 6 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 1000 || err != nil {
		t.FailNow()
	}
	totalflow = 7 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 1000 || err != nil {
		t.FailNow()
	}
	totalflow = 11 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 500 || err != nil {
		t.FailNow()
	}
	totalflow = 11*1024*1024 + 20000
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 500 || err != nil {
		t.FailNow()
	}
	totalflow = 16 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 300 || err != nil {
		t.FailNow()
	}
	totalflow = 21 * 1024 * 1024
	if limit, err := SearchLimit(limitArray, flowArray, totalflow); limit != 200 || err != nil {
		t.FailNow()
	}

	limitArray2 := []int64{2, 3}
	flowArray2 := []int64{1}
	var totalflow2 int64
	totalflow2 = 3 * 1024 * 1024
	if _, err := SearchLimit(limitArray2, flowArray2, totalflow2); err == nil {
		t.FailNow()
	}

	limitArray3 := []int64{1}
	flowArray3 := []int64{}
	var totalflow3 int64
	totalflow3 = 3 * 1024 * 1024
	if _, err := SearchLimit(limitArray3, flowArray3, totalflow3); err == nil {
		t.FailNow()
	}

	limitArray4 := []int64{}
	flowArray4 := []int64{1}
	var totalflow4 int64
	totalflow3 = 3 * 1024 * 1024
	if _, err := SearchLimit(limitArray4, flowArray4, totalflow4); err == nil {
		t.FailNow()
	}
}

func TestSearchLimit2(t *testing.T) {
	limitArray := []int64{0}
	flowArray := []int64{0}
	if limit, err := SearchLimit(limitArray, flowArray, 1); limit != 0 || err != nil {
		t.FailNow()
	}
}
