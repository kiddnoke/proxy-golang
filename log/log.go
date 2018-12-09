package log

import (
	"github.com/google/logger"
	"log"
	"os"
)

type SsLogger struct {
	*logger.Logger
	level int
}

func NewLogger(instance, filepath string, level int) *SsLogger {
	lf, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	logger.SetFlags(log.Ldate | log.Lmicroseconds)
	loggerInstance := logger.Init(instance, false, false, lf)
	return &SsLogger{loggerInstance, level}
}
