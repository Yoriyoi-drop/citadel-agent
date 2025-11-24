package middleware

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay before first retry
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Backoff multiplier
	Jitter       bool          // Add random jitter to prevent thundering herd
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.Retryable
	}
	// By default, consider errors retryable
	return true
}

// Retry executes a function with exponential backoff retry logic
func Retry(fn func() error, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt, config)
		time.Sleep(delay)
	}

	return lastErr
}

// RetryWithContext executes a function with retry logic and context support
func RetryWithContext(fn func() error, config RetryConfig, shouldRetry func(error) bool) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if shouldRetry != nil && !shouldRetry(err) {
			return err
		}

		if attempt == config.MaxAttempts-1 {
			break
		}

		delay := calculateDelay(attempt, config)
		time.Sleep(delay)
	}

	return lastErr
}

// calculateDelay calculates the delay for the next retry with exponential backoff
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Calculate exponential backoff
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter if enabled
	if config.Jitter {
		jitter := rand.Float64() * delay * 0.1 // 10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

// RetryableFunc wraps a function with retry logic
type RetryableFunc struct {
	fn     func() error
	config RetryConfig
}

// NewRetryableFunc creates a new retryable function
func NewRetryableFunc(fn func() error, config RetryConfig) *RetryableFunc {
	return &RetryableFunc{
		fn:     fn,
		config: config,
	}
}

// Execute runs the function with retry logic
func (r *RetryableFunc) Execute() error {
	return Retry(r.fn, r.config)
}

// WithCircuitBreaker combines retry logic with circuit breaker
func WithCircuitBreaker(fn func() error, retryConfig RetryConfig, cb *CircuitBreaker) error {
	return Retry(func() error {
		return cb.Execute(fn)
	}, retryConfig)
}

// RetryStrategy defines different retry strategies
type RetryStrategy int

const (
	StrategyExponential RetryStrategy = iota
	StrategyLinear
	StrategyConstant
)

// AdvancedRetryConfig provides more control over retry behavior
type AdvancedRetryConfig struct {
	RetryConfig
	Strategy      RetryStrategy
	RetryableFunc func(error) bool
	OnRetry       func(attempt int, err error)
}

// AdvancedRetry executes with advanced retry configuration
func AdvancedRetry(fn func() error, config AdvancedRetryConfig) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if retryable
		if config.RetryableFunc != nil && !config.RetryableFunc(err) {
			return err
		}

		if attempt == config.MaxAttempts-1 {
			break
		}

		// Call retry callback
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		// Calculate delay based on strategy
		var delay time.Duration
		switch config.Strategy {
		case StrategyExponential:
			delay = calculateDelay(attempt, config.RetryConfig)
		case StrategyLinear:
			delay = config.InitialDelay * time.Duration(attempt+1)
		case StrategyConstant:
			delay = config.InitialDelay
		}

		time.Sleep(delay)
	}

	return lastErr
}
