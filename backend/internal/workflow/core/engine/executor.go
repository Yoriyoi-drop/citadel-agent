package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/types"
)

// Executor manages the execution of workflows and their nodes
type Executor struct {
	nodeRegistry interfaces.NodeFactory
	storage      Storage
	logger       Logger
	config       *ExecutorConfig
}

// ExecutorConfig holds the configuration for the executor
type ExecutorConfig struct {
	MaxConcurrentExecutions int
	MaxConcurrentNodes      int
	DefaultTimeout          time.Duration
	MaxRetries              int
	RetryDelay              time.Duration
	EnableProfiling         bool
	EnableCaching           bool
	CacheTTL                time.Duration
}

// NewExecutor creates a new workflow executor
func NewExecutor(storage Storage, logger Logger, nodeRegistry interfaces.NodeFactory, config *ExecutorConfig) *Executor {
	if config == nil {
		config = &ExecutorConfig{
			MaxConcurrentExecutions: 100,
			MaxConcurrentNodes:      50,
			DefaultTimeout:          30 * time.Second,
			MaxRetries:              3,
			RetryDelay:              1 * time.Second,
			EnableProfiling:         false,
			EnableCaching:           false,
			CacheTTL:                10 * time.Minute,
		}
	}

	return &Executor{
		nodeRegistry: nodeRegistry,
		storage:      storage,
		logger:       logger,
		config:       config,
	}
}

// ExecuteWorkflow executes a workflow instance
func (e *Executor) ExecuteWorkflow(ctx context.Context, workflow *types.Workflow, triggerParams map[string]interface{}) (string, error) {
	executionID := generateExecutionID()
	startTime := time.Now()

	e.logger.Info("Starting execution", map[string]interface{}{
		"execution_id":  executionID,
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
	})

	// Create execution record
	execution := &types.Execution{
		ID:            executionID,
		WorkflowID:    workflow.ID,
		Status:        types.ExecutionCreated,
		Variables:     make(map[string]interface{}),
		NodeResults:   make(map[string]*types.NodeResult),
		StartedAt:     startTime,
		TriggeredBy:   "manual", // This can be from scheduler, API, etc.
		TriggerParams: triggerParams,
	}

	// Initialize variables with trigger params and workflow config
	for k, v := range workflow.Variables {
		execution.Variables[k] = v
	}
	for k, v := range triggerParams {
		execution.Variables[k] = v
	}

	// Persist execution
	if err := e.storage.CreateExecution(execution); err != nil {
		e.logger.Error("Failed to create execution", map[string]interface{}{
			"execution_id": executionID,
			"error":        err,
		})
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	// Update execution status to queued
	execution.Status = types.ExecutionQueued
	if err := e.storage.UpdateExecution(execution); err != nil {
		return "", fmt.Errorf("failed to update execution: %w", err)
	}

	// Execute in background to return execution ID immediately
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour) // Long-running execution timeout
		defer cancel()

		err := e.runExecution(ctx, execution, workflow)
		if err != nil {
			e.logger.Error("Execution failed", map[string]interface{}{
				"execution_id": executionID,
				"error":        err,
			})
		}
	}()

	return executionID, nil
}

// runExecution runs the actual execution of the workflow
func (e *Executor) runExecution(ctx context.Context, execution *types.Execution, workflow *types.Workflow) error {
	execution.Status = types.ExecutionRunning
	startTime := time.Now()

	if err := e.storage.UpdateExecution(execution); err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	e.logger.Info("Execution started", map[string]interface{}{
		"execution_id": execution.ID,
		"workflow_id":  workflow.ID,
	})

	// Execute nodes following dependency graph
	result := e.executeNodes(ctx, execution, workflow)

	// Complete execution
	execution.CompletedAt = &startTime
	execution.ExecutionTime = time.Since(startTime)

	if result.success {
		execution.Status = types.ExecutionSucceeded
	} else {
		execution.Status = types.ExecutionFailed
		execution.Error = &result.errorMsg
	}

	if err := e.storage.UpdateExecution(execution); err != nil {
		return fmt.Errorf("failed to update completed execution: %w", err)
	}

	e.logger.Info("Execution completed", map[string]interface{}{
		"execution_id": execution.ID,
		"status":       execution.Status,
		"duration":     execution.ExecutionTime,
	})

	return nil
}

// nodeExecutionResult holds the result of node execution
type nodeExecutionResult struct {
	success  bool
	errorMsg string
}

// executeNodes executes workflow nodes with dependency resolution
func (e *Executor) executeNodes(ctx context.Context, execution *types.Execution, workflow *types.Workflow) nodeExecutionResult {
	// Build dependency graph
	graph, err := e.buildDependencyGraph(workflow)
	if err != nil {
		return nodeExecutionResult{
			success:  false,
			errorMsg: fmt.Sprintf("failed to build dependency graph: %v", err),
		}
	}

	// Prepare execution state
	executionState := &executionState{
		execution: execution,
		workflow:  workflow,
		graph:     graph,
		nodes:     make(map[string]*types.Node),
		results:   make(map[string]*types.NodeResult),
		lock:      sync.Mutex{},
	}

	// Map nodes by ID for quick lookup
	for _, node := range workflow.Nodes {
		executionState.nodes[node.ID] = node
	}

	// Track ready and completed nodes
	readyNodes := make(chan *types.Node, len(workflow.Nodes))
	completedNodes := make(chan string, len(workflow.Nodes))
	errorChan := make(chan error, len(workflow.Nodes))

	// Goroutine to manage scheduling of ready nodes
	go e.manageNodeScheduling(executionState, readyNodes, completedNodes, errorChan)

	// Start with nodes that have no dependencies
	for _, node := range workflow.Nodes {
		if len(node.Dependencies) == 0 {
			readyNodes <- node
		}
	}

	// Wait for all nodes to complete or encounter an error
	completedCount := 0
	expectedCount := len(workflow.Nodes)

	for completedCount < expectedCount {
		select {
		case nodeID := <-completedNodes:
			completedCount++
			e.logger.Debug("Node completed", map[string]interface{}{
				"execution_id":    execution.ID,
				"node_id":         nodeID,
				"completed_count": completedCount,
				"total_nodes":     expectedCount,
			})

			// Schedule dependent nodes that are now ready
			for _, dependentNodeID := range graph[nodeID] {
				dependentNode := executionState.nodes[dependentNodeID]

				// Check if all dependencies are satisfied
				dependencySatisfied := true
				for _, depID := range dependentNode.Dependencies {
					if _, exists := executionState.results[depID]; !exists {
						dependencySatisfied = false
						break
					}
				}

				if dependencySatisfied {
					readyNodes <- dependentNode
				}
			}

		case err := <-errorChan:
			return nodeExecutionResult{
				success:  false,
				errorMsg: fmt.Sprintf("node execution error: %v", err),
			}

		case <-ctx.Done():
			return nodeExecutionResult{
				success:  false,
				errorMsg: fmt.Sprintf("execution context cancelled: %v", ctx.Err()),
			}
		}
	}

	// Close the readyNodes channel as no more nodes will be scheduled
	close(readyNodes)

	return nodeExecutionResult{
		success:  true,
		errorMsg: "",
	}
}

// manageNodeScheduling manages the scheduling and execution of nodes
func (e *Executor) manageNodeScheduling(
	state *executionState,
	readyNodes chan *types.Node,
	completedNodes chan string,
	errorChan chan error,
) {
	// Limit concurrent node execution
	semaphore := make(chan struct{}, e.config.MaxConcurrentNodes)

	var wg sync.WaitGroup

	for node := range readyNodes {
		wg.Add(1)
		go func(currentNode *types.Node) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute the node
			result, err := e.executeSingleNode(state.execution, state.workflow, currentNode, state.results)
			if err != nil {
				errorChan <- err
				return
			}

			// Store result
			state.lock.Lock()
			state.results[currentNode.ID] = result
			state.execution.NodeResults[currentNode.ID] = result
			state.lock.Unlock()

			// Update execution with result
			if err := e.storage.UpdateExecution(state.execution); err != nil {
				errorChan <- fmt.Errorf("failed to update execution with node result: %w", err)
				return
			}

			// Notify completion
			completedNodes <- currentNode.ID
		}(node)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(completedNodes)
}

// executeSingleNode executes a single node
func (e *Executor) executeSingleNode(execution *types.Execution, workflow *types.Workflow, node *types.Node, previousResults map[string]*types.NodeResult) (*types.NodeResult, error) {
	startTime := time.Now()

	nodeResult := &types.NodeResult{
		ID:          generateNodeResultID(),
		ExecutionID: execution.ID,
		NodeID:      node.ID,
		Status:      types.NodeRunning,
		StartedAt:   startTime,
		RetryCount:  0,
	}

	// Prepare inputs by evaluating expressions and dependencies
	inputs, err := e.prepareNodeInputs(node, execution, previousResults)
	if err != nil {
		errorMsg := err.Error()
		nodeResult.Error = &errorMsg
		nodeResult.Status = types.NodeFailed
		nodeResult.CompletedAt = &startTime
		nodeResult.ExecutionTime = time.Since(startTime)

		return nodeResult, err
	}

	// Add execution context to inputs
	inputs["execution_context"] = map[string]interface{}{
		"execution_id": execution.ID,
		"workflow_id":  workflow.ID,
		"node_id":      node.ID,
		"timestamp":    startTime.Unix(),
		"variables":    execution.Variables,
	}

	e.logger.Info("Executing node", map[string]interface{}{
		"execution_id": execution.ID,
		"node_id":      node.ID,
		"node_type":    node.Type,
		"inputs_size":  len(inputs),
	})

	// Create and configure node instance
	nodeInstance, err := e.nodeRegistry.CreateInstance(node.Type, node.Config)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to create node instance: %v", err)
		nodeResult.Error = &errorMsg
		nodeResult.Status = types.NodeFailed
		nodeResult.CompletedAt = &startTime
		nodeResult.ExecutionTime = time.Since(startTime)

		return nodeResult, err
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.DefaultTimeout)
	defer cancel()

	output, err := nodeInstance.Execute(ctx, inputs)

	// Record completion time
	completedAt := time.Now()
	nodeResult.CompletedAt = &completedAt
	nodeResult.ExecutionTime = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		nodeResult.Error = &errorMsg
		nodeResult.Status = types.NodeFailed

		e.logger.Error("Node execution failed", map[string]interface{}{
			"execution_id": execution.ID,
			"node_id":      node.ID,
			"error":        err,
		})
	} else {
		nodeResult.Output = output
		nodeResult.Status = types.NodeCompleted

		e.logger.Info("Node executed successfully", map[string]interface{}{
			"execution_id": execution.ID,
			"node_id":      node.ID,
			"output_size":  len(output),
			"duration":     nodeResult.ExecutionTime,
		})
	}

	return nodeResult, nil
}

// prepareNodeInputs prepares inputs for a node based on dependencies and expressions
func (e *Executor) prepareNodeInputs(node *types.Node, execution *types.Execution, previousResults map[string]*types.NodeResult) (map[string]interface{}, error) {
	// Start with the node's configured inputs
	inputs := make(map[string]interface{})
	for k, v := range node.Inputs {
		inputs[k] = v
	}

	// Merge with execution variables
	for k, v := range execution.Variables {
		// Only add if not overridden by node inputs
		if _, exists := inputs[k]; !exists {
			inputs[k] = v
		}
	}

	// Process dependency outputs
	for _, depID := range node.Dependencies {
		result, exists := previousResults[depID]
		if !exists {
			return nil, fmt.Errorf("dependency node %s result not found", depID)
		}

		if result.Status != types.NodeCompleted {
			return nil, fmt.Errorf("dependency node %s did not complete successfully", depID)
		}

		// Add dependency outputs to inputs using namespaced access
		for k, v := range result.Output {
			inputs[fmt.Sprintf("%s_%s", depID, k)] = v
		}
	}

	return inputs, nil
}

// buildDependencyGraph creates a dependency graph from the workflow
func (e *Executor) buildDependencyGraph(workflow *types.Workflow) (map[string][]string, error) {
	graph := make(map[string][]string)

	// Create adjacency list from connections
	for _, conn := range workflow.Connections {
		graph[conn.SourceNodeID] = append(graph[conn.SourceNodeID], conn.TargetNodeID)
	}

	return graph, nil
}

// executionState holds the state of a running execution
type executionState struct {
	execution *types.Execution
	workflow  *types.Workflow
	graph     map[string][]string
	nodes     map[string]*types.Node
	results   map[string]*types.NodeResult
	lock      sync.Mutex
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// generateNodeResultID generates a unique node result ID
func generateNodeResultID() string {
	return fmt.Sprintf("nr_%d", time.Now().UnixNano())
}
