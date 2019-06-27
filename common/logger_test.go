package common

import "testing"

var l *Logger

func init() {
}

func TestNewLogger(t *testing.T) {
	l = NewLogger(LOG_DEBUG, "lslslslsl")
}
func TestLogger_SetLevel(t *testing.T) {
	l.SetLevel(LOG_WARN)
	l.Debug("TestLogger_SetLevel DEBUG 不应该被看见")
	l.Warn("TestLogger_SetLevel WARN 可以被看见")
}
func TestLogger_SetPrefix(t *testing.T) {
	l.SetLevel(LOG_INFO)
	l.SetPrefix("墨绿的夜")
	l.Debug("TestLogger_SetLevel DEBUG 不应该被看见")
	l.Warn("TestLogger_SetLevel WARN 可以被看见")
}
