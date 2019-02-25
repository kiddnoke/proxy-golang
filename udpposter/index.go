package udpposter

import (
	"fmt"
	"net"
	"strconv"

	pb "proxy-golang/proto"

	"github.com/golang/protobuf/proto"
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
	return Post(pb.ConvertMapToRecordByReflect(item))
}
func PostParams(user_id, sn_id int64,
	device_id, app_version, os, user_type, carrier_operator string,
	ip, websit string, time_stamp, rate, connect_time, traffic int64) (err error) {
	return Post(pb.Record{
		UserId:          user_id,
		SnId:            sn_id,
		DeviceId:        device_id,
		AppVersion:      app_version,
		Os:              os,
		UserType:        user_type,
		CarrierOperator: carrier_operator,
		Ip:              ip,
		Website:         websit,
		Time:            time_stamp,
		Rate:            rate,
		ConnectTime:     connect_time,
		Traffic:         traffic,
	})
}