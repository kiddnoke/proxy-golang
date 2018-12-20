package main

import (
	"../../comm/wswarpper"
	"log"
	"time"
)

type Config struct {
	BeginPort      int    `json:"beginport"`
	EndPort        int    `json:"endport"`
	ManagerPort    int    `json:"manager_port"`
	ControllerPort int    `json:"controller_port"`
	State          string `json:"state"`
	Area           int    `json:"area"`
}

func main() {
	client := wswarpper.New()
	client.Connect("127.0.0.1", 7001)
	client.Request("echo", "1212", func(msg string) {
		log.Println(msg)
	})
	for {
		time.Sleep(time.Second * 5)
	}
}
