package logger

import (
	"fmt"
	"os"
	"path/filepath"

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

// InitWithConfig initializes the logger with configuration settings
func InitWithConfig(level, format, output, filePath string) error {
	log = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// Set formatter
	switch format {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			FullTimestamp:   true,
		})
	case "json":
		fallthrough
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	}

	// Set output
	switch output {
	case "file":
		if filePath == "" {
			return fmt.Errorf("file_path must be specified when output is 'file'")
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		log.SetOutput(file)
	case "stdout":
		fallthrough
	default:
		log.SetOutput(os.Stdout)
	}

	return nil
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

// RotateLog closes the current log file and opens a new one
// This is useful for log rotation systems
func RotateLog(filePath string) error {
	if log == nil {
		return fmt.Errorf("logger not initialized")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	log.SetOutput(file)
	return nil
}

// LogRequest logs HTTP request information with structured fields
func LogRequest(method, path, clientIP string, statusCode int, latency string) {
	WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"status_code": statusCode,
		"latency":     latency,
		"type":        "request",
	}).Info("HTTP request processed")
}

// LogDatabase logs database operation information
func LogDatabase(operation, table string, duration string, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration,
		"type":      "database",
	}

	if err != nil {
		fields["error"] = err.Error()
		WithFields(fields).Error("Database operation failed")
	} else {
		WithFields(fields).Debug("Database operation completed")
	}
}

// LogCache logs cache operation information
func LogCache(operation, key string, hit bool, duration string) {
	WithFields(logrus.Fields{
		"operation": operation,
		"key":       key,
		"hit":       hit,
		"duration":  duration,
		"type":      "cache",
	}).Debug("Cache operation completed")
}

// LogQueue logs queue operation information
func LogQueue(operation string, jobType string, workerID int, duration string, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"job_type":  jobType,
		"worker_id": workerID,
		"duration":  duration,
		"type":      "queue",
	}

	if err != nil {
		fields["error"] = err.Error()
		WithFields(fields).Error("Queue operation failed")
	} else {
		WithFields(fields).Info("Queue operation completed")
	}
}
