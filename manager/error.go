package manager

import "errors"

var (
	KeyNotExist = errors.New("ProxyRelayNotExist")
	KeyExist    = errors.New("ProxyRelayExist")
)
