package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/google/uuid"
)

// PluginAwareEngine wraps the original engine to support plugin nodes
type PluginAwareEngine struct {
	baseEngine    *engine.Engine
	pluginManager *NodeManager
	registry      *PluginAwareNodeRegistry
	mutex         sync.RWMutex
}

// NewPluginAwareEngine creates a new engine that supports both local and plugin nodes
func NewPluginAwareEngine(baseEngine *engine.Engine, pluginManager *NodeManager) *PluginAwareEngine {
	pluginRegistry := NewPluginAwareNodeRegistry(pluginManager)
	
	return &PluginAwareEngine{
		baseEngine:    baseEngine,
		pluginManager: pluginManager,
		registry:      pluginRegistry,
	}
}

// RegisterLocalNodeType registers a local node type (same as before)
func (p *PluginAwareEngine) RegisterLocalNodeType(nodeType string, constructor func(map[string]interface{}) (interfaces.NodeInstance, error)) {
	// Register with the base engine too if needed
	p.baseEngine.GetNodeRegistry().RegisterNodeType(nodeType, constructor)
	
	// Also register with our plugin-aware registry
	p.registry.RegisterNodeType(nodeType, constructor)
}

// RegisterPluginNodeType registers a plugin node type
func (p *PluginAwareEngine) RegisterPluginNodeType(pluginID string) error {
	return p.registry.RegisterPluginNode(pluginID)
}

