// backend/internal/api/workflow_handler.go
package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/citadel-agent/backend/internal/services"
	"github.com/citadel-agent/backend/internal/models"
)

// WorkflowHandler handles HTTP requests for workflow operations
type WorkflowHandler struct {
	service *services.WorkflowService
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(service *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
	}
}

// CreateWorkflow handles POST /api/v1/workflows
func (wh *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	var workflow models.Workflow
	if err := c.BodyParser(&workflow); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default values if not provided
	if workflow.ID != "" {
		workflow.ID = "" // Don't allow client to set ID
	}

	createdWorkflow, err := wh.service.CreateWorkflow(c.Context(), &workflow)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create workflow: %v", err),
		})
	}

	return c.Status(201).JSON(createdWorkflow)
}

// GetWorkflow handles GET /api/v1/workflows/:id
func (wh *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	
	workflow, err := wh.service.GetWorkflow(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Workflow not found: %v", err),
		})
	}

	return c.JSON(workflow)
}

// UpdateWorkflow handles PUT /api/v1/workflows/:id
func (wh *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var workflow models.Workflow
	if err := c.BodyParser(&workflow); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updatedWorkflow, err := wh.service.UpdateWorkflow(c.Context(), id, &workflow)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update workflow: %v", err),
		})
	}

	return c.JSON(updatedWorkflow)
}

// DeleteWorkflow handles DELETE /api/v1/workflows/:id
func (wh *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	
	err := wh.service.DeleteWorkflow(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete workflow: %v", err),
		})
	}

	return c.SendStatus(204)
}

// ListWorkflows handles GET /api/v1/workflows
func (wh *WorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil || page < 0 {
		page = 0
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	workflows, err := wh.service.ListWorkflows(c.Context(), page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list workflows: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"data": workflows,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// ExecuteWorkflow handles POST /api/v1/workflows/:id/run
func (wh *WorkflowHandler) ExecuteWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var params map[string]interface{}
	if err := c.BodyParser(&params); err != nil {
		// If body parsing fails, continue with empty params
		params = make(map[string]interface{})
	}

	executionID, err := wh.service.ExecuteWorkflow(c.Context(), id, params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to execute workflow: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"execution_id": executionID,
		"message":      "Workflow execution started",
	})
}

// GetExecution handles GET /api/v1/executions/:id
func (wh *WorkflowHandler) GetExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	
	execution, err := wh.service.GetExecution(c.Context(), executionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Execution not found: %v", err),
		})
	}

	return c.JSON(execution)
}

// UpdateWorkflowStatus handles PUT /api/v1/workflows/:id/status
func (wh *WorkflowHandler) UpdateWorkflowStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var req struct {
		Status models.WorkflowStatus `json:"status"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := wh.service.UpdateWorkflowStatus(c.Context(), id, req.Status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update workflow status: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Workflow status updated successfully",
		"status":  req.Status,
	})
}

// GetWorkflowExecutions handles GET /api/v1/workflows/:id/executions
func (wh *WorkflowHandler) GetWorkflowExecutions(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil || page < 0 {
		page = 0
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// Get workflow to verify it exists
	_, err = wh.service.GetWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Workflow not found: %v", err),
		})
	}

	// This is a simplified implementation - in a real system you would have a separate
	// execution service to handle this query
	// For now, we'll just return an empty list with pagination info
	executions := make([]*models.Execution, 0)
	
	return c.JSON(fiber.Map{
		"data": executions,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// GetExecutionLogs handles GET /api/v1/executions/:id/logs
func (wh *WorkflowHandler) GetExecutionLogs(c *fiber.Ctx) error {
	executionID := c.Params("id")
	
	// Verify execution exists
	_, err := wh.service.GetExecution(c.Context(), executionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Execution not found: %v", err),
		})
	}
	
	// For now, return an empty logs list
	logs := make([]*models.ExecutionLog, 0)
	
	return c.JSON(fiber.Map{
		"data": logs,
		"pagination": fiber.Map{
			"page":  0,
			"limit": 20,
		},
	})
}

// RetryExecution handles POST /api/v1/executions/:id/retry
func (wh *WorkflowHandler) RetryExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	
	// Verify execution exists
	execution, err := wh.service.GetExecution(c.Context(), executionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Execution not found: %v", err),
		})
	}

	// Check if execution is in a state that can be retried
	if execution.Status != models.ExecutionStatusFailed && 
	   execution.Status != models.ExecutionStatusCancelled {
		return c.Status(400).JSON(fiber.Map{
			"error": "Execution cannot be retried in current state",
		})
	}

	// For now, just return success - in a real implementation this would
	// trigger a retry of the failed execution
	return c.JSON(fiber.Map{
		"execution_id": execution.ID,
		"message":      "Execution retry initiated",
		"status":       models.ExecutionStatusRetrying,
	})
}

// CancelExecution handles POST /api/v1/executions/:id/cancel
func (wh *WorkflowHandler) CancelExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	
	// Verify execution exists
	execution, err := wh.service.GetExecution(c.Context(), executionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Execution not found: %v", err),
		})
	}

	// Check if execution is in a state that can be cancelled
	if execution.Status != models.ExecutionStatusRunning && 
	   execution.Status != models.ExecutionStatusPending {
		return c.Status(400).JSON(fiber.Map{
			"error": "Execution cannot be cancelled in current state",
		})
	}

	// For now, just return success - in a real implementation this would
	// trigger cancellation of the running execution
	return c.JSON(fiber.Map{
		"execution_id": execution.ID,
		"message":      "Execution cancellation initiated",
		"status":       models.ExecutionStatusCancelled,
	})
}

// RegisterRoutes registers workflow handler routes
func (wh *WorkflowHandler) RegisterRoutes(app *fiber.App) {
	// Public workflow routes
	v1 := app.Group("/api/v1")
	
	v1.Post("/workflows", wh.CreateWorkflow)
	v1.Get("/workflows", wh.ListWorkflows)
	v1.Get("/workflows/:id", wh.GetWorkflow)
	v1.Put("/workflows/:id", wh.UpdateWorkflow)
	v1.Delete("/workflows/:id", wh.DeleteWorkflow)
	v1.Put("/workflows/:id/status", wh.UpdateWorkflowStatus)
	
	// Workflow execution routes
	v1.Post("/workflows/:id/run", wh.ExecuteWorkflow)
	v1.Get("/workflows/:id/executions", wh.GetWorkflowExecutions)
	
	// Execution-specific routes
	v1.Get("/executions/:id", wh.GetExecution)
	v1.Post("/executions/:id/retry", wh.RetryExecution)
	v1.Post("/executions/:id/cancel", wh.CancelExecution)
	v1.Get("/executions/:id/logs", wh.GetExecutionLogs)
}

// HealthCheck handles GET /health
func (wh *WorkflowHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "workflow-api",
	})
}