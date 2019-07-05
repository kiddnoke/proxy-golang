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
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			conn.WriteToUDP([]byte("1"), remoteAddr)
			wg.Add(-1)
		}
	}()
	r1 := pb.Record{1, 1, "1", "1", "1", "1", 1, "www.baidu.com", 1, 1, 1, "zhongguoyidong", "vip", struct{}{}, nil, 1}
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
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			conn.WriteToUDP([]byte("1"), remoteAddr)
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
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			r := pb.Record{}
			if err := proto.Unmarshal(data[:n], &r); err != nil {
				t.FailNow()
			}
			conn.WriteToUDP([]byte("1"), remoteAddr)
			time.Sleep(time.Millisecond * 200)
			wg.Add(-1)
		}
	}()
	time_stamp := time.Now().UnixNano() / 1000000
	err = PostParams(1, 1, "1", "v1.1", "ios", "free", "zhongguoyidong", "192.168.1.1", "www.baiud.com", time_stamp, 1212, 1212, 1212)
	if err == nil {
		wg.Add(1)
	}
	wg.Wait()
}
