package shadowsocks

import (
	"net"
	"time"
)

type TcpListener struct {
	*net.TCPListener
	config       Config
	speedlimiter *Bucket
}

func NewTcpListener(config Config) *TcpListener {
	speedlimiter := NewBucket(time.Second, config.Limit*1024)
	return &TcpListener{speedlimiter:speedlimiter}
}
