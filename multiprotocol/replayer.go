package multiprotocol

import "time"

type Limiter interface {
	WaitN(n int) (err error)
	SetLimit(bytesPerSec int)
	Burst() int
}
type TrafficStatistic interface {
	GetTraffic() (tu, td, uu, ud int64)
	AddTraffic(tu, td, uu, ud int)
	GetStartTimeStamp() time.Time
	GetLastTimeStamp() time.Time
	Clear()
}
type Switcher interface {
	Start()
	Stop()
	Close()
}
type Panduanqi interface {
	IsTimeout() bool
	IsExpire() bool
	IsOverflow() bool
	IsNotify() bool
	IsStairCase() (limit int, flag bool)
}

type Relayer interface {
	Switcher
	Panduanqi
	TrafficStatistic
	Limiter
}
