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
	GetMaxRate() (float64, float64)
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
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}
type Relayer interface {
	Switcher
	Isser
	TrafficStatistic
	Limiter
	Getter
	Logger
}
