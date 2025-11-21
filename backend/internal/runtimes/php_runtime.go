// backend/internal/runtimes/php_runtime.go
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

// PHPRuntime implements the Runtime for PHP
type PHPRuntime struct {
	initialised bool
	stats       RuntimeStats
}

func (php *PHPRuntime) ExecuteCode(ctx context.Context, code string, inputs map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Create a temporary file for the PHP code
	tempFile, err := os.CreateTemp("", "php-exec-*.php")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Prepare the PHP code with input access
	inputsJSON, err := json.Marshal(inputs)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to marshal inputs: %v", err)
	}

	phpCode := fmt.Sprintf(`
<?php
// Safely decode input
$input = json_decode('%s', true);

try {
	// Execute the provided code in a safer context
	ob_start();
	
	// Define a closure to execute the code in an isolated scope
	$result = call_user_func(function() use ($input) {
		// Make inputs available as $input variable
		%s
		
		// Return value if defined, otherwise return a default
		return isset($result) ? $result : 'No result variable defined';
	});
	
	$output = ob_get_contents();
	ob_end_clean();
	
	// Output the result in JSON format
	echo json_encode([
		'result' => $result,
		'output' => $output,
		'inputs' => $input
	]);
} catch (Exception $e) {
	fwrite(STDERR, $e->getMessage());
	exit(1);
}
?>
`, string(inputsJSON), code)

	if _, err := tempFile.WriteString(phpCode); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write code to file: %v", err)
	}
	tempFile.Close()

	// Execute with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "php", tempFile.Name())
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

func (php *PHPRuntime) ValidateCode(code string) error {
	// Basic validation for PHP code
	unsafePatterns := []string{
		"eval(", "exec(", "system(", "shell_exec(", "passthru(",
		"popen(", "proc_open(", "include(", "require(",
		"file_get_contents(", "fopen(", "file_put_contents(",
		"assert(", "create_function(", "unserialize(",
	}
	
	for _, pattern := range unsafePatterns {
		if strings.Contains(code, pattern) {
			return fmt.Errorf("unsafe pattern '%s' detected in PHP code", pattern)
		}
	}
	return nil
}

func (php *PHPRuntime) Initialize() error {
	cmd := exec.Command("php", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("php runtime not available: %v", err)
	}

	php.initialised = true
	return nil
}

func (php *PHPRuntime) Dispose() error {
	php.initialised = false
	return nil
}

func (php *PHPRuntime) GetInfo() RuntimeInfo {
	return RuntimeInfo{
		Name:    "PHP Runtime",
		Version: "PHP 8.x",
		Status:  fmt.Sprintf("initialized: %v", php.initialised),
		Stats:   php.stats,
	}
}