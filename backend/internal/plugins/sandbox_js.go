package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

// JavascriptSandbox provides a secure environment for executing JavaScript code
type JavascriptSandbox struct {
	defaultTimeout time.Duration
	vmPool         chan *otto.Otto
	maxVMs         int
}

// NewJavascriptSandbox creates a new JavaScript sandbox
func NewJavascriptSandbox(timeout time.Duration, maxVMs int) *JavascriptSandbox {
	sb := &JavascriptSandbox{
		defaultTimeout: timeout,
		vmPool:         make(chan *otto.Otto, maxVMs),
		maxVMs:         maxVMs,
	}

	// Pre-populate the VM pool
	for i := 0; i < maxVMs; i++ {
		vm := otto.New()
		sb.vmPool <- vm
	}

	return sb
}

// Execute executes JavaScript code in a sandboxed environment
func (sb *JavascriptSandbox) Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error) {
	startTime := time.Now()

	// Get a VM from the pool (with timeout to avoid blocking indefinitely)
	vmCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var vm *otto.Otto
	select {
	case vm = <-sb.vmPool:
		// Got a VM from the pool
	case <-vmCtx.Done():
		return &PluginResult{
			Success: false,
			Error:   "Timeout getting VM from pool",
		}, nil
	}

	// Ensure VM is returned to the pool when done
	defer func() {
		// Clear the VM to reset its state
		vm = otto.New()
		select {
		case sb.vmPool <- vm:
			// VM returned to pool
		default:
			// Pool is full, just discard the VM
		}
	}()

	// Set up interrupt mechanism for timeout
	done := make(chan struct{})
	defer close(done)

	// Set timeout for execution
	execCtx, cancel := context.WithTimeout(ctx, sb.defaultTimeout)
	defer cancel()

	// Create a channel to receive the result
	resultChan := make(chan *PluginResult, 1)

	// Run execution in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- &PluginResult{
					Success: false,
					Error:   fmt.Sprintf("Panic during execution: %v", r),
				}
			}
		}()

		// Add input data to the VM
		for k, v := range input {
			err := vm.Set(k, v)
			if err != nil {
				resultChan <- &PluginResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to set input variable %s: %v", k, err),
				}
				return
			}
		}

		// Execute the code
		value, err := vm.Run(code)
		if err != nil {
			resultChan <- &PluginResult{
				Success: false,
				Error:   fmt.Sprintf("JavaScript execution error: %v", err),
			}
			return
		}

		// Get the result
		result, err := value.Export()
		if err != nil {
			resultChan <- &PluginResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to export result: %v", err),
			}
			return
		}

		resultChan <- &PluginResult{
			Success:  true,
			Data:     result,
			ExecTime: time.Since(startTime),
		}
	}()

	// Wait for the result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-execCtx.Done():
		// Interrupt the VM
		vm.Interrupt <- func() {
			// This will cause the VM to panic and be recovered in the goroutine
			panic("Execution timeout")
		}
		
		// Wait a bit more for the interrupt to take effect
		select {
		case result := <-resultChan:
			return result, nil
		case <-time.After(100 * time.Millisecond):
			return &PluginResult{
				Success: false,
				Error:   "Code execution timed out and could not be interrupted",
			}, nil
		}
	}
}

// ValidateCode validates JavaScript code for security
func (sb *JavascriptSandbox) ValidateCode(code string) error {
	// In a real implementation, you would validate the code for
	// potentially unsafe operations like:
	// - Accessing global objects that could be dangerous
	// - Making network requests
	// - Accessing the file system
	//
	// For this example, we'll do a basic check for dangerous patterns
	dangerousPatterns := []string{
		"require(",      // Node.js require
		"import(",       // Import statements
		"eval(",         // Eval function
		"Function(",     // Function constructor
		"process.",      // Node.js process object
		"global.",       // Global object
		"window.",       // Browser window object
		"document.",     // Browser document object
		"XMLHttpRequest", // Browser AJAX
		"fetch(",        // Fetch API
		"File",          // File API
	}

	for _, pattern := range dangerousPatterns {
		if contains(code, pattern) {
			return fmt.Errorf("code contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    len(str) > len(substr) && 
		    (str[:len(substr)] == substr || 
		     str[len(str)-len(substr):] == substr ||
		     contains(str[1:], substr)))
}