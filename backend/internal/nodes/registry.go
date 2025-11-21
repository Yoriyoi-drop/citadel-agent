// backend/internal/nodes/registry.go
package nodes

import (
	"fmt"
	"sync"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// NodeType represents different types of nodes
type NodeType string

const (
	HTTPRequestNodeType    NodeType = "http_request"
	ConditionNodeType      NodeType = "condition"
	DelayNodeType          NodeType = "delay"
	DatabaseQueryNodeType  NodeType = "database_query"
	ScriptExecutionNodeType NodeType = "script_execution"
	AIAgentNodeType        NodeType = "ai_agent"
	DataTransformerNodeType NodeType = "data_transformer"
	NotificationNodeType   NodeType = "notification"
	LoopNodeType           NodeType = "loop"
	ErrorHandlerNodeType   NodeType = "error_handler"
)

// NodeFactory creates node instances based on type
type NodeFactory struct {
	registry map[NodeType]NodeConstructor
	mutex    sync.RWMutex
}

// NodeConstructor is a function that creates a new node instance
type NodeConstructor func(config map[string]interface{}) (engine.NodeInstance, error)

// Global node factory
var globalNodeFactory *NodeFactory
var once sync.Once

// GetNodeFactory returns the singleton instance of NodeFactory
func GetNodeFactory() *NodeFactory {
	once.Do(func() {
		globalNodeFactory = NewNodeFactory()
	})
	return globalNodeFactory
}

// NewNodeFactory creates a new node factory with all node types registered
func NewNodeFactory() *NodeFactory {
	nf := &NodeFactory{
		registry: make(map[NodeType]NodeConstructor),
	}

	// Register all node types
	nf.RegisterNodeType(HTTPRequestNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewHTTPRequestNode(config)
	})

	nf.RegisterNodeType(ConditionNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewConditionNode(config)
	})

	nf.RegisterNodeType(DelayNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewDelayNode(config)
	})

	nf.RegisterNodeType(DatabaseQueryNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewDatabaseQueryNode(config)
	})

	nf.RegisterNodeType(ScriptExecutionNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewScriptExecutionNode(config)
	})

	nf.RegisterNodeType(AIAgentNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewAIAgentNode(config)
	})

	nf.RegisterNodeType(DataTransformerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewDataTransformerNode(config)
	})

	nf.RegisterNodeType(NotificationNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewNotificationNode(config)
	})

	nf.RegisterNodeType(LoopNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewLoopNode(config)
	})

	nf.RegisterNodeType(ErrorHandlerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return engine.NewErrorHandlerNode(config)
	})

	return nf
}

// RegisterNodeType registers a new node type with its constructor
func (nf *NodeFactory) RegisterNodeType(nodeType NodeType, constructor NodeConstructor) {
	nf.mutex.Lock()
	defer nf.mutex.Unlock()

	nf.registry[nodeType] = constructor
}

// CreateNode creates a new node instance based on the node type and configuration
func (nf *NodeFactory) CreateNode(nodeType NodeType, config map[string]interface{}) (engine.NodeInstance, error) {
	nf.mutex.RLock()
	constructor, exists := nf.registry[nodeType]
	nf.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node type %s is not registered", nodeType)
	}

	return constructor(config)
}

// ListNodeTypes returns all registered node types
func (nf *NodeFactory) ListNodeTypes() []NodeType {
	nf.mutex.RLock()
	defer nf.mutex.RUnlock()

	types := make([]NodeType, 0, len(nf.registry))
	for nodeType := range nf.registry {
		types = append(types, nodeType)
	}

	return types
}

// IsNodeTypeRegistered checks if a node type is registered
func (nf *NodeFactory) IsNodeTypeRegistered(nodeType NodeType) bool {
	nf.mutex.RLock()
	defer nf.mutex.RUnlock()

	_, exists := nf.registry[nodeType]
	return exists
}