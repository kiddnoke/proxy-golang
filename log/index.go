package log

import "log"

func Logf(f string, v ...interface{}) {
	log.Printf(f, v...)
}
