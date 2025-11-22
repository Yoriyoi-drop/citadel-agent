package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sony/gobreaker"

	"github.com/citadel-agent/backend/internal/engine"
)

// CircuitBreakerNodeConfig represents the configuration for a Circuit Breaker node
type CircuitBreakerNodeConfig struct {
	Name             string                 `json:"name"`                    // Name of the circuit breaker
	MaxRequests      uint32                 `json:"max_requests"`            // Max requests allowed in half-open state
	Interval         time.Duration          `json:"interval"`                // Interval for steady state tests
	Timeout          time.Duration          `json:"timeout"`                 // Timeout for the circuit breaker
	ReadyToTrip      func(counts gobreaker.Counts) bool `json:"-"`          // Function to decide if circuit should trip
	OnStateChange    func(name string, from gobreaker.State, to gobreaker.State) `json:"-"` // Function to call when state changes
	BackOffPolicy    string                 `json:"backoff_policy"`          // "exponential", "linear", "fixed"
	InitialTimeout   time.Duration          `json:"initial_timeout"`         // Initial timeout for backoff
	MaxTimeout       time.Duration          `json:"max_timeout"`             // Maximum timeout for backoff
	FailureThreshold float64                `json:"failure_threshold"`       // Threshold for failure rate (0-1)
	MinRequests      uint32                 `json:"min_requests"`            // Minimum requests before circuit can trip
	TaskPayload      map[string]interface{} `json:"task_payload"`            // Payload for the task to execute
	Enabled          bool                   `json:"enabled"`                 // Whether the circuit breaker is enabled
	Description      string                 `json:"description"`             // Description of the circuit breaker
}

// CircuitBreakerNode represents a node that implements circuit breaker pattern
type CircuitBreakerNode struct {
	config  *CircuitBreakerNodeConfig
	cb      *gobreaker.CircuitBreaker
}

// NewCircuitBreakerNode creates a new Circuit Breaker node
func NewCircuitBreakerNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var cbConfig CircuitBreakerNodeConfig
	err = json.Unmarshal(jsonData, &cbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if cbConfig.Name == "" {
		cbConfig.Name = fmt.Sprintf("cb_%d", time.Now().UnixNano())
	}

	if cbConfig.FailureThreshold == 0 {
		cbConfig.FailureThreshold = 0.6 // default to 60% failure rate
	}

	if cbConfig.MinRequests == 0 {
		cbConfig.MinRequests = 5 // default to minimum 5 requests
	}

	if cbConfig.MaxRequests == 0 {
		cbConfig.MaxRequests = 3 // default max requests in half-open state
	}

	if cbConfig.Interval == 0 {
		cbConfig.Interval = 60 * time.Second // default 1 minute
	}

	if cbConfig.Timeout == 0 {
		cbConfig.Timeout = 10 * time.Second // default 10 seconds
	}

	// Create the circuit breaker settings
	settings := gobreaker.Settings{
		Name:        cbConfig.Name,
		MaxRequests: cbConfig.MaxRequests,
		Interval:    cbConfig.Interval,
		Timeout:     cbConfig.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= cbConfig.MinRequests && failureRatio >= cbConfig.FailureThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
			fmt.Printf("Circuit breaker %s changed from %s to %s\n", name, from, to)
		},
	}

	// Create the circuit breaker
	cb := gobreaker.NewCircuitBreaker(settings)

	return &CircuitBreakerNode{
		config: &cbConfig,
		cb:     cb,
	}, nil
}

// Execute implements the NodeInstance interface
func (c *CircuitBreakerNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	taskPayload := c.config.TaskPayload
	if inputPayload, ok := input["task_payload"].(map[string]interface{}); ok {
		taskPayload = inputPayload
	}

	enabled := c.config.Enabled
	if inputEnabled, ok := input["enabled"].(bool); ok {
		enabled = inputEnabled
	}

	// Check if circuit breaker should be enabled
	if !enabled {
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"message":   "circuit breaker disabled, executing directly",
				"enabled":   false,
				"task_executed": true,
				"result":    taskPayload, // Return the payload as result when disabled
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Execute the function with circuit breaker protection
	result, err := c.cb.Execute(func() (interface{}, error) {
		// Simulate the actual function we're protecting
		// In a real implementation, this would be the actual service call
		executionResult, executionErr := c.executeProtectedTask(taskPayload)
		return executionResult, executionErr
	})

	if err != nil {
		// Circuit breaker is open or the function failed
		return &engine.ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("circuit breaker error: %v", err),
			Data: map[string]interface{}{
				"circuit_state": string(c.cb.State()),
				"failures":      c.cb.Counts().TotalFailures,
				"successes":     c.cb.Counts().TotalSuccesses,
				"requests":      c.cb.Counts().Requests,
			},
			Timestamp: time.Now(),
		}, nil
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"message":       "task executed successfully",
			"circuit_state": string(c.cb.State()),
			"result":        result,
			"failures":      c.cb.Counts().TotalFailures,
			"successes":     c.cb.Counts().TotalSuccesses,
			"requests":      c.cb.Counts().Requests,
			"timestamp":     time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// executeProtectedTask simulates executing the protected task
func (c *CircuitBreakerNode) executeProtectedTask(payload map[string]interface{}) (interface{}, error) {
	// In a real implementation, this would execute the actual task based on the payload
	// For example, calling an external service, database, etc.
	
	// Simulate a task that can fail based on input
	if shouldFail, exists := payload["should_fail"]; exists {
		if fail, ok := shouldFail.(bool); ok && fail {
			return nil, fmt.Errorf("simulated failure based on payload")
		}
	}
	
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	
	// Return success result
	result := map[string]interface{}{
		"executed":  true,
		"timestamp": time.Now().Unix(),
		"input":     payload,
		"protected": true,
	}

	return result, nil
}

// GetType returns the type of the node
func (c *CircuitBreakerNode) GetType() string {
	return "circuit_breaker"
}

// GetID returns a unique ID for the node instance
func (c *CircuitBreakerNode) GetID() string {
	return "circuit_breaker_" + fmt.Sprintf("%d", time.Now().Unix())
}