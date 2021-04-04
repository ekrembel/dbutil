package logger

import (
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"
)

var Logger *log.Entry

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	Logger = log.WithFields(log.Fields{
		"application": "coin-dbutil",
	})
}

func Info(logMsg ...interface{}) {
	_, fileName, lineNumber, _ := runtime.Caller(1)

	Logger.WithFields(log.Fields{
		"file": path.Base(fileName),
		"line": lineNumber,
	}).Info(logMsg)
}

func Error(logMsg ...interface{}) {
	_, fileName, lineNumber, _ := runtime.Caller(1)

	Logger.WithFields(log.Fields{
		"file": path.Base(fileName),
		"line": lineNumber,
	}).Error(logMsg)
}
