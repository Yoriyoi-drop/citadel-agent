// backend/internal/engine/runner.go
package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/google/uuid"
)

// Runner manages the execution of workflows
type Runner struct {
	executor    *Executor
	workflowManager *WorkflowManager
	aiManager   interfaces.AIManagerInterface // Use interface instead of concrete type
}

// NewRunner creates a new instance of Runner
func NewRunner(executor *Executor, workflowMgr *WorkflowManager, aiManager interfaces.AIManagerInterface) *Runner {
	return &Runner{
		executor:    executor,
		workflowManager: workflowMgr,
		aiManager:   aiManager,
	}
}

// RunWorkflow runs a workflow with the given inputs
func (r *Runner) RunWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) (*Execution, error) {
	// Retrieve workflow from manager
	workflow, err := r.workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Create execution record
	executionID := uuid.New()
	execution := &Execution{
		ID:          executionID.String(),
		WorkflowID:  workflow.ID,
		Status:      "running",
		StartedAt:   time.Now(),
		Inputs:      inputs,
		NodeResults: make(map[string]interface{}),
	}

	// Trigger the workflow execution
	go r.executeWorkflow(ctx, workflow, execution)

	return execution, nil
}

// executeWorkflow executes the workflow in the background
func (r *Runner) executeWorkflow(ctx context.Context, workflow *Workflow, execution *Execution) {
	log.Printf("Starting execution of workflow %s with ID %s", workflow.ID, execution.ID)

	// Create dependency resolver
	depResolver := NewDependencyResolver(workflow.Nodes, workflow.Edges)

	// Validate the workflow for cycles and other issues
	if err := depResolver.ValidateWorkflow(); err != nil {
		r.updateExecutionStatus(execution, "failed", fmt.Sprintf("workflow validation error: %v", err))
		return
	}

	// Resolve execution order using topological sort
	executionOrder, err := depResolver.ResolveExecutionOrder()
	if err != nil {
		r.updateExecutionStatus(execution, "failed", fmt.Sprintf("dependency resolution error: %v", err))
		return
	}

	// Keep track of results from each node
	nodeResults := make(map[string]interface{})
	var mu sync.Mutex

	// Execute nodes in the resolved order
	wg := sync.WaitGroup{}
	errorChan := make(chan error, 1)
	executionCtx, cancel := context.WithCancel(ctx)

	defer cancel()

	for _, nodeID := range executionOrder {
		wg.Add(1)
		go func(nodeID string) {
			defer wg.Done()

			node, exists := workflow.GetNode(nodeID)
			if !exists {
				select {
				case errorChan <- fmt.Errorf("node %s not found in workflow", nodeID):
				default:
				}
				return
			}

			// Check if execution was cancelled
			select {
			case <-executionCtx.Done():
				return
			default:
			}

			// Prepare input data for the node
			inputData, err := r.prepareNodeInput(node, nodeResults, execution.Inputs)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to prepare input for node %s: %v", nodeID, err):
				default:
				}
				return
			}

			// Execute the node
			result, err := r.executor.ExecuteNode(executionCtx, node.Type, inputData)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to execute node %s: %v", nodeID, err):
				default:
				}
				return
			}

			// Store result
			mu.Lock()
			nodeResults[nodeID] = result.Data
			mu.Unlock()

			// Update execution status with node result
			r.updateNodeResult(execution, nodeID, result)
		}(nodeID)
	}

	// Wait for all nodes to complete or an error to occur
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Wait for error or completion
	select {
	case err, ok := <-errorChan:
		if ok && err != nil {
			r.updateExecutionStatus(execution, "failed", err.Error())
			cancel()
			return
		}
	case <-executionCtx.Done():
		// Context cancelled
		r.updateExecutionStatus(execution, "cancelled", "execution cancelled")
		return
	}

	r.updateExecutionStatus(execution, "completed", "execution completed successfully")
}

// prepareNodeInput prepares input data for a node based on its dependencies
func (r *Runner) prepareNodeInput(node *Node, nodeResults map[string]interface{}, globalInputs map[string]interface{}) (map[string]interface{}, error) {
	inputData := make(map[string]interface{})

	// Add global inputs
	for k, v := range globalInputs {
		inputData[k] = v
	}

	// Add inputs from dependency nodes
	for _, dependencyNodeID := range node.Inputs {
		if result, exists := nodeResults[dependencyNodeID]; exists {
			// Add the result from the dependency node
			dependencyKey := fmt.Sprintf("input_from_%s", dependencyNodeID)
			inputData[dependencyKey] = result
		}
	}

	// Add node-specific settings
	for k, v := range node.Settings {
		inputData[k] = v
	}

	return inputData, nil
}

// updateExecutionStatus updates the status of an execution
func (r *Runner) updateExecutionStatus(execution *Execution, status, message string) {
	execution.Status = status
	execution.Message = message
	execution.EndedAt = time.Now()

	// In a real implementation, this would update the database
	// For now, we'll just log the status change
	log.Printf("Execution %s status updated to: %s - %s", execution.ID, status, message)
}

// updateNodeResult updates the result of a node execution
func (r *Runner) updateNodeResult(execution *Execution, nodeID string, result *ExecutionResult) {
	execution.NodeResults[nodeID] = result

	// In a real implementation, this might send updates to a WebSocket or event bus
	// For now, log the node completion
	log.Printf("Node %s in execution %s completed with status: %s", nodeID, execution.ID, result.Status)
}

// StartContinuousWorkflow starts a continuous workflow that runs indefinitely
func (r *Runner) StartContinuousWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) error {
	workflow, err := r.workflowManager.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	if !workflow.IsContinuous {
		return errors.New("workflow is not configured as continuous")
	}

	// Run the workflow continuously in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Continuous workflow %s stopped", workflowID)
				return
			default:
				// Execute the workflow
				_, err := r.RunWorkflow(ctx, workflowID, inputs)
				if err != nil {
					log.Printf("Error running continuous workflow %s: %v", workflowID, err)
					// Wait before retrying
					time.Sleep(5 * time.Second)
					continue
				}

				// Wait based on the workflow's schedule configuration
				if workflow.Schedule.Interval > 0 {
					time.Sleep(time.Duration(workflow.Schedule.Interval) * time.Second)
				} else {
					// Default to 30 seconds if no interval specified
					time.Sleep(30 * time.Second)
				}
			}
		}
	}()

	return nil
}

// ExecuteWithTimeout executes a workflow with a timeout
func (r *Runner) ExecuteWithTimeout(ctx context.Context, workflowID string, inputs map[string]interface{}, timeout time.Duration) (*Execution, error) {
	executionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	execution, err := r.RunWorkflow(executionCtx, workflowID, inputs)
	if err != nil {
		return nil, err
	}

	// Wait for execution to complete or timeout
	select {
	case <-executionCtx.Done():
		if errors.Is(executionCtx.Err(), context.DeadlineExceeded) {
			r.updateExecutionStatus(execution, "timeout", "execution exceeded timeout limit")
			return execution, nil
		}
	}

	return execution, nil
}

// StopExecution stops a running workflow execution
func (r *Runner) StopExecution(executionID string) error {
	// In a real implementation, this would stop a running workflow
	// by cancelling its context and updating its status
	// For now, we'll just log the action
	log.Printf("Stopping execution: %s", executionID)
	
	// Update execution status to cancelled
	// This would typically involve storing this in the database
	// and cancelling any running goroutines associated with the execution
	
	return nil
}