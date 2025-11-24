package engine

import (
	"fmt"
	"sync"

	"github.com/citadel-agent/backend/internal/workflow/core/types"
)

// NodeTypeRegistryImpl implements NodeTypeRegistry
type NodeTypeRegistryImpl struct {
	mu      sync.RWMutex
	nodeTypes map[string]func() types.NodeInstance
	metadata map[string]types.NodeMetadata
}

// NewNodeTypeRegistry creates a new node type registry
func NewNodeTypeRegistry() *NodeTypeRegistryImpl {
	return &NodeTypeRegistryImpl{
		nodeTypes: make(map[string]func() types.NodeInstance),
		metadata:  make(map[string]types.NodeMetadata),
	}
}

// RegisterNodeType registers a new node type with the registry
func (r *NodeTypeRegistryImpl) RegisterNodeType(id string, creator func() types.NodeInstance, metadata types.NodeMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodeTypes[id]; exists {
		return fmt.Errorf("node type %s already registered", id)
	}

	r.nodeTypes[id] = creator
	r.metadata[id] = metadata
	return nil
}

// GetNodeType returns the creator function for a given node type
func (r *NodeTypeRegistryImpl) GetNodeType(id string) (func() types.NodeInstance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	creator, exists := r.nodeTypes[id]
	return creator, exists
}

// GetNodeMetadata returns the metadata for a given node type
func (r *NodeTypeRegistryImpl) GetNodeMetadata(id string) (types.NodeMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.metadata[id]
	return metadata, exists
}

// ListNodeTypes returns all registered node types
func (r *NodeTypeRegistryImpl) ListNodeTypes() []types.NodeMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]types.NodeMetadata, 0, len(r.metadata))
	for _, m := range r.metadata {
		metadata = append(metadata, m)
	}
	return metadata
}

// Global registry instance
var globalRegistry = NewNodeTypeRegistry()

// Global functions to access the registry
func RegisterNodeType(id string, creator func() types.NodeInstance, metadata types.NodeMetadata) error {
	return globalRegistry.RegisterNodeType(id, creator, metadata)
}

func GetNodeType(id string) (func() types.NodeInstance, bool) {
	return globalRegistry.GetNodeType(id)
}

func GetNodeMetadata(id string) (types.NodeMetadata, bool) {
	return globalRegistry.GetNodeMetadata(id)
}

func ListNodeTypes() []types.NodeMetadata {
	return globalRegistry.ListNodeTypes()
}