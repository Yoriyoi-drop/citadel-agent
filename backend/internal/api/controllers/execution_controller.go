package controllers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/services"
)

// ExecutionController handles execution-related HTTP requests
type ExecutionController struct {
	service *services.ExecutionService
	manager *services.ExecutionManagerService
}

// NewExecutionController creates a new execution controller
func NewExecutionController(executionService *services.ExecutionService, executionManagerService *services.ExecutionManagerService) *ExecutionController {
	return &ExecutionController{
		service: executionService,
		manager: executionManagerService,
	}
}

// GetExecution retrieves an execution by ID
func (c *ExecutionController) GetExecution(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	execution, err := c.service.GetExecution(id)
	if err != nil {
		log.Printf("Error getting execution %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Execution not found",
		})
	}

	return ctx.JSON(execution)
}

// GetExecutions retrieves all executions with optional filtering
func (c *ExecutionController) GetExecutions(ctx *fiber.Ctx) error {
	// Get query parameters for filtering
	workflowID := ctx.Query("workflowId")
	status := ctx.Query("status")

	var executions []*models.Execution
	var err error

	switch {
	case workflowID != "" && status != "":
		// Call the correct method
		executions, err = c.service.GetExecutionsByWorkflowAndStatus(workflowID, status)
	case workflowID != "":
		executions, err = c.service.GetExecutions(workflowID)
	case status != "":
		executions, err = c.service.GetExecutionsByStatus(status)
	default:
		// Return all executions - in a real implementation, you'd want pagination
		// For now, this is not implemented to avoid returning too many records
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "workflowId or status parameter is required for this endpoint",
		})
	}

	if err != nil {
		log.Printf("Error getting executions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve executions",
		})
	}

	return ctx.JSON(executions)
}

// GetExecutionResults retrieves the results of a specific execution
func (c *ExecutionController) GetExecutionResults(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	results, err := c.service.GetExecutionResults(id)
	if err != nil {
		log.Printf("Error getting execution results for %s: %v", id, err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Execution results not found",
		})
	}

	return ctx.JSON(fiber.Map{
		"execution_id": id,
		"results":      results,
	})
}

// CancelExecution cancels a running execution
func (c *ExecutionController) CancelExecution(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if err := c.manager.CancelExecution(id); err != nil {
		log.Printf("Error cancelling execution %s: %v", id, err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot cancel execution",
		})
	}

	return ctx.JSON(fiber.Map{
		"execution_id": id,
		"status":      "cancelled",
		"message":     "Execution cancelled successfully",
	})
}

// RetryExecution retries a failed execution
func (c *ExecutionController) RetryExecution(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	// In a real implementation, you might want to accept new variables in the request body
	newId, err := c.manager.RetryExecution(id, nil)
	if err != nil {
		log.Printf("Error retrying execution %s: %v", id, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retry execution",
		})
	}

	return ctx.JSON(fiber.Map{
		"original_execution_id": id,
		"new_execution_id":      newId,
		"message":              "Execution retry initiated successfully",
	})
}

// ExecuteWorkflow executes a workflow (this is for ad-hoc execution)
func (c *ExecutionController) ExecuteWorkflow(ctx *fiber.Ctx) error {
	// Parse request body to get workflow ID and variables
	var req struct {
		WorkflowID string                 `json:"workflowId"`
		Variables  map[string]interface{} `json:"variables"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if req.WorkflowID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Workflow ID is required",
		})
	}

	executionID, err := c.manager.ExecuteWorkflow(req.WorkflowID, req.Variables)
	if err != nil {
		log.Printf("Error executing workflow %s: %v", req.WorkflowID, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot execute workflow",
		})
	}

	return ctx.JSON(fiber.Map{
		"execution_id": executionID,
		"status":      "initiated",
		"message":     "Workflow execution initiated successfully",
	})
}

// GetRecentExecutions retrieves recent executions
func (c *ExecutionController) GetRecentExecutions(ctx *fiber.Ctx) error {
	limitStr := ctx.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	executions, err := c.manager.GetRecentExecutions(limit)
	if err != nil {
		log.Printf("Error getting recent executions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve recent executions",
		})
	}

	return ctx.JSON(executions)
}

// GetRunningExecutions retrieves all currently running executions
func (c *ExecutionController) GetRunningExecutions(ctx *fiber.Ctx) error {
	executions, err := c.manager.GetRunningExecutions()
	if err != nil {
		log.Printf("Error getting running executions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve running executions",
		})
	}

	return ctx.JSON(executions)
}

// GetExecutionStats retrieves execution statistics for a workflow
func (c *ExecutionController) GetExecutionStats(ctx *fiber.Ctx) error {
	workflowId := ctx.Params("workflowId")

	stats, err := c.manager.GetExecutionStats(workflowId)
	if err != nil {
		log.Printf("Error getting execution stats for workflow %s: %v", workflowId, err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve execution statistics",
		})
	}

	return ctx.JSON(stats)
}