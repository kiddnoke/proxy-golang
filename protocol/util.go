package protocol

import (
	"../log"
	"net/url"
)

func logf(f string, v ...interface{}) {
	log.Logf(f, v...)
}

func parseURL(s string) (addr, cipher, password string, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}
