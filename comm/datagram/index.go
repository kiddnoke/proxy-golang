package datagram

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net"
	"proxy-golang/comm"
	"strconv"

	"github.com/kataras/go-events"
)

const VERSION = "v1.0.0"

type Params struct {
	Open   *interface{} `json:"open"`
	Close  *interface{} `json:"close"`
	Remove *interface{} `json:"remove"`
	Limit  *interface{} `json:"limit"`
	Ping   int64        `json:"ping"`
}
type datagramer interface {
}
type datagram struct {
	events.EventEmmiter
	conn    *net.UDPConn
	clients map[int]*net.UDPAddr
	running bool
	comm.Community
}

func DataGram(addr string) (*datagram, error) {
	if _, port, err := net.SplitHostPort(addr); err == nil {
		Port, _ := strconv.Atoi(port)
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("localhost"), Port: Port})
		return &datagram{clients: make(map[int]*net.UDPAddr), running: false, conn: conn, EventEmmiter: events.New()}, err
	} else {
		return nil, errors.New("fatal")
	}
}
func (d *datagram) loop() {
	conn := d.conn
	clients := d.clients
	for d.running {
		var params Params
		data := make([]byte, 300)
		_, remote, err := conn.ReadFromUDP(data)
		clients[remote.Port] = remote
		if err != nil {
			log.Printf("Failed to read UDP manage msg, error: %s", err.Error())
			continue
		}
		if err := json.Unmarshal(bytes.Trim(data, "\x00\r\n "), &params); err != nil {
			log.Printf("Failed to Unmarshal json, error: %s", err.Error())
		}
		var res []byte

		_, err = conn.WriteToUDP(res, remote)
		if err != nil {
			log.Printf("Failed to write UDP manage msg, error:[%s]", err.Error())
			continue
		}
	}
}
