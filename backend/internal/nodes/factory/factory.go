package factory

import (
	"fmt"

	"github.com/citadel-agent/backend/internal/nodes/base"
	"github.com/citadel-agent/backend/internal/nodes/registry"
)

// Factory creates node instances
type Factory struct {
	registry *registry.Registry
}

// NewFactory creates a new node factory
func NewFactory() *Factory {
	return &Factory{
		registry: registry.GetRegistry(),
	}
}

// Create creates a new node instance with configuration
func (f *Factory) Create(nodeID string, config map[string]interface{}) (base.Node, error) {
	// Create node instance
	node, err := f.registry.CreateInstance(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Validate configuration
	if err := node.Validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return node, nil
}

// CreateWithDefaults creates a node with default configuration
func (f *Factory) CreateWithDefaults(nodeID string) (base.Node, error) {
	reg, err := f.registry.Get(nodeID)
	if err != nil {
		return nil, err
	}

	// Build default config
	config := make(map[string]interface{})
	for _, cfg := range reg.Metadata.Config {
		if cfg.Default != nil {
			config[cfg.Name] = cfg.Default
		}
	}

	return f.Create(nodeID, config)
}

// ValidateConfig validates configuration without creating node
func (f *Factory) ValidateConfig(nodeID string, config map[string]interface{}) error {
	node, err := f.registry.CreateInstance(nodeID)
	if err != nil {
		return err
	}

	return node.Validate(config)
}

// GetMetadata returns node metadata
func (f *Factory) GetMetadata(nodeID string) (base.NodeMetadata, error) {
	reg, err := f.registry.Get(nodeID)
	if err != nil {
		return base.NodeMetadata{}, err
	}

	return reg.Metadata, nil
}

// ListAvailable returns all available node types
func (f *Factory) ListAvailable() []base.NodeMetadata {
	return f.registry.List()
}

// Global factory instance
var globalFactory *Factory

// GetFactory returns the global factory instance
func GetFactory() *Factory {
	if globalFactory == nil {
		globalFactory = NewFactory()
	}
	return globalFactory
}
