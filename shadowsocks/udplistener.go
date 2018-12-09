package shadowsocks

import (
	"context"
	"log"
	"net"
)

type UdpListener struct {
	*net.UDPConn
	config  SSconfig
	cipher  *Cipher // 加密子
	limiter *speedlimiter
	running bool
	ctx     context.Context
}

func NewUdpListener(l *net.UDPConn, config SSconfig) *UdpListener {
	ctx := context.Background()
	cipher, err := NewCipher(config.Method, config.Password)
	if err != nil {
		log.Printf("Error generating cipher for port: %d %v\n", config.ServerPort, err)
	}
	return &UdpListener{UDPConn: l, limiter: util.NewSpeedLimiterWithContext(ctx, config.Limit*1024), config: config, cipher: cipher, ctx: ctx}
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
			log.Printf("udp transfer btye len[%d] ", i)
		}); err != nil {
			log.Printf("udp read error:[%s]", err.Error())
			break
		}
	}
	log.Printf("UdpRelayer port:[%d] Uid:[%d] Sid:[%d] Close", l.config.ServerPort, l.config.Uid, l.config.Sid)
}
func (l *UdpListener) Start() {
	l.running = true
	go l.Listening()
}
func (l *UdpListener) Stop() {
	l.running = false
	l.Close()
}
func (l *UdpListener) Accept() error {
	buf := leakyBuf.Get()
	_, _, err := l.ReadFrom(buf[0:])
	if err != nil {
		return err
	}
	return nil
}
