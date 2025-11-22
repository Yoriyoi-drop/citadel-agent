// workflow/core/engine/state.go
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExecutionState represents the state of a workflow execution
type ExecutionState struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        ExecutionStatus        `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at,omitempty"`  // Add this field
	Variables     map[string]interface{} `json:"variables"`
	NodeResults   map[string]*NodeResult `json:"node_results"`
	Error         *string                `json:"error,omitempty"`
	TriggeredBy   string                 `json:"triggered_by"`
	TriggerParams map[string]interface{} `json:"trigger_params,omitempty"`
	Progress      ExecutionProgress      `json:"progress"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ExecutionProgress tracks execution progress
type ExecutionProgress struct {
	TotalNodes       int `json:"total_nodes"`
	CompletedNodes   int `json:"completed_nodes"`
	FailedNodes      int `json:"failed_nodes"`
	RunningNodes     int `json:"running_nodes"`
	SkippedNodes     int `json:"skipped_nodes"`
	CompletionPercent int `json:"completion_percent"`
}


// NodeAttempt represents a single attempt of a node execution
type NodeAttempt struct {
	AttemptNumber int                   `json:"attempt_number"`
	Status        NodeStatus            `json:"status"`
	Output        map[string]interface{} `json:"output"`
	Error         *string               `json:"error,omitempty"`
	StartedAt     time.Time             `json:"started_at"`
	CompletedAt   time.Time             `json:"completed_at"`
	ExecutionTime time.Duration         `json:"execution_time"`
}

// StateStorage interface defines the methods for state persistence
type StateStorage interface {
	// Execution management
	CreateExecution(ctx context.Context, execution *ExecutionState) error
	UpdateExecution(ctx context.Context, execution *ExecutionState) error
	GetExecution(ctx context.Context, executionID string) (*ExecutionState, error)
	DeleteExecution(ctx context.Context, executionID string) error
	ListExecutions(ctx context.Context, workflowID string, filters *ExecutionFilters) ([]*ExecutionState, error)
	UpdateExecutionStatus(ctx context.Context, executionID string, status ExecutionStatus) error
	
	// Node result management
	CreateNodeResult(ctx context.Context, executionID string, result *NodeResult) error
	UpdateNodeResult(ctx context.Context, executionID string, result *NodeResult) error
	GetNodeResult(ctx context.Context, executionID, nodeID string) (*NodeResult, error)
	ListNodeResults(ctx context.Context, executionID string) ([]*NodeResult, error)
	
	// State snapshots for long-running executions
	CreateStateSnapshot(ctx context.Context, executionID string, snapshot *StateSnapshot) error
	GetLatestStateSnapshot(ctx context.Context, executionID string) (*StateSnapshot, error)
	ListStateSnapshots(ctx context.Context, executionID string) ([]*StateSnapshot, error)
	
	// Cleanup operations
	CleanupOldExecutions(ctx context.Context, before time.Time) error
	CleanupNodeResults(ctx context.Context, executionID string, before time.Time) error
}

// ExecutionFilters defines filters for querying executions
type ExecutionFilters struct {
	Status      []ExecutionStatus `json:"status,omitempty"`
	StartDate   *time.Time        `json:"start_date,omitempty"`
	EndDate     *time.Time        `json:"end_date,omitempty"`
	Limit       *int              `json:"limit,omitempty"`
	Offset      *int              `json:"offset,omitempty"`
	OrderBy     string            `json:"order_by,omitempty"` // created_at, completed_at, etc.
	OrderDir    string            `json:"order_dir,omitempty"` // asc, desc
}

// StateSnapshot represents a snapshot of execution state at a point in time
type StateSnapshot struct {
	ID           string                 `json:"id"`
	ExecutionID  string                 `json:"execution_id"`
	CreatedAt    time.Time              `json:"created_at"`
	State        map[string]interface{} `json:"state"`
	NodeResults  map[string]*NodeResult `json:"node_results"`
	Progress     ExecutionProgress      `json:"progress"`
	Checkpoint   string                 `json:"checkpoint"` // Named checkpoint for recovery
	Metadata     map[string]interface{} `json:"metadata"`
	Size         int64                  `json:"size"` // Size in bytes
}

// DefaultStateStorage provides a default implementation of StateStorage
type DefaultStateStorage struct {
	// In a real implementation, this would connect to a database
	// For this example, we'll use in-memory storage as a placeholder
	executions   map[string]*ExecutionState
	nodeResults  map[string]map[string]*NodeResult // executionID -> nodeID -> result
	snapshots    map[string][]*StateSnapshot      // executionID -> snapshots
	mutex        sync.RWMutex
	logger       Logger
}

// NewDefaultStateStorage creates a new default state storage
func NewDefaultStateStorage(logger Logger) *DefaultStateStorage {
	return &DefaultStateStorage{
		executions:  make(map[string]*ExecutionState),
		nodeResults: make(map[string]map[string]*NodeResult),
		snapshots:   make(map[string][]*StateSnapshot),
		logger:      logger,
	}
}

// CreateExecution creates a new execution state
func (s *DefaultStateStorage) CreateExecution(ctx context.Context, execution *ExecutionState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.executions[execution.ID]; exists {
		return fmt.Errorf("execution with ID %s already exists", execution.ID)
	}

	s.executions[execution.ID] = execution
	
	if s.logger != nil {
		s.logger.Info("Created execution %s for workflow %s", execution.ID, execution.WorkflowID)
	}

	return nil
}

// UpdateExecution updates an existing execution state
func (s *DefaultStateStorage) UpdateExecution(ctx context.Context, execution *ExecutionState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.executions[execution.ID]; !exists {
		return fmt.Errorf("execution with ID %s does not exist", execution.ID)
	}

	// Update progress calculation
	execution.Progress = s.calculateProgress(execution)

	s.executions[execution.ID] = execution
	
	if s.logger != nil {
		s.logger.Info("Updated execution %s status to %s", execution.ID, execution.Status)
	}

	return nil
}

// GetExecution retrieves an execution state by ID
func (s *DefaultStateStorage) GetExecution(ctx context.Context, executionID string) (*ExecutionState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution with ID %s not found", executionID)
	}

	// Copy to avoid race conditions
	result := *execution
	result.Variables = make(map[string]interface{})
	for k, v := range execution.Variables {
		result.Variables[k] = v
	}
	
	result.NodeResults = make(map[string]*NodeResult)
	for k, v := range execution.NodeResults {
		result.NodeResults[k] = v
	}

	return &result, nil
}

// DeleteExecution deletes an execution state
func (s *DefaultStateStorage) DeleteExecution(ctx context.Context, executionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.executions, executionID)
	delete(s.nodeResults, executionID)
	delete(s.snapshots, executionID)
	
	if s.logger != nil {
		s.logger.Info("Deleted execution %s", executionID)
	}

	return nil
}

// ListExecutions returns a list of executions for a workflow with optional filters
func (s *DefaultStateStorage) ListExecutions(ctx context.Context, workflowID string, filters *ExecutionFilters) ([]*ExecutionState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var results []*ExecutionState
	
	for _, execution := range s.executions {
		// Filter by workflow ID
		if workflowID != "" && execution.WorkflowID != workflowID {
			continue
		}
		
		// Apply status filters
		if filters != nil && len(filters.Status) > 0 {
			found := false
			for _, status := range filters.Status {
				if execution.Status == status {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Apply date filters
		if filters != nil {
			if filters.StartDate != nil && execution.StartedAt.Before(*filters.StartDate) {
				continue
			}
			if filters.EndDate != nil && execution.StartedAt.After(*filters.EndDate) {
				continue
			}
		}
		
		results = append(results, execution)
	}
	
	// Apply ordering and pagination
	if filters != nil {
		// Sort results based on OrderBy and OrderDir
		// For simplicity in this example, we'll skip this step
		// In a real implementation, you would implement sorting
		
		// Apply pagination
		if filters.Offset != nil {
			offset := *filters.Offset
			if offset < len(results) {
				results = results[offset:]
			} else {
				results = []*ExecutionState{} // No results after offset
			}
		}
		
		if filters.Limit != nil {
			limit := *filters.Limit
			if limit < len(results) {
				results = results[:limit]
			}
		}
	}

	return results, nil
}

// UpdateExecutionStatus updates only the status of an execution
func (s *DefaultStateStorage) UpdateExecutionStatus(ctx context.Context, executionID string, status ExecutionStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	execution, exists := s.executions[executionID]
	if !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}
	
	execution.Status = status
	execution.UpdatedAt = time.Now()
	
	if s.logger != nil {
		s.logger.Info("Updated execution %s status to %s", executionID, status)
	}

	return nil
}

// CreateNodeResult creates a new node result
func (s *DefaultStateStorage) CreateNodeResult(ctx context.Context, executionID string, result *NodeResult) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Ensure execution exists
	if _, exists := s.executions[executionID]; !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}

	// Initialize node results map if needed
	if s.nodeResults[executionID] == nil {
		s.nodeResults[executionID] = make(map[string]*NodeResult)
	}

	s.nodeResults[executionID][result.NodeID] = result
	
	if s.logger != nil {
		s.logger.Info("Created node result for execution %s, node %s", executionID, result.NodeID)
	}

	// Update parent execution
	execution := s.executions[executionID]
	if execution.NodeResults == nil {
		execution.NodeResults = make(map[string]*NodeResult)
	}
	execution.NodeResults[result.NodeID] = result
	execution.Progress = s.calculateProgress(execution)

	return nil
}

// UpdateNodeResult updates an existing node result
func (s *DefaultStateStorage) UpdateNodeResult(ctx context.Context, executionID string, result *NodeResult) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Ensure execution exists
	if _, exists := s.executions[executionID]; !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}

	// Ensure node results map exists
	if s.nodeResults[executionID] == nil {
		s.nodeResults[executionID] = make(map[string]*NodeResult)
	}

	// Check if node result exists
	_, exists := s.nodeResults[executionID][result.NodeID]
	if !exists {
		return fmt.Errorf("node result for node %s in execution %s does not exist", result.NodeID, executionID)
	}

	s.nodeResults[executionID][result.NodeID] = result
	
	if s.logger != nil {
		s.logger.Info("Updated node result for execution %s, node %s", executionID, result.NodeID)
	}

	// Update parent execution
	execution := s.executions[executionID]
	execution.NodeResults[result.NodeID] = result
	execution.Progress = s.calculateProgress(execution)

	return nil
}

// GetNodeResult retrieves a node result
func (s *DefaultStateStorage) GetNodeResult(ctx context.Context, executionID, nodeID string) (*NodeResult, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodeResults, exists := s.nodeResults[executionID]
	if !exists {
		return nil, fmt.Errorf("no node results found for execution %s", executionID)
	}

	result, exists := nodeResults[nodeID]
	if !exists {
		return nil, fmt.Errorf("node result for node %s in execution %s not found", nodeID, executionID)
	}

	return result, nil
}

// ListNodeResults retrieves all node results for an execution
func (s *DefaultStateStorage) ListNodeResults(ctx context.Context, executionID string) ([]*NodeResult, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodeResults, exists := s.nodeResults[executionID]
	if !exists {
		return []*NodeResult{}, nil // Return empty slice instead of error
	}

	results := make([]*NodeResult, 0, len(nodeResults))
	for _, result := range nodeResults {
		results = append(results, result)
	}

	return results, nil
}

// CreateStateSnapshot creates a state snapshot
func (s *DefaultStateStorage) CreateStateSnapshot(ctx context.Context, executionID string, snapshot *StateSnapshot) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Ensure execution exists
	if _, exists := s.executions[executionID]; !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}

	// Initialize snapshots slice if needed
	if s.snapshots[executionID] == nil {
		s.snapshots[executionID] = make([]*StateSnapshot, 0)
	}

	snapshot.ID = fmt.Sprintf("%s-snapshot-%d", executionID, time.Now().UnixNano())
	snapshot.CreatedAt = time.Now()
	
	// Calculate size for the snapshot
	snapshot.Size = int64(len(snapshot.State)) * 100 // Rough estimation

	s.snapshots[executionID] = append(s.snapshots[executionID], snapshot)
	
	if s.logger != nil {
		s.logger.Info("Created snapshot %s for execution %s at %s", snapshot.ID, executionID, snapshot.CreatedAt.Format(time.RFC3339))
	}

	return nil
}

// GetLatestStateSnapshot retrieves the latest state snapshot
func (s *DefaultStateStorage) GetLatestStateSnapshot(ctx context.Context, executionID string) (*StateSnapshot, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	snapshotSlice, exists := s.snapshots[executionID]
	if !exists || len(snapshotSlice) == 0 {
		return nil, fmt.Errorf("no snapshots found for execution %s", executionID)
	}

	// Find the latest snapshot
	latest := snapshotSlice[0]
	for _, snapshot := range snapshotSlice {
		if snapshot.CreatedAt.After(latest.CreatedAt) {
			latest = snapshot
		}
	}

	return latest, nil
}

// ListStateSnapshots retrieves all state snapshots for an execution
func (s *DefaultStateStorage) ListStateSnapshots(ctx context.Context, executionID string) ([]*StateSnapshot, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	snapshots, exists := s.snapshots[executionID]
	if !exists {
		return []*StateSnapshot{}, nil // Return empty slice instead of error
	}

	// Return a copy to prevent external modification
	result := make([]*StateSnapshot, len(snapshots))
	copy(result, snapshots)

	return result, nil
}

// CleanupOldExecutions removes executions older than the specified time
func (s *DefaultStateStorage) CleanupOldExecutions(ctx context.Context, before time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	count := 0
	for id, execution := range s.executions {
		if execution.StartedAt.Before(before) {
			delete(s.executions, id)
			delete(s.nodeResults, id)
			delete(s.snapshots, id)
			count++
		}
	}
	
	if s.logger != nil {
		s.logger.Info("Cleaned up %d old executions before %s", count, before.Format(time.RFC3339))
	}

	return nil
}

// CleanupNodeResults removes node results older than the specified time for an execution
func (s *DefaultStateStorage) CleanupNodeResults(ctx context.Context, executionID string, before time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nodeResults, exists := s.nodeResults[executionID]
	if !exists {
		return nil // Nothing to clean up
	}

	for nodeID, result := range nodeResults {
		if result.CompletedAt != nil && (*result.CompletedAt).Before(before) {
			delete(nodeResults, nodeID)
		}
	}
	
	if s.logger != nil {
		s.logger.Info("Cleaned up node results for execution %s before %s", executionID, before.Format(time.RFC3339))
	}

	return nil
}

// calculateProgress calculates the progress of an execution
func (s *DefaultStateStorage) calculateProgress(execution *ExecutionState) ExecutionProgress {
	progress := ExecutionProgress{
		TotalNodes:    len(execution.NodeResults),
		CompletedNodes: 0,
		FailedNodes:    0,
		RunningNodes:   0,
		SkippedNodes:   0,
	}

	for _, result := range execution.NodeResults {
		switch result.Status {
		case NodeSuccess:
			progress.CompletedNodes++
		case NodeFailed:
			progress.FailedNodes++
		case NodeRunning:
			progress.RunningNodes++
		case NodeSkipped:
			progress.SkippedNodes++
		}
	}

	if progress.TotalNodes > 0 {
		progress.CompletionPercent = (progress.CompletedNodes + progress.FailedNodes + progress.SkippedNodes) * 100 / progress.TotalNodes
	}

	return progress
}