// ExecuteWithPlugins executes a workflow with support for plugin nodes
func (p *PluginAwareEngine) ExecuteWithPlugins(ctx context.Context, workflow *engine.Workflow, triggerParams map[string]interface{}) (string, error) {
	executionID := uuid.New().String()

	p.baseEngine.GetLogger().Info("Starting execution %s for workflow %s", executionID, workflow.ID)

	// Create execution instance
	execution := &engine.Execution{
		ID:            executionID,
		WorkflowID:    workflow.ID,
		Status:        engine.ExecutionPending,
		StartedAt:     time.Now(),
		Variables:     make(map[string]interface{}),
		NodeResults:   make(map[string]*engine.NodeResult),
		TriggeredBy:   "manual", // This can be from scheduler, API, etc
		TriggerParams: triggerParams,
	}

	// Persist execution
	if err := p.baseEngine.GetStorage().CreateExecution(execution); err != nil {
		p.baseEngine.GetLogger().Error("Failed to create execution: %v", err)
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	// Update in-memory cache
	p.mutex.Lock()
	// We'll need to use the base engine's methods for execution management
	p.mutex.Unlock()

	// Start execution asynchronously
	go p.runExecutionWithPlugins(ctx, execution, workflow)

	return executionID, nil
}

// runExecutionWithPlugins runs the actual execution of the workflow with plugin support
func (p *PluginAwareEngine) runExecutionWithPlugins(ctx context.Context, execution *engine.Execution, workflow *engine.Workflow) {
	p.mutex.Lock()
	execution.Status = engine.ExecutionRunning
	p.mutex.Unlock()

	// Update execution status in storage
	if err := p.baseEngine.GetStorage().UpdateExecution(execution); err != nil {
		p.baseEngine.GetLogger().Error("Failed to update execution status: %v", err)
		p.failExecution(execution, err.Error())
		return
	}

	// Execute nodes with dependency resolution and plugin support
	if err := p.executeNodesWithPlugins(ctx, execution, workflow); err != nil {
		p.baseEngine.GetLogger().Error("Execution failed: %v", err)
		p.failExecution(execution, err.Error())
		return
	}

	// Complete execution
	p.completeExecution(execution)
}

// executeNodesWithPlugins executes workflow nodes with dependency resolution and plugin support
func (p *PluginAwareEngine) executeNodesWithPlugins(ctx context.Context, execution *engine.Execution, workflow *engine.Workflow) error {
	// Build dependency graph
	graph, err := p.buildDependencyGraph(workflow)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Execute nodes in parallel respecting dependencies
	semaphore := make(chan struct{}, p.baseEngine.GetParallelism())
	errChan := make(chan error, len(workflow.Nodes))
	doneChan := make(chan string, len(workflow.Nodes))

	// Track which nodes are ready to execute
	readyNodes := make(map[string]bool)
	for _, node := range workflow.Nodes {
		readyNodes[node.ID] = len(node.Dependencies) == 0
	}

	for {
		// Find ready nodes
		var nodeToExecute *engine.Node
		for _, node := range workflow.Nodes {
			if readyNodes[node.ID] && !p.isNodeExecuted(execution, node.ID) {
				nodeToExecute = node
				break
			}
		}

		if nodeToExecute == nil {
			// Check if all nodes are completed
			if p.allNodesCompleted(execution, workflow.Nodes) {
				break
			}

			// No ready nodes but not all completed - check for cycles or errors
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Mark node as not ready anymore
		readyNodes[nodeToExecute.ID] = false

		// Execute node in goroutine
		go func(node *engine.Node) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := p.executeSingleNodeWithPlugins(ctx, execution, node, workflow); err != nil {
				errChan <- fmt.Errorf("node %s execution failed: %w", node.ID, err)
				return
			}

			doneChan <- node.ID
		}(nodeToExecute)

		// Update ready nodes when a node completes
		go func() {
			nodeID := <-doneChan

			// Mark node as executed
			p.markNodeAsExecuted(execution, nodeID)

			// Update ready nodes based on dependencies
			for _, node := range workflow.Nodes {
				if !p.isNodeExecuted(execution, node.ID) && !readyNodes[node.ID] {
					if p.dependenciesSatisfied(execution, node.Dependencies) {
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

// executeSingleNodeWithPlugins executes a single node with plugin support
func (p *PluginAwareEngine) executeSingleNodeWithPlugins(ctx context.Context, execution *engine.Execution, node *engine.Node, workflow *engine.Workflow) error {
	startTime := time.Now()

	p.baseEngine.GetLogger().Info("Executing node %s for execution %s", node.ID, execution.ID)

	// Create node result
	nodeResult := &engine.NodeResult{
		NodeID:    node.ID,
		Status:    engine.NodeRunning,
		StartedAt: startTime,
	}

	// Create node instance using the plugin-aware registry
	nodeInstance, err := p.registry.CreateInstance(node.Type, node.Config)
	if err != nil {
		return fmt.Errorf("failed to create node instance (local or plugin): %w", err)
	}

	// Prepare inputs by evaluating expressions and dependencies
	inputs, err := p.prepareNodeInputs(execution, node, workflow)
	if err != nil {
		return fmt.Errorf("failed to prepare inputs: %w", err)
	}

	// Execute the node
	output, err := nodeInstance.Execute(ctx, inputs)

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Update node result
	completedAt := time.Now()
	nodeResult.CompletedAt = &completedAt
	nodeResult.ExecutionTime = executionTime

	if err != nil {
		nodeResult.Status = engine.NodeFailed
		errStr := err.Error()
		nodeResult.Error = &errStr
	} else {
		nodeResult.Status = engine.NodeSuccess
		nodeResult.Output = output
	}

	// Save node result to storage
	if err := p.baseEngine.GetStorage().CreateNodeResult(nodeResult); err != nil {
		return fmt.Errorf("failed to save node result: %w", err)
	}

	// Update execution with node result
	p.mutex.Lock()
	execution.NodeResults[node.ID] = nodeResult
	p.mutex.Unlock()

	// Update execution in storage
	if err := p.baseEngine.GetStorage().UpdateExecution(execution); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	if err != nil {
		return err
	}

	p.baseEngine.GetLogger().Info("Node %s completed with status %s for execution %s", node.ID, nodeResult.Status, execution.ID)
	return nil
}

// buildDependencyGraph builds dependency graph for the workflow
func (p *PluginAwareEngine) buildDependencyGraph(workflow *engine.Workflow) (map[string][]string, error) {
	graph := make(map[string][]string)

	// Create adjacency list from connections
	for _, conn := range workflow.Connections {
		graph[conn.SourceNodeID] = append(graph[conn.SourceNodeID], conn.TargetNodeID)
	}

	return graph, nil
}

// prepareNodeInputs prepares inputs for a node based on dependencies
func (p *PluginAwareEngine) prepareNodeInputs(execution *engine.Execution, node *engine.Node, workflow *engine.Workflow) (map[string]interface{}, error) {
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

		if result.Status != engine.NodeSuccess {
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

// Helper methods
func (p *PluginAwareEngine) isNodeExecuted(execution *engine.Execution, nodeID string) bool {
	_, exists := execution.NodeResults[nodeID]
	return exists
}

func (p *PluginAwareEngine) markNodeAsExecuted(execution *engine.Execution, nodeID string) {
	// This will be called after successful node execution
	// The node result should already be in execution.NodeResults
}

func (p *PluginAwareEngine) dependenciesSatisfied(execution *engine.Execution, dependencies []string) bool {
	for _, depID := range dependencies {
		result, exists := execution.NodeResults[depID]
		if !exists || result.Status != engine.NodeSuccess {
			return false
		}
	}
	return true
}

func (p *PluginAwareEngine) allNodesCompleted(execution *engine.Execution, nodes []*engine.Node) bool {
	for _, node := range nodes {
		if !p.isNodeExecuted(execution, node.ID) {
			return false
		}
	}
	return true
}

func (p *PluginAwareEngine) failExecution(execution *engine.Execution, errorMsg string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	execution.Status = engine.ExecutionFailed
	execution.Error = &errorMsg
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Update in storage
	if err := p.baseEngine.GetStorage().UpdateExecution(execution); err != nil {
		p.baseEngine.GetLogger().Error("Failed to update failed execution: %v", err)
	}
}

func (p *PluginAwareEngine) completeExecution(execution *engine.Execution) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	execution.Status = engine.ExecutionSuccess
	completedAt := time.Now()
	execution.CompletedAt = &completedAt

	// Update in storage
	if err := p.baseEngine.GetStorage().UpdateExecution(execution); err != nil {
		p.baseEngine.GetLogger().Error("Failed to update completed execution: %v", err)
	}

	p.baseEngine.GetLogger().Info("Execution %s completed successfully", execution.ID)
}

// Getters to access internal components
func (p *PluginAwareEngine) GetBaseEngine() *engine.Engine {
	return p.baseEngine
}

func (p *PluginAwareEngine) GetPluginManager() *NodeManager {
	return p.pluginManager
}

func (p *PluginAwareEngine) GetRegistry() *PluginAwareNodeRegistry {
	return p.registry
}