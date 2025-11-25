package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"citadel-agent/backend/internal/workflow/core/engine"
)

// WorkflowHandler handles workflow-related API requests
type WorkflowHandler struct {
	executor *engine.WorkflowExecutor
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(executor *engine.WorkflowExecutor) *WorkflowHandler {
	return &WorkflowHandler{
		executor: executor,
	}
}

// ExecuteWorkflowHandler handles workflow execution requests
func (wh *WorkflowHandler) ExecuteWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var workflow engine.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "Invalid workflow format", http.StatusBadRequest)
		return
	}

	// Get inputs from request
	var inputs map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		// If inputs are not provided in the body, use an empty map
		inputs = make(map[string]interface{})
	}

	// Execute workflow
	ctx := r.Context()
	results, err := wh.executor.ExecuteWorkflow(ctx, &workflow, inputs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Workflow execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"results": results,
		"workflow_id": workflow.ID,
	})
}

// GetWorkflowHandler returns a workflow by ID
func (wh *WorkflowHandler) GetWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement retrieving a workflow from storage
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// SaveWorkflowHandler saves a workflow
func (wh *WorkflowHandler) SaveWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement saving a workflow to storage
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// ListWorkflowsHandler lists all available workflows
func (wh *WorkflowHandler) ListWorkflowsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement listing workflows from storage
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}