// backend/internal/runtimes/multi_runtime.go
package runtimes

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// MultiRuntimeManager manages multiple language runtimes
type MultiRuntimeManager struct {
	runtimes map[RuntimeType]Runtime
}

// NewMultiRuntimeManager creates a new multi-runtime manager
func NewMultiRuntimeManager() *MultiRuntimeManager {
	manager := &MultiRuntimeManager{
		runtimes: make(map[RuntimeType]Runtime),
	}

	// Initialize all runtime managers
	manager.runtimes[RuntimeGo] = &GoRuntime{}
	manager.runtimes[RuntimeJS] = &JavaScriptRuntime{}
	manager.runtimes[RuntimePython] = &PythonRuntime{}
	manager.runtimes[RuntimeJava] = &JavaRuntime{}
	manager.runtimes[RuntimeRuby] = &RubyRuntime{}
	manager.runtimes[RuntimePHP] = &PHPRuntime{}
	manager.runtimes[RuntimeRust] = &RustRuntime{}
	manager.runtimes[RuntimeCSharp] = &CSharpRuntime{}
	manager.runtimes[RuntimeShell] = &ShellRuntime{}

	return manager
}

// ExecuteCode executes code in the specified runtime
func (mrm *MultiRuntimeManager) ExecuteCode(ctx context.Context, runtimeType RuntimeType, code string, input map[string]interface{}, timeout time.Duration) (*interfaces.ExecutionResult, error) {
	runtime, exists := mrm.runtimes[runtimeType]
	if !exists {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("Runtime type %s not supported", runtimeType),
			Timestamp: time.Now(),
		}, nil
	}

	// Validate the code first
	if err := runtime.ValidateCode(code); err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("Code validation failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	// Execute the code with the provided timeout
	result, err := runtime.ExecuteCode(ctx, code, input, timeout)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("Execution failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &interfaces.ExecutionResult{
		Status:    "success",
		Data:      result,
		Timestamp: time.Now(),
	}, nil
}

// ValidateCode validates code for a specific runtime
func (mrm *MultiRuntimeManager) ValidateCode(runtimeType RuntimeType, code string) error {
	runtime, exists := mrm.runtimes[runtimeType]
	if !exists {
		return fmt.Errorf("runtime type %s not supported", runtimeType)
	}

	return runtime.ValidateCode(code)
}

// GetRuntimeInfo returns information about a specific runtime
func (mrm *MultiRuntimeManager) GetRuntimeInfo(runtimeType RuntimeType) (RuntimeInfo, error) {
	runtime, exists := mrm.runtimes[runtimeType]
	if !exists {
		return RuntimeInfo{}, fmt.Errorf("runtime type %s not found", runtimeType)
	}

	return runtime.GetInfo(), nil
}

// ListRuntimes returns a list of available runtimes
func (mrm *MultiRuntimeManager) ListRuntimes() []RuntimeType {
	runtimes := make([]RuntimeType, 0, len(mrm.runtimes))
	for runtimeType := range mrm.runtimes {
		runtimes = append(runtimes, runtimeType)
	}
	return runtimes
}

// ExecuteWithRuntime executes code in a specific runtime - convenience method
func (mrm *MultiRuntimeManager) ExecuteWithRuntime(ctx context.Context, language string, code string, input map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Map language string to RuntimeType
	var runtimeType RuntimeType
	switch language {
	case "go", "golang":
		runtimeType = RuntimeGo
	case "javascript", "js", "node":
		runtimeType = RuntimeJS
	case "python", "py":
		runtimeType = RuntimePython
	case "java":
		runtimeType = RuntimeJava
	case "ruby", "rb":
		runtimeType = RuntimeRuby
	case "php":
		runtimeType = RuntimePHP
	case "rust", "rs":
		runtimeType = RuntimeRust
	case "csharp", "c#", "cs":
		runtimeType = RuntimeCSharp
	case "shell", "bash", "sh":
		runtimeType = RuntimeShell
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	result, err := mrm.ExecuteCode(ctx, runtimeType, code, input, timeout)
	if err != nil {
		return nil, err
	}

	if result.Status == "error" {
		return nil, fmt.Errorf("execution error: %s", result.Error)
	}

	return result.Data.(map[string]interface{}), nil
}

// InitializeAll initializes all available runtimes
func (mrm *MultiRuntimeManager) InitializeAll() error {
	for _, runtime := range mrm.runtimes {
		if err := runtime.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize runtime: %v", err)
		}
	}
	return nil
}

// DisposeAll disposes all available runtimes
func (mrm *MultiRuntimeManager) DisposeAll() error {
	for _, runtime := range mrm.runtimes {
		if err := runtime.Dispose(); err != nil {
			// Don't fail on individual dispose errors, just log them
			fmt.Printf("Warning: failed to dispose runtime: %v\n", err)
		}
	}
	return nil
}