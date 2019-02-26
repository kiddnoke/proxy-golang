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
	limitArray := []int64{20, 40, 60}
	flowArray := []int64{10240, 20480, 30720}
	var totalflow int64
	totalflow = 0
	if limit, err := SearchLimit(0, limitArray, flowArray, totalflow); limit != 0 || err != nil {
		t.FailNow()
	}
	totalflow = 10240
	if limit, err := SearchLimit(30, limitArray, flowArray, totalflow); limit != 20 || err != nil {
		t.FailNow()
	}
	totalflow = 10240 + 1
	if limit, err := SearchLimit(10, limitArray, flowArray, totalflow); limit != 10 || err != nil {
		t.FailNow()
	}
	totalflow = 20480
	if limit, err := SearchLimit(50, limitArray, flowArray, totalflow); limit != 40 || err != nil {
		t.FailNow()
	}
	totalflow = 20480 + 1
	if limit, err := SearchLimit(30, limitArray, flowArray, totalflow); limit != 30 || err != nil {
		t.FailNow()
	}
	totalflow = 30720
	if limit, err := SearchLimit(100, limitArray, flowArray, totalflow); limit != 60 || err != nil {
		t.FailNow()
	}
	totalflow = 30720 + 1
	if limit, err := SearchLimit(0, limitArray, flowArray, totalflow); limit != 60 || err != nil {
		t.FailNow()
	}

	limitArray2 := []int64{2, 3}
	flowArray2 := []int64{1}
	var totalflow2 int64
	totalflow2 = 3 * 1024 * 1024
	if _, err := SearchLimit(0, limitArray2, flowArray2, totalflow2); err == nil {
		t.FailNow()
	}

	limitArray3 := []int64{1}
	flowArray3 := []int64{}
	var totalflow3 int64
	totalflow3 = 3 * 1024 * 1024
	if _, err := SearchLimit(0, limitArray3, flowArray3, totalflow3); err == nil {
		t.FailNow()
	}

	limitArray4 := []int64{}
	flowArray4 := []int64{1}
	var totalflow4 int64
	totalflow3 = 3 * 1024 * 1024
	if _, err := SearchLimit(0, limitArray4, flowArray4, totalflow4); err == nil {
		t.FailNow()
	}
}

func TestSearchLimit2(t *testing.T) {
	limitArray := []int64{0}
	flowArray := []int64{0}
	if limit, err := SearchLimit(0, limitArray, flowArray, 1); limit != 0 || err != nil {
		t.FailNow()
	}
}

func TestSearchLimit3(t *testing.T) {
	limitArray := []int64{10, 50, 100}
	flowArray := []int64{10240, 20480, 30720}
	if limit, err := SearchLimit(100, limitArray, flowArray, 583257); limit != 100 || err != nil {
		t.FailNow()
	}
}
