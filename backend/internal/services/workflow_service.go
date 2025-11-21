// backend/internal/services/workflow_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"
)

// WorkflowService handles business logic for workflow operations
type WorkflowService struct {
	repo      *repositories.WorkflowRepository
	engine    *engine.Engine
	validator *WorkflowValidator
}

// WorkflowValidator validates workflow data
type WorkflowValidator struct {
	maxNodesPerWorkflow int
	maxWorkflowSize     int
	allowedNodeTypes    []engine.NodeType
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(
	repo *repositories.WorkflowRepository,
	workflowEngine *engine.Engine,
) *WorkflowService {
	return &WorkflowService{
		repo:   repo,
		engine: workflowEngine,
		validator: &WorkflowValidator{
			maxNodesPerWorkflow: 100,
			maxWorkflowSize:     1024 * 1024, // 1MB
			allowedNodeTypes:    []engine.NodeType{"http_request", "condition", "delay", "database_query", "script_execution", "ai_agent", "data_transformer", "notification", "loop", "error_handler"},
		},
	}
}

// CreateWorkflow creates a new workflow
func (ws *WorkflowService) CreateWorkflow(ctx context.Context, workflow *models.Workflow) (*models.Workflow, error) {
	// Validate workflow
	if err := ws.validator.ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Set creation timestamp
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()
	workflow.Status = models.WorkflowStatusActive

	// Save to repository
	createdWorkflow, err := ws.repo.Create(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return createdWorkflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (ws *WorkflowService) GetWorkflow(ctx context.Context, id string) (*models.Workflow, error) {
	workflow, err := ws.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (ws *WorkflowService) UpdateWorkflow(ctx context.Context, id string, workflow *models.Workflow) (*models.Workflow, error) {
	// Validate workflow
	if err := ws.validator.ValidateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Set update timestamp
	workflow.UpdatedAt = time.Now()

	// Save to repository
	updatedWorkflow, err := ws.repo.Update(ctx, id, workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return updatedWorkflow, nil
}

// DeleteWorkflow deletes a workflow by ID
func (ws *WorkflowService) DeleteWorkflow(ctx context.Context, id string) error {
	// Check if workflow has active executions
	activeExecutions, err := ws.repo.GetActiveExecutions(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check active executions: %w", err)
	}

	if len(activeExecutions) > 0 {
		return fmt.Errorf("cannot delete workflow with %d active executions", len(activeExecutions))
	}

	// Delete from repository
	if err := ws.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// ListWorkflows retrieves a list of workflows with pagination
func (ws *WorkflowService) ListWorkflows(ctx context.Context, page, limit int) ([]*models.Workflow, error) {
	workflows, err := ws.repo.List(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows, nil
}

// ExecuteWorkflow triggers the execution of a workflow
func (ws *WorkflowService) ExecuteWorkflow(ctx context.Context, id string, params map[string]interface{}) (string, error) {
	// Get the workflow
	workflow, err := ws.GetWorkflow(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow: %w", err)
	}

	if workflow.Status != models.WorkflowStatusActive {
		return "", fmt.Errorf("workflow is not active: %s", workflow.Status)
	}

	// Execute using the engine
	executionID, err := ws.engine.ExecuteWorkflow(ctx, &engine.Workflow{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Nodes:       convertNodesToEngineFormat(workflow.Nodes),
		Connections: convertConnectionsToEngineFormat(workflow.Connections),
		Config:      workflow.Config,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	}, params)
	
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	// Log the execution
	if err := ws.repo.LogExecution(ctx, &models.ExecutionLog{
		WorkflowID:  id,
		ExecutionID: executionID,
		Status:      models.ExecutionStatusRunning,
		StartedAt:   time.Now(),
		Parameters:  params,
	}); err != nil {
		// Log error but don't fail the execution
		fmt.Printf("Failed to log execution: %v\n", err)
	}

	return executionID, nil
}

// GetExecution retrieves execution details by ID
func (ws *WorkflowService) GetExecution(ctx context.Context, executionID string) (*engine.Execution, error) {
	execution, err := ws.engine.GetExecution(executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return execution, nil
}

// ValidateWorkflow validates workflow structure and content
func (wv *WorkflowValidator) ValidateWorkflow(workflow *models.Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Name) > 100 {
		return fmt.Errorf("workflow name too long (max 100 chars)")
	}

	if workflow.Description != "" && len(workflow.Description) > 1000 {
		return fmt.Errorf("workflow description too long (max 1000 chars)")
	}

	if len(workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	if len(workflow.Nodes) > wv.maxNodesPerWorkflow {
		return fmt.Errorf("workflow has too many nodes (max %d)", wv.maxNodesPerWorkflow)
	}

	// Check node types
	for i, node := range workflow.Nodes {
		isAllowed := false
		for _, allowedType := range wv.allowedNodeTypes {
			if string(allowedType) == node.Type {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return fmt.Errorf("node at index %d has invalid type: %s", i, node.Type)
		}
	}

	// Check for circular dependencies
	if hasCycle := ws.hasCircularDependencies(workflow); hasCycle {
		return fmt.Errorf("workflow has circular dependencies")
	}

	return nil
}

// hasCircularDependencies checks if the workflow has circular dependencies
func (ws *WorkflowService) hasCircularDependencies(workflow *models.Workflow) bool {
	// This is a simplified implementation
	// In a real system, you would use graph algorithms to detect cycles
	// For now, we'll use a basic check
	
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	nodesMap := make(map[string]*models.Node)
	for _, node := range workflow.Nodes {
		nodesMap[node.ID] = node
	}
	
	for _, node := range workflow.Nodes {
		if !visited[node.ID] {
			if ws.hasCycleDFS(node.ID, nodesMap, visited, recStack, workflow.Connections) {
				return true
			}
		}
	}
	
	return false
}

func (ws *WorkflowService) hasCycleDFS(nodeID string, nodesMap map[string]*models.Node, visited, recStack map[string]bool, connections []*models.Connection) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Find all connections from this node
	for _, conn := range connections {
		if conn.SourceNodeID == nodeID {
			targetID := conn.TargetNodeID
			if !visited[targetID] {
				if ws.hasCycleDFS(targetID, nodesMap, visited, recStack, connections) {
					return true
				}
			} else if recStack[targetID] {
				return true // Cycle detected
			}
		}
	}

	recStack[nodeID] = false
	return false
}

// convertNodesToEngineFormat converts models nodes to engine format
func convertNodesToEngineFormat(nodes []*models.Node) []*engine.Node {
	engineNodes := make([]*engine.Node, len(nodes))
	for i, node := range nodes {
		engineNodes[i] = &engine.Node{
			ID:          node.ID,
			Type:        node.Type,
			Name:        node.Name,
			Config:      node.Config,
			Dependencies: node.Dependencies,
			Inputs:      node.Inputs,
			Outputs:     make(map[string]interface{}), // Initially empty
			Status:      engine.NodePending,
		}
	}
	return engineNodes
}

// convertConnectionsToEngineFormat converts models connections to engine format
func convertConnectionsToEngineFormat(connections []*models.Connection) []*engine.Connection {
	engineConnections := make([]*engine.Connection, len(connections))
	for i, conn := range connections {
		engineConnections[i] = &engine.Connection{
			SourceNodeID: conn.SourceNodeID,
			TargetNodeID: conn.TargetNodeID,
			Port:         conn.Port,
			Condition:    conn.Condition,
		}
	}
	return engineConnections
}

// UpdateWorkflowStatus updates the status of a workflow
func (ws *WorkflowService) UpdateWorkflowStatus(ctx context.Context, id string, status models.WorkflowStatus) error {
	workflow, err := ws.GetWorkflow(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	workflow.Status = status
	workflow.UpdatedAt = time.Now()

	_, err = ws.repo.Update(ctx, id, workflow)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	return nil
}