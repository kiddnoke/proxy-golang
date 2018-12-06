package shadowsocks

import (
	"net"
	"time"
)

type UdpListener struct {
	*net.UDPConn
	config       Config
	cipher       *Cipher // 加密子
	speedlimiter *Bucket //
}

func NewUdpListener(config Config) *UdpListener {
	speedlimiter := NewBucket(time.Second, config.Limit*1024)
	return &UdpListener{speedlimiter: speedlimiter}
}
