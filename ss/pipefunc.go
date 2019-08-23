package ss

import "net"

func PipeThenClose(left, right net.Conn, addTraffic func(n int)) (netErr error) {
	buf := leakyBuf.Get()
	defer leakyBuf.Put(buf)
	return PipeCacheThenClose(buf, left, right, nil, addTraffic)
}
func PipeCacheThenClose(buffcache []byte, left, right net.Conn, pipecache []byte, transter_callback func(n int)) (netErr error) {
	for {
		n, err := left.Read(buffcache)
		if err != nil {
			return
		}
		if n > 0 {
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
func PipeNonBlocking(buffcache []byte, left, right net.Conn, pipecache *[]byte, transter_callback func(n int)) (netErr error) {
	setTcpConNonBlocking(left)
	setTcpConNonBlocking(right)
	for {
		nr, err := left.Read(buffcache)
		if nr == 0 || err != nil {
			return err
		}
		if nr > 0 {
			*pipecache = append(*pipecache, buffcache[:nr]...)

			//right.SetWriteDeadline(time.Now().Add(WriteDeadlineDuration))
			nw, err := right.Write(*cache)
			if nw > 0 {
				*cache = (*cache)[nw:]
				addTraffic(nr, len(*cache))
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
