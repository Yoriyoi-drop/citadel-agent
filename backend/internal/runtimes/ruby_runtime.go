// backend/internal/runtimes/ruby_runtime.go
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

// RubyRuntime implements the Runtime for Ruby
type RubyRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (rr *RubyRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary file for the Ruby code
	tempFile, err := os.CreateTemp("", "rb-exec-*.rb")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Prepare the Ruby code with input access
	inputsJSON, err := json.Marshal(inputs)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to marshal inputs: %v", err)
	}

	rubyCode := fmt.Sprintf(`
require 'json'

inputs = JSON.parse('%s')

begin
  # Execute the provided code
  result = eval(<<-'CODE'
    %s
  CODE
  )

  puts JSON.dump({
    'result' => result || 'No result variable defined',
    'inputs' => inputs
  })
rescue => e
  STDERR.puts e.message
  exit 1
end
`, string(inputsJSON), strings.ReplaceAll(code, "'", "\\'"))

	if _, err := tempFile.WriteString(rubyCode); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}
	tempFile.Close()

	// Execute with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "ruby", tempFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()

	if ctxWithTimeout.Err() == context.DeadlineExceeded {
		cancel() // Make sure to cancel the context
		return nil, fmt.Errorf("execution timed out after %v", timeout)
	}

	if execErr != nil {
		return nil, fmt.Errorf("execution failed: %v, stderr: %s", execErr, stderr.String())
	}

	result := map[string]interface{}{
		"output":       stdout.String(),
		"error_output": stderr.String(),
		"inputs":       inputs,
	}

	return result, nil
}

func (rr *RubyRuntime) ValidateCode(code string) error {
	// Basic validation for Ruby code
	unsafePatterns := []string{
		"eval(", "exec(", "system(", "open(", "syscall(",
		"require(", "load(", "binding", "TOPLEVEL_BINDING",
		"Kernel.", "ObjectSpace.", "TracePoint",
	}
	
	for _, pattern := range unsafePatterns {
		if strings.Contains(code, pattern) {
			return fmt.Errorf("unsafe pattern '%s' detected in Ruby code", pattern)
		}
	}
	return nil
}

func (rr *RubyRuntime) Initialize() error {
	cmd := exec.Command("ruby", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ruby runtime not available: %v", err)
	}

	rr.initialised = true
	return nil
}

func (rr *RubyRuntime) Dispose() error {
	rr.initialised = false
	return nil
}

func (rr *RubyRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "Ruby Runtime",
		Version: "Ruby 3.x",
		Status:  fmt.Sprintf("initialized: %v", rr.initialised),
		Stats:   rr.stats,
	}
}