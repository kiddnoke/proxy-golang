package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"strconv"
)

const delay_udp_port = 7666

func delayer() {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(delay_udp_port))
	if err != nil {
		log.Println(err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println(errors.WithMessage(err, "已经有其他实例开启udp:7666的监听了"))
		return
	}
	go func() {
		for {
			data := make([]byte, 4096)
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("failed to read UDP msg because of ", err.Error())
				return
			}
			conn.WriteToUDP(data[0:n], remoteAddr)
		}
	}()
}
