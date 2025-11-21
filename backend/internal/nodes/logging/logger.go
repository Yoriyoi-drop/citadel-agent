// backend/internal/nodes/logging/logger.go
package logging

import (
	"context"
	"fmt"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogNodeConfig represents the configuration for a log node
type LogNodeConfig struct {
	Level      LogLevel            `json:"level"`
	Message    string             `json:"message"`
	Fields     map[string]string  `json:"fields"`
	Format     string             `json:"format"` // json, text, etc.
	OutputType string             `json:"output_type"` // console, file, http
	OutputDest string             `json:"output_dest"` // file path or URL
	Enabled    bool               `json:"enabled"`
}

// LogNode represents a logging node
type LogNode struct {
	config *LogNodeConfig
}

// NewLogNode creates a new log node
func NewLogNode(config *LogNodeConfig) *LogNode {
	// Set defaults if not provided
	if config.Level == "" {
		config.Level = LogLevelInfo
	}
	if config.Format == "" {
		config.Format = "json"
	}
	if config.OutputType == "" {
		config.OutputType = "console"
	}
	if config.Enabled {
		config.Enabled = true
	}

	return &LogNode{
		config: config,
	}
}

// Execute executes the logging node
func (ln *LogNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	message := ln.config.Message
	if msg, exists := inputs["message"]; exists {
		if msgStr, ok := msg.(string); ok {
			message = msgStr
		}
	}

	level := ln.config.Level
	if lvl, exists := inputs["level"]; exists {
		if lvlStr, ok := lvl.(string); ok {
			level = LogLevel(lvlStr)
		}
	}

	// Add additional fields from inputs
	fields := make(map[string]interface{})
	for k, v := range ln.config.Fields {
		fields[k] = v
	}

	// Add fields from inputs
	for k, v := range inputs {
		if k != "message" && k != "level" {
			fields[k] = v
		}
	}

	// Only execute if logging is enabled
	if !ln.config.Enabled {
		return map[string]interface{}{
			"success": true,
			"message": "logging disabled",
			"level":   string(level),
		}, nil
	}

	// Log the message based on output type
	err := ln.logMessage(level, message, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to log message: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"level":   string(level),
		"message": message,
		"fields":  fields,
		"output_type": ln.config.OutputType,
		"timestamp": time.Now().Unix(),
	}, nil
}

// logMessage logs the message based on the output type
func (ln *LogNode) logMessage(level LogLevel, message string, fields map[string]interface{}) error {
	// In a real implementation, we would route to different loggers based on output type
	// For now, we'll simulate the behavior
	
	switch ln.config.OutputType {
	case "console":
		return ln.logToConsole(level, message, fields)
	case "file":
		return ln.logToFile(ln.config.OutputDest, level, message, fields)
	case "http":
		return ln.logToHTTP(ln.config.OutputDest, level, message, fields)
	default:
		return ln.logToConsole(level, message, fields)
	}
}

// logToConsole logs to console output
func (ln *LogNode) logToConsole(level LogLevel, message string, fields map[string]interface{}) error {
	// Simulate console logging
	// In a real implementation, we would format and output to console
	logEntry := fmt.Sprintf("[%s] %s | %s | Fields: %v", 
		level, 
		time.Now().Format("2006-01-02 15:04:05"), 
		message, 
		fields)
	
	fmt.Println(logEntry) // This would be replaced with proper logger in real implementation
	return nil
}

// logToFile logs to a file
func (ln *LogNode) logToFile(filePath, level LogLevel, message string, fields map[string]interface{}) error {
	// For simplicity, we'll just return nil
	// In a real implementation, we would write to the specified file
	return nil
}

// logToHTTP logs to an HTTP endpoint
func (ln *LogNode) logToHTTP(endpoint, level LogLevel, message string, fields map[string]interface{}) error {
	// For simplicity, we'll just return nil
	// In a real implementation, we would send an HTTP request to the endpoint
	return nil
}

// RegisterLogNode registers the log node type with the engine
func RegisterLogNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("logging", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var level LogLevel
		if levelVal, exists := config["level"]; exists {
			if levelStr, ok := levelVal.(string); ok {
				level = LogLevel(levelStr)
			}
		}

		var message string
		if msgVal, exists := config["message"]; exists {
			if msgStr, ok := msgVal.(string); ok {
				message = msgStr
			}
		}

		var outputType string
		if outputVal, exists := config["output_type"]; exists {
			if outputStr, ok := outputVal.(string); ok {
				outputType = outputStr
			}
		}

		var outputDest string
		if destVal, exists := config["output_dest"]; exists {
			if destStr, ok := destVal.(string); ok {
				outputDest = destStr
			}
		}

		var format string
		if formatVal, exists := config["format"]; exists {
			if formatStr, ok := formatVal.(string); ok {
				format = formatStr
			}
		}

		var enabled bool
		if enabledVal, exists := config["enabled"]; exists {
			if enabledBool, ok := enabledVal.(bool); ok {
				enabled = enabledBool
			}
		} else {
			enabled = true // default to enabled
		}

		var fields map[string]string
		if fieldsVal, exists := config["fields"]; exists {
			if fieldsMap, ok := fieldsVal.(map[string]interface{}); ok {
				fields = make(map[string]string)
				for k, v := range fieldsMap {
					if vStr, ok := v.(string); ok {
						fields[k] = vStr
					} else {
						fields[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}

		nodeConfig := &LogNodeConfig{
			Level:      level,
			Message:    message,
			Fields:     fields,
			Format:     format,
			OutputType: outputType,
			OutputDest: outputDest,
			Enabled:    enabled,
		}

		return NewLogNode(nodeConfig), nil
	})
}