package registry

import (
	"fmt"
	"sync"

	"github.com/citadel-agent/backend/internal/nodes/base"
)

// NodeCreator is a function that creates a new node instance
type NodeCreator func() base.Node

// Registry manages all registered node types
type Registry struct {
	nodes map[string]NodeRegistration
	mu    sync.RWMutex
}

// NodeRegistration contains node registration info
type NodeRegistration struct {
	Metadata base.NodeMetadata
	Creator  NodeCreator
}

// Global registry instance
var globalRegistry *Registry
var once sync.Once

// GetRegistry returns the global registry instance
func GetRegistry() *Registry {
	once.Do(func() {
		globalRegistry = &Registry{
			nodes: make(map[string]NodeRegistration),
		}
	})
	return globalRegistry
}

// Register registers a new node type
func (r *Registry) Register(nodeID string, creator NodeCreator, metadata base.NodeMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; exists {
		return fmt.Errorf("node %s already registered", nodeID)
	}

	r.nodes[nodeID] = NodeRegistration{
		Metadata: metadata,
		Creator:  creator,
	}

	return nil
}

// Unregister removes a node type from registry
func (r *Registry) Unregister(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(r.nodes, nodeID)
	return nil
}

// Get retrieves a node registration
func (r *Registry) Get(nodeID string) (NodeRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, exists := r.nodes[nodeID]
	if !exists {
		return NodeRegistration{}, fmt.Errorf("node %s not found", nodeID)
	}

	return reg, nil
}

// CreateInstance creates a new instance of a node
func (r *Registry) CreateInstance(nodeID string) (base.Node, error) {
	reg, err := r.Get(nodeID)
	if err != nil {
		return nil, err
	}

	return reg.Creator(), nil
}

// List returns all registered nodes
func (r *Registry) List() []base.NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]base.NodeMetadata, 0, len(r.nodes))
	for _, reg := range r.nodes {
		result = append(result, reg.Metadata)
	}

	return result
}

// ListByCategory returns nodes filtered by category
func (r *Registry) ListByCategory(category string) []base.NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]base.NodeMetadata, 0)
	for _, reg := range r.nodes {
		if reg.Metadata.Category == category {
			result = append(result, reg.Metadata)
		}
	}

	return result
}

// Search searches nodes by name, description, or tags
func (r *Registry) Search(query string) []base.NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]base.NodeMetadata, 0)
	for _, reg := range r.nodes {
		// Simple search - can be enhanced with fuzzy matching
		if contains(reg.Metadata.Name, query) ||
			contains(reg.Metadata.Description, query) ||
			containsTag(reg.Metadata.Tags, query) {
			result = append(result, reg.Metadata)
		}
	}

	return result
}

// Count returns the number of registered nodes
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.nodes)
}

// Categories returns all unique categories
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make(map[string]bool)
	for _, reg := range r.nodes {
		categories[reg.Metadata.Category] = true
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}

	return result
}

// Helper functions

func contains(s, substr string) bool {
	// Case-insensitive contains
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			(len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if contains(tag, query) {
			return true
		}
	}
	return false
}

// RegisterNode is a convenience function to register a node
func RegisterNode(nodeID string, creator NodeCreator, metadata base.NodeMetadata) error {
	return GetRegistry().Register(nodeID, creator, metadata)
}

// GetNode is a convenience function to get a node
func GetNode(nodeID string) (base.Node, error) {
	return GetRegistry().CreateInstance(nodeID)
}

// ListNodes is a convenience function to list all nodes
func ListNodes() []base.NodeMetadata {
	return GetRegistry().List()
}
