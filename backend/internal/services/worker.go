package services

import (
	"context"
	"log"
	"time"

	"citadel-agent/backend/internal/engine"
)

// Worker handles workflow execution jobs
type Worker struct {
	executionService *ExecutionService
	runner           *engine.Runner
	nodeRegistry     *engine.NodeRegistry
	jobQueue         chan *engine.Workflow
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewWorker creates a new worker instance
func NewWorker(executionService *ExecutionService, runner *engine.Runner, nodeRegistry *engine.NodeRegistry) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Worker{
		executionService: executionService,
		runner:           runner,
		nodeRegistry:     nodeRegistry,
		jobQueue:         make(chan *engine.Workflow, 100), // Buffered channel
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start starts the worker to process jobs
func (w *Worker) Start(ctx context.Context) error {
	log.Println("Starting worker...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker context cancelled")
			return ctx.Err()
		case workflow := <-w.jobQueue:
			go w.processWorkflow(workflow)
		}
	}
}

// SubmitWorkflow submits a workflow to be executed
func (w *Worker) SubmitWorkflow(workflow *engine.Workflow) {
	select {
	case w.jobQueue <- workflow:
		log.Printf("Workflow %s submitted to worker queue", workflow.ID)
	default:
		log.Printf("Worker queue is full, cannot submit workflow %s", workflow.ID)
	}
}

// processWorkflow processes a single workflow
func (w *Worker) processWorkflow(workflow *engine.Workflow) {
	log.Printf("Processing workflow: %s", workflow.ID)
	
	// Create execution context
	execution, err := w.runner.RunWorkflow(w.ctx, workflow, nil)
	if err != nil {
		log.Printf("Error running workflow %s: %v", workflow.ID, err)
		return
	}
	
	log.Printf("Workflow %s completed with status: %s", workflow.ID, execution.Status)
}

// Stop stops the worker gracefully
func (w *Worker) Stop(ctx context.Context) error {
	log.Println("Stopping worker...")
	
	// Cancel the worker context
	w.cancel()
	
	// Wait for any running operations to complete or context to timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second): // Wait max 10 seconds
		log.Println("Worker stopped with timeout")
		return nil
	}
}