package logger

// Logger defines the interface for logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, keysAndValues ...interface{})
	
	// Info logs an info message
	Info(msg string, keysAndValues ...interface{})
	
	// Warn logs a warning message
	Warn(msg string, keysAndValues ...interface{})
	
	// Error logs an error message
	Error(msg string, keysAndValues ...interface{})
	
	// Fatal logs a fatal message and then calls os.Exit(1)
	Fatal(msg string, keysAndValues ...interface{})
	
	// With returns a logger with the specified key-value pairs
	With(keysAndValues ...interface{}) Logger
	
	// WithCorrelationID returns a logger with the correlation ID field
	WithCorrelationID(correlationID string) Logger
}
