package udpposter

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
	pb "proxy-golang/proto"
	"strconv"
)

const _port = 6666

var conn *net.UDPConn

func init() {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(_port))
	if err != nil {
	}
	conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Can't dial: ", err)
	}
}
func Post(record pb.Record) (err error) {
	data, _ := proto.Marshal(&record)
	_, err = conn.Write(data)
	return
}
func PostDict(item map[string]interface{}) (err error) {
	return Post(pb.ConvertMaptoRecordByReflect(item))
}
