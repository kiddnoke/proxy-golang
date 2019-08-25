// +build linux

package ss

import (
	"net"
	"syscall"

	"github.com/pkg/errors"
)

func setTcpConNonBlocking(conn net.Conn) (err error) {
	c, ok := conn.(*net.TCPConn)
	if !ok {
		return errors.New("only work with TCP connection")
	}
	f, err := c.File()
	if err != nil {
		return err
	}

	fd := f.Fd()

	// The File() call above puts both the original socket fd and the file fd in blocking mode.
	// Set the file fd back to non-blocking mode and the original socket fd will become non-blocking as well.
	// Otherwise blocking I/O will waste OS threads.
	if err := syscall.SetNonblock(int(fd), true); err != nil {
		return err
	}
	return
}
