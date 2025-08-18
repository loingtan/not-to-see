package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(verbose bool) {
	log = logrus.New()

	log.SetOutput(os.Stdout)

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

func GetLogger() *logrus.Logger {
	if log == nil {
		Init(false)
	}
	return log
}

func Debug(format string, args ...any) {
	GetLogger().Debugf(format, args...)
}

func Info(format string, args ...any) {
	GetLogger().Infof(format, args...)
}

func Warn(format string, args ...any) {
	GetLogger().Warnf(format, args...)
}

func Error(format string, args ...any) {
	GetLogger().Errorf(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}
