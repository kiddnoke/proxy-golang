package main

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func TestHttpSrv(t *testing.T) {
	go HttpSrv(10001)
	tsp := transport.GetDefaultWebsocketTransport()
	tsp.RequestHeader = http.Header{}
	tsp.RequestHeader.Add("EventId", "1")
	tsp.RequestHeader.Add("Uid", "1")

	c, err := gosocketio.Dial(
		gosocketio.GetUrl("localhost", 10001, false),
		tsp)

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
	})
	if err != nil {
		t.FailNow()
	}

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
	})
	if err != nil {
		t.FailNow()
	}
	time.Sleep(2 * time.Second)
	_ = c.Emit("delay", Message{1, time.Now().Unix() / 10000})
	_ = c.On("delay", func(h *gosocketio.Channel, msg Message) {
		if msg.Id != 1 {
			t.FailNow()
		}
		log.Printf("delay: %v", msg.Body)
	})
	time.Sleep(time.Second)
	return
}
