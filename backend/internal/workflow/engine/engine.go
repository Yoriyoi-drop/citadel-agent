// backend/internal/workflow/engine/engine.go
package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/models"
	"github.com/citadel-agent/backend/internal/workflow/nodes"
)

// Engine handles workflow execution
type Engine struct {
	nodeRegistry *nodes.NodeRegistry
	executions   map[string]*models.Execution
	mutex        sync.RWMutex
}

// NewEngine creates a new workflow engine instance
func NewEngine() *Engine {
	return &Engine{
		nodeRegistry: nodes.GetNodeRegistry(),
		executions:   make(map[string]*models.Execution),
	}
}

// ExecuteWorkflow executes a complete workflow
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *models.Workflow) (*models.Execution, error) {
	execution := &models.Execution{
		ID:         generateExecutionID(), // You would need an ID generation function
		WorkflowID: workflow.ID,
		Mode:       "manual",
		Status:     "running",
		StartedAt:  time.Now(),
		Finished:   false,
	}

	// Store execution
	e.mutex.Lock()
	e.executions[execution.ID] = execution
	e.mutex.Unlock()

	// Execute the workflow
	go e.executeWorkflowInternal(ctx, workflow, execution)

	return execution, nil
}

// executeWorkflowInternal handles the actual workflow execution
func (e *Engine) executeWorkflowInternal(ctx context.Context, workflow *models.Workflow, execution *models.Execution) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Workflow execution panicked: %v", r)
			e.updateExecutionStatus(execution.ID, "error", fmt.Sprintf("Panicked: %v", r))
		}
	}()

	// Find the starting node(s)
	startNodes := e.findStartNodes(workflow)
	
	// Execute nodes in the correct order
	for _, node := range startNodes {
		if err := e.executeNode(ctx, node, workflow, execution); err != nil {
			log.Printf("Error executing node %s: %v", node.ID, err)
			e.updateExecutionStatus(execution.ID, "error", err.Error())
			return
		}
	}

	// Mark execution as finished
	e.updateExecutionStatus(execution.ID, "success", "")
}

// findStartNodes finds the starting node(s) of a workflow
func (e *Engine) findStartNodes(workflow *models.Workflow) []*models.Node {
	// Find nodes that have no incoming connections
	// This is a simplified approach
	var startNodes []*models.Node
	
	// In a real implementation, you'd identify trigger nodes
	// For now, let's assume the first node is the start
	if len(workflow.Nodes) > 0 {
		for i := range workflow.Nodes {
			startNodes = append(startNodes, &workflow.Nodes[i])
		}
		// In a real n8n-like system, you'd identify trigger nodes specifically
	}
	
	return startNodes
}

// executeNode executes a single node
func (e *Engine) executeNode(ctx context.Context, node *models.Node, workflow *models.Workflow, execution *models.Execution) error {
	// Get the node implementation from registry
	nodeImpl, err := e.nodeRegistry.GetNode(node.Type)
	if err != nil {
		return fmt.Errorf("node type %s not found: %w", node.Type, err)
	}

	// Execute the node
	result, err := nodeImpl.Execute(node.Parameters)
	if err != nil {
		return fmt.Errorf("error executing node %s: %w", node.ID, err)
	}

	// Process outputs
	node.Outputs = result

	// Find and execute connected nodes
	connectedNodes := e.findConnectedNodes(node, workflow)
	for _, connectedNode := range connectedNodes {
		// Pass the result to the connected node as input
		connectedNode.Parameters["input"] = result
		if err := e.executeNode(ctx, connectedNode, workflow, execution); err != nil {
			return err
		}
	}

	return nil
}

// findConnectedNodes finds all nodes connected to the given node
func (e *Engine) findConnectedNodes(fromNode *models.Node, workflow *models.Workflow) []*models.Node {
	var connectedNodes []*models.Node
	
	for _, connection := range workflow.Connections {
		if connection.SourceNode == fromNode.ID {
			// Find the target node in the workflow
			for i := range workflow.Nodes {
				if workflow.Nodes[i].ID == connection.TargetNode {
					connectedNodes = append(connectedNodes, &workflow.Nodes[i])
					break
				}
			}
		}
	}
	
	return connectedNodes
}

// updateExecutionStatus updates the status of an execution
func (e *Engine) updateExecutionStatus(executionID, status, message string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if execution, exists := e.executions[executionID]; exists {
		execution.Status = status
		now := time.Now()
		execution.StoppedAt = &now
		execution.Finished = true

		if status == "error" && message != "" {
			execution.Error = &message
		}
	}
}

// GetExecution returns the status of an execution
func (e *Engine) GetExecution(executionID string) (*models.Execution, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	execution, exists := e.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}
	
	return execution, nil
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	// In a real implementation, you'd use UUID or similar
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}