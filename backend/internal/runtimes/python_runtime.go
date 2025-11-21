// backend/internal/runtimes/python_runtime.go
package runtimes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PythonRuntime implements the Runtime interface for Python
type PythonRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (pr *PythonRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary file for the Python code
	tempFile, err := os.CreateTemp("", "py-exec-*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Prepare the Python code with input access
	inputsJSON, err := json.Marshal(inputs)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to marshal inputs: %v", err)
	}

	pythonCode := fmt.Sprintf(`
import json
import sys

inputs = %s

try:
    # Create a local namespace for execution
    local_vars = {}
    exec_code = """
%s
    """
    exec(exec_code, {"inputs": inputs, "__builtins__": {
        "__import__": lambda x: __builtins__.__import__(x) if x in ["json", "sys", "math", "datetime", "re", "collections"] else None,
        "print": print,
        "len": len,
        "str": str,
        "int": int,
        "float": float,
        "bool": bool,
        "range": range,
        "enumerate": enumerate,
        "zip": zip,
        "map": map,
        "filter": filter,
        "json": json
    }}, local_vars)
    
    # Look for a result variable that might be defined
    result = local_vars.get('result', globals().get('result', 'No result variable defined in execution'))
    
    print(json.dumps({
        'result': result,
        'inputs': inputs
    }))
except Exception as e:
    print(json.dumps({'error': str(e)}), file=sys.stderr)
    sys.exit(1)
`, strconv.Quote(string(inputsJSON)), strings.ReplaceAll(code, "\"", "\\\""))

	if _, err := tempFile.WriteString(pythonCode); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}
	tempFile.Close()

	// Execute with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "python3", tempFile.Name())
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

func (pr *PythonRuntime) ValidateCode(code string) error {
	// Basic validation for Python code
	unsafePatterns := []string{
		"__import__", "exec(", "eval(", "compile(", "open(", "file(",
		"os.", "subprocess.", "sys.", "importlib.", "pickle.",
		"execfile(", "getattr(", "setattr(", "hasattr(",
	}
	
	for _, pattern := range unsafePatterns {
		if strings.Contains(code, pattern) {
			return fmt.Errorf("unsafe pattern '%s' detected in Python code", pattern)
		}
	}
	return nil
}

func (pr *PythonRuntime) Initialize() error {
	cmd := exec.Command("python3", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("python runtime not available: %v", err)
	}

	pr.initialised = true
	return nil
}

func (pr *PythonRuntime) Dispose() error {
	pr.initialised = false
	return nil
}

func (pr *PythonRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "Python Runtime",
		Version: "Python 3.x",
		Status:  fmt.Sprintf("initialized: %v", pr.initialised),
		Stats:   pr.stats,
	}
}