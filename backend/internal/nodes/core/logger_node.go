package core

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// LoggerNodeConfig represents the configuration for a Logger node
type LoggerNodeConfig struct {
	Level     string                 `json:"level"`      // debug, info, warn, error, fatal
	Message   string                 `json:"message"`    // log message template
	Fields    map[string]interface{} `json:"fields,omitempty"` // additional fields to log
	Output    string                 `json:"output"`     // stdout, file, or custom
	FilePath  string                 `json:"file_path,omitempty"` // path to log file if output is file
	Format    string                 `json:"format"`     // json or console
	WithTrace bool                   `json:"with_trace"` // include trace information
}

// LoggerNode handles logging with various configurations
type LoggerNode struct {
	config LoggerNodeConfig
	logger zerolog.Logger
}

// NewLoggerNode creates a new Logger node with the given configuration
func NewLoggerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Extract config values
	level := getStringValue(config["level"], "info")
	message := getStringValue(config["message"], "")
	output := getStringValue(config["output"], "stdout")
	filePath := getStringValue(config["file_path"], "")
	format := getStringValue(config["format"], "json")
	withTrace := getBoolValue(config["with_trace"], false)

	// Extract fields
	fields := make(map[string]interface{})
	if fieldsVal, exists := config["fields"]; exists {
		if fieldsMap, ok := fieldsVal.(map[string]interface{}); ok {
			fields = fieldsMap
		}
	}

	loggerConfig := LoggerNodeConfig{
		Level:     level,
		Message:   message,
		Fields:    fields,
		Output:    output,
		FilePath:  filePath,
		Format:    format,
		WithTrace: withTrace,
	}

	// Configure logger based on settings
	var logger zerolog.Logger

	switch loggerConfig.Output {
	case "file":
		if loggerConfig.FilePath == "" {
			return nil, fmt.Errorf("file_path is required when output is file")
		}
		file, err := os.OpenFile(loggerConfig.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger = zerolog.New(file).With().Timestamp().Logger()
	default:
		// Default to stdout
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Set log level
	var zLevel zerolog.Level
	switch loggerConfig.Level {
	case "debug":
		zLevel = zerolog.DebugLevel
	case "info":
		zLevel = zerolog.InfoLevel
	case "warn":
		zLevel = zerolog.WarnLevel
	case "error":
		zLevel = zerolog.ErrorLevel
	case "fatal":
		zLevel = zerolog.FatalLevel
	default:
		zLevel = zerolog.InfoLevel
	}
	logger = logger.Level(zLevel)

	// Set format
	if loggerConfig.Format == "console" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	return &LoggerNode{
		config: loggerConfig,
		logger: logger,
	}, nil
}

// Execute implements the NodeInstance interface
func (l *LoggerNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Prepare log fields
	logFields := make(map[string]interface{})

	// Add fields from config
	for k, v := range l.config.Fields {
		logFields[k] = v
	}

	// Add fields from input
	for k, v := range input {
		// Don't overwrite fields from config, but add new ones
		if _, exists := logFields[k]; !exists {
			logFields[k] = v
		}
	}

	// Add trace information if requested
	if l.config.WithTrace {
		logFields["trace"] = map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"node_type": "logger",
		}
	}

	// Create log event
	logEvent := l.logger.With()

	// Add all fields to the log event
	for k, v := range logFields {
		logEvent = logEvent.Interface(k, v)
	}

	event := logEvent.Logger()

	// Format message using input if template contains placeholders
	message := l.config.Message
	if message == "" {
		message = "Log message from workflow"
	} else {
		// Simple placeholder replacement (e.g., {field_name})
		for k, v := range input {
			placeholder := "{" + k + "}"
			valueStr := fmt.Sprintf("%v", v)
			message = replaceAll(message, placeholder, valueStr)
		}
	}

	// Log based on the configured level
	switch l.config.Level {
	case "debug":
		event.Debug().Msg(message)
	case "warn":
		event.Warn().Msg(message)
	case "error":
		event.Error().Msg(message)
	case "fatal":
		event.Fatal().Msg(message)
	default:
		event.Info().Msg(message)
	}

	// Return success result
	return map[string]interface{}{
		"success":     true,
		"logged":      true,
		"message":     message,
		"log_fields":  logFields,
		"level":       l.config.Level,
		"output_type": l.config.Output,
		"timestamp":   time.Now().Unix(),
	}, nil
}

// replaceAll is a simple string replacement function
func replaceAll(s, old, new string) string {
	result := ""
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			// Check if it's a placeholder
			end := -1
			for j := i; j < len(s) && s[j] != '}'; j++ {
				if s[j] == '}' {
					end = j
					break
				}
			}
			if end != -1 {
				placeholder := s[i+1 : end]
				if placeholder == old {
					result += new
					i = end + 1
					continue
				}
			}
		}
		result += string(s[i])
		i++
	}
	return result
}

// getStringValue safely extracts a string value with default fallback
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// getBoolValue safely extracts a boolean value with default fallback
func getBoolValue(v interface{}, defaultValue bool) bool {
	if v == nil {
		return defaultValue
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return defaultValue
}