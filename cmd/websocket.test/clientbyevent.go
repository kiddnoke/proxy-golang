package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"os"
)

func main() {
	//currtimestamp := strconv.FormatInt(time.Now().UTC().Unix()*1000,10)
	currtimestamp := "1545120729565"
	hasher := md5.New()
	hasher.Write([]byte(currtimestamp))
	hasher.Write([]byte("VpnMgrCore"))
	opts := &socketio_client.Options{
		Transport: "websocket",
		Query:     make(map[string]string),
	}
	opts.Query["keys"] = hex.EncodeToString(hasher.Sum(nil))
	opts.Query["timestamp"] = currtimestamp
	uri := "http://localhost:7001/socket.io/"

	client, err := socketio_client.NewClient(uri, opts)
	if err != nil {
		log.Printf("NewClient error:%v\n", err)
		return
	}

	client.On("error", func() {
		log.Printf("on error\n")
	})
	client.On("connection", func() {
		log.Printf("on connect\n")
	})
	client.On("message", func(msg string) {
		log.Printf("on message:%v\n", msg)
	})
	client.On("disconnection", func() {
		log.Printf("on disconnect\n")
	})

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		client.Emit("message", command)
		log.Printf("send message:%v\n", command)
	}
}