package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// WorkflowInput represents the input for a Citadel Agent workflow
type WorkflowInput struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TriggeredBy string                 `json:"triggered_by"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// WorkflowOutput represents the output of a Citadel Agent workflow
type WorkflowOutput struct {
	ID     string                 `json:"id"`
	Result map[string]interface{} `json:"result"`
	Error  string                 `json:"error,omitempty"`
	Status string                 `json:"status"`
}

// NodeInput represents the input for a single node execution
type NodeInput struct {
	NodeID    string                 `json:"node_id"`
	NodeType  string                 `json:"node_type"`
	Config    map[string]interface{} `json:"config"`
	Inputs    map[string]interface{} `json:"inputs"`
	Variables map[string]interface{} `json:"variables"`
}

// NodeOutput represents the output of a single node execution
type NodeOutput struct {
	NodeID   string                 `json:"node_id"`
	Status   string                 `json:"status"`
	Output   map[string]interface{} `json:"output"`
	Error    string                 `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
}

// NodeExecutionOptions contains options for node execution
type NodeExecutionOptions struct {
	RetryAttempts   int           `json:"retry_attempts"`
	Timeout         time.Duration `json:"timeout"`
	CircuitBreaker  bool          `json:"circuit_breaker"`
	RetryOnFailure  bool          `json:"retry_on_failure"`
	MaxConcurrent   int           `json:"max_concurrent"`
}

// WorkflowDefinition represents a complete workflow definition
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []NodeDefinition       `json:"nodes"`
	Connections []ConnectionDefinition `json:"connections"`
	Options     WorkflowOptions        `json:"options"`
}

// NodeDefinition represents a single node in the workflow
type NodeDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Inputs      map[string]interface{} `json:"inputs"`
	Options     NodeExecutionOptions   `json:"options"`
}

// ConnectionDefinition represents a connection between nodes
type ConnectionDefinition struct {
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	Condition    string `json:"condition,omitempty"` // Conditional execution
	Expression   string `json:"expression,omitempty"` // Expression for data transformation
}

// WorkflowOptions contains options for workflow execution
type WorkflowOptions struct {
	Parallelism      int           `json:"parallelism"`
	Timeout          time.Duration `json:"timeout"`
	RetryAttempts    int           `json:"retry_attempts"`
	CircuitBreaker   bool          `json:"circuit_breaker"`
	ErrorHandling    string        `json:"error_handling"` // continue, stop, retry
	MaxConcurrent    int           `json:"max_concurrent"`
	RetryPolicy      RetryPolicy   `json:"retry_policy"`
}

// RetryPolicy defines the retry policy for workflow execution
type RetryPolicy struct {
	InitialInterval    time.Duration `json:"initial_interval"`
	BackoffCoefficient float64       `json:"backoff_coefficient"`
	MaximumInterval    time.Duration `json:"maximum_interval"`
	MaximumAttempts    int           `json:"maximum_attempts"`
	NonRetryableErrors []string      `json:"non_retryable_errors"`
}

// CitadelAgentWorkflow is the main workflow function for Citadel Agent
func CitadelAgentWorkflow(ctx workflow.Context, input WorkflowInput) (WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Citadel Agent workflow", "WorkflowID", workflow.GetInfo(ctx).WorkflowExecution.ID)

	// Set workflow execution options
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		HeartbeatTimeout:    time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, opts)

	// Parse workflow definition from input or external source
	workflowDef := parseWorkflowDefinition(input.ID)

	// Execute the workflow nodes
	result, err := executeNodes(ctx, workflowDef, input.Parameters)
	if err != nil {
		logger.Error("Workflow execution failed", "Error", err)
		return WorkflowOutput{
			ID:     input.ID,
			Error:  err.Error(),
			Status: "failed",
		}, nil
	}

	logger.Info("Workflow completed successfully", "WorkflowID", workflow.GetInfo(ctx).WorkflowExecution.ID)
	return WorkflowOutput{
		ID:     input.ID,
		Result: result,
		Status: "completed",
	}, nil
}

