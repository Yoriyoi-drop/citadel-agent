package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// TaskQueueNodeConfig represents the configuration for a Task Queue node
type TaskQueueNodeConfig struct {
	RedisAddr      string                 `json:"redis_addr"`       // Redis address for asynq
	QueueName      string                 `json:"queue_name"`       // Name of the queue
	TaskType       string                 `json:"task_type"`        // Type of task to enqueue
	TaskPayload    map[string]interface{} `json:"task_payload"`     // Payload for the task
	MaxRetry       int                    `json:"max_retry"`        // Maximum number of retries
	Timeout        int                    `json:"timeout"`          // Task timeout in seconds
	Deadline       string                 `json:"deadline"`         // Task deadline in RFC3339 format
	UniqueKey      string                 `json:"unique_key"`       // Unique key to prevent duplicate tasks
	Priority       int                    `json:"priority"`         // Task priority (0-9, 9 is highest)
	TaskGroup      string                 `json:"task_group"`       // Task group for ordered processing
	ScheduleAt     string                 `json:"schedule_at"`      // Schedule task at specific time
	ProcessNow     bool                   `json:"process_now"`      // Whether to process task immediately
	Options        map[string]interface{} `json:"options"`          // Additional options for the task
}

// TaskQueueNode represents a node that handles task queues using asynq
type TaskQueueNode struct {
	config *TaskQueueNodeConfig
	client *asynq.Client
	server *asynq.Server
}

// NewTaskQueueNode creates a new Task Queue node
func NewTaskQueueNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var taskQueueConfig TaskQueueNodeConfig
	err = json.Unmarshal(jsonData, &taskQueueConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if taskQueueConfig.RedisAddr == "" {
		return nil, fmt.Errorf("redis_addr is required for Task Queue node")
	}

	if taskQueueConfig.TaskType == "" {
		taskQueueConfig.TaskType = "default_task" // default task type
	}

	// Create asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: taskQueueConfig.RedisAddr})

	return &TaskQueueNode{
		config: &taskQueueConfig,
		client: client,
	}, nil
}

// Execute implements the NodeInstance interface
func (t *TaskQueueNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	taskType := t.config.TaskType
	if inputTaskType, ok := input["task_type"].(string); ok && inputTaskType != "" {
		taskType = inputTaskType
	}

	taskPayload := t.config.TaskPayload
	if inputPayload, ok := input["task_payload"].(map[string]interface{}); ok {
		taskPayload = inputPayload
	}

	queueName := t.config.QueueName
	if inputQueue, ok := input["queue_name"].(string); ok && inputQueue != "" {
		queueName = inputQueue
	}

	maxRetry := t.config.MaxRetry
	if inputRetry, ok := input["max_retry"].(float64); ok {
		maxRetry = int(inputRetry)
	}

	timeout := t.config.Timeout
	if inputTimeout, ok := input["timeout"].(float64); ok {
		timeout = int(inputTimeout)
	}

	priority := t.config.Priority
	if inputPriority, ok := input["priority"].(float64); ok {
		priority = int(inputPriority)
	}

	uniqueKey := t.config.UniqueKey
	if inputKey, ok := input["unique_key"].(string); ok {
		uniqueKey = inputKey
	}

	scheduleAt := t.config.ScheduleAt
	if inputSchedule, ok := input["schedule_at"].(string); ok {
		scheduleAt = inputSchedule
	}

	processNow := t.config.ProcessNow
	if inputProcessNow, ok := input["process_now"].(bool); ok {
		processNow = inputProcessNow
	}

	// Create task
	task := asynq.NewTask(taskType, taskPayload)

	// Build options for the task
	var opts []asynq.Option

	if maxRetry > 0 {
		opts = append(opts, asynq.MaxRetry(maxRetry))
	}

	if timeout > 0 {
		opts = append(opts, asynq.Timeout(time.Duration(timeout)*time.Second))
	}

	if priority > 0 {
		opts = append(opts, asynq.MaxRetry(priority))
	}

	if queueName != "" {
		opts = append(opts, asynq.Queue(queueName))
	}

	if uniqueKey != "" {
		opts = append(opts, asynq.Unique(time.Hour)) // Keep unique for 1 hour
	}

	// Enqueue the task
	var info *asynq.TaskInfo
	var err error

	if scheduleAt != "" {
		// Schedule the task for later execution
		schedTime, parseErr := time.Parse(time.RFC3339, scheduleAt)
		if parseErr != nil {
			return map[string]interface{}{
				"status":    "error",
				"error":     fmt.Sprintf("failed to parse schedule time: %v", parseErr),
				"timestamp": time.Now().Unix(),
			}, nil
		}
		
		info, err = t.client.EnqueueAt(schedTime, task, opts...)
	} else if !processNow {
		// Enqueue immediately for processing
		info, err = t.client.Enqueue(task, opts...)
	} else {
		// Just return info about the task without enqueuing
		return map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"message":      "task ready to be enqueued",
				"task_type":    taskType,
				"task_payload": taskPayload,
				"queue":        queueName,
				"options":      opts,
				"timestamp":    time.Now().Unix(),
			},
			"timestamp": time.Now().Unix(),
		}, nil
	}

	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"error":     fmt.Sprintf("failed to enqueue task: %v", err),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"message":      "task enqueued successfully",
			"task_id":      info.ID,
			"task_type":    taskType,
			"queue":        info.Queue,
			"status":       string(info.State),
			"task_payload": taskPayload,
			"timestamp":    time.Now().Unix(),
		},
		"timestamp": time.Now().Unix(),
	}, nil
}

// GetType returns the type of the node
func (t *TaskQueueNode) GetType() string {
	return "task_queue"
}

// GetID returns a unique ID for the node instance
func (t *TaskQueueNode) GetID() string {
	return "task_queue_" + fmt.Sprintf("%d", time.Now().Unix())
}