package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

// PluginType defines the type of plugin
type PluginType string

const (
	JavascriptPlugin PluginType = "javascript"
	PythonPlugin     PluginType = "python"
)

// Plugin represents a single plugin
type Plugin struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        PluginType  `json:"type"`
	Code        string      `json:"code"`
	Schema      interface{} `json:"schema"` // JSON schema for plugin configuration
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// PluginResult represents the result of plugin execution
type PluginResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PluginExecutor defines the interface for executing plugins
type PluginExecutor interface {
	Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error)
}

// JavascriptExecutor executes JavaScript code in a sandbox
type JavascriptExecutor struct {
	timeout time.Duration
}

// NewJavascriptExecutor creates a new JavaScript executor
func NewJavascriptExecutor(timeout time.Duration) *JavascriptExecutor {
	return &JavascriptExecutor{
		timeout: timeout,
	}
}

// Execute executes JavaScript code in a sandbox
func (j *JavascriptExecutor) Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error) {
	// Create a new VM
	vm := otto.New()
	
	// Set a timeout for execution
	done := make(chan bool, 1)
	go func() {
		select {
		case <-time.After(j.timeout):
			vm.Interrupt <- func() {
				panic("Execution timeout")
			}
		case <-done:
			// Execution completed within time
		}
	}()

	// Add input data to the VM
	for k, v := range input {
		err := vm.Set(k, v)
		if err != nil {
			done <- true
			return &PluginResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to set input variable %s: %v", k, err),
			}, nil
		}
	}

	// Execute the code
	value, err := vm.Run(code)
	done <- true
	
	if err != nil {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("JavaScript execution error: %v", err),
		}, nil
	}

	// Get the result
	result, err := value.Export()
	if err != nil {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to export result: %v", err),
		}, nil
	}

	return &PluginResult{
		Success: true,
		Data:    result,
	}, nil
}

// PythonExecutor executes Python code in a sandbox
// Note: This is a simplified version. In a real implementation,
// you would need to call out to a Python process with proper sandboxing.
type PythonExecutor struct {
	timeout time.Duration
}

// NewPythonExecutor creates a new Python executor
func NewPythonExecutor(timeout time.Duration) *PythonExecutor {
	return &PythonExecutor{
		timeout: timeout,
	}
}

// Execute executes Python code in a sandbox
func (p *PythonExecutor) Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error) {
	// In a real implementation, this would involve:
	// 1. Writing the code and input to a temporary file
	// 2. Calling a Python subprocess with proper sandboxing
	// 3. Reading the result from the subprocess
	//
	// For this example, we'll simulate execution
	fmt.Printf("Executing Python code: %s\nWith input: %+v\n", code, input)
	
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	
	// For this example, return a success result
	return &PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"result": "Python code executed successfully",
			"input":  input,
		},
	}, nil
}

// PluginManager manages different types of plugins
type PluginManager struct {
	jsExecutor  *JavascriptExecutor
	pyExecutor  *PythonExecutor
	plugins     map[string]*Plugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(jsTimeout, pyTimeout time.Duration) *PluginManager {
	return &PluginManager{
		jsExecutor: NewJavascriptExecutor(jsTimeout),
		pyExecutor: NewPythonExecutor(pyTimeout),
		plugins:    make(map[string]*Plugin),
	}
}

// RegisterPlugin registers a new plugin
func (pm *PluginManager) RegisterPlugin(plugin *Plugin) {
	pm.plugins[plugin.ID] = plugin
}

// ExecutePlugin executes a registered plugin
func (pm *PluginManager) ExecutePlugin(ctx context.Context, pluginID string, input map[string]interface{}) (*PluginResult, error) {
	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Plugin with ID %s not found", pluginID),
		}, nil
	}

	switch plugin.Type {
	case JavascriptPlugin:
		return pm.jsExecutor.Execute(ctx, plugin.Code, input)
	case PythonPlugin:
		return pm.pyExecutor.Execute(ctx, plugin.Code, input)
	default:
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported plugin type: %s", plugin.Type),
		}, nil
	}
}

// ExecuteCode executes arbitrary code in the appropriate sandbox
func (pm *PluginManager) ExecuteCode(ctx context.Context, pluginType PluginType, code string, input map[string]interface{}) (*PluginResult, error) {
	switch pluginType {
	case JavascriptPlugin:
		return pm.jsExecutor.Execute(ctx, code, input)
	case PythonPlugin:
		return pm.pyExecutor.Execute(ctx, code, input)
	default:
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported plugin type: %s", pluginType),
		}, nil
	}
}