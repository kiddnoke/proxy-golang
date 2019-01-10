package comm

type Community interface {
	Connect(host string, port int) (err error)
	/**
	 * manager_port
	 * beginport
	 * endport
	 * controller_port
	 * health
	 * state
	 * area
	 */
	Login(manager_port, beginport, endport, controrller_port int, state, area string)
	/**
	 * socket id
	 */
	Logout()
	/**
	 * health
	 */
	Health(health int)
	/*
	 * nil
	 */
	HeartBeat()
	/*
	 * [ {sid , transfer}]
	 */
	Transfer(sid int, transfer []int64)
	/*
	 * { sid, uid, transfer, active }
	 */
	Timeout(sid, uid int, transfer []int64, activestamp int64)
	/*
	 * { sid, uid, limitup, limitdown }
	 */
	Overflow(sid, uid int, limit int)
	/*
	 * { sid, uid, transfer }
	 */
	Expire(sid, uid int, transfer []int64)
	/*
	 * { uid, sid, FreeUid, Time }
	 */
	Balance(sid, uid int, duration int)
	Echo(json interface{})
	OnOpened(callback func(msg []byte))
	OnClosed(callback func(msg []byte))
	OnLimit(callback func(msg map[string]interface{}))
}
