// backend/internal/nodes/multi_language_nodes.go
package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/ai"
	"github.com/citadel-agent/backend/internal/runtimes"
	"github.com/citadel-agent/backend/internal/interfaces" // Changed to use interface instead of engine
)

// GoNode executes Go code
type GoNode struct {
	Code string `json:"code"`
}

func (g *GoNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	// Create a new runtime manager for this execution
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	// Get timeout from settings or default to 30 seconds
	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeGo, g.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// JavaScriptNode executes JavaScript code
type JavaScriptNode struct {
	Code string `json:"code"`
}

func (js *JavaScriptNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeJS, js.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// PythonNode executes Python code
type PythonNode struct {
	Code string `json:"code"`
}

func (p *PythonNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimePython, p.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// JavaNode executes Java code
type JavaNode struct {
	Code string `json:"code"`
}

func (j *JavaNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeJava, j.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// RubyNode executes Ruby code
type RubyNode struct {
	Code string `json:"code"`
}

func (r *RubyNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeRuby, r.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// PHPNode executes PHP code
type PHPNode struct {
	Code string `json:"code"`
}

func (php *PHPNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimePHP, php.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// RustNode executes Rust code
type RustNode struct {
	Code string `json:"code"`
}

func (rust *RustNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeRust, rust.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// CSharpNode executes C# code
type CSharpNode struct {
	Code string `json:"code"`
}

func (c *CSharpNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeCSharp, c.Code, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// ShellNode executes shell commands
type ShellNode struct {
	Commands string `json:"commands"`
}

func (s *ShellNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	runtimeMgr := runtimes.NewMultiRuntimeManager()

	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	result, err := runtimeMgr.ExecuteCode(ctx, runtimes.RuntimeShell, s.Commands, input, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}

// Validate functions to ensure the code is safe to execute
func (g *GoNode) Validate() error {
	if g.Code == "" {
		return fmt.Errorf("Go code cannot be empty")
	}

	// Basic validation for Go code
	if len(g.Code) < 10 {
		return fmt.Errorf("Go code appears too short to be valid")
	}

	return nil
}

func (js *JavaScriptNode) Validate() error {
	if js.Code == "" {
		return fmt.Errorf("JavaScript code cannot be empty")
	}

	// Basic validation for JavaScript code
	if len(js.Code) < 5 {
		return fmt.Errorf("JavaScript code appears too short to be valid")
	}

	// Check for unsafe functions
	if containsUnsafeJS(js.Code) {
		return fmt.Errorf("JavaScript code contains unsafe functions")
	}

	return nil
}

func (p *PythonNode) Validate() error {
	if p.Code == "" {
		return fmt.Errorf("Python code cannot be empty")
	}

	// Basic validation for Python code
	if len(p.Code) < 5 {
		return fmt.Errorf("Python code appears too short to be valid")
	}

	// Check for unsafe functions
	if containsUnsafePython(p.Code) {
		return fmt.Errorf("Python code contains unsafe functions")
	}

	return nil
}

func (j *JavaNode) Validate() error {
	if j.Code == "" {
		return fmt.Errorf("Java code cannot be empty")
	}

	// Basic validation for Java code
	if len(j.Code) < 20 {
		return fmt.Errorf("Java code appears too short to be valid")
	}

	return nil
}

func (r *RubyNode) Validate() error {
	if r.Code == "" {
		return fmt.Errorf("Ruby code cannot be empty")
	}

	// Basic validation for Ruby code
	if len(r.Code) < 5 {
		return fmt.Errorf("Ruby code appears too short to be valid")
	}

	// Check for unsafe functions
	if containsUnsafeRuby(r.Code) {
		return fmt.Errorf("Ruby code contains unsafe functions")
	}

	return nil
}

func (php *PHPNode) Validate() error {
	if php.Code == "" {
		return fmt.Errorf("PHP code cannot be empty")
	}

	// Basic validation for PHP code
	if len(php.Code) < 10 {
		return fmt.Errorf("PHP code appears too short to be valid")
	}

	// Check for unsafe functions
	if containsUnsafePHP(php.Code) {
		return fmt.Errorf("PHP code contains unsafe functions")
	}

	return nil
}

func (rust *RustNode) Validate() error {
	if rust.Code == "" {
		return fmt.Errorf("Rust code cannot be empty")
	}

	// Basic validation for Rust code
	if len(rust.Code) < 20 {
		return fmt.Errorf("Rust code appears too short to be valid")
	}

	return nil
}

func (c *CSharpNode) Validate() error {
	if c.Code == "" {
		return fmt.Errorf("C# code cannot be empty")
	}

	// Basic validation for C# code
	if len(c.Code) < 20 {
		return fmt.Errorf("C# code appears too short to be valid")
	}

	return nil
}

func (s *ShellNode) Validate() error {
	if s.Commands == "" {
		return fmt.Errorf("Shell commands cannot be empty")
	}

	// Basic validation for shell commands
	if len(s.Commands) < 2 {
		return fmt.Errorf("Shell commands appear too short to be valid")
	}

	// Check for unsafe commands
	if containsUnsafeShell(s.Commands) {
		return fmt.Errorf("Shell commands contain unsafe operations")
	}

	return nil
}

// Helper functions to detect unsafe code patterns
func containsUnsafeJS(code string) bool {
	unsafePatterns := []string{
		"eval(", "Function(", "setTimeout(", "setInterval(",
		"import(", "require(", "__proto__", "constructor",
	}

	for _, pattern := range unsafePatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
}

func containsUnsafePython(code string) bool {
	unsafePatterns := []string{
		"eval(", "exec(", "compile(", "__import__", "open(",
		"os.", "subprocess.", "sys.", "importlib.",
	}

	for _, pattern := range unsafePatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
}

func containsUnsafeRuby(code string) bool {
	unsafePatterns := []string{
		"eval(", "exec(", "system(", "open(", "syscall(",
		"require(", "load(", "binding.", "TOPLEVEL_BINDING",
	}

	for _, pattern := range unsafePatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
}

func containsUnsafePHP(code string) bool {
	unsafePatterns := []string{
		"eval(", "exec(", "system(", "shell_exec(", "passthru(",
		"popen(", "proc_open(", "include(", "require(",
	}

	for _, pattern := range unsafePatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
}

func containsUnsafeShell(code string) bool {
	unsafePatterns := []string{
		"rm ", "mv ", "cp ", "ln ", "dd ", "mount ", "umount ",
		"chmod ", "chown ", "useradd ", "userdel ", "passwd ",
		"su ", "sudo ", "/dev/", "/proc/", "/sys/",
	}

	for _, pattern := range unsafePatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(text, substr string) bool {
	return containsIgnoreCaseHelper(text, substr)
}

func containsIgnoreCaseHelper(text, substr string) bool {
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

func toLowerCase(s string) string {
	// Simple lowercase implementation
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

// AIAgentNode executes an AI agent
type AIAgentNode struct {
	AIManager *ai.AIManager
}

func (a *AIAgentNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	agentID, exists := input["agent_id"].(string)
	if !exists {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  "agent_id is required for AI agent node",
			Timestamp: time.Now(),
		}, nil
	}

	agentInput, exists := input["input"].(string)
	if !exists {
		// If input is not a string, convert it to string
		agentInput = fmt.Sprintf("%v", input)
	}

	// Use the AI manager to execute the agent
	if a.AIManager == nil {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  "AI manager not initialized for AIAgentNode",
			Timestamp: time.Now(),
		}, nil
	}

	result, err := a.AIManager.ExecuteAgent(ctx, agentID, map[string]interface{}{
		"input": agentInput,
		"context": input,
	})
	if err != nil {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("AI agent execution failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	executionResult := &interfaces.ExecutionResult{
		Status: "success",
		Data:   result,
		Timestamp: time.Now(),
	}

	return executionResult, nil
}

// MultiRuntimeNode executes code in a specific language runtime
type MultiRuntimeNode struct {
	RuntimeMgr *runtimes.MultiRuntimeManager
}

func (m *MultiRuntimeNode) Execute(ctx context.Context, input map[string]interface{}) (*interfaces.ExecutionResult, error) {
	// Extract runtime configuration from input
	runtimeTypeStr, exists := input["runtime_type"].(string)
	if !exists {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  "runtime_type is required for multi-runtime node",
			Timestamp: time.Now(),
		}, nil
	}

	runtimeType := runtimes.RuntimeType(runtimeTypeStr)

	code, exists := input["code"].(string)
	if !exists {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  "code is required for multi-runtime node",
			Timestamp: time.Now(),
		}, nil
	}

	// Get timeout configuration
	timeoutDuration := 30 * time.Second
	if timeoutVal, exists := input["timeout"]; exists {
		if timeoutSecs, ok := timeoutVal.(float64); ok {
			timeoutDuration = time.Duration(timeoutSecs) * time.Second
		}
	}

	// Remove runtime configuration from input to pass to the code as data
	passThroughInput := make(map[string]interface{})
	for k, v := range input {
		if k != "runtime_type" && k != "code" && k != "timeout" {
			passThroughInput[k] = v
		}
	}

	// Execute the code in the appropriate runtime using the provided manager
	if m.RuntimeMgr == nil {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  "runtime manager not initialized for MultiRuntimeNode",
			Timestamp: time.Now(),
		}, nil
	}

	result, err := m.RuntimeMgr.ExecuteCode(ctx, runtimeType, code, passThroughInput, timeoutDuration)
	if err != nil {
		return &interfaces.ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("runtime execution failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &result.ExecutionResult, nil
}