package engine

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// BasicLogger provides a simple logging implementation
type BasicLogger struct {
	level   LogLevel
	logger  *log.Logger
}

// NewBasicLogger creates a new basic logger instance
func NewBasicLogger(level LogLevel) *BasicLogger {
	return &BasicLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
	}
}

// Debug logs a debug message
func (bl *BasicLogger) Debug(msg string, fields ...map[string]interface{}) {
	if bl.level > DebugLevel {
		return
	}
	bl.logMessage(DebugLevel, msg, fields...)
}

// Info logs an info message
func (bl *BasicLogger) Info(msg string, fields ...map[string]interface{}) {
	if bl.level > InfoLevel {
		return
	}
	bl.logMessage(InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (bl *BasicLogger) Warn(msg string, fields ...map[string]interface{}) {
	if bl.level > WarnLevel {
		return
	}
	bl.logMessage(WarnLevel, msg, fields...)
}

// Error logs an error message
func (bl *BasicLogger) Error(msg string, fields ...map[string]interface{}) {
	if bl.level > ErrorLevel {
		return
	}
	bl.logMessage(ErrorLevel, msg, fields...)
}

// logMessage formats and writes the log message
func (bl *BasicLogger) logMessage(level LogLevel, msg string, fields ...map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000000")
	levelStr := bl.levelToString(level)

	fieldStr := ""
	if len(fields) > 0 {
		combinedFields := make(map[string]interface{})
		for _, fieldMap := range fields {
			for k, v := range fieldMap {
				combinedFields[k] = v
			}
		}
		fieldStr = fmt.Sprintf(" %v", combinedFields)
	}

	logMsg := fmt.Sprintf("[%s] [%s] %s%s", timestamp, levelStr, msg, fieldStr)
	bl.logger.Println(logMsg)
}

// levelToString converts LogLevel to string representation
func (bl *BasicLogger) levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}