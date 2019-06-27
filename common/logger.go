package common

import (
	"fmt"
	golog "log"
	"os"
)

var defaultLevel int

const (
	LOG_TRACE = iota
	LOG_DEBUG
	LOG_INFO
	LOG_WARN
	LOG_ERROR
	LOG_PANIC
)

var levelstr = map[int]string{
	LOG_TRACE: "TRACE",
	LOG_DEBUG: "DEBUG",
	LOG_INFO:  "INFO",
	LOG_WARN:  "WARN",
	LOG_ERROR: "ERROR",
	LOG_PANIC: "PANIC",
}

type Logger struct {
	level  int
	prefix string
	ll     golog.Logger
}

func NewLogger(level int, prefix string) *Logger {

	return &Logger{
		level:  level,
		prefix: prefix,
		ll:     *golog.New(os.Stdout, prefix+" ", golog.LstdFlags|golog.Lshortfile),
	}
}
func (l *Logger) SetLevel(level int) {
	l.level = level
}
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
	l.ll.SetPrefix(prefix + " ")
}
func (l *Logger) Trace(format string, v ...interface{}) {
	l.output(LOG_TRACE, fmt.Sprintf(format, v...))
}
func (l *Logger) Debug(format string, v ...interface{}) {
	l.output(LOG_DEBUG, fmt.Sprintf(format, v...))
}
func (l *Logger) Info(format string, v ...interface{}) {
	l.output(LOG_INFO, fmt.Sprintf(format, v...))

}
func (l *Logger) Warn(format string, v ...interface{}) {
	l.output(LOG_WARN, fmt.Sprintf(format, v...))
}
func (l *Logger) Error(format string, v ...interface{}) {
	l.output(LOG_ERROR, fmt.Sprintf(format, v...))
}
func (l *Logger) Panic(format string, v ...interface{}) {
	l.output(LOG_PANIC, fmt.Sprintf(format, v...))
	panic(fmt.Sprintf(format, v...))
}
func (l *Logger) output(level int, s string) error {
	if l.level <= level {
		t := levelstr[level] + " "
		t += s
		return l.ll.Output(3, t)
	}
	return nil
}
func LoggerLevel(level int) string {
	if LOG_TRACE <= level && level <= LOG_PANIC {
		return levelstr[level]
	}
	panic(fmt.Sprintf("it do no have %d level", level))
}
func SetDefaultLevel(level int) {
	if LOG_TRACE <= level && level <= LOG_PANIC {
		defaultLevel = level
		return
	}
	panic(fmt.Sprintf("it do no have %d level", level))
}
func GetDefaultLevel() (string, int) {
	str := LoggerLevel(defaultLevel)
	return str, defaultLevel
}
