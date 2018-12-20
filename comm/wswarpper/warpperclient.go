package wswarpper

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

const (
	webSocketProtocol       = "ws://"
	webSocketSecureProtocol = "wss://"
	socketioUrl             = "/socket.io/?EIO=3&transport=websocket"
)

type Message struct {
	Id   int64       `json:"id"`
	Body interface{} `json:"body"`
}

func getUrlWithOpt(host string, port int, values url.Values, secure bool) (retUrl string) {
	var prefix string
	if secure {
		prefix = webSocketSecureProtocol
	} else {
		prefix = webSocketProtocol
	}
	retUrl = prefix + host + ":" + strconv.Itoa(port) + socketioUrl + "&" + values.Encode()
	return
}

type WarpperClient struct {
	client    *gosocketio.Client
	seqid     int64
	callbacks map[int64]interface{}
	keys      string
	timestamp string
}
type Channel *gosocketio.Channel

func Make() (client WarpperClient) {
	return *New()
}
func New() (client *WarpperClient) {
	currtimestamp := strconv.FormatInt(time.Now().UTC().Unix()*1000, 10)
	hasher := md5.New()
	hasher.Write([]byte(currtimestamp))
	hasher.Write([]byte("VpnMgrCore"))
	return &WarpperClient{
		seqid:     0,
		keys:      hex.EncodeToString(hasher.Sum(nil)),
		timestamp: currtimestamp,
		callbacks: make(map[int64]interface{})}
}
func (w *WarpperClient) Connect(host string, port int) (err error) {
	query := &url.Values{}
	query.Add("keys", w.keys)
	query.Add("timestamp", w.timestamp)
	w.client, err = gosocketio.Dial(
		getUrlWithOpt(host, port, *query, false),
		&transport.WebsocketTransport{
			PingInterval:   20 * time.Second,
			PingTimeout:    20 * time.Second,
			ReceiveTimeout: transport.WsDefaultReceiveTimeout,
			SendTimeout:    transport.WsDefaultSendTimeout,
			BufferSize:     transport.WsDefaultBufferSize,
		})
	return
}
func (w *WarpperClient) Request(router string, msg interface{}, callback interface{}) {
	w.seqid++
	Id := w.seqid
	message := Message{Id: Id, Body: msg}
	_ = w.client.Emit(router, message)
	w.callbacks[Id] = callback
	_ = w.client.On(router, func(channel Channel, Msg Message) {
		if Msg.Id == Id {
			args := []reflect.Value{reflect.ValueOf(Msg.Body)}
			Caller := reflect.ValueOf(w.callbacks[Id])
			Caller.Call(args)
			delete(w.callbacks, Id)
		}
	})
}
func (w *WarpperClient) Notify(router string, msg interface{}) {
	message := Message{Id: 0, Body: msg}
	_ = w.client.Emit(router, message)
}
func (w *WarpperClient) SocketId() (id string) {
	return w.client.Id()
}
func (w *WarpperClient) OnDisconnect(callback func(c Channel)) {
	_ = w.client.On(gosocketio.OnDisconnection, callback)
}
func (w *WarpperClient) OnConnect(callback func(c Channel)) {
	_ = w.client.On(gosocketio.OnConnection, callback)
}
func (w *WarpperClient) OnError(callback func(c Channel)) {
	_ = w.client.On(gosocketio.OnError, callback)
}
func (w *WarpperClient) OnOpen(callback interface{}) {
	_ = w.client.On("open", func(ch Channel, msg interface{}) {
		caller := reflect.ValueOf(callback)
		args := []reflect.Value{reflect.ValueOf(msg)}
		caller.Call(args)
	})
}
func (w *WarpperClient) OnClose(callback interface{}) {
	_ = w.client.On("close", func(ch Channel, msg interface{}) {
		caller := reflect.ValueOf(callback)
		args := []reflect.Value{reflect.ValueOf(msg)}
		caller.Call(args)
	})
}
func (w *WarpperClient) OnEcho(callback interface{}) {
	_ = w.client.On("echo", func(ch Channel, Msg Message) {
		caller := reflect.ValueOf(callback)
		args := []reflect.Value{reflect.ValueOf(Msg.Body)}
		caller.Call(args)
	})
}
