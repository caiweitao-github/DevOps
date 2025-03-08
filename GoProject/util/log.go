package util

import (
	"log"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func NewInitLog(fileName string) *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
	})
	file, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	log.SetOutput(file)
	return log
}

func InitLog(filename, prefix, prefix1 string) (logInfo, logError *log.Logger) {
	logInfo = LogConf(filename, prefix)
	logError = LogConf(filename, prefix1)
	return
}

func LogConf(filename, prefix string) *log.Logger {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate)
	logger := log.New(logFile, prefix, log.LstdFlags)
	return logger
}
