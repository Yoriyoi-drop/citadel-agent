// backend/internal/runtimes/javascript_runtime.go
package runtimes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// JavaScriptRuntime implements the Runtime interface for JavaScript
type JavaScriptRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (jr *JavaScriptRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary file for the JavaScript code
	tempFile, err := os.CreateTemp("", "js-exec-*.js")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Prepare the JavaScript code with input access
	inputsJSON, err := json.Marshal(inputs)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to marshal inputs: %v", err)
	}

	jsCode := fmt.Sprintf(`
const inputs = %s;

try {
	// Execute the provided code
	%s
	
	// Look for a result variable that might be defined
	const result = typeof result !== 'undefined' ? result : 'No result variable defined';
	
	console.log(JSON.stringify({ 
		result: result,
		inputs: inputs
	}));
} catch (error) {
	console.error(error.message);
	process.exit(1);
}
`, string(inputsJSON), code)

	if _, err := tempFile.WriteString(jsCode); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}
	tempFile.Close()

	// Execute with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "node", tempFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if ctxWithTimeout.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("execution timed out after %v", timeout)
	}

	if err != nil {
		return nil, fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	result := map[string]interface{}{
		"output":       stdout.String(),
		"error_output": stderr.String(),
		"inputs":       inputs,
	}

	return result, nil
}

func (jr *JavaScriptRuntime) ValidateCode(code string) error {
	// Basic validation for JavaScript code
	if strings.Contains(code, "eval(") || strings.Contains(code, "Function(") ||
	   strings.Contains(code, "__proto__") || strings.Contains(code, "constructor") {
		return fmt.Errorf("unsafe function usage detected in JavaScript code")
	}
	return nil
}

func (jr *JavaScriptRuntime) Initialize() error {
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("javascript runtime (node) not available: %v", err)
	}

	jr.initialised = true
	return nil
}

func (jr *JavaScriptRuntime) Dispose() error {
	jr.initialised = false
	return nil
}

func (jr *JavaScriptRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "JavaScript Runtime",
		Version: "Node.js",
		Status:  fmt.Sprintf("initialized: %v", jr.initialised),
		Stats:   jr.stats,
	}
}