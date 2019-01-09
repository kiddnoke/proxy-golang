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
	Transfer(sid int, transfer []int)
	/*
	 * { sid, uid, transfer, active }
	 */
	Timeout(sid, uid int, transfer []int, activestamp int)
	/*
	 * { sid, uid, limitup, limitdown }
	 */
	Overflow(sid, uid int, limit int)
	/*
	 * { sid, uid, transfer }
	 */
	Expire(sid, uid int, transfer []int)
	/*
	 * { uid, sid, FreeUid, Time }
	 */
	Balance(sid, uid int, FreeUid string, duration int)
	Echo(json interface{})
	OnOpened(callback func(msg interface{}))
	OnClosed(callback func(msg interface{}))
}
