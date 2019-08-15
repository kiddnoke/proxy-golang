package udpposter

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	pb "proxy-golang/proto"

	"github.com/golang/protobuf/proto"
)

func TestPost(t *testing.T) {
	var wg sync.WaitGroup
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(_port))
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
	}
	defer conn.Close()
	go func() {
		for {
			data := make([]byte, 4096)
			n, _, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			wg.Add(-1)
		}
	}()
	r1 := pb.Record{}
	if err := postRecordProtoBuf(r1); err == nil {
		wg.Add(1)
	}
	wg.Wait()

}

func TestPostDict(t *testing.T) {
	var wg sync.WaitGroup
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(_port))
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
	}
	defer conn.Close()
	go func() {
		for {
			data := make([]byte, 4096)
			n, _, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			wg.Add(-1)
		}
	}()
	r := make(map[string]interface{})
	r["sn_id"] = int64(1)
	r["user_id"] = int64(1)
	r["website"] = "www.baidu.com"
	if err := PostDict(r); err == nil {
		wg.Add(1)
	}
	wg.Wait()
}

func TestPostParams(t *testing.T) {
	var wg sync.WaitGroup
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(_port))
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
	}
	defer conn.Close()
	go func() {
		for {
			data := make([]byte, 4096)
			n, _, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			time.Sleep(time.Millisecond * 200)
			wg.Done()
		}
	}()
	err = PostParams(1, 1, 1, "1", "1", "1", "1", "1", "1", "1", "1", 111, 111, 111, 1, "1", "1", "1", 1)
	if err == nil {
		wg.Add(1)
	}
	wg.Wait()
}
