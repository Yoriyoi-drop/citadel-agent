package core

import (
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/citadel-agent/backend/internal/interfaces"
)

// RegisterCoreNodes registers all core backend & HTTP nodes
func RegisterCoreNodes(registry *engine.NodeRegistry) {
	// Register HTTP node
	registry.RegisterNodeType("http_request", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewHTTPNode(config)
	})

	// Register Validator node
	registry.RegisterNodeType("validator", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewValidatorNode(config)
	})

	// Register Logger node
	registry.RegisterNodeType("logger", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewLoggerNode(config)
	})

	// Register Config Manager node
	registry.RegisterNodeType("config_manager", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewConfigManagerNode(config)
	})

	// Register UUID Generator node
	registry.RegisterNodeType("uuid_generator", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewUUIDGeneratorNode(config)
	})
}