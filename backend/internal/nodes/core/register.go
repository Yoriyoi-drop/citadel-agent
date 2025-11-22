package core

import (
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/citadel-agent/backend/internal/interfaces"
)

// RegisterCoreNodes registers all core backend & HTTP nodes
func RegisterCoreNodes(registry *engine.NodeRegistry) error {
	// Register HTTP node
	err := registry.RegisterNodeType("http_request", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewHTTPNode(config)
	})
	if err != nil {
		return err
	}

	// Register Validator node
	err = registry.RegisterNodeType("validator", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewValidatorNode(config)
	})
	if err != nil {
		return err
	}

	// Register Logger node
	err = registry.RegisterNodeType("logger", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewLoggerNode(config)
	})
	if err != nil {
		return err
	}

	// Register Config Manager node
	err = registry.RegisterNodeType("config_manager", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewConfigManagerNode(config)
	})
	if err != nil {
		return err
	}

	// Register UUID Generator node
	err = registry.RegisterNodeType("uuid_generator", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewUUIDGeneratorNode(config)
	})
	if err != nil {
		return err
	}

	return nil
}