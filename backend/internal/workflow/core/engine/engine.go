// workflow/core/engine/engine.go
package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/citadel-agent/backend/internal/interfaces"
)

// NodeRegistry manages node definitions
type NodeRegistry struct {
	nodes map[string]func(config map[string]interface{}) (interfaces.NodeInstance, error)
}

// NewNodeRegistry creates a new node registry
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		nodes: make(map[string]func(config map[string]interface{}) (interfaces.NodeInstance, error)),
	}
}

// RegisterNodeType registers a new node type with its constructor
func (r *NodeRegistry) RegisterNodeType(nodeType string, constructor func(map[string]interface{}) (interfaces.NodeInstance, error)) {
	r.nodes[nodeType] = constructor
}

// CreateInstance creates a new instance of the specified node type
func (r *NodeRegistry) CreateInstance(nodeType string, config map[string]interface{}) (interfaces.NodeInstance, error) {
	constructor, exists := r.nodes[nodeType]
	if !exists {
		return nil, fmt.Errorf("node type %s not registered", nodeType)
	}
	return constructor(config)
}

// Workflow represents a complete workflow definition
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []*Node                `json:"nodes"`
	Connections []*Connection          `json:"connections"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Node represents a single node in the workflow
type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Status      NodeStatus             `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       *string                `json:"error,omitempty"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodePending   NodeStatus = "pending"
	NodeRunning   NodeStatus = "running"
	NodeSuccess   NodeStatus = "success"
	NodeFailed    NodeStatus = "failed"
	NodeSkipped   NodeStatus = "skipped"
	NodeCancelled NodeStatus = "cancelled"
)

// Connection represents a connection between nodes
type Connection struct {
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	Port         string `json:"port,omitempty"` // For complex connections
	Condition    string `json:"condition,omitempty"` // Conditional execution
}

// Execution represents a running instance of a workflow
type Execution struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        ExecutionStatus        `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	NodeResults   map[string]*NodeResult `json:"node_results"`
	Error         *string                `json:"error,omitempty"`
	TriggeredBy   string                 `json:"triggered_by"`
	TriggerParams map[string]interface{} `json:"trigger_params,omitempty"`
}

// NodeResult represents the result of a node execution
type NodeResult struct {
	NodeID      string                 `json:"node_id"`
	Status      NodeStatus             `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Error       *string                `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	ExecutionTime time.Duration        `json:"execution_time"`
}

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	ExecutionPending   ExecutionStatus = "pending"
	ExecutionRunning   ExecutionStatus = "running"
	ExecutionSuccess   ExecutionStatus = "success"
	ExecutionFailed    ExecutionStatus = "failed"
	ExecutionCancelled ExecutionStatus = "cancelled"
	ExecutionPaused    ExecutionStatus = "paused"
)

// Engine represents the workflow engine
type Engine struct {
	mutex        sync.RWMutex
	executions   map[string]*Execution
	storage      Storage
	scheduler    *Scheduler
	nodeRegistry *NodeRegistry
	parallelism  int
	logger       Logger
	securityMgr  *SecurityManager  // Added security manager
	monitoring   *MonitoringSystem // Added monitoring system
	aiAgentMgr   *AIManager        // Added AI agent manager
}

// SecurityManager handles security aspects of workflow execution
type SecurityManager struct {
	runtimeValidator *RuntimeValidator
	permissionChecker *PermissionChecker
	resourceLimiter  *ResourceLimiter
}

// MonitoringSystem handles workflow monitoring and observability
type MonitoringSystem struct {
	metricsCollector *MetricsCollector
	tracer          *TraceCollector
	alerter         *Alerter
}

