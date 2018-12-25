package main

import (
	"../../comm/wswarpper"
	"log"
	"time"
)

func main() {
	client := wswarpper.New()

	client.Connect("127.0.0.1", 7001)
	client.OnConnect(func(c wswarpper.Channel) {
		log.Println(c)
	})
	client.Request("echo", "121212", func(ack string) {
		log.Println(ack)
	})
	for {
		time.Sleep(time.Second)
	}
}
