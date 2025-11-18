package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Node represents a single node in the workflow
type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
	Inputs   []string               `json:"inputs"`
	Outputs  []string               `json:"outputs"`
}

// Edge represents a connection between two nodes
type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// Workflow represents the entire workflow
type Workflow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Execution represents a single run of a workflow
type Execution struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"` // "running", "completed", "failed", "cancelled"
	StartedAt  time.Time              `json:"started_at"`
	EndedAt    *time.Time             `json:"ended_at,omitempty"`
	Results    map[string]interface{} `json:"results"`
	Error      string                 `json:"error,omitempty"`
	Variables  map[string]interface{} `json:"variables"`
}

// Runner manages the execution of entire workflows
type Runner struct {
	executor *Executor
}

// NewRunner creates a new instance of Runner
func NewRunner(executor *Executor) *Runner {
	return &Runner{
		executor: executor,
	}
}

// RunWorkflow executes an entire workflow
func (r *Runner) RunWorkflow(ctx context.Context, workflow *Workflow, variables map[string]interface{}) (*Execution, error) {
	execution := &Execution{
		ID:         uuid.New().String(),
		WorkflowID: workflow.ID,
		Status:     "running",
		StartedAt:  time.Now(),
		Results:    make(map[string]interface{}),
		Variables:  variables,
	}

	// Execute the workflow
	err := r.executeWorkflow(ctx, workflow, execution)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
	} else {
		execution.Status = "completed"
	}

	endTime := time.Now()
	execution.EndedAt = &endTime

	return execution, err
}

// executeWorkflow executes the workflow logic
func (r *Runner) executeWorkflow(ctx context.Context, workflow *Workflow, execution *Execution) error {
	// Build a dependency graph from nodes and edges
	nodeMap := make(map[string]*Node)
	for i := range workflow.Nodes {
		node := &workflow.Nodes[i]
		nodeMap[node.ID] = node
	}

	// Find start nodes (nodes with no incoming edges)
	startNodes := r.findStartNodes(workflow)
	
	// Execute the workflow starting from start nodes
	for _, startNode := range startNodes {
		err := r.executeNodeRecursive(ctx, workflow, execution, startNode, nodeMap)
		if err != nil {
			return err
		}
	}

	return nil
}

// findStartNodes finds nodes with no incoming edges
func (r *Runner) findStartNodes(workflow *Workflow) []*Node {
	// Create a set of all target nodes (nodes that have incoming edges)
	targetNodes := make(map[string]bool)
	for _, edge := range workflow.Edges {
		targetNodes[edge.Target] = true
	}

	var startNodes []*Node
	for i := range workflow.Nodes {
		node := &workflow.Nodes[i]
		if !targetNodes[node.ID] {
			startNodes = append(startNodes, node)
		}
	}

	return startNodes
}

// executeNodeRecursive executes a node and its dependents
func (r *Runner) executeNodeRecursive(ctx context.Context, workflow *Workflow, execution *Execution, node *Node, nodeMap map[string]*Node) error {
	// Prepare input data for the node
	inputData, err := r.prepareNodeInput(node, execution)
	if err != nil {
		return fmt.Errorf("failed to prepare input for node %s: %w", node.ID, err)
	}

	// Execute the node
	result, err := r.executor.ExecuteNode(ctx, node.Type, inputData)
	if err != nil {
		return fmt.Errorf("failed to execute node %s: %w", node.ID, err)
	}

	// Store result
	execution.Results[node.ID] = result

	// Send update via WebSocket
	r.executor.SendUpdate(result)

	// Check if execution failed
	if result.Status == "error" {
		return fmt.Errorf("node %s failed: %s", node.ID, result.Error)
	}

	// Find dependent nodes (nodes that depend on this node)
	dependentNodes := r.findDependentNodes(node.ID, workflow.Edges)
	
	// Execute dependent nodes
	for _, dependentNode := range dependentNodes {
		depNode := nodeMap[dependentNode]
		err := r.executeNodeRecursive(ctx, workflow, execution, depNode, nodeMap)
		if err != nil {
			return err
		}
	}

	return nil
}

// prepareNodeInput prepares input data for a node based on its dependencies
func (r *Runner) prepareNodeInput(node *Node, execution *Execution) (map[string]interface{}, error) {
	inputData := make(map[string]interface{})

	// Copy execution variables
	for k, v := range execution.Variables {
		inputData[k] = v
	}

	// Get input from dependent nodes
	for _, inputNodeID := range node.Inputs {
		if result, exists := execution.Results[inputNodeID]; exists {
			// Add input data from previous node result
			if resultData, ok := result.(*ExecutionResult); ok {
				if resultData.Data != nil {
					if resultMap, ok := resultData.Data.(map[string]interface{}); ok {
						for k, v := range resultMap {
							inputData[k] = v
						}
					} else {
						// If not a map, store it as 'data' field
						inputData["data"] = resultData.Data
					}
				}
			}
		}
	}

	return inputData, nil
}

// findDependentNodes finds nodes that depend on the current node
func (r *Runner) findDependentNodes(nodeID string, edges []Edge) []string {
	var dependentNodes []string
	for _, edge := range edges {
		if edge.Source == nodeID {
			dependentNodes = append(dependentNodes, edge.Target)
		}
	}
	return dependentNodes
}