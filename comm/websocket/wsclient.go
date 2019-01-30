package wswarpper

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"proxy-golang/comm"

	"github.com/CHH/eventemitter"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	webSocketProtocol       = "ws://"
	webSocketSecureProtocol = "wss://"
	socketioUrl             = "/socket.io/?EIO=3&transport=websocket"
)
const VERSION = "v1.1.0"

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
	*gosocketio.Client
	seqid     int64
	callbacks map[int64]interface{}
	keys      string
	timestamp string
	emmiter   *eventemitter.EventEmitter
	comm.Community
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
		callbacks: make(map[int64]interface{}),
		emmiter:   eventemitter.New(),
	}
}
func (w *WarpperClient) connect(host string, port int) (err error) {
	query := &url.Values{}
	query.Add("keys", w.keys)
	query.Add("timestamp", w.timestamp)
	w.Client, err = gosocketio.Dial(
		getUrlWithOpt(host, port, *query, false),
		&transport.WebsocketTransport{
			PingInterval:   5 * time.Second,
			PingTimeout:    10 * time.Second,
			ReceiveTimeout: 15 * time.Second,
			SendTimeout:    20 * time.Second,
			BufferSize:     transport.WsDefaultBufferSize,
		})
	if err != nil {
		return
	}
	_ = w.Client.On(gosocketio.OnConnection, func(c Channel) {
		w.emmiter.Emit("connect", c)
	})
	_ = w.Client.On(gosocketio.OnDisconnection, func(c Channel) {
		w.emmiter.Emit("disconnect", c)
	})
	_ = w.Client.On(gosocketio.OnError, func(c Channel) {
		w.emmiter.Emit("error", c)
	})
	return
}
func (w *WarpperClient) Request(router string, msg interface{}, callback interface{}) {
	w.seqid++
	Id := w.seqid
	message := Message{Id: Id, Body: msg}
	_ = w.Client.Emit(router, message)
	w.callbacks[Id] = callback
	_ = w.Client.On(router, func(channel Channel, Msg Message) {
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
	_ = w.Emit(router, message)
}
func (w *WarpperClient) SocketId() (id string) {
	return w.Id()
}
func (w *WarpperClient) OnDisconnect(callback func(c Channel)) {
	w.emmiter.On("disconnect", callback)
}
func (w *WarpperClient) OnConnect(callback func(c Channel)) {
	_ = w.On(gosocketio.OnConnection, callback)
	w.emmiter.On("connect", callback)
}
func (w *WarpperClient) OnError(callback func(c Channel)) {
	_ = w.On(gosocketio.OnError, callback)
	w.emmiter.On("error", callback)
}

func (w *WarpperClient) Connect(host string, port int) (err error) {
	for {
		time.Sleep(time.Second)
		err := w.connect(host, port)
		if err != nil {
			log.Printf("wsclient connecting error :[%s]", err.Error())
			continue
		} else {
			break
		}
	}
	return nil
}
func (w *WarpperClient) Login(manager_port, beginport, endport, controrller_port int, state, area string) {
	request := make(map[string]interface{})
	request["manager_port"] = manager_port
	request["controller_port"] = controrller_port
	request["beginport"] = beginport
	request["endport"] = endport
	request["state"] = state
	request["area"] = area
	w.Notify("login", request)
	return
}
func (w *WarpperClient) Logout() {
	w.Notify("logout", nil)
}
func (w *WarpperClient) Health(health int) {
	w.Notify("health", health)
}
func (w *WarpperClient) HeartBeat() {
	w.Notify("heartbeat", nil)
}
func (w *WarpperClient) Transfer(sid int64, transfer []int64) {
	request := make(map[string]interface{})
	request["sid"] = sid
	request["transfer"] = transfer
	w.Notify("transfer", []interface{}{request})
}
func (w *WarpperClient) TransferList(transferList []interface{}) {
	w.Notify("transfer", transferList)
}
func (w *WarpperClient) Timeout(sid, uid int64, transfer []int64, activestamp int64) {
	request := make(map[string]interface{})
	request["sid"] = sid
	request["uid"] = uid
	request["transfer"] = transfer
	request["activestamp"] = activestamp
	w.Notify("timeout", request)
}
func (w *WarpperClient) Overflow(sid, uid int64, limit int) {
	request := make(map[string]interface{})
	request["sid"] = sid
	request["uid"] = uid
	request["limitup"] = limit
	request["limitdown"] = limit
	w.Notify("overflow", request)
}
func (w *WarpperClient) Expire(sid, uid int64, transfer []int64) {
	request := make(map[string]interface{})
	request["sid"] = sid
	request["uid"] = uid
	request["transfer"] = transfer
	w.Notify("expire", request)
}
func (w *WarpperClient) Balance(sid, uid int64, duration int) {
	request := make(map[string]interface{})
	request["sid"] = sid
	request["uid"] = uid
	request["balancenotifytime"] = duration
	w.Notify("balance", request)
}
func (w *WarpperClient) Echo(json interface{}) {
	w.Notify("echo", json)
}
func (w *WarpperClient) OnOpened(callback func(msg []byte)) {
	_ = w.On("open", func(channel Channel, Msg interface{}) {
		jsonstr, _ := json.Marshal(Msg)
		callback(jsonstr)
	})
}
func (w *WarpperClient) OnClosed(callback func(msg []byte)) {
	_ = w.On("close", func(channel Channel, Msg interface{}) {
		jsonstr, _ := json.Marshal(Msg)
		callback(jsonstr)
	})
}
func (w *WarpperClient) OnLimit(callback func(msg map[string]interface{})) {
	_ = w.On("limit", func(channel Channel, Msg interface{}) {
		callback(Msg.(map[string]interface{}))
	})
}
