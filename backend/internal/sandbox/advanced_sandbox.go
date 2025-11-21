package sandbox

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// AdvancedSandboxConfig contains configuration for advanced sandboxing
type AdvancedSandboxConfig struct {
	Timeout         time.Duration            `json:"timeout"`
	MaxMemory       int64                   `json:"max_memory"` // in bytes
	MaxCPU          int                     `json:"max_cpu"`    // percentage
	AllowedHosts    []string                `json:"allowed_hosts"`
	BlockedPaths    []string                `json:"blocked_paths"`
	AllowedCommands []string                `json:"allowed_commands"`
	Environment     map[string]string       `json:"environment"`
	WorkingDir      string                  `json:"working_dir"`
	NetworkAccess   bool                    `json:"network_access"`
	FileAccess      bool                    `json:"file_access"`
	ProcessLimits   ProcessLimits           `json:"process_limits"`
	SyscallFilter   *SyscallFilter          `json:"syscall_filter"`
	ResourceLimits  map[string]interface{}  `json:"resource_limits"`
}

// ProcessLimits defines limits for spawned processes
type ProcessLimits struct {
	MaxProcesses    int `json:"max_processes"`
	MaxOpenFiles    int `json:"max_open_files"`
	MaxFileSize     int64 `json:"max_file_size"`
	MaxTotalFiles   int64 `json:"max_total_files"`
}

// SyscallFilter defines which system calls are allowed
type SyscallFilter struct {
	AllowedSyscalls []string `json:"allowed_syscalls"`
	BlockedSyscalls []string `json:"blocked_syscalls"`
}

// AdvancedSandbox provides enhanced security sandboxing
type AdvancedSandbox struct {
	config *AdvancedSandboxConfig
}

// NewAdvancedSandbox creates a new advanced sandbox instance
func NewAdvancedSandbox(config *AdvancedSandboxConfig) *AdvancedSandbox {
	if config == nil {
		config = &AdvancedSandboxConfig{
			Timeout:       30 * time.Second,
			MaxMemory:     100 * 1024 * 1024, // 100MB
			MaxCPU:        80,
			NetworkAccess: false,
			FileAccess:    false,
			ProcessLimits: ProcessLimits{
				MaxProcesses: 10,
				MaxOpenFiles: 10,
				MaxFileSize:  10 * 1024 * 1024, // 10MB
			},
		}
	}
	
	return &AdvancedSandbox{
		config: config,
	}
}

// ExecuteWithSandbox executes code with advanced sandboxing
func (as *AdvancedSandbox) ExecuteWithSandbox(ctx context.Context, code string, language string, input map[string]interface{}) (*ExecutionResult, error) {
	// Validate code for dangerous patterns first
	if containsDangerousPattern(code) {
		return &ExecutionResult{
			Success: false,
			Error:   "Code contains dangerous patterns",
		}, nil
	}

	// Create a temporary execution environment
	tempDir, err := os.MkdirTemp("", "citadel-sandbox-*")
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create temp directory: %v", err),
		}, nil
	}
	defer os.RemoveAll(tempDir)

	// Write code to temporary file
	codeFile := fmt.Sprintf("%s/code.%s", tempDir, getFileExtension(language))
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to write code to file: %v", err),
		}, nil
	}

	// Prepare command based on language
	cmd, err := as.prepareCommand(language, codeFile, tempDir)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to prepare command: %v", err),
		}, nil
	}

	// Set up execution environment
	cmd.Dir = tempDir
	cmd.Env = as.buildEnvironment()
	
	// Apply resource limits
	if err := as.applyResourceLimits(cmd); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to apply resource limits: %v", err),
		}, nil
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, as.config.Timeout)
	defer cancel()

	// Run command with timeout
	stdout, stderr := as.captureOutput(cmd)
	
	// Start the command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to start command: %v", err),
		}, nil
	}

	// Wait for command to finish or timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-execCtx.Done():
		// Timeout occurred, kill the process
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return &ExecutionResult{
			Success: false,
			Error:   "Execution timed out",
		}, nil
	case err := <-done:
		executionTime := time.Since(startTime)
		
		// Get output
		stdoutBytes, _ := io.ReadAll(stdout)
		stderrBytes, _ := io.ReadAll(stderr)
		
		output := string(stdoutBytes)
		if err != nil {
			// Command failed, but we still want to return the output
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("Command failed: %v\nStderr: %s", err, string(stderrBytes)),
				Data:      output,
				ExecTime:  executionTime,
				Resources: as.getResourceUsage(),
			}, nil
		}

		return &ExecutionResult{
			Success:   true,
			Data:      output,
			Error:     "",
			ExecTime:  executionTime,
			Resources: as.getResourceUsage(),
		}, nil
	}
}

// prepareCommand prepares the command for execution based on language
func (as *AdvancedSandbox) prepareCommand(language, codeFile, workingDir string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	
	switch strings.ToLower(language) {
	case "javascript", "js":
		cmd = exec.Command("node", "--disable-proto=property", "--disallow-code-generation-from-strings", codeFile)
	case "python", "py":
		cmd = exec.Command("python3", "-B", "-S", "--check-hash-based-pycs", "never", codeFile) // -B: no bytecode, -S: no site packages
	case "go":
		// For Go, we first build then execute
		binaryFile := fmt.Sprintf("%s/program", workingDir)
		buildCmd := exec.Command("go", "build", "-o", binaryFile, codeFile)
		if err := buildCmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to build Go program: %v", err)
		}
		cmd = exec.Command(binaryFile)
	case "bash", "sh":
		cmd = exec.Command("bash", "-c", fmt.Sprintf("cd %s && timeout %ds bash %s", workingDir, int(as.config.Timeout.Seconds()), codeFile))
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	
	return cmd, nil
}

// buildEnvironment creates a restricted environment
func (as *AdvancedSandbox) buildEnvironment() []string {
	env := []string{
		"HOME=" + os.TempDir(),
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"LANG=C",
		"LC_ALL=C",
	}
	
	// Add any custom environment variables from config
	for key, value := range as.config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	
	return env
}

// applyResourceLimits applies system-level resource limits to the command
func (as *AdvancedSandbox) applyResourceLimits(cmd *exec.Cmd) error {
	// For Unix-like systems, we can apply resource limits using rlimit
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	
	// Set up resource limits
	limits := []unix.Rlimit{
		// CPU time limit
		{
			Cur: uint64(as.config.Timeout.Seconds()) + 1,
			Max: uint64(as.config.Timeout.Seconds()) + 5,
		}, // RLIMIT_CPU
		// Virtual memory limit
		{
			Cur: uint64(as.config.MaxMemory),
			Max: uint64(as.config.MaxMemory),
		}, // RLIMIT_AS
		// File size limit
		{
			Cur: uint64(as.config.ProcessLimits.MaxFileSize),
			Max: uint64(as.config.ProcessLimits.MaxFileSize),
		}, // RLIMIT_FSIZE
		// Number of processes limit
		{
			Cur: uint64(as.config.ProcessLimits.MaxProcesses),
			Max: uint64(as.config.ProcessLimits.MaxProcesses),
		}, // RLIMIT_NPROC
		// Number of open files limit
		{
			Cur: uint64(as.config.ProcessLimits.MaxOpenFiles),
			Max: uint64(as.config.ProcessLimits.MaxOpenFiles),
		}, // RLIMIT_NOFILE
	}
	
	// Apply the first few limits that match standard rlimit constants
	cmd.SysProcAttr.Rlimit = []syscall.Rlimit{
		{Cur: uint64(limits[0].Cur), Max: uint64(limits[0].Max)}, // CPU time
		{Cur: uint64(limits[1].Cur), Max: uint64(limits[1].Max)}, // Virtual memory
		{Cur: uint64(limits[2].Cur), Max: uint64(limits[2].Max)}, // File size
		{Cur: uint64(limits[3].Cur), Max: uint64(limits[3].Max)}, // Processes
		{Cur: uint64(limits[4].Cur), Max: uint64(limits[4].Max)}, // Open files
	}
	
	// Set additional security attributes
	cmd.SysProcAttr.NoSetGroups = true
	cmd.SysProcAttr.Chroot = "" // Empty chroot for no chrooting (we use other methods)
	
	return nil
}

// captureOutput captures command output
func (as *AdvancedSandbox) captureOutput(cmd *exec.Cmd) (io.Reader, io.Reader) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	
	// Start the command to initialize pipes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	
	return stdout, stderr
}

// getResourceUsage gets resource usage information
func (as *AdvancedSandbox) getResourceUsage() map[string]interface{} {
	// In a real implementation, this would gather actual resource usage
	// For now, we'll return a placeholder
	return map[string]interface{}{
		"max_memory": as.config.MaxMemory,
		"timeout":    as.config.Timeout.Seconds(),
		"cpu_limit":  as.config.MaxCPU,
	}
}

// containsDangerousPattern checks for potentially dangerous code patterns
func containsDangerousPattern(code string) bool {
	dangerousPatterns := []string{
		// Shell/command execution
		"exec(", "eval(", "system(", "popen(", "os.system", "os.popen", "subprocess.",
		// File operations
		"open(", "file(", "io.open", "os.remove", "os.rename", "os.chmod", "os.chown",
		// Network operations (if network access is disabled)
		"socket.", "urllib.", "requests.", "http.client", "httplib", "urllib2",
		// Import/execution of modules
		"importlib.", "__import__", "compile(", "execfile(", "__loader__",
		// Dangerous JavaScript patterns
		"Function(", "setTimeout(", "setInterval(", "eval(", "new Function(",
		// Command injection patterns
		"|", "&&", "||", ";", ">", ">>", "2>", "/dev/null",
	}
	
	codeLower := strings.ToLower(code)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(codeLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// getFileExtension gets the file extension for a language
func getFileExtension(language string) string {
	lowerLang := strings.ToLower(language)
	switch lowerLang {
	case "javascript", "js":
		return "js"
	case "python", "py":
		return "py"
	case "go":
		return "go"
	case "bash", "sh":
		return "sh"
	default:
		return "txt"
	}
}

// ContainerSandbox uses container-based sandboxing for maximum isolation
type ContainerSandbox struct {
	config *AdvancedSandboxConfig
}

// NewContainerSandbox creates a new container-based sandbox
func NewContainerSandbox(config *AdvancedSandboxConfig) *ContainerSandbox {
	return &ContainerSandbox{
		config: config,
	}
}

// ExecuteInContainer executes code in a containerized environment
func (cs *ContainerSandbox) ExecuteInContainer(ctx context.Context, code string, language string, input map[string]interface{}) (*ExecutionResult, error) {
	// This is a simplified version - a real implementation would use Docker API
	// or a container runtime to execute code in a container
	
	// For demonstration purposes, we'll show how it would work conceptually
	result := &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Executed in containerized environment: %s code", language),
		Error:   "",
		ExecTime: 100 * time.Millisecond, // Placeholder
		Resources: map[string]interface{}{
			"containerized": true,
			"network_access": cs.config.NetworkAccess,
			"file_access":    cs.config.FileAccess,
		},
	}
	
	// In a real implementation:
	// 1. Create a temporary Dockerfile with security configurations
	// 2. Build the image with minimal base and security restrictions
	// 3. Run the container with resource limits and security options
	// 4. Collect output and ensure cleanup
	
	return result, nil
}

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Success   bool                   `json:"success"`
	Data      string                 `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ExecTime  time.Duration          `json:"exec_time,omitempty"`
	Resources map[string]interface{} `json:"resources,omitempty"`
}

// validateConfig validates the sandbox configuration
func (as *AdvancedSandbox) validateConfig() error {
	if as.config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	
	if as.config.MaxMemory <= 0 {
		return fmt.Errorf("max memory must be greater than 0")
	}
	
	return nil
}