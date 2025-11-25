package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"citadel-agent/backend/internal/interfaces"
	"citadel-agent/backend/internal/workflow/core/types"
	"github.com/google/uuid"
)

// Engine represents the workflow engine
type Engine struct {
	mutex                 sync.RWMutex
	executions            map[string]*types.Execution
	storage               Storage
	scheduler             *Scheduler
	nodeRegistry          interfaces.NodeFactory
	parallelism           int
	logger                Logger
	securityMgr           *SecurityManager       // Added security manager
	monitoring            *MonitoringSystem      // Added monitoring system
	aiAgentMgr            AIManager              // Added AI agent manager
	retryManager          *RetryManager          // Added retry manager
	circuitBreakerManager *CircuitBreakerManager // Added circuit breaker manager
}

// SecurityManager handles security aspects of workflow execution
type SecurityManager struct {
	runtimeValidator  RuntimeValidator
	permissionChecker PermissionChecker
	resourceLimiter   ResourceLimiter
}

// MonitoringSystem handles workflow monitoring and observability
type MonitoringSystem struct {
	metricsCollector MetricsCollector
	tracer           TraceCollector
	alerter          Alerter
}

// Config for the engine
type Config struct {
	Parallelism  int
	Logger       Logger
	Storage      Storage
	NodeRegistry interfaces.NodeFactory
}

// NewEngine creates a new workflow engine
func NewEngine(config *Config) *Engine {
	if config.Parallelism <= 0 {
		config.Parallelism = 10 // default parallelism
	}

	// Initialize new components
	securityMgr := &SecurityManager{
		runtimeValidator:  NewRuntimeValidator(),
		permissionChecker: NewPermissionChecker(),
		resourceLimiter:   NewResourceLimiter(),
	}

	monitoring := &MonitoringSystem{
		metricsCollector: NewMetricsCollector(),
		tracer:           NewTraceCollector(),
		alerter:          NewAlerter(),
	}

	nodeRegistry := config.NodeRegistry
	if nodeRegistry == nil {
		nodeRegistry = interfaces.NewNodeRegistry()
	}

	// Create and register core node types
	// TODO: Re-enable when node constructors are fully implemented
	// RegisterCoreNodes(nodeRegistry)

	// Initialize managers - using nil initially, they may be initialized later
	retryManager := &RetryManager{}
	circuitBreakerManager := &CircuitBreakerManager{}

	engine := &Engine{
		executions:            make(map[string]*types.Execution),
		storage:               config.Storage,
		scheduler:             nil, // TODO: Implement scheduler
		nodeRegistry:          nodeRegistry,
		parallelism:           config.Parallelism,
		logger:                config.Logger,
		securityMgr:           securityMgr,
		monitoring:            monitoring,
		aiAgentMgr:            NewAIManager(),
		retryManager:          retryManager,
		circuitBreakerManager: circuitBreakerManager,
	}

	return engine
}

// ExecuteWorkflow executes a workflow
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *types.Workflow, triggerParams map[string]interface{}) (string, error) {
	executionID := uuid.New().String()

	execution := &types.Execution{
		ID:            executionID,
		WorkflowID:    workflow.ID,
		Status:        types.ExecutionCreated,
		StartedAt:     time.Now(),
		Variables:     make(map[string]interface{}),
		NodeResults:   make(map[string]*types.NodeResult),
		TriggeredBy:   "api",
		TriggerParams: triggerParams,
	}

	// Add trigger params to variables
	for k, v := range triggerParams {
		execution.Variables[k] = v
	}

	// Save execution to storage
	if err := e.storage.CreateExecution(execution); err != nil {
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	// Add to in-memory cache
	e.mutex.Lock()
	e.executions[executionID] = execution
	e.mutex.Unlock()

	// Execute workflow in background
	go e.runExecution(ctx, execution, workflow)

	return executionID, nil
}

// runExecution runs the actual execution
func (e *Engine) runExecution(ctx context.Context, execution *types.Execution, workflow *types.Workflow) {
	// Implementation for running execution would go here
	// This would handle dependency resolution, node execution, etc.
}

// GetExecution gets an execution by ID
func (e *Engine) GetExecution(id string) (*types.Execution, error) {
	e.mutex.RLock()
	execution, exists := e.executions[id]
	e.mutex.RUnlock()

	if exists {
		return execution, nil
	}

	// Try to get from storage
	return e.storage.GetExecution(id)
}

// RegisterCoreNodes registers all core node types
// TODO: Re-enable when node constructors are fully implemented
/*
func RegisterCoreNodes(registry interfaces.NodeFactory) {
	// Register core node constructors
	registry.RegisterNodeType("http_request", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewHTTPRequestNode(config)
	})
	registry.RegisterNodeType("database_query", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewDatabaseNode(config)
	})
	registry.RegisterNodeType("text_generator", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewTextGeneratorNode(config)
	})
	registry.RegisterNodeType("data_transformer", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewTransformerNode(config)
	})
	registry.RegisterNodeType("encryption", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewEncryptionNode(config)
	})
	registry.RegisterNodeType("notification", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return NewNotificationNode(config)
	})
	// Register more core nodes as needed
}
*/
