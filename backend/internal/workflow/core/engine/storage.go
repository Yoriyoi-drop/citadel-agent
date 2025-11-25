package engine

import (
	"sync"

	"citadel-agent/backend/internal/workflow/core/types"
)



// BasicStorage provides a basic in-memory implementation for testing
type BasicStorage struct {
	executions    map[string]*types.Execution
	nodeResults   map[string]*types.NodeResult
	workflows     map[string]*types.Workflow
	variables     map[string]map[string]interface{} // execution_id -> key -> value
	mutex         sync.RWMutex
}

// NewBasicStorage creates a new in-memory storage for testing
func NewBasicStorage() *BasicStorage {
	return &BasicStorage{
		executions:  make(map[string]*types.Execution),
		nodeResults: make(map[string]*types.NodeResult),
		workflows:   make(map[string]*types.Workflow),
		variables:   make(map[string]map[string]interface{}),
	}
}

// Implementation of Storage interface methods would go here
// For brevity, I'll implement the most essential ones:

func (bs *BasicStorage) CreateExecution(execution *types.Execution) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.executions[execution.ID] = execution
	return nil
}

func (bs *BasicStorage) UpdateExecution(execution *types.Execution) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.executions[execution.ID] = execution
	return nil
}

func (bs *BasicStorage) GetExecution(id string) (*types.Execution, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	execution, exists := bs.executions[id]
	if !exists {
		return nil, &types.WorkflowValidationError{
			Errors: []types.ValidationError{
				{
					Field:   "execution_id",
					Message: "execution not found",
					Code:    "EXECUTION_NOT_FOUND",
					Value:   id,
				},
			},
		}
	}
	
	return execution, nil
}

func (bs *BasicStorage) ListExecutions(workflowID string, limit, offset int) ([]*types.Execution, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	var results []*types.Execution
	for _, exec := range bs.executions {
		if exec.WorkflowID == workflowID {
			results = append(results, exec)
		}
	}
	
	// Apply pagination
	if offset < len(results) {
		results = results[offset:]
	}
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	
	return results, nil
}

func (bs *BasicStorage) CreateNodeResult(result *types.NodeResult) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.nodeResults[result.ID] = result
	return nil
}

func (bs *BasicStorage) GetNodeResult(executionID, nodeID string) (*types.NodeResult, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	for _, result := range bs.nodeResults {
		if result.ExecutionID == executionID && result.NodeID == nodeID {
			return result, nil
		}
	}
	
	return nil, &types.WorkflowValidationError{
		Errors: []types.ValidationError{
			{
				Field:   "node_result",
				Message: "node result not found",
				Code:    "NODE_RESULT_NOT_FOUND",
				Value:   map[string]string{"execution_id": executionID, "node_id": nodeID},
			},
		},
	}
}

func (bs *BasicStorage) CreateWorkflow(workflow *types.Workflow) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.workflows[workflow.ID] = workflow
	return nil
}

func (bs *BasicStorage) GetWorkflow(id string) (*types.Workflow, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	workflow, exists := bs.workflows[id]
	if !exists {
		return nil, &types.WorkflowValidationError{
			Errors: []types.ValidationError{
				{
					Field:   "workflow_id",
					Message: "workflow not found",
					Code:    "WORKFLOW_NOT_FOUND",
					Value:   id,
				},
			},
		}
	}
	
	return workflow, nil
}

func (bs *BasicStorage) GetVariable(executionID, key string) (interface{}, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	varMap, exists := bs.variables[executionID]
	if !exists {
		return nil, nil
	}
	
	value, exists := varMap[key]
	if !exists {
		return nil, nil
	}
	
	return value, nil
}

func (bs *BasicStorage) SetVariable(executionID, key string, value interface{}) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	if bs.variables[executionID] == nil {
		bs.variables[executionID] = make(map[string]interface{})
	}
	
	bs.variables[executionID][key] = value
	return nil
}

