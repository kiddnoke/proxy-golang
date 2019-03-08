package pushService

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func generatorKey(args ...interface{}) (keystr string) {
	keystr = ""
	for _, value := range args[:len(args)-1] {
		keystr += fmt.Sprintf("%v", value)
		keystr += "-"
	}
	keystr += fmt.Sprintf("%v", args[len(args)-1])
	return keystr
}

type Message struct {
	Id   int         `json:"id"`
	Body interface{} `json:"body"`
}

type Pusher interface {
	Push(key, method string, body interface{}) (err error)
}

type PushService struct {
	*gosocketio.Server
	UserSids sync.Map
}

var Id int

func NewPushService() (push *PushService, err error) {
	Id = 0
	p := &PushService{}
	p.Server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	err = p.Server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Printf("Connected Client:EventId[%s] Uid[%s] Port[%s]",
			c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		key := generatorKey(c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		sid := c.Id()
		if _, loaded := p.UserSids.LoadOrStore(key, sid); loaded {
			log.Printf("load")
		} else {
			log.Printf("stroed")
		}
	})
	if err != nil {
		goto End
	}
	err = p.Server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Printf("Disconnected Client:EventId[%s] Uid[%s] Port[%s]",
			c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		key := generatorKey(c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		p.UserSids.Delete(key)
	})
	if err != nil {
		goto End
	}
	err = p.Server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Printf("OnError Client:EventId[%s] Uid[%s] Port[%s]",
			c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		key := generatorKey(c.RequestHeader().Get("EventId"), c.RequestHeader().Get("Uid"), c.RequestHeader().Get("Port"))
		p.UserSids.Delete(key)
	})
	if err != nil {
		goto End
	}
	err = p.Server.On("delay", func(c *gosocketio.Channel, msg Message) (err error) {
		err = c.Emit("delay", msg)
		return
	})
End:
	return p, nil
}
func (self *PushService) Push(key, method string, body interface{}) (err error) {
	if sid, ok := self.UserSids.Load(key); ok {
		Id++
		if cn, err := self.GetChannel(sid.(string)); err == nil {
			_ = cn.Emit(method, Message{Id: Id, Body: body})
		}
	} else {
		err = errors.New("Key does not Existed")
	}
	return
}
