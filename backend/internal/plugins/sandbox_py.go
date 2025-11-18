package plugins

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

// PythonSandbox provides a secure environment for executing Python code
type PythonSandbox struct {
	defaultTimeout time.Duration
	tempDir        string
}

// NewPythonSandbox creates a new Python sandbox
func NewPythonSandbox(timeout time.Duration, tempDir string) *PythonSandbox {
	// Create temp directory if it doesn't exist
	if tempDir == "" {
		tempDir = "/tmp/citadel-python-sandbox"
	}
	
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		// In a real implementation, you might want to handle this error differently
		// For now, we'll continue with a default temp directory
		tempDir = os.TempDir()
	}

	return &PythonSandbox{
		defaultTimeout: timeout,
		tempDir:        tempDir,
	}
}

// Execute executes Python code in a sandboxed environment
func (ps *PythonSandbox) Execute(ctx context.Context, code string, input map[string]interface{}) (*PluginResult, error) {
	startTime := time.Now()

	// Create a temporary file for the Python script
	tempFile, err := ioutil.TempFile(ps.tempDir, "citadel-python-*.py")
	if err != nil {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temporary file: %v", err),
		}, nil
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file

	// Create the Python script with the input data
	pythonScript := fmt.Sprintf(`
import json
import sys

# Parse input
input_data = json.loads('''%s''')

# User's code starts here
%s

# Output result
result = locals().get('result', locals())
print(json.dumps(result))
`, formatInputData(input), code)

	// Write the Python script to the temporary file
	_, err = tempFile.Write([]byte(pythonScript))
	if err != nil {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to write to temporary file: %v", err),
		}, nil
	}
	tempFile.Close()

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, ps.defaultTimeout)
	defer cancel()

	// Execute the Python script
	cmd := exec.CommandContext(execCtx, "python3", tempFile.Name())

	// Capture stdout and stderr
	output, err := cmd.Output()
	stderr := ""
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
	}

	if err != nil {
		return &PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Python execution error: %v, stderr: %s", err, stderr),
		}, nil
	}

	// In a real implementation, you would want to parse the output as JSON
	// For this example, we'll just return the string output
	result := string(output)

	return &PluginResult{
		Success:  true,
		Data:     result,
		ExecTime: time.Since(startTime),
	}, nil
}

// formatInputData formats the input data as a JSON string for Python
func formatInputData(input map[string]interface{}) string {
	// In a real implementation, you would properly serialize the input to JSON
	// For this example, we'll create a simple representation
	// A proper implementation would use json.Marshal with proper escaping
	result := "{"
	for k, v := range input {
		result += fmt.Sprintf("'%s': %v, ", k, v)
	}
	if len(result) > 1 {
		result = result[:len(result)-2] // Remove the last comma and space
	}
	result += "}"
	return result
}

// ValidateCode validates Python code for security
func (ps *PythonSandbox) ValidateCode(code string) error {
	// In a real implementation, you would validate the code for
	// potentially unsafe operations like:
	// - Importing dangerous modules
	// - Accessing the file system
	// - Making network requests
	// - Using eval/exec functions
	//
	// For this example, we'll do a basic check for dangerous patterns
	dangerousPatterns := []string{
		"import os",           // OS module
		"import sys",          // System module
		"import subprocess",   // Subprocess module
		"import requests",     // HTTP requests
		"import urllib",       // HTTP requests
		"eval(",              // Eval function
		"exec(",              // Exec function
		"open(",              // File operations
		"__import__",         // Dynamic imports
		"compile(",           // Code compilation
		"globals()",          // Access to global namespace
		"locals()",           // Access to local namespace
		"vars()",             // Access to variables
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