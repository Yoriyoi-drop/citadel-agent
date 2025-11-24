package utility

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// LoggerNode implements a node that logs data
type LoggerNode struct {
	id      string
	nodeType string
	level   string
	message string
	config  map[string]interface{}
}

// Initialize sets up the logger node with configuration
func (l *LoggerNode) Initialize(config map[string]interface{}) error {
	l.config = config

	if level, ok := config["level"]; ok {
		if lvl, ok := level.(string); ok {
			l.level = lvl
		} else {
			return fmt.Errorf("level must be a string")
		}
	} else {
		l.level = "info" // default level
	}

	if message, ok := config["message"]; ok {
		if msg, ok := message.(string); ok {
			l.message = msg
		} else {
			return fmt.Errorf("message must be a string")
		}
	}

	return nil
}

// Execute logs the input data
func (l *LoggerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Create log message
	logMsg := l.message
	if logMsg == "" {
		// If no custom message, use the inputs data
		inputBytes, err := json.Marshal(inputs)
		if err != nil {
			return inputs, fmt.Errorf("failed to marshal input data for logging: %v", err)
		}
		logMsg = string(inputBytes)
	} else {
		// If message contains template variables, replace them with input data
		logMsg = l.replaceTemplateVariables(logMsg, inputs)
	}

	// Log based on level
	switch l.level {
	case "debug":
		log.Printf("[DEBUG] %s", logMsg)
	case "warn":
		log.Printf("[WARN] %s", logMsg)
	case "error":
		log.Printf("[ERROR] %s", logMsg)
	default:
		log.Printf("[INFO] %s", logMsg)
	}

	// Return the input data unchanged for passthrough
	return inputs, nil
}

// replaceTemplateVariables replaces template variables in the message with values from input data
func (l *LoggerNode) replaceTemplateVariables(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		var valueStr string
		if valBytes, err := json.Marshal(value); err == nil {
			valueStr = string(valBytes)
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	return result
}

// GetType returns the type of the node
func (l *LoggerNode) GetType() string {
	return l.nodeType
}

// GetID returns the unique identifier for this node instance
func (l *LoggerNode) GetID() string {
	return l.id
}

// NewLoggerNode creates a new logger node constructor for the registry
func NewLoggerNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	node := &LoggerNode{
		id:       fmt.Sprintf("logger_%d", time.Now().UnixNano()),
		nodeType: "logger",
	}

	if err := node.Initialize(config); err != nil {
		return nil, err
	}

	return node, nil
}