package wswarpper

import (
	"Vpn-golang/comm"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"log"
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
	*gosocketio.Client
	seqid     int64
	callbacks map[int64]interface{}
	keys      string
	timestamp string
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
	return
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
	_ = w.On(gosocketio.OnDisconnection, callback)
}
func (w *WarpperClient) OnConnect(callback func(c Channel)) {
	_ = w.On(gosocketio.OnConnection, callback)
}
func (w *WarpperClient) OnError(callback func(c Channel)) {
	_ = w.On(gosocketio.OnError, callback)
}

func (w *WarpperClient) Login(manager_port, beginport, endport, controrller_port int, state, area string) {
	request := struct {
		ManagerPort    int    `json:"manager_port"`
		ControllerPort int    `json:"controller_port"`
		Beginport      int    `json:"beginport"`
		Endport        int    `json:"endport"`
		State          string `json:"state"`
		Area           string `json:"area"`
	}{
		ManagerPort:    manager_port,
		ControllerPort: controrller_port,
		Beginport:      beginport,
		Endport:        endport,
		State:          state, Area: area,
	}
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
func (w *WarpperClient) Transfer(sid int, transfer []int64) {
	request := struct {
		Sid      int     `json:"sid"`
		Transfer []int64 `json:"transfer"`
	}{
		Sid:      sid,
		Transfer: transfer,
	}
	w.Notify("transfer", request)

}
func (w *WarpperClient) Timeout(sid, uid int, transfer []int64, activestamp int64) {
	request := struct {
		Sid         int     `json:"sid"`
		Uid         int     `json:"uid"`
		Transfer    []int64 `json:"transfer"`
		Activectamp int64   `json:"activestamp"`
	}{
		Sid:         sid,
		Uid:         uid,
		Transfer:    transfer,
		Activectamp: activestamp,
	}
	w.Notify("transfer", request)
}
func (w *WarpperClient) Overflow(sid, uid int, limit int) {
	request := struct {
		Sid       int `json:"sid"`
		Uid       int `json:"uid"`
		Limitup   int `json:"limitup"`
		Limitdown int `json:"limitdown"`
	}{
		Sid:       sid,
		Uid:       uid,
		Limitup:   limit,
		Limitdown: limit,
	}
	w.Notify("overflow", request)
}
func (w *WarpperClient) Expire(sid, uid int, transfer []int64) {
	request := struct {
		Sid      int     `json:"sid"`
		Uid      int     `json:"uid"`
		Transfer []int64 `json:"transfer"`
	}{
		Sid:      sid,
		Uid:      uid,
		Transfer: transfer,
	}
	w.Notify("overflow", request)
}
func (w *WarpperClient) Balance(sid, uid int, duration int) {
	request := struct {
		Sid  int `json:"sid"`
		Uid  int `json:"uid"`
		Time int `json:"NoticeTime"`
	}{
		Sid:  sid,
		Uid:  uid,
		Time: duration,
	}
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
