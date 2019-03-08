package pushService

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

var PushServer *PushService

func init() {
	go func(port int) {
		PushServer, _ = NewPushService()
		router := http.NewServeMux()

		router.Handle("/socket.io/", PushServer)
		srv := &http.Server{
			Handler:      router,
			Addr:         "0.0.0.0:" + strconv.Itoa(port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())
	}(50000)
}
func TestDelay(t *testing.T) {
	tsp := transport.GetDefaultWebsocketTransport()
	tsp.RequestHeader = http.Header{}
	tsp.RequestHeader.Add("EventId", "1")
	tsp.RequestHeader.Add("Uid", "1")
	tsp.RequestHeader.Add("Port", "1")
	var wg sync.WaitGroup
	wg.Add(2)
	c, _ := gosocketio.Dial(
		gosocketio.GetUrl("localhost", 50000, false),
		tsp)

	err := c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
		_ = c.Emit("delay", Message{1, time.Now().Unix()})
		wg.Done()
	})
	if err != nil {
		t.FailNow()
	}
	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Println("OnDisconnection")

	})
	if err != nil {
		t.FailNow()
	}
	err = c.On("delay", func(h *gosocketio.Channel, msg Message) {
		if msg.Id != 1 {
			t.FailNow()
		}
		log.Printf("delay: %v", msg.Body)
		wg.Done()
	})
	if err != nil {
		t.FailNow()
	}

	wg.Wait()

	return
}
func TestPush(t *testing.T) {
	tsp := transport.GetDefaultWebsocketTransport()
	tsp.RequestHeader = http.Header{}
	tsp.RequestHeader.Add("EventId", "2")
	tsp.RequestHeader.Add("Uid", "2")
	tsp.RequestHeader.Add("Port", "2")
	var wg sync.WaitGroup
	wg.Add(3)
	CTestPush, err := gosocketio.Dial(
		gosocketio.GetUrl("localhost", 50000, false),
		tsp)

	err = CTestPush.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
		PushServer.Push("2-2-2", "test", 2)
		wg.Done()
	})
	if err != nil {
		t.FailNow()
	}
	err = CTestPush.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {

	})
	if err != nil {
		t.FailNow()
	}
	err = CTestPush.On("test", func(h *gosocketio.Channel, msg Message) {
		if msg.Body.(float64) == 2 {
			wg.Done()
			CTestPush.Close()
			time.AfterFunc(time.Second*5, func() {
				if _, ok := PushServer.UserSids.Load("2-2-2"); ok {
					t.FailNow()
				} else {
					wg.Done()
				}
			})
		} else {
			t.FailNow()
		}
	})
	wg.Wait()
	return
}
