// backend/internal/runtimes/go_runtime.go
package runtimes

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GoRuntime implements the Runtime interface for Go
type GoRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (gr *GoRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary directory for this execution
	tempDir, err := os.MkdirTemp("", "go-exec-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the code to a temporary file
	mainFile := tempDir + "/main.go"
	if err := os.WriteFile(mainFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}

	// Create a simple go.mod file
	goModContent := `module temp

go 1.19
`
	if err := os.WriteFile(tempDir+"/go.mod", []byte(goModContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create go.mod: %v", err)
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the command
	cmd := exec.CommandContext(ctxWithTimeout, "go", "run", "main.go")
	cmd.Dir = tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Check for context cancellation (timeout)
	if ctxWithTimeout.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("execution timed out after %v", timeout)
	}

	if err != nil {
		return nil, fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	result := map[string]interface{}{
		"output":         stdout.String(),
		"error_output":   stderr.String(),
		"execution_time": fmt.Sprintf("%v", timeout),
		"inputs":         inputs,
	}

	return result, nil
}

func (gr *GoRuntime) ValidateCode(code string) error {
	// Basic validation for Go code
	if !strings.Contains(code, "package main") {
		return fmt.Errorf("go code must include package main declaration")
	}
	return nil
}

func (gr *GoRuntime) Initialize() error {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go runtime not available: %v", err)
	}

	gr.initialised = true
	return nil
}

func (gr *GoRuntime) Dispose() error {
	gr.initialised = false
	return nil
}

func (gr *GoRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "Go Runtime",
		Version: "1.19+",
		Status:  fmt.Sprintf("initialized: %v", gr.initialised),
		Stats:   gr.stats,
	}
}