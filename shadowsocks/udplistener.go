package shadowsocks

import (
	"log"
	"net"
	"time"
)

type UdpListener struct {
	*net.UDPConn
	config       SSconfig
	cipher       *Cipher // 加密子
	speedlimiter *Bucket //
	running      bool
}

func NewUdpListener(l *net.UDPConn, config SSconfig) *UdpListener {
	speedlimiter := NewBucket(time.Second, config.Limit*1024)
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}

	return &UdpListener{UDPConn: l, speedlimiter: speedlimiter, config: config, cipher: cipher}
}
func makeUdpListener(l *net.UDPConn, config SSconfig) UdpListener {
	return *NewUdpListener(l, config)
}
func (l *UdpListener) Listening() {
	defer l.Close()
	log.Printf("SS listening at udp port[%d]", l.config.ServerPort)
	SecurePacketConn := NewSecurePacketConn(l, l.cipher.Copy())
	for l.running {
		if err := ReadAndHandleUDPReq(SecurePacketConn, func(i int) {

		}); err != nil {
			log.Printf("udp read error: %v\n", err)
			return
		}
	}
}
func (l *UdpListener) Start() {
	l.running = true
	go l.Listening()
}
func (l *UdpListener) Stop() {
	l.running = false
	l.Close()
}
