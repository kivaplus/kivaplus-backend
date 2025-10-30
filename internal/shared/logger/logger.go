package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Logger interface defines logging methods
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// JSONLogger implements structured JSON logging
type JSONLogger struct {
	logger *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new JSON logger
func New() Logger {
	return &JSONLogger{
		logger: log.New(os.Stdout, "", 0),
	}
}

// Info logs an info message
func (l *JSONLogger) Info(msg string, fields ...interface{}) {
	l.log("INFO", msg, fields...)
}

// Error logs an error message
func (l *JSONLogger) Error(msg string, fields ...interface{}) {
	l.log("ERROR", msg, fields...)
}

// Debug logs a debug message
func (l *JSONLogger) Debug(msg string, fields ...interface{}) {
	l.log("DEBUG", msg, fields...)
}

// Warn logs a warning message
func (l *JSONLogger) Warn(msg string, fields ...interface{}) {
	l.log("WARN", msg, fields...)
}

// log is the internal logging method
func (l *JSONLogger) log(level, msg string, fields ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Fields:    make(map[string]interface{}),
	}

	// Parse fields as key-value pairs
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			entry.Fields[key] = value
		}
	}

	// Marshal to JSON and log
	if jsonBytes, err := json.Marshal(entry); err == nil {
		l.logger.Println(string(jsonBytes))
	} else {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("[%s] %s %v", level, msg, fields)
	}
}
