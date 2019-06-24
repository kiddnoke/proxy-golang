package multiprotocol

import (
	"time"
)

type Limiter interface {
	SetLimit(bytesPerSec int)
	Burst() int
}
type TrafficStatistic interface {
	GetTraffic() (tu, td, uu, ud int64)
	GetStartTimeStamp() time.Time
	GetLastTimeStamp() time.Time
	Clear()
}
type Switcher interface {
	Start()
	Stop()
	Close()
}
type Isser interface {
	IsTimeout() bool
	IsExpire() bool
	IsOverflow() bool
	IsNotify() bool
	IsStairCase() (limit int, flag bool)
}
type Getter interface {
	GetConfig() *Config
}
type Logger interface {
	Printf(format string, v ...interface{})
}
type Relayer interface {
	Switcher
	Isser
	TrafficStatistic
	Limiter
	Getter
}
