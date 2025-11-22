// workflow/core/engine/retry_circuit_breaker.go
package engine

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// RetryStrategy defines how retries should be handled
type RetryStrategy struct {
	MaxRetries      int           `json:"max_retries"`
	BackoffStrategy BackoffStrategy `json:"backoff_strategy"`
	Conditions      []RetryCondition `json:"conditions"`
	Timeout         time.Duration   `json:"timeout"`
}

// BackoffStrategy defines the backoff pattern for retries
type BackoffStrategy struct {
	Type       BackoffType `json:"type"`
	BaseDelay  time.Duration `json:"base_delay"`
	MaxDelay   time.Duration `json:"max_delay"`
	Jitter     bool        `json:"jitter"`
	Multiplier float64     `json:"multiplier"` // For exponential backoff
}

// BackoffType represents different backoff strategies
type BackoffType string

const (
	FixedBackoff     BackoffType = "fixed"
	ExponentialBackoff BackoffType = "exponential"
	LinearBackoff    BackoffType = "linear"
	RandomBackoff    BackoffType = "random"
)

// RetryCondition specifies when a retry should happen
type RetryCondition struct {
	ErrorMatch   string `json:"error_match"`   // Regexp pattern to match error
	StatusCode   []int `json:"status_code"`    // HTTP status codes to retry
	NodeTypes    []string `json:"node_types"` // Apply to specific node types
	MaxRetries   *int   `json:"max_retries,omitempty"` // Override global max retries for this condition
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Name           string        `json:"name"`
	MaxRequests    uint32        `json:"max_requests"`
	Interval       time.Duration `json:"interval"`
	Timeout        time.Duration `json:"timeout"`
	ReadyToTrip    func(counts gobreaker.Counts) bool `json:"-"`
	OnStateChange  func(name string, from, to gobreaker.State) `json:"-"`
}

// RetryManager manages retry logic
type RetryManager struct {
	strategies map[string]*RetryStrategy
	mutex      sync.RWMutex
	logger     Logger
}

// NewRetryManager creates a new retry manager
func NewRetryManager(logger Logger) *RetryManager {
	return &RetryManager{
		strategies: make(map[string]*RetryStrategy),
		logger:     logger,
	}
}

// AddRetryStrategy adds a retry strategy for a specific workflow or node type
func (rm *RetryManager) AddRetryStrategy(key string, strategy *RetryStrategy) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.strategies[key] = strategy
	
	if rm.logger != nil {
		rm.logger.Info("Added retry strategy for %s", key)
	}
}

// GetRetryStrategy retrieves a retry strategy
func (rm *RetryManager) GetRetryStrategy(key string) (*RetryStrategy, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	strategy, exists := rm.strategies[key]
	return strategy, exists
}

// ShouldRetry determines if a retry should be attempted
func (rm *RetryManager) ShouldRetry(ctx context.Context, strategy *RetryStrategy, attempt int, err error, nodeType string) (bool, time.Duration) {
	if attempt >= strategy.MaxRetries {
		return false, 0
	}

	// Check conditions
	for _, condition := range strategy.Conditions {
		// Check if this condition applies to this node type
		if len(condition.NodeTypes) > 0 {
			applies := false
			for _, allowedType := range condition.NodeTypes {
				if allowedType == nodeType {
					applies = true
					break
				}
			}
			if !applies {
				continue
			}
		}

		// Check error matching
		if condition.ErrorMatch != "" {
			// In a real implementation, we would match the error against a regexp
			// For now, we'll just check if error contains the string
			if err != nil && strings.Contains(err.Error(), condition.ErrorMatch) {
				maxRetries := condition.MaxRetries
				if maxRetries != nil && attempt >= *maxRetries {
					return false, 0
				}
				return true, rm.calculateBackoff(strategy.BackoffStrategy, attempt)
			}
		}
	}

	// Default behavior - retry on error if within limits
	if err != nil {
		return true, rm.calculateBackoff(strategy.BackoffStrategy, attempt)
	}

	return false, 0
}

// calculateBackoff calculates the delay before the next retry based on the strategy
func (rm *RetryManager) calculateBackoff(strategy BackoffStrategy, attempt int) time.Duration {
	var delay time.Duration

	switch strategy.Type {
	case FixedBackoff:
		delay = strategy.BaseDelay
	case LinearBackoff:
		delay = time.Duration(attempt+1) * strategy.BaseDelay
	case ExponentialBackoff:
		delay = time.Duration(float64(strategy.BaseDelay) * math.Pow(strategy.Multiplier, float64(attempt)))
		if strategy.MaxDelay > 0 && delay > strategy.MaxDelay {
			delay = strategy.MaxDelay
		}
	case RandomBackoff:
		base := float64(strategy.BaseDelay)
		max := float64(strategy.MaxDelay)
		if max == 0 {
			max = base * 10 // Default max if not specified
		}
		delay = time.Duration(base + (max-base)*rand.Float64())
	default:
		delay = strategy.BaseDelay
	}

	// Add jitter if requested
	if strategy.Jitter {
		jitter := rand.Float64() * 0.1 // Up to 10% jitter
		delay = time.Duration(float64(delay) * (1 + jitter))
	}

	return delay
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, strategyKey string, nodeType string, fn func() error) (err error) {
	strategy, exists := rm.GetRetryStrategy(strategyKey)
	if !exists {
		// Use default strategy
		strategy = &RetryStrategy{
			MaxRetries: 3,
			BackoffStrategy: BackoffStrategy{
				Type:      ExponentialBackoff,
				BaseDelay: 1 * time.Second,
				MaxDelay:  30 * time.Second,
				Multiplier: 2.0,
				Jitter:    true,
			},
			Conditions: []RetryCondition{
				{ErrorMatch: "timeout", NodeTypes: []string{}},
				{ErrorMatch: "connection refused", NodeTypes: []string{}},
				{ErrorMatch: "5xx", NodeTypes: []string{}},
			},
		}
	}

	attempt := 0
	for {
		// Create a context with timeout for this attempt
		timeoutCtx := ctx
		if strategy.Timeout > 0 {
			var cancel context.CancelFunc
			timeoutCtx, cancel = context.WithTimeout(ctx, strategy.Timeout)
			defer cancel()
		}

		err = fn()
		
		if err == nil {
			if rm.logger != nil {
				rm.logger.Info("Function succeeded on attempt %d", attempt+1)
			}
			return nil
		}

		retry, delay := rm.ShouldRetry(timeoutCtx, strategy, attempt, err, nodeType)
		if !retry {
			if rm.logger != nil {
				rm.logger.Info("Not retrying after %d attempts, last error: %v", attempt+1, err)
			}
			return err
		}

		if rm.logger != nil {
			rm.logger.Info("Attempt %d failed, retrying in %v: %v", attempt+1, delay, err)
		}

		// Wait for the specified delay or until context is cancelled
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}

		attempt++
	}
}

// CircuitBreakerManager manages circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mutex    sync.RWMutex
	configs  map[string]*CircuitBreakerConfig
	logger   Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		configs:  make(map[string]*CircuitBreakerConfig),
		logger:   logger,
	}
}

// AddCircuitBreaker adds a circuit breaker with the specified config
func (cm *CircuitBreakerManager) AddCircuitBreaker(name string, config *CircuitBreakerConfig) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	})

	cm.breakers[name] = cb
	cm.configs[name] = config
	
	if cm.logger != nil {
		cm.logger.Info("Added circuit breaker: %s", name)
	}
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (cm *CircuitBreakerManager) ExecuteWithCircuitBreaker(ctx context.Context, name string, fn func() (interface{}, error)) (interface{}, error) {
	cm.mutex.RLock()
	cb, exists := cm.breakers[name]
	cm.mutex.RUnlock()

	if !exists {
		// Create a default circuit breaker if one doesn't exist
		defaultConfig := &CircuitBreakerConfig{
			Name:        name,
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     60 * time.Second,
		}
		
		cm.AddCircuitBreaker(name, defaultConfig)
		
		cb = cm.breakers[name]
	}

	return cb.Execute(fn)
}

// IsCircuitOpen returns whether the circuit breaker for the given name is open
func (cm *CircuitBreakerManager) IsCircuitOpen(name string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cb, exists := cm.breakers[name]
	if !exists {
		return false
	}

	return cb.State() == gobreaker.StateOpen
}

// ResetCircuit resets the circuit breaker to closed state
func (cm *CircuitBreakerManager) ResetCircuit(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cb, exists := cm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	// We can't directly reset the circuit breaker, but we can create a new one
	config, configExists := cm.configs[name]
	if !configExists {
		return fmt.Errorf("circuit breaker config %s not found", name)
	}

	cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	})

	cm.breakers[name] = cb
	
	if cm.logger != nil {
		cm.logger.Info("Reset circuit breaker: %s", name)
	}

	return nil
}

// GetCircuitState returns the current state of the circuit breaker
func (cm *CircuitBreakerManager) GetCircuitState(name string) (gobreaker.State, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cb, exists := cm.breakers[name]
	if !exists {
		return gobreaker.State(0), fmt.Errorf("circuit breaker %s not found", name)
	}

	return cb.State(), nil
}

// Integration with the workflow engine
func (e *Engine) initializeRetryAndCircuitBreaker() {
	// Initialize retry manager
	e.retryManager = NewRetryManager(e.logger)
	
	// Initialize circuit breaker manager
	e.circuitBreakerManager = NewCircuitBreakerManager(e.logger)
	
	// Add default retry strategy
	defaultRetryStrategy := &RetryStrategy{
		MaxRetries: 3,
		BackoffStrategy: BackoffStrategy{
			Type:       ExponentialBackoff,
			BaseDelay:  1 * time.Second,
			MaxDelay:   30 * time.Second,
			Multiplier: 2.0,
			Jitter:     true,
		},
		Conditions: []RetryCondition{
			{
				ErrorMatch: "timeout",
				NodeTypes:  []string{"http_request", "database_query"},
			},
			{
				ErrorMatch: "connection refused",
				NodeTypes:  []string{"http_request", "database_query"},
			},
			{
				StatusCode: []int{500, 502, 503, 504},
				NodeTypes:  []string{"http_request"},
			},
		},
	}
	
	e.retryManager.AddRetryStrategy("default", defaultRetryStrategy)
	
	// Add default circuit breaker
	defaultCBConfig := &CircuitBreakerConfig{
		Name:        "default",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	}
	
	e.circuitBreakerManager.AddCircuitBreaker("default", defaultCBConfig)
	
	// Add circuit breakers for different node types
	for _, nodeType := range []string{"http_request", "database_query", "external_api"} {
		cbConfig := &CircuitBreakerConfig{
			Name:        fmt.Sprintf("node_%s", nodeType),
			MaxRequests: 3,
			Interval:    30 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 2
			},
		}
		e.circuitBreakerManager.AddCircuitBreaker(fmt.Sprintf("node_%s", nodeType), cbConfig)
	}
}

// ExecuteNodeWithRetryAndCircuitBreaker executes a node with retry and circuit breaker protection
func (e *Engine) ExecuteNodeWithRetryAndCircuitBreaker(ctx context.Context, execution *Execution, node *Node, workflow *Workflow) error {
	// Check if circuit breaker is open for this node type
	cbName := fmt.Sprintf("node_%s", node.Type)
	if e.circuitBreakerManager.IsCircuitOpen(cbName) {
		return fmt.Errorf("circuit breaker is open for node type %s, skipping execution", node.Type)
	}

	// Wrap node execution with retry and circuit breaker
	_, err := e.circuitBreakerManager.ExecuteWithCircuitBreaker(ctx, cbName, func() (interface{}, error) {
		return nil, e.retryManager.ExecuteWithRetry(ctx, "default", node.Type, func() error {
			return e.executeSingleNode(ctx, execution, node, workflow)
		})
	})

	return err
}