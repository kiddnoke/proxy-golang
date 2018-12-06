package shadowsocks

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

type Proxy struct {
	TcpInstance     TcpListener
	UdpInstance     UdpListener
	Conf            Config
	SpeedLimiter    Bucket
	TransferChannel chan interface{}
	master          *Manager
}

type Manager struct {
	sync.Mutex
	proxyTable map[int]Proxy
	conn       *net.UDPConn
	running    bool
}

func NewManager(port int) (m *Manager) {
	// 端口是否在用
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("localhost"), Port: port})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listening:", err)
		os.Exit(1)
	}
	return &Manager{proxyTable: make(map[int]Proxy), conn: conn ,running:false}
}
func (m *Manager) Add(port int, config Config) (e error, ok bool) {
	m.Lock()
	defer m.Unlock()
	// 1 有没有
	// 没有就加
	// 有就报错
	return
}
func (m *Manager) Get(port int) (p *Proxy, ok bool) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有就返回
	// 没有就报错
	return
}
func (m *Manager) Remove(port int) (e error, ok bool) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有就删除
	// 没有就说没这个key
	return
}
func (m *Manager) Update(port int, config Config) (e error, ok bool) {
	m.Lock()
	defer m.Unlock()
	// 有没有
	// 有的话，先删除， 在添加
	// 没有的话， 啥都不干 报个错
	return
}

type Command struct {
	Cmd string `json:"Cmd"`
	Config
}

func (m *Manager) Loop() {
	var params struct {
		Cmd string `json:"Cmd"`
		Config
	}
	conn := m.conn
	for m.running {
		data := make([]byte, 300)
		_, remote, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to read UDP manage msg, error: ", err.Error())
			continue
		}
		_ = json.Unmarshal(data, params)
		var res []byte
		switch params.Cmd {
		//case strings.HasPrefix(command, "add:"):
		case "open":
			_, _ = m.Add(params.Config.ServerPort, params.Config)
		case "close":
			_, _ = m.Remove(params.Config.ServerPort)
		//case strings.HasPrefix(command, "remove:"):
		case "remove":
			_, _ = m.Remove(params.Config.ServerPort)
			//case strings.HasPrefix(command, "update"):
			//case strings.HasPrefix(command, "query"):
			//case strings.HasPrefix(command, "ping"):
			//case strings.HasPrefix(command, "ping-stop"): // add the stop ping command
		}
		if len(res) == 0 {
			continue
		}
		_, err = conn.WriteToUDP(res, remote)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write UDP manage msg, error: ", err.Error())
			continue
		}
	}

}

func ManagerDaemon(m *Manager) {
	m.Loop()
}
func (m *Manager) Stop() {
	m.conn.Close()
	m.running = false
}
func (m *Manager) Start() {
	m.running = true
}
func (m *Manager) Run() {
	m.Loop()
}