// executeNodes executes the nodes in the workflow according to dependencies
func executeNodes(ctx workflow.Context, workflowDef *WorkflowDefinition, initialInputs map[string]interface{}) (map[string]interface{}, error) {
	logger := workflow.GetLogger(ctx)
	
	// Create a map to track node results
	nodeResults := make(map[string]*NodeOutput)
	
	// Execute nodes based on dependencies
	remainingNodes := make(map[string]NodeDefinition)
	for _, node := range workflowDef.Nodes {
		remainingNodes[node.ID] = node
	}
	
	// Track completed dependencies per node
	depTracker := make(map[string]int)
	
	for len(remainingNodes) > 0 {
		// Find nodes that have all their dependencies satisfied
		readyNodes := findReadyNodes(remainingNodes, depTracker, nodeResults)
		
		if len(readyNodes) == 0 {
			// Check if there are remaining nodes but no ready ones (circular dependency)
			if len(remainingNodes) > 0 {
				return nil, workflow.NewApplicationError("circular dependency detected in workflow", "CircularDependency", nil)
			}
			break
		}
		
		// Execute ready nodes in parallel up to the max concurrent limit
		for _, node := range readyNodes {
			nodeInput := NodeInput{
				NodeID:    node.ID,
				NodeType:  node.Type,
				Config:    node.Config,
				Inputs:    node.Inputs,
				Variables: getNodeInputs(node.ID, workflowDef.Connections, nodeResults, initialInputs),
			}
			
			// Execute the node activity
			var nodeOutput NodeOutput
			err := workflow.ExecuteActivity(ctx, ExecuteNodeActivity, nodeInput).Get(ctx, &nodeOutput)
			if err != nil {
				if workflowDef.Options.ErrorHandling == "continue" {
					logger.Error("Node execution failed, continuing workflow", "NodeID", node.ID, "Error", err)
					nodeOutput = NodeOutput{
						NodeID: node.ID,
						Status: "failed",
						Error:  err.Error(),
					}
				} else {
					logger.Error("Node execution failed, stopping workflow", "NodeID", node.ID, "Error", err)
					return nil, err
				}
			}
			
			// Store the result
			nodeResults[node.ID] = &nodeOutput
			
			// Remove from remaining nodes
			delete(remainingNodes, node.ID)
			
			// Update dependency tracker
			for _, conn := range workflowDef.Connections {
				if conn.SourceNodeID == node.ID {
					depTracker[conn.TargetNodeID]++
				}
			}
		}
	}
	
	// Compile final result
	finalResult := make(map[string]interface{})
	for nodeID, result := range nodeResults {
		finalResult[nodeID] = result.Output
	}
	
	return finalResult, nil
}

// findReadyNodes finds nodes that have all their dependencies satisfied
func findReadyNodes(remainingNodes map[string]NodeDefinition, depTracker map[string]int, nodeResults map[string]*NodeOutput) []NodeDefinition {
	var readyNodes []NodeDefinition
	
	for nodeID, node := range remainingNodes {
		// Count dependencies for this node
		depCount := 0
		satisfiedCount := 0
		
		for _, conn := range getConnectionsToNode(nodeID) {
			depCount++
			if _, exists := nodeResults[conn.SourceNodeID]; exists {
				satisfiedCount++
			}
		}
		
		// If all dependencies are satisfied, node is ready
		if satisfiedCount >= depCount {
			readyNodes = append(readyNodes, node)
		}
	}
	
	return readyNodes
}

// getNodeInputs compiles inputs for a node based on its dependencies
func getNodeInputs(nodeID string, connections []ConnectionDefinition, nodeResults map[string]*NodeOutput, initialInputs map[string]interface{}) map[string]interface{} {
	inputs := make(map[string]interface{})
	
	// Start with initial inputs
	for k, v := range initialInputs {
		inputs[k] = v
	}
	
	// Add outputs from dependency nodes
	for _, conn := range connections {
		if conn.TargetNodeID == nodeID {
			if depResult, exists := nodeResults[conn.SourceNodeID]; exists && depResult.Status == "success" {
				for k, v := range depResult.Output {
					inputs[conn.SourceNodeID+"_"+k] = v
				}
			}
		}
	}
	
	return inputs
}

// getConnectionsToNode gets all connections that point to a specific node
func getConnectionsToNode(targetNodeID string) []ConnectionDefinition {
	// This would typically come from workflow definition
	// For now, return an empty slice - this will be implemented in a real scenario
	return []ConnectionDefinition{}
}

// parseWorkflowDefinition parses a workflow definition
func parseWorkflowDefinition(workflowID string) *WorkflowDefinition {
	// This would typically fetch the workflow definition from a database or file
	// For now, return a minimal example - this will be implemented in a real scenario
	return &WorkflowDefinition{
		ID:          workflowID,
		Name:        "Default Workflow",
		Description: "Default workflow for demonstration",
		Options: WorkflowOptions{
			Parallelism:   5,
			Timeout:       time.Minute * 30,
			RetryAttempts: 3,
			ErrorHandling: "continue",
		},
	}
}