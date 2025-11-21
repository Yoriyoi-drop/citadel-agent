// backend/internal/runtimes/java_runtime.go
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

// JavaRuntime implements the Runtime interface for Java
type JavaRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (jr *JavaRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary directory for this execution
	tempDir, err := os.MkdirTemp("", "java-exec-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract class name from the Java code
	className := "Main"
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "public class ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "class" && i+1 < len(parts) {
					className = strings.TrimSpace(parts[i+1])
					if strings.Contains(className, "{") {
						className = strings.Split(className, "{")[0]
					}
					break
				}
			}
			break
		}
	}

	// Write the Java code to a temporary file
	javaFile := tempDir + "/" + className + ".java"
	if err := os.WriteFile(javaFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}

	// Compile the Java code with a timeout
	compileCtx, cancelCompile := context.WithTimeout(ctx, timeout/2) // Use half the timeout for compilation
	defer cancelCompile()

	compileCmd := exec.CommandContext(compileCtx, "javac", javaFile)
	compileCmd.Dir = tempDir
	compileOutput, compileErr := compileCmd.CombinedOutput()

	if compileCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("compilation timed out after %v", timeout/2)
	}

	if compileErr != nil {
		return nil, fmt.Errorf("compilation failed: %v, output: %s", compileErr, string(compileOutput))
	}

	// Execute the Java code with the remaining timeout
	executeCtx, cancelEx := context.WithTimeout(ctx, timeout/2) // Use remaining timeout for execution
	defer cancelEx()

	cmd := exec.CommandContext(executeCtx, "java", "-cp", tempDir, className)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()

	if executeCtx.Err() == context.DeadlineExceeded {
		cancelEx() // Make sure to cancel the context
		return nil, fmt.Errorf("execution timed out after %v", timeout/2)
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

func (jr *JavaRuntime) ValidateCode(code string) error {
	// Basic validation for Java code
	if !strings.Contains(code, "public class") && !strings.Contains(code, "class Main") {
		return fmt.Errorf("java code should contain a public class or Main class definition")
	}
	return nil
}

func (jr *JavaRuntime) Initialize() error {
	cmd := exec.Command("java", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("java runtime not available: %v", err)
	}

	jr.initialised = true
	return nil
}

func (jr *JavaRuntime) Dispose() error {
	jr.initialised = false
	return nil
}

func (jr *JavaRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "Java Runtime",
		Version: "Java SE",
		Status:  fmt.Sprintf("initialized: %v", jr.initialised),
		Stats:   jr.stats,
	}
}