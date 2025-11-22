package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// WorkerPoolNodeConfig represents the configuration for a Worker Pool node
type WorkerPoolNodeConfig struct {
	NumWorkers     int                    `json:"num_workers"`     // Number of workers in the pool
	WorkerType     string                 `json:"worker_type"`     // Type of worker
	QueueSize      int                    `json:"queue_size"`      // Size of the task queue
	MaxRetries     int                    `json:"max_retries"`     // Maximum number of retries per task
	Timeout        int                    `json:"timeout"`         // Timeout for each task in seconds
	TaskPayload    map[string]interface{} `json:"task_payload"`    // Default payload for tasks
	Enabled        bool                   `json:"enabled"`         // Whether the worker pool is enabled
	AutoScale      bool                   `json:"auto_scale"`      // Whether to auto-scale workers
	MaxWorkers     int                    `json:"max_workers"`     // Maximum number of workers for auto-scaling
	Description    string                 `json:"description"`     // Description of the worker pool
}

// Task represents a single task to be executed by a worker
type Task struct {
	ID      string
	Type    string
	Payload map[string]interface{}
	Retries int
	Result  chan TaskResult
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID string
	Error  error
	Data   interface{}
}

// WorkerPoolNode represents a node that manages a pool of workers
type WorkerPoolNode struct {
	config     *WorkerPoolNodeConfig
	taskQueue  chan Task
	workersWg  sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
	mutex      sync.RWMutex
}

// NewWorkerPoolNode creates a new Worker Pool node
func NewWorkerPoolNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var workerConfig WorkerPoolNodeConfig
	err = json.Unmarshal(jsonData, &workerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if workerConfig.NumWorkers <= 0 {
		workerConfig.NumWorkers = 1 // default to 1 worker
	}

	if workerConfig.QueueSize <= 0 {
		workerConfig.QueueSize = 10 // default queue size
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPoolNode{
		config:    &workerConfig,
		taskQueue: make(chan Task, workerConfig.QueueSize),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Execute implements the NodeInstance interface
func (w *WorkerPoolNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	numWorkers := w.config.NumWorkers
	if inputWorkers, ok := input["num_workers"].(float64); ok {
		numWorkers = int(inputWorkers)
		if numWorkers <= 0 {
			numWorkers = 1
		}
	}

	queueSize := w.config.QueueSize
	if inputQueueSize, ok := input["queue_size"].(float64); ok {
		queueSize = int(inputQueueSize)
		if queueSize <= 0 {
			queueSize = 10
		}
	}

	enabled := w.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	autoScale := w.config.AutoScale
	if inputAutoScale, ok := input["auto_scale"].(bool); ok {
		autoScale = inputAutoScale
	}

	maxWorkers := w.config.MaxWorkers
	if inputMaxWorkers, ok := input["max_workers"].(float64); ok {
		maxWorkers = int(inputMaxWorkers)
	}

	taskPayload := w.config.TaskPayload
	if inputPayload, ok := input["task_payload"].(map[string]interface{}); ok {
		taskPayload = inputPayload
	}

	// Check if worker pool should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "worker pool disabled, not started",
				"enabled": false,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Lock to prevent race conditions
	w.mutex.Lock()
	if w.isRunning {
		// If already running, just add tasks to the queue
		w.mutex.Unlock()
		
		// Add tasks to the queue if there are any in the input
		if len(taskPayload) > 0 {
			task := Task{
				ID:      fmt.Sprintf("task_%d", time.Now().UnixNano()),
				Type:    "default",
				Payload: taskPayload,
				Retries: 0,
				Result:  make(chan TaskResult, 1),
			}
			
			select {
			case w.taskQueue <- task:
				return &engine.ExecutionResult{
					Status: "success",
					Data: map[string]interface{}{
						"message":   "task added to worker pool queue",
						"task_id":   task.ID,
						"queue_len": len(w.taskQueue),
					},
					Timestamp: time.Now(),
				}, nil
			case <-time.After(5 * time.Second):
				return &engine.ExecutionResult{
					Status: "error",
					Error:  "worker pool queue is full, task not added",
					Timestamp: time.Now(),
				}, nil
			}
		}
		
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message": "worker pool already running",
				"enabled": true,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Start the worker pool
	w.isRunning = true
	w.mutex.Unlock()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		w.workersWg.Add(1)
		go w.worker(i)
	}

	// If there's a task payload in the input, create and add a task
	if len(taskPayload) > 0 {
		task := Task{
			ID:      fmt.Sprintf("task_%d", time.Now().UnixNano()),
			Type:    "default",
			Payload: taskPayload,
			Retries: 0,
			Result:  make(chan TaskResult, 1),
		}
		
		select {
		case w.taskQueue <- task:
			// Task added successfully, wait for result or timeout
			select {
			case result := <-task.Result:
				if result.Error != nil {
					return &engine.ExecutionResult{
						Status:    "error",
						Error:     result.Error.Error(),
						Timestamp: time.Now(),
					}, nil
				}
				return &engine.ExecutionResult{
					Status: "success",
					Data: map[string]interface{}{
						"message": "task completed",
						"task_id": result.TaskID,
						"result":  result.Data,
					},
					Timestamp: time.Now(),
				}, nil
			case <-time.After(30 * time.Second): // Wait up to 30 seconds for result
				return &engine.ExecutionResult{
					Status: "error",
					Error:  "task timeout waiting for result",
					Timestamp: time.Now(),
				}, nil
			}
		case <-time.After(5 * time.Second):
			return &engine.ExecutionResult{
				Status: "error",
				Error:  "worker pool queue is full, task not added",
				Timestamp: time.Now(),
			}, nil
		}
	}

	// Return success indicating the worker pool is running
	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":      "worker pool started",
			"num_workers":  numWorkers,
			"queue_size":   queueSize,
			"auto_scale":   autoScale,
			"max_workers":  maxWorkers,
			"is_running":   true,
			"timestamp":    time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// worker is the function that runs in each worker goroutine
func (w *WorkerPoolNode) worker(workerID int) {
	defer w.workersWg.Done()

	for {
		select {
		case task := <-w.taskQueue:
			// Execute the task
			result := w.executeTask(task)
			
			// Send the result back
			select {
			case task.Result <- result:
				// Result sent successfully
			case <-time.After(5 * time.Second):
				// Result channel is not being received, log the issue
				fmt.Printf("Warning: Task result not received for task %s\n", task.ID)
			}
			
		case <-w.ctx.Done():
			// Context cancelled, worker should stop
			return
		}
	}
}

// executeTask executes a single task
func (w *WorkerPoolNode) executeTask(task Task) TaskResult {
	// Simulate task execution
	// In a real implementation, this would execute the actual task based on its type
	resultData := map[string]interface{}{
		"worker_id":    task.ID,
		"task_id":      task.ID,
		"executed_at":  time.Now().Unix(),
		"input_data":   task.Payload,
		"task_type":    task.Type,
	}

	// Add any processing delay based on the payload or task type
	processingDelay := time.Duration(100) * time.Millisecond // Default processing delay
	if delayInterface, exists := task.Payload["processing_delay"]; exists {
		if delayFloat, ok := delayInterface.(float64); ok {
			processingDelay = time.Duration(delayFloat) * time.Millisecond
		}
	}
	
	time.Sleep(processingDelay)

	return TaskResult{
		TaskID: task.ID,
		Error:  nil,
		Data:   resultData,
	}
}

// GetType returns the type of the node
func (w *WorkerPoolNode) GetType() string {
	return "worker_pool"
}

// GetID returns a unique ID for the node instance
func (w *WorkerPoolNode) GetID() string {
	return "worker_pool_" + fmt.Sprintf("%d", time.Now().Unix())
}