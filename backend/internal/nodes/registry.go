// backend/internal/nodes/registry.go
package nodes

import (
	"fmt"
	"sync"

	"citadel-agent/backend/internal/workflow/core/engine"
	"citadel-agent/backend/internal/nodes/core"
	"citadel-agent/backend/internal/nodes/database"
	"citadel-agent/backend/internal/nodes/workflow"
	"citadel-agent/backend/internal/nodes/security"
	"citadel-agent/backend/internal/nodes/debug"
	"citadel-agent/backend/internal/nodes/utilities"
	"citadel-agent/backend/internal/nodes/basic"
	"citadel-agent/backend/internal/nodes/plugins"
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

	// Core Backend & HTTP Node Types
	ValidatorNodeType      NodeType = "validator"
	LoggerNodeType         NodeType = "logger"
	ConfigManagerNodeType  NodeType = "config_manager"
	UUIDGeneratorNodeType  NodeType = "uuid_generator"

	// Database & ORM Node Types
	GORMDatabaseNodeType   NodeType = "gorm_database"
	BunDatabaseNodeType    NodeType = "bun_database"
	EntDatabaseNodeType    NodeType = "ent_database"
	SQLCDatabaseNodeType   NodeType = "sqlc_database"
	MigrateDatabaseNodeType NodeType = "migrate_database"

	// Workflow & Scheduling Node Types
	CronSchedulerNodeType  NodeType = "cron_scheduler"
	TaskQueueNodeType      NodeType = "task_queue"
	JobSchedulerNodeType   NodeType = "job_scheduler"
	WorkerPoolNodeType     NodeType = "worker_pool"
	CircuitBreakerNodeType NodeType = "circuit_breaker"

	// Security Node Types
	FirewallManagerNodeType      NodeType = "firewall_manager"
	EncryptionNodeType           NodeType = "encryption"
	AccessControlNodeType        NodeType = "access_control"
	APIKeyManagerNodeType        NodeType = "api_key_manager"
	JWTHandlerNodeType           NodeType = "jwt_handler"
	OAuth2ProviderNodeType       NodeType = "oauth2_provider"
	SecurityOperationNodeType    NodeType = "security_operation"

	// Debug & Logging Node Types
	DebugNodeType                NodeType = "debug"
	LoggingNodeType              NodeType = "logging"

	// Utility Node Types
	UtilityNodeType              NodeType = "utility"

	// Basic Node Types
	BasicNodeType                NodeType = "basic"

	// Plugin Node Types
	PluginNodeType               NodeType = "plugin"
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

	// Register Core Backend & HTTP nodes
	nf.RegisterNodeType(ValidatorNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return core.NewValidatorNode(config)
	})

	nf.RegisterNodeType(LoggerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return core.NewLoggerNode(config)
	})

	nf.RegisterNodeType(ConfigManagerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return core.NewConfigManagerNode(config)
	})

	nf.RegisterNodeType(UUIDGeneratorNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return core.NewUUIDGeneratorNode(config)
	})

	// Register Database & ORM nodes
	nf.RegisterNodeType(GORMDatabaseNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return database.NewGORMDatabaseNode(config)
	})

	nf.RegisterNodeType(BunDatabaseNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return database.NewBunDatabaseNode(config)
	})

	nf.RegisterNodeType(EntDatabaseNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return database.NewEntDatabaseNode(config)
	})

	nf.RegisterNodeType(SQLCDatabaseNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return database.NewSQLCNode(config)
	})

	nf.RegisterNodeType(MigrateDatabaseNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return database.NewMigrateNode(config)
	})

	// Register Workflow & Scheduling nodes
	nf.RegisterNodeType(CronSchedulerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return workflow.NewCronSchedulerNode(config)
	})

	nf.RegisterNodeType(TaskQueueNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return workflow.NewTaskQueueNode(config)
	})

	nf.RegisterNodeType(JobSchedulerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return workflow.NewJobSchedulerNode(config)
	})

	nf.RegisterNodeType(WorkerPoolNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return workflow.NewWorkerPoolNode(config)
	})

	nf.RegisterNodeType(CircuitBreakerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return workflow.NewCircuitBreakerNode(config)
	})

	// Register Security nodes
	nf.RegisterNodeType(FirewallManagerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.FirewallManagerNodeFromConfig(config)
	})

	nf.RegisterNodeType(EncryptionNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.EncryptionNodeFromConfig(config)
	})

	nf.RegisterNodeType(AccessControlNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.AccessControlNodeFromConfig(config)
	})

	nf.RegisterNodeType(APIKeyManagerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.APIKeyManagerNodeFromConfig(config)
	})

	nf.RegisterNodeType(JWTHandlerNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.JWTHandlerNodeFromConfig(config)
	})

	nf.RegisterNodeType(OAuth2ProviderNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.OAuth2ProviderNodeFromConfig(config)
	})

	nf.RegisterNodeType(SecurityOperationNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return security.NewSecurityNodeFromConfig(config)
	})

	// Register Debug & Logging nodes
	nf.RegisterNodeType(DebugNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return debug.DebugNodeFromConfig(config)
	})

	nf.RegisterNodeType(LoggingNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return core.NewLoggerNode(config)
	})

	// Register Utility nodes
	nf.RegisterNodeType(UtilityNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return utilities.UtilityNodeFromConfig(config)
	})

	// Register Basic nodes
	nf.RegisterNodeType(BasicNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return basic.BasicNodeFromConfig(config)
	})

	// Register Plugin nodes
	nf.RegisterNodeType(PluginNodeType, func(config map[string]interface{}) (engine.NodeInstance, error) {
		return plugins.PluginNodeFromConfig(config)
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