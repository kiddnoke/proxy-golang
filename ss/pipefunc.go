package ss

import (
	"net"
	"time"
)

var WriteDeadlineDuration = time.Millisecond * 5

func PipeThenClose(left, right net.Conn, addTraffic func(n int)) (netErr error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	return PipeCacheThenClose(buf, left, right, addTraffic)
}
func PipeCacheThenClose(buffcache []byte, left, right net.Conn, transter_callback func(n int)) (netErr error) {
	for {
		n, err := left.Read(buffcache)
		if err != nil {
			return
		}
		if n > 0 {
			transter_callback(n)
			if _, err := right.Write(buffcache[0:n]); err != nil {
				netErr = err
				break
			}
		}
		if err != nil {
			netErr = err
			break
		}
	}
	return
}
func PipeNonBlocking(pipecache *[]byte, left, right net.Conn, transter_callback func(nr int, nw int, pipelength int)) (netErr error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	return PipeCacheNonBlocking(buf, pipecache, left, right, transter_callback)
}

func PipeCacheNonBlocking(buffcache []byte, pipecache *[]byte, left, right net.Conn, transter_callback func(nr int, nw int, pipelength int)) (netErr error) {
	setTcpConNonBlocking(right)
	for {
		nr, err := left.Read(buffcache)
		if nr == 0 || err != nil {
			return err
		}
		if nr > 0 {
			*pipecache = append(*pipecache, buffcache[:nr]...)
			//right.SetWriteDeadline(time.Now().Add(WriteDeadlineDuration))
			nw, err := right.Write(*pipecache)
			if nw > 0 {
				*pipecache = (*pipecache)[nw:]
				transter_callback(nr, nw, len(*pipecache))
			}
			if err != nil {
				if eo, ok := err.(*net.OpError); ok && eo.Timeout() {
					continue
				}
				netErr = err
				break
			}
		}
	}
	return
}
