// backend/internal/security/sandbox.go
package security

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/models"
)

// RuntimeSandbox provides sandboxing for code execution
type RuntimeSandbox struct {
	config *SandboxConfig
}

// SandboxConfig holds configuration for the sandbox
type SandboxConfig struct {
	MaxExecutionTime    time.Duration
	MaxMemory           int64
	MaxOutputLength     int
	AllowedHosts        []string
	BlockedPaths        []string
	AllowedCapabilities []string
	EnableNetwork       bool
	EnableFileAccess    bool
}

// NewRuntimeSandbox creates a new runtime sandbox
func NewRuntimeSandbox(config *SandboxConfig) *RuntimeSandbox {
	if config == nil {
		config = &SandboxConfig{
			MaxExecutionTime: 30 * time.Second,
			MaxMemory:        100 * 1024 * 1024, // 100MB
			MaxOutputLength:  10000,
			AllowedHosts:     []string{"api.github.com", "api.openai.com"},
			BlockedPaths:     []string{"/etc/", "/proc/", "/sys/"},
			EnableNetwork:    false,
			EnableFileAccess: false,
		}
	}

	return &RuntimeSandbox{
		config: config,
	}
}

// ExecuteCode executes code in a secure sandbox
func (rs *RuntimeSandbox) ExecuteCode(ctx context.Context, code string, runtimeType string, inputs map[string]interface{}) (*ExecutionResult, error) {
	// Validate code before execution
	if err := rs.validateCode(code, runtimeType); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Code validation failed: %v", err),
		}, nil
	}

	// Check if runtime type is allowed
	if !rs.isRuntimeAllowed(runtimeType) {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Runtime type %s is not allowed", runtimeType),
		}, nil
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, rs.config.MaxExecutionTime)
	defer cancel()

	// Execute based on runtime type
	switch runtimeType {
	case "javascript":
		return rs.executeJavaScript(execCtx, code, inputs)
	case "python":
		return rs.executePython(execCtx, code, inputs)
	case "go":
		return rs.executeGo(execCtx, code, inputs)
	case "shell":
		return rs.executeShell(execCtx, code, inputs)
	default:
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported runtime type: %s", runtimeType),
		}, nil
	}
}

// validateCode validates code for security issues
func (rs *RuntimeSandbox) validateCode(code, runtimeType string) error {
	switch runtimeType {
	case "javascript":
		return rs.validateJavaScript(code)
	case "python":
		return rs.validatePython(code)
	case "shell":
		return rs.validateShell(code)
	default:
		return nil
	}
}

// validateJavaScript validates JavaScript code
func (rs *RuntimeSandbox) validateJavaScript(code string) error {
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"eval(", "Function(", "setTimeout(", "setInterval(",
		"import(", "require(", "__proto__", "constructor",
		"process.", "global.", "window.", "document.",
		"XMLHttpRequest", "fetch(", "File", "FileReader",
		"atob(", "btoa(", "unescape(", "escape(",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// validatePython validates Python code
func (rs *RuntimeSandbox) validatePython(code string) error {
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"eval(", "exec(", "compile(", "__import__",
		"open(", "file(", "input(", "raw_input(",
		"os.", "subprocess.", "sys.", "importlib.",
		"__builtin__", "builtins.", "execfile(",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// validateShell validates shell code
func (rs *RuntimeSandbox) validateShell(code string) error {
	// Check for dangerous patterns
	dangerousPatterns := []string{
		"rm ", "mv ", "cp ", "ln ", "dd ", "mount ", "umount ",
		"chmod ", "chown ", "useradd ", "userdel ", "passwd ",
		"su ", "sudo ", "/dev/", "/proc/", "/sys/",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// executeJavaScript executes JavaScript code in sandbox
func (rs *RuntimeSandbox) executeJavaScript(ctx context.Context, code string, inputs map[string]interface{}) (*ExecutionResult, error) {
	// This is a placeholder implementation
	// In a real system, you would use a JavaScript VM with proper sandboxing
	// such as Otto, goja, or by running in a container

	// For this example, we'll just return a mock result
	result := map[string]interface{}{
		"output": fmt.Sprintf("Executed JavaScript: %s", code),
		"inputs": inputs,
	}

	return &ExecutionResult{
		Success: true,
		Data:    result,
	}, nil
}

// executePython executes Python code in sandbox
func (rs *RuntimeSandbox) executePython(ctx context.Context, code string, inputs map[string]interface{}) (*ExecutionResult, error) {
	// This is a placeholder implementation
	// In a real system, you would execute Python in a subprocess with restrictions

	result := map[string]interface{}{
		"output": fmt.Sprintf("Executed Python: %s", code),
		"inputs": inputs,
	}

	return &ExecutionResult{
		Success: true,
		Data:    result,
	}, nil
}

// executeGo executes Go code in sandbox
func (rs *RuntimeSandbox) executeGo(ctx context.Context, code string, inputs map[string]interface{}) (*ExecutionResult, error) {
	// Go code execution would typically involve compiling and running in a container
	// This is a complex operation that requires proper sandboxing

	result := map[string]interface{}{
		"output": fmt.Sprintf("Executed Go: %s", code),
		"inputs": inputs,
	}

	return &ExecutionResult{
		Success: true,
		Data:    result,
	}, nil
}

// executeShell executes shell code in sandbox
func (rs *RuntimeSandbox) executeShell(ctx context.Context, code string, inputs map[string]interface{}) (*ExecutionResult, error) {
	// Shell execution is dangerous and should be heavily restricted
	// In a real system, this would run in a restricted container

	result := map[string]interface{}{
		"output": fmt.Sprintf("Executed Shell: %s", code),
		"inputs": inputs,
	}

	return &ExecutionResult{
		Success: true,
		Data:    result,
	}, nil
}

// isRuntimeAllowed checks if a runtime type is allowed
func (rs *RuntimeSandbox) isRuntimeAllowed(runtimeType string) bool {
	allowedRuntimes := []string{"javascript", "python", "go", "shell"}
	for _, allowed := range allowedRuntimes {
		if runtimeType == allowed {
			return true
		}
	}
	return false
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	ExecTime time.Duration         `json:"exec_time,omitempty"`
}

// containsIgnoreCase checks if text contains substring ignoring case
func containsIgnoreCase(text, substr string) bool {
	textLower := toLowerCase(text)
	substrLower := toLowerCase(substr)

	for i := 0; i <= len(textLower)-len(substrLower); i++ {
		match := true
		for j := 0; j < len(substrLower); j++ {
			if textLower[i+j] != substrLower[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLowerCase converts string to lowercase
func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}