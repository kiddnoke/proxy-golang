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
func postRecordProtoBuf(record pb.Record) (err error) {
	data, _ := proto.Marshal(&record)
	_, err = conn.Write(data)
	return
}
func PostDict(item map[string]interface{}) (err error) {
	return postRecordProtoBuf(pb.ConvertMapToRecordByReflect(item))
}
func PostParams(app_id, user_id, sn_id int64,
	device_id, app_version, os, user_type, carrier_operator, network_type string,
	ip, websit string, time_stamp, rate, connect_time, traffic int64, serverip string, state string, chargeType string) (err error) {
	return postRecordProtoBuf(pb.Record{
		AppId:           app_id,
		NetworkType:     network_type,
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
		Code:            state,
		ServerIp:        serverip,
		ServerType:      chargeType,
	})
}
func PostMaxRate(app_id, user_id, sn_id int64,
	device_id, app_version, os, user_type, carrier_operator, network_type string,
	rate, connect_time, traffic int64, serverip string, state string, chargeType string) error {

	return postRecordProtoBuf(pb.Record{
		AppId:           app_id,
		NetworkType:     network_type,
		UserId:          user_id,
		SnId:            sn_id,
		DeviceId:        device_id,
		AppVersion:      app_version,
		Os:              os,
		UserType:        user_type,
		CarrierOperator: carrier_operator,
		Ip:              "",
		Website:         "maxrate",
		Time:            0,
		Rate:            rate,
		ConnectTime:     connect_time,
		Traffic:         traffic,
		Code:            state,
		ServerIp:        serverip,
		ServerType:      chargeType,
	})
}
