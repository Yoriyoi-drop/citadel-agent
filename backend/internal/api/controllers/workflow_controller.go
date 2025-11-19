package controllers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"citadel-agent/backend/internal/models"
	"citadel-agent/backend/internal/services"
)

// WorkflowController handles workflow-related HTTP requests
type WorkflowController struct {
	service *services.WorkflowService
	executionService *services.ExecutionService
	executionManagerService *services.ExecutionManagerService
}

// NewWorkflowController creates a new workflow controller
func NewWorkflowController(workflowService *services.WorkflowService, executionService *services.ExecutionService, executionManagerService *services.ExecutionManagerService) *WorkflowController {
	return &WorkflowController{
		service: workflowService,
		executionService: executionService,
		executionManagerService: executionManagerService,
	}
}

// CreateWorkflow creates a new workflow
func (c *WorkflowController) CreateWorkflow(ctx *fiber.Ctx) error {
	workflow := new(models.Workflow)
	if err := ctx.BodyParser(workflow); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if err := c.service.CreateWorkflow(workflow); err != nil {
		log.Printf("Error creating workflow: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot create workflow",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(workflow)
}

// GetWorkflow retrieves a workflow by ID
func (c *WorkflowController) GetWorkflow(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	workflow, err := c.service.GetWorkflow(id)
	if err != nil {
		log.Printf("Error getting workflow %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	return ctx.JSON(workflow)
}

// GetWorkflows retrieves all workflows
func (c *WorkflowController) GetWorkflows(ctx *fiber.Ctx) error {
	// Get query parameters for pagination
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// In a real implementation, we would apply pagination
	// For now, just get all workflows
	workflows, err := c.service.GetWorkflows()
	if err != nil {
		log.Printf("Error getting workflows: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve workflows",
		})
	}

	return ctx.JSON(workflows)
}

// UpdateWorkflow updates a workflow
func (c *WorkflowController) UpdateWorkflow(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	workflow, err := c.service.GetWorkflow(id)
	if err != nil {
		log.Printf("Error getting workflow %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	if err := ctx.BodyParser(workflow); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	// Keep the original ID
	workflow.ID = id

	if err := c.service.UpdateWorkflow(workflow); err != nil {
		log.Printf("Error updating workflow %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot update workflow",
		})
	}

	return ctx.JSON(workflow)
}

// DeleteWorkflow deletes a workflow by ID
func (c *WorkflowController) DeleteWorkflow(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if err := c.service.DeleteWorkflow(id); err != nil {
		log.Printf("Error deleting workflow %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot delete workflow",
		})
	}

	return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "Workflow deleted successfully",
	})
}

// ExecuteWorkflow executes a workflow
func (c *WorkflowController) ExecuteWorkflow(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	// Parse request body to get variables (optional)
	var req struct {
		Variables map[string]interface{} `json:"variables"`
	}

	// Only parse body if it's provided
	if ctx.Body() != nil && len(ctx.Body()) > 0 {
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
	}

	_, err := c.service.GetWorkflow(id) // Use underscore to indicate we're not using the workflow variable
	if err != nil {
		log.Printf("Error getting workflow %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	// In a real implementation, this would submit the workflow to a worker
	// For now, we'll just return a success message
	log.Printf("Workflow %s submitted for execution", id)

	return ctx.JSON(fiber.Map{
		"message":    "Workflow submitted for execution",
		"workflowId": id,
		"status":     "submitted",
	})
}

// GetWorkflowExecutions retrieves all executions for a workflow
func (c *WorkflowController) GetWorkflowExecutions(ctx *fiber.Ctx) error {
	workflowID := ctx.Params("id")

	executions, err := c.executionService.GetExecutions(workflowID)
	if err != nil {
		log.Printf("Error getting executions for workflow %s: %v", workflowID, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve executions",
		})
	}

	return ctx.JSON(executions)
}

// GetWorkflowStats retrieves execution statistics for a workflow
func (c *WorkflowController) GetWorkflowStats(ctx *fiber.Ctx) error {
	workflowID := ctx.Params("id")

	stats, err := c.executionManagerService.GetExecutionStats(workflowID)
	if err != nil {
		log.Printf("Error getting execution stats for workflow %s: %v", workflowID, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve execution statistics",
		})
	}

	return ctx.JSON(stats)
}