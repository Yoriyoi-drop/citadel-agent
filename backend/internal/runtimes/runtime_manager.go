// citadel-agent/backend/internal/runtimes/runtime_manager.go
package runtimes

import (
	"context"
	"fmt"
	"time"
)

// RuntimeManager manages different language runtimes for node execution
type RuntimeManager struct {
	runtimes map[string]Runtime
}

// NewRuntimeManager creates a new runtime manager
func NewRuntimeManager() *RuntimeManager {
	rm := &RuntimeManager{
		runtimes: make(map[string]Runtime),
	}

	// Initialize default runtimes
	jsRuntime := &JavaScriptRuntime{name: "javascript", version: "ECMAScript 2023", status: "initialized"}
	pyRuntime := &PythonRuntime{name: "python", version: "3.9+", status: "initialized"}
	goRuntime := &GoRuntime{name: "go", version: "1.22", status: "initialized"}

	rm.runtimes["javascript"] = jsRuntime
	rm.runtimes["python"] = pyRuntime
	rm.runtimes["go"] = goRuntime

	return rm
}

// GetRuntime returns a specific runtime by name
func (rm *RuntimeManager) GetRuntime(language string) (Runtime, error) {
	runtime, exists := rm.runtimes[language]
	if !exists {
		return nil, fmt.Errorf("runtime for language '%s' not found", language)
	}

	return runtime, nil
}

// ExecuteCode executes code in the specified language runtime
func (rm *RuntimeManager) ExecuteCode(ctx context.Context, language string, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	runtime, err := rm.GetRuntime(language)
	if err != nil {
		return nil, err
	}

	// Validate the code first
	if err := runtime.ValidateCode(code); err != nil {
		return nil, fmt.Errorf("code validation failed: %w", err)
	}

	// Execute the code
	result, err := runtime.ExecuteCode(ctx, code, inputs, timeout)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return result, nil
}

// InitializeAll initializes all runtimes
func (rm *RuntimeManager) InitializeAll() error {
	for name, runtime := range rm.runtimes {
		if err := runtime.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize %s runtime: %w", name, err)
		}
	}
	return nil
}

// DisposeAll disposes all runtimes
func (rm *RuntimeManager) DisposeAll() error {
	for name, runtime := range rm.runtimes {
		if err := runtime.Dispose(); err != nil {
			fmt.Printf("Warning: failed to dispose %s runtime: %v\n", name, err)
		}
	}
	return nil
}

// ListRuntimes returns a list of available runtimes
func (rm *RuntimeManager) ListRuntimes() []string {
	runtimes := make([]string, 0, len(rm.runtimes))
	for name := range rm.runtimes {
		runtimes = append(runtimes, name)
	}
	return runtimes
}

// GetRuntimeInfo returns information about a specific runtime
func (rm *RuntimeManager) GetRuntimeInfo(language string) (RuntimeInfo, error) {
	runtime, err := rm.GetRuntime(language)
	if err != nil {
		return RuntimeInfo{}, err
	}

	return runtime.GetInfo(), nil
}

// ExecuteWithRuntime executes code in the specified language runtime
func (rm *RuntimeManager) ExecuteWithRuntime(ctx context.Context, language string, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	runtime, err := rm.GetRuntime(language)
	if err != nil {
		return nil, err
	}
	
	// Validate the code first
	if err := runtime.ValidateCode(code); err != nil {
		return nil, fmt.Errorf("code validation failed: %w", err)
	}
	
	// Execute the code
	startTime := time.Now()
	result, err := runtime.ExecuteCode(ctx, code, inputs, timeout)
	executionTime := time.Since(startTime)
	
	// Update stats
	rm.updateStats(runtime.GetInfo().Name, err, executionTime)
	
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	
	return result, nil
}

// updateStats updates runtime statistics
func (rm *RuntimeManager) updateStats(runtimeName string, err error, executionTime time.Duration) {
	// In a real implementation, this would update the runtime's stats
	// For now, just a placeholder
}