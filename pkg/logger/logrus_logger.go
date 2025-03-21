package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger defines the interface for logging

// logrusLogger implements the Logger interface using logrus
type logrusLogger struct {
	logger *logrus.Logger
}

// NewLogrusLogger creates a new instance of Logger using logrus
func NewLogrusLogger() Logger {
	// Create a new logrus logger
	logger := logrus.New()

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Use text formatter for more readable logs
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
		DisableSorting:  false,
		ForceColors:     true,
		PadLevelText:    true,
	})

	// Determine log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Enable caller info
	logger.SetReportCaller(true)

	return &logrusLogger{
		logger: logger,
	}
}

// convertToFields converts key-value pairs to logrus fields
func convertToFields(keysAndValues []interface{}) logrus.Fields {
	fields := logrus.Fields{}

	// Process the key-value pairs
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if ok {
				fields[key] = keysAndValues[i+1]
			}
		}
	}

	return fields
}

// Debug logs a message at debug level
func (l *logrusLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(convertToFields(keysAndValues)).Debug(msg)
}

// Info logs a message at info level
func (l *logrusLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(convertToFields(keysAndValues)).Info(msg)
}

// Warn logs a message at warn level
func (l *logrusLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(convertToFields(keysAndValues)).Warn(msg)
}

// Error logs a message at error level
func (l *logrusLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(convertToFields(keysAndValues)).Error(msg)
}

// Fatal logs a message at fatal level and then exits
func (l *logrusLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(convertToFields(keysAndValues)).Fatal(msg)
}