// Storage interface for persistence
type Storage interface {
	CreateExecution(execution *Execution) error
	UpdateExecution(execution *Execution) error
	GetExecution(id string) (*Execution, error)
	CreateNodeResult(result *NodeResult) error
	UpdateNodeResult(result *NodeResult) error
	GetNodeResult(executionID, nodeID string) (*NodeResult, error)
	ListExecutions(workflowID string) ([]*Execution, error)
	GetExecutionHistory(workflowID string, limit, offset int) ([]*Execution, error)  // Added execution history
	GetNodeExecutionStats(nodeID string) (*NodeExecutionStats, error)               // Added node stats
	SaveWorkflowSnapshot(workflow *Workflow) error                                 // Added workflow snapshot
	RestoreWorkflowSnapshot(workflowID string) (*Workflow, error)                  // Added workflow restore
	GetActiveExecutions() ([]*Execution, error)                                    // Added active executions
	UpdateExecutionWithVariables(executionID string, variables map[string]interface{}) error // Added variable updates
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Config for the engine
type Config struct {
	Parallelism int
	Logger      Logger
	Storage     Storage
}

// NodeExecutionStats provides statistics about node execution
type NodeExecutionStats struct {
	NodeID          string        `json:"node_id"`
	TotalExecutions int           `json:"total_executions"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	LastExecutedAt  time.Time     `json:"last_executed_at"`
}

// NewEngine creates a new workflow engine
func NewEngine(config *Config) *Engine {
	if config.Parallelism <= 0 {
		config.Parallelism = 10 // default parallelism
	}

	// Initialize new components
	securityMgr := &SecurityManager{
		runtimeValidator: NewRuntimeValidator(),
		permissionChecker: NewPermissionChecker(),
		resourceLimiter:  NewResourceLimiter(),
	}

	monitoring := &MonitoringSystem{
		metricsCollector: NewMetricsCollector(),
		tracer:          NewTraceCollector(),
		alerter:         NewAlerter(),
	}

	aiAgentMgr := NewAIManager() // Assuming this exists or will be created

	engine := &Engine{
		executions:   make(map[string]*Execution),
		storage:      config.Storage,
		scheduler:    NewScheduler(),
		nodeRegistry: NewNodeRegistry(),
		parallelism:  config.Parallelism,
		logger:       config.Logger,
		securityMgr:  securityMgr,
		monitoring:   monitoring,
		aiAgentMgr:   aiAgentMgr,
	}

	// Initialize basic node types
	engine.nodeRegistry.RegisterNodeType("http_request", NewHTTPRequestNode)
	engine.nodeRegistry.RegisterNodeType("condition", NewConditionNode)
	engine.nodeRegistry.RegisterNodeType("delay", NewDelayNode)
	engine.nodeRegistry.RegisterNodeType("database_query", NewDatabaseQueryNode)
	engine.nodeRegistry.RegisterNodeType("script_execution", NewScriptExecutionNode)

	// Added advanced node types
	engine.nodeRegistry.RegisterNodeType("ai_agent", NewAIAgentNode)
	engine.nodeRegistry.RegisterNodeType("data_transformer", NewDataTransformerNode)
	engine.nodeRegistry.RegisterNodeType("notification", NewNotificationNode)
	engine.nodeRegistry.RegisterNodeType("loop", NewLoopNode)
	engine.nodeRegistry.RegisterNodeType("error_handler", NewErrorHandlerNode)

	return engine
}

// ExecuteWorkflow starts the execution of a workflow
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *Workflow, triggerParams map[string]interface{}) (string, error) {
	executionID := uuid.New().String()

	e.logger.Info("Starting execution %s for workflow %s", executionID, workflow.ID)

	// Create execution instance
	execution := &Execution{
		ID:            executionID,
		WorkflowID:    workflow.ID,
		Status:        ExecutionPending,
		StartedAt:     time.Now(),
		Variables:     make(map[string]interface{}),
		NodeResults:   make(map[string]*NodeResult),
		TriggeredBy:   "manual", // This can be from scheduler, API, etc
		TriggerParams: triggerParams,
	}

	// Persist execution
	if err := e.storage.CreateExecution(execution); err != nil {
		e.logger.Error("Failed to create execution: %v", err)
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	// Update in-memory cache
	e.mutex.Lock()
	e.executions[executionID] = execution
	e.mutex.Unlock()

	// Start execution asynchronously
	go e.runExecution(ctx, execution, workflow)

	return executionID, nil
}

// runExecution runs the actual execution of the workflow
func (e *Engine) runExecution(ctx context.Context, execution *Execution, workflow *Workflow) {
	e.mutex.Lock()
	execution.Status = ExecutionRunning
	e.mutex.Unlock()

	// Update execution status in storage
	if err := e.storage.UpdateExecution(execution); err != nil {
		e.logger.Error("Failed to update execution status: %v", err)
		e.failExecution(execution, err.Error())
		return
	}

	// Execute nodes with dependency resolution
	if err := e.executeNodes(ctx, execution, workflow); err != nil {
		e.logger.Error("Execution failed: %v", err)
		e.failExecution(execution, err.Error())
		return
	}

	// Complete execution
	e.completeExecution(execution)
}

// executeNodes executes workflow nodes with dependency resolution
func (e *Engine) executeNodes(ctx context.Context, execution *Execution, workflow *Workflow) error {
	// Build dependency graph
	graph, err := e.buildDependencyGraph(workflow)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Execute nodes in parallel respecting dependencies
	semaphore := make(chan struct{}, e.parallelism)
	errChan := make(chan error, len(workflow.Nodes))
	doneChan := make(chan string, len(workflow.Nodes))

	// Track which nodes are ready to execute
	readyNodes := make(map[string]bool)
	for _, node := range workflow.Nodes {
		readyNodes[node.ID] = len(node.Dependencies) == 0
	}

	for {
		// Find ready nodes
		var nodeToExecute *Node
		for _, node := range workflow.Nodes {
			if readyNodes[node.ID] && !e.isNodeExecuted(execution, node.ID) {
				nodeToExecute = node
				break
			}
		}

		if nodeToExecute == nil {
			// Check if all nodes are completed
			if e.allNodesCompleted(execution, workflow.Nodes) {
				break
			}

			// No ready nodes but not all completed - check for cycles or errors
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Mark node as not ready anymore
		readyNodes[nodeToExecute.ID] = false

		// Execute node in goroutine
		go func(node *Node) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := e.executeSingleNode(ctx, execution, node, workflow); err != nil {
				errChan <- fmt.Errorf("node %s execution failed: %w", node.ID, err)
				return
			}

			doneChan <- node.ID
		}(nodeToExecute)

		// Update ready nodes when a node completes
		go func() {
			nodeID := <-doneChan

			// Mark node as executed
			e.markNodeAsExecuted(execution, nodeID)

			// Update ready nodes based on dependencies
			for _, node := range workflow.Nodes {
				if !e.isNodeExecuted(execution, node.ID) && !readyNodes[node.ID] {
					if e.dependenciesSatisfied(execution, node.Dependencies) {
						readyNodes[node.ID] = true
					}
				}
			}
		}()
	}

	// Wait for all goroutines to complete or catch errors
	completedCount := 0
	for completedCount < len(workflow.Nodes) {
		select {
		case err := <-errChan:
			return err
		case <-doneChan:
			completedCount++
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// buildDependencyGraph builds dependency graph for the workflow
func (e *Engine) buildDependencyGraph(workflow *Workflow) (map[string][]string, error) {
	graph := make(map[string][]string)

	// Create adjacency list from connections
	for _, conn := range workflow.Connections {
		graph[conn.SourceNodeID] = append(graph[conn.SourceNodeID], conn.TargetNodeID)
	}

	return graph, nil
}

// executeSingleNode executes a single node
func (e *Engine) executeSingleNode(ctx context.Context, execution *Execution, node *Node, workflow *Workflow) error {
	startTime := time.Now()

	e.logger.Info("Executing node %s for execution %s", node.ID, execution.ID)

	// Create node result
	nodeResult := &NodeResult{
		NodeID:    node.ID,
		Status:    NodeRunning,
		StartedAt: startTime,
	}

	// Create node instance
	nodeInstance, err := e.nodeRegistry.CreateInstance(node.Type, node.Config)
	if err != nil {
		return fmt.Errorf("failed to create node instance: %w", err)
	}

	// Prepare inputs by evaluating expressions and dependencies
	inputs, err := e.prepareNodeInputs(execution, node, workflow)
	if err != nil {
		return fmt.Errorf("failed to prepare inputs: %w", err)
	}

	// Execute the node
	output, err := nodeInstance.Execute(ctx, inputs)

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Update node result
	nodeResult.CompletedAt = time.Now()
	nodeResult.ExecutionTime = executionTime

	if err != nil {
		nodeResult.Status = NodeFailed
		errStr := err.Error()
		nodeResult.Error = &errStr
	} else {
		nodeResult.Status = NodeSuccess
		nodeResult.Output = output
	}

	// Save node result to storage
	if err := e.storage.CreateNodeResult(nodeResult); err != nil {
		return fmt.Errorf("failed to save node result: %w", err)
	}

	// Update execution with node result
	e.mutex.Lock()
	execution.NodeResults[node.ID] = nodeResult
	e.mutex.Unlock()

	// Update execution in storage
	if err := e.storage.UpdateExecution(execution); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	if err != nil {
		return err
	}

	e.logger.Info("Node %s completed with status %s for execution %s", node.ID, nodeResult.Status, execution.ID)
	return nil
}

// prepareNodeInputs prepares inputs for a node based on dependencies
func (e *Engine) prepareNodeInputs(execution *Execution, node *Node, workflow *Workflow) (map[string]interface{}, error) {
	inputs := make(map[string]interface{})

	// Copy the original inputs
	for k, v := range node.Inputs {
		inputs[k] = v
	}

	// Add outputs from dependent nodes
	for _, depNodeID := range node.Dependencies {
		result, exists := execution.NodeResults[depNodeID]
		if !exists {
			return nil, fmt.Errorf("dependency node %s not executed", depNodeID)
		}

		if result.Status != NodeSuccess {
			return nil, fmt.Errorf("dependency node %s did not succeed", depNodeID)
		}

		// Add dependency outputs to inputs
		for k, v := range result.Output {
			// Use namespace to avoid conflicts
			inputs[fmt.Sprintf("%s_%s", depNodeID, k)] = v
		}
	}

	return inputs, nil
}

// isNodeExecuted checks if a node has been executed
func (e *Engine) isNodeExecuted(execution *Execution, nodeID string) bool {
	_, exists := execution.NodeResults[nodeID]
	return exists
}

// markNodeAsExecuted marks a node as executed
func (e *Engine) markNodeAsExecuted(execution *Execution, nodeID string) {
	// This will be called after successful node execution
	// The node result should already be in execution.NodeResults
}

// dependenciesSatisfied checks if all dependencies for a node are satisfied
func (e *Engine) dependenciesSatisfied(execution *Execution, dependencies []string) bool {
	for _, depID := range dependencies {
		result, exists := execution.NodeResults[depID]
		if !exists || result.Status != NodeSuccess {
			return false
		}
	}
	return true
}

// allNodesCompleted checks if all nodes in workflow are completed
func (e *Engine) allNodesCompleted(execution *Execution, nodes []*Node) bool {
	for _, node := range nodes {
		if !e.isNodeExecuted(execution, node.ID) {
			return false
		}
	}
	return true
}

// failExecution marks execution as failed
func (e *Engine) failExecution(execution *Execution, errorMsg string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execution.Status = ExecutionFailed
	execution.Error = &errorMsg
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Update in storage
	if err := e.storage.UpdateExecution(execution); err != nil {
		e.logger.Error("Failed to update failed execution: %v", err)
	}
}

// completeExecution marks execution as completed successfully
func (e *Engine) completeExecution(execution *Execution) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execution.Status = ExecutionSuccess
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Update in storage
	if err := e.storage.UpdateExecution(execution); err != nil {
		e.logger.Error("Failed to update completed execution: %v", err)
	}

	e.logger.Info("Execution %s completed successfully", execution.ID)
}

// GetExecution retrieves an execution by ID
func (e *Engine) GetExecution(id string) (*Execution, error) {
	e.mutex.RLock()
	execution, exists := e.executions[id]
	e.mutex.RUnlock()

	if exists {
		return execution, nil
	}

	// Try to get from storage
	return e.storage.GetExecution(id)
}

// NewHTTPRequestNode creates a new HTTP request node instance
func NewHTTPRequestNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// This would return an HTTP request node implementation
	// For now, returning a simple struct that satisfies the interface
	return &HTTPNode{Config: config}, nil
}

// NewConditionNode creates a new condition node instance
func NewConditionNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &ConditionNode{Config: config}, nil
}

// NewDelayNode creates a new delay node instance
func NewDelayNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &DelayNode{Config: config}, nil
}

// NewDatabaseQueryNode creates a new database query node instance
func NewDatabaseQueryNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &DatabaseNode{Config: config}, nil
}

// NewScriptExecutionNode creates a new script execution node instance
func NewScriptExecutionNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	return &ScriptNode{Config: config}, nil
}

// HTTPNode represents an HTTP request node
type HTTPNode struct {
	Config map[string]interface{}
}

func (n *HTTPNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate HTTP request execution
	result := map[string]interface{}{
		"result":    "HTTP request executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// ConditionNode represents a condition node
type ConditionNode struct {
	Config map[string]interface{}
}

func (n *ConditionNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate condition evaluation
	result := map[string]interface{}{
		"result":    "Condition evaluated",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// DelayNode represents a delay/sleep node
type DelayNode struct {
	Config map[string]interface{}
}

func (n *DelayNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate delay
	delaySecs := 1.0
	if secs, ok := n.Config["seconds"].(float64); ok {
		delaySecs = secs
	}

	time.Sleep(time.Duration(delaySecs) * time.Second)

	result := map[string]interface{}{
		"result":       "Delay completed",
		"delayed_by":   delaySecs,
		"config":       n.Config,
		"inputs":       inputs,
		"timestamp":    time.Now().Unix(),
	}
	return result, nil
}

// DatabaseNode represents a database query node
type DatabaseNode struct {
	Config map[string]interface{}
}

func (n *DatabaseNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate database query
	result := map[string]interface{}{
		"result":    "Database query executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}

// ScriptNode represents a script execution node
type ScriptNode struct {
	Config map[string]interface{}
}

func (n *ScriptNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate script execution
	result := map[string]interface{}{
		"result":    "Script executed",
		"config":    n.Config,
		"inputs":    inputs,
		"timestamp": time.Now().Unix(),
	}
	return result, nil
}