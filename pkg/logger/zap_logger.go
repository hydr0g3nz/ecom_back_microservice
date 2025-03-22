package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implements the Logger interface using Zap
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// NewZapLogger creates a new instance of ZapLogger
func NewZapLogger() *ZapLogger {
	// Create a production encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	
	// Determine if we should use JSON or console format
	var encoder zapcore.Encoder
	logFormat := os.Getenv("LOG_FORMAT")
	if strings.ToLower(logFormat) == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	
	// Set log level based on environment
	level := zapcore.InfoLevel
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	
	// Create a core with stdout
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)
	
	// Create the logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	).Sugar()
	
	return &ZapLogger{logger: logger}
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugw(msg, keysAndValues...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warnw(msg, keysAndValues...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and then calls os.Exit(1)
func (l *ZapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalw(msg, keysAndValues...)
}

// With returns a logger with the specified key-value pairs
func (l *ZapLogger) With(keysAndValues ...interface{}) Logger {
	return &ZapLogger{logger: l.logger.With(keysAndValues...)}
}

// WithCorrelationID returns a logger with the correlation ID field
func (l *ZapLogger) WithCorrelationID(correlationID string) Logger {
	return &ZapLogger{logger: l.logger.With("correlation_id", correlationID)}
}
