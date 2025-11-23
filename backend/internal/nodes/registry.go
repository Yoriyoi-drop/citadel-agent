package nodes

import (
	"fmt"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/nodes/ai"
	"github.com/citadel-agent/backend/internal/nodes/database"
	"github.com/citadel-agent/backend/internal/nodes/http"
	"github.com/citadel-agent/backend/internal/nodes/integration"
	"github.com/citadel-agent/backend/internal/nodes/security"
	"github.com/citadel-agent/backend/internal/nodes/utility"
)

// NodeType represents different types of nodes
type NodeType string

const (
	// Core HTTP Node Types
	HTTPRequestNodeType NodeType = "http_request"

	// Database Node Types
	DatabaseQueryNodeType NodeType = "database_query"

	// AI Node Types
	TextGeneratorNodeType NodeType = "text_generator"

	// Utility Node Types
	DataTransformerNodeType NodeType = "data_transformer"

	// Security Node Types
	EncryptionNodeType NodeType = "encryption"

	// Integration Node Types
	NotificationNodeType NodeType = "notification"
)

// NodeFactory creates node instances based on type
type NodeFactory struct {
	registry map[NodeType]NodeConstructor
}

// NodeConstructor is a function that creates a new node instance
type NodeConstructor func(config map[string]interface{}) (interfaces.NodeInstance, error)

// Global node factory
var globalNodeFactory *NodeFactory

// GetNodeFactory returns the singleton instance of NodeFactory
func GetNodeFactory() *NodeFactory {
	if globalNodeFactory == nil {
		globalNodeFactory = NewNodeFactory()
	}
	return globalNodeFactory
}

// NewNodeFactory creates a new node factory with all node types registered
func NewNodeFactory() *NodeFactory {
	nf := &NodeFactory{
		registry: make(map[NodeType]NodeConstructor),
	}

	// Register all node types
	nf.registerNodeType(HTTPRequestNodeType, http.NewHTTPRequestNode)
	nf.registerNodeType(DatabaseQueryNodeType, database.NewDatabaseNode)
	nf.registerNodeType(TextGeneratorNodeType, ai.NewTextGeneratorNode)
	nf.registerNodeType(DataTransformerNodeType, utility.NewTransformerNode)
	nf.registerNodeType(EncryptionNodeType, security.NewEncryptionNode)
	nf.registerNodeType(NotificationNodeType, integration.NewNotificationNode)

	return nf
}

// RegisterNodeType registers a new node type with its constructor (internal version)
func (nf *NodeFactory) registerNodeType(nodeType NodeType, constructor NodeConstructor) {
	nf.registry[nodeType] = constructor
}

// RegisterNodeType implements interfaces.NodeFactory (string version)
func (nf *NodeFactory) RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (interfaces.NodeInstance, error)) error {
	nf.registry[NodeType(nodeType)] = constructor
	return nil
}

// CreateNode creates a new node instance based on the node type and configuration
func (nf *NodeFactory) CreateNode(nodeType NodeType, config map[string]interface{}) (interfaces.NodeInstance, error) {
	constructor, exists := nf.registry[nodeType]
	if !exists {
		return nil, fmt.Errorf("node type %s is not registered", nodeType)
	}

	return constructor(config)
}

// CreateInstance implements interfaces.NodeFactory
func (nf *NodeFactory) CreateInstance(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	return nf.CreateNode(NodeType(nodeType), config)
}

// ListNodeTypes returns all registered node types as strings (implements interfaces.NodeFactory)
func (nf *NodeFactory) ListNodeTypes() []string {
	types := make([]string, 0, len(nf.registry))
	for nodeType := range nf.registry {
		types = append(types, string(nodeType))
	}
	return types
}

// IsNodeTypeRegistered checks if a node type is registered
func (nf *NodeFactory) IsNodeTypeRegistered(nodeType NodeType) bool {
	_, exists := nf.registry[nodeType]
	return exists
}

// GetNodeDefinition implements interfaces.NodeFactory
func (nf *NodeFactory) GetNodeDefinition(nodeType string) (*interfaces.NodeDefinition, bool) {
	// TODO: Implement node definitions
	return nil, false
}
