// backend/internal/api/handlers/workflow_handler.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/citadel-agent/backend/internal/workflow/engine"
	"github.com/citadel-agent/backend/internal/workflow/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	engine *engine.Engine
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(engine *engine.Engine) *WorkflowHandler {
	return &WorkflowHandler{
		engine: engine,
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

	// Generate ID if not provided
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	workflow.Status = "active" // Default status
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	// In a real implementation, you would save to database
	// For now, we'll just return the created workflow
	return c.Status(201).JSON(workflow)
}

// GetWorkflow handles GET /api/v1/workflows/:id
func (wh *WorkflowHandler) GetWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	
	// In a real implementation, you would fetch from database
	// For now, return a placeholder
	workflow := models.Workflow{
		ID:   workflowID,
		Name: fmt.Sprintf("Workflow %s", workflowID),
		Status: "active",
	}
	
	return c.JSON(workflow)
}

// UpdateWorkflow handles PUT /api/v1/workflows/:id
func (wh *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	
	var workflow models.Workflow
	if err := c.BodyParser(&workflow); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify the ID matches the URL parameter
	if workflow.ID != workflowID {
		return c.Status(400).JSON(fiber.Map{
			"error": "Workflow ID in body doesn't match URL parameter",
		})
	}

	workflow.UpdatedAt = time.Now()

	// In a real implementation, you would update in database
	// For now, we'll just return the updated workflow
	return c.JSON(workflow)
}

// DeleteWorkflow handles DELETE /api/v1/workflows/:id
func (wh *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	
	// In a real implementation, you would delete from database
	// For now, just return success
	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Workflow %s deleted successfully", workflowID),
	})
}

// ExecuteWorkflow handles POST /api/v1/workflows/:id/execute
func (wh *WorkflowHandler) ExecuteWorkflow(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	
	// In a real implementation, you would fetch the workflow from database
	// For now, create a placeholder workflow with a simple structure
	workflow := &models.Workflow{
		ID: workflowID,
		Name: fmt.Sprintf("Workflow %s", workflowID),
		Nodes: []models.Node{
			{
				ID: "node-1",
				Name: "Start Node",
				Type: "http_request",
				Parameters: map[string]interface{}{
					"url": "https://jsonplaceholder.typicode.com/posts/1",
					"method": "GET",
				},
			},
		},
		Connections: []models.Connection{},
	}
	
	ctx := context.Background()
	execution, err := wh.engine.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to execute workflow: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"execution_id": execution.ID,
		"message": "Workflow execution started",
	})
}

// GetExecution handles GET /api/v1/executions/:id
func (wh *WorkflowHandler) GetExecution(c *fiber.Ctx) error {
	executionID := c.Params("id")
	
	execution, err := wh.engine.GetExecution(executionID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Execution not found: %v", err),
		})
	}

	return c.JSON(execution)
}

// ListWorkflows handles GET /api/v1/workflows
func (wh *WorkflowHandler) ListWorkflows(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// In a real implementation, you would fetch from database
	// For now, return placeholder data
	workflows := []models.Workflow{
		{
			ID: uuid.New().String(),
			Name: "Example Workflow",
			Description: "This is an example workflow",
			Status: "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return c.JSON(fiber.Map{
		"data": workflows,
		"pagination": fiber.Map{
			"page": page,
			"limit": limit,
			"total": len(workflows), // In real app, this would be total from DB
		},
	})
}

// RegisterRoutes registers workflow handler routes
func (wh *WorkflowHandler) RegisterRoutes(router fiber.Router) {
	// Public routes (might need authentication in real app)
	router.Post("/workflows", wh.CreateWorkflow)
	router.Get("/workflows", wh.ListWorkflows) 
	router.Get("/workflows/:id", wh.GetWorkflow)
	router.Put("/workflows/:id", wh.UpdateWorkflow)
	router.Delete("/workflows/:id", wh.DeleteWorkflow)
	
	// Execution routes
	router.Post("/workflows/:id/execute", wh.ExecuteWorkflow)
	router.Get("/executions/:id", wh.GetExecution)
}