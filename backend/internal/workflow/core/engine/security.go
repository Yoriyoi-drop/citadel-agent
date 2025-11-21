package engine

import (
	"context"
	"fmt"
	"time"
)

// RuntimeValidator validates runtime execution for security
type RuntimeValidator struct {
	allowedHosts []string
	blockedPaths []string
	maxExecTime  time.Duration
	maxMemory    int64
}

// NewRuntimeValidator creates a new runtime validator
func NewRuntimeValidator() *RuntimeValidator {
	return &RuntimeValidator{
		allowedHosts: []string{"api.github.com", "api.openai.com", "httpbin.org"}, // Default allowed hosts
		blockedPaths: []string{"/etc/", "/proc/", "/sys/"}, // Default blocked paths
		maxExecTime:  30 * time.Second,
		maxMemory:    100 * 1024 * 1024, // 100MB
	}
}

// ValidateRuntime checks if runtime execution is allowed
func (rv *RuntimeValidator) ValidateRuntime(runtimeType string, code string, config map[string]interface{}) error {
	// Check runtime type is allowed
	allowedRuntimes := []string{"go", "javascript", "python", "shell", "http"}
	isAllowed := false
	for _, allowed := range allowedRuntimes {
		if runtimeType == allowed {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return fmt.Errorf("runtime type %s is not allowed", runtimeType)
	}

	// Check for dangerous patterns in code
	if containsDangerousPattern(code) {
		return fmt.Errorf("code contains dangerous patterns")
	}

	// Check configuration for security issues
	if err := rv.validateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// validateConfig validates runtime configuration for security issues
func (rv *RuntimeValidator) validateConfig(config map[string]interface{}) error {
	if timeout, exists := config["timeout"]; exists {
		if timeoutVal, ok := timeout.(float64); ok {
			if time.Duration(timeoutVal) > rv.maxExecTime {
				return fmt.Errorf("timeout exceeds maximum allowed value of %v", rv.maxExecTime)
			}
		}
	}

	if memory, exists := config["memory_limit"]; exists {
		if memoryVal, ok := memory.(float64); ok {
			if int64(memoryVal) > rv.maxMemory {
				return fmt.Errorf("memory limit exceeds maximum allowed value of %d bytes", rv.maxMemory)
			}
		}
	}

	return nil
}

// PermissionChecker handles permission checking for workflows and nodes
type PermissionChecker struct {
	permissions map[string][]string // userID -> []permission
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{
		permissions: make(map[string][]string),
	}
}

// CheckPermission checks if a user has permission to perform an action
func (pc *PermissionChecker) CheckPermission(userID, action string) bool {
	userPerms, exists := pc.permissions[userID]
	if !exists {
		return false
	}

	for _, perm := range userPerms {
		if perm == action || perm == "*" { // "*" means all permissions
			return true
		}
	}

	return false
}

// AddPermission adds a permission to a user
func (pc *PermissionChecker) AddPermission(userID, permission string) {
	if _, exists := pc.permissions[userID]; !exists {
		pc.permissions[userID] = []string{}
	}
	pc.permissions[userID] = append(pc.permissions[userID], permission)
}

// ResourceLimiter limits resource usage for workflows
type ResourceLimiter struct {
	currentUsage map[string]*ResourceUsage
	limitConfig  *ResourceLimits
}

// ResourceUsage tracks current resource usage
type ResourceUsage struct {
	CPUUsage      float64
	MemoryUsage   int64
	NetworkUsage  int64
	ExecutionTime time.Duration
	RequestCount  int
}

// ResourceLimits defines resource limits
type ResourceLimits struct {
	MaxCPUUsage      float64
	MaxMemoryUsage   int64
	MaxNetworkUsage  int64
	MaxExecutionTime time.Duration
	MaxRequestCount  int
}

// NewResourceLimiter creates a new resource limiter
func NewResourceLimiter() *ResourceLimiter {
	return &ResourceLimiter{
		currentUsage: make(map[string]*ResourceUsage),
		limitConfig: &ResourceLimits{
			MaxCPUUsage:      80.0,      // 80%
			MaxMemoryUsage:   512 * 1024 * 1024, // 512MB
			MaxNetworkUsage:  100 * 1024 * 1024, // 100MB
			MaxExecutionTime: 10 * time.Minute,
			MaxRequestCount:  1000,
		},
	}
}

// CheckResourceLimits checks if resources are within limits
func (rl *ResourceLimiter) CheckResourceLimits(userID string, additionalUsage *ResourceUsage) error {
	currentUsage, exists := rl.currentUsage[userID]
	if !exists {
		currentUsage = &ResourceUsage{}
		rl.currentUsage[userID] = currentUsage
	}

	// Check if adding additional usage would exceed limits
	if currentUsage.CPUUsage+additionalUsage.CPUUsage > rl.limitConfig.MaxCPUUsage {
		return fmt.Errorf("CPU usage would exceed limit of %.2f%%", rl.limitConfig.MaxCPUUsage)
	}

	if currentUsage.MemoryUsage+additionalUsage.MemoryUsage > rl.limitConfig.MaxMemoryUsage {
		return fmt.Errorf("memory usage would exceed limit of %d bytes", rl.limitConfig.MaxMemoryUsage)
	}

	if currentUsage.NetworkUsage+additionalUsage.NetworkUsage > rl.limitConfig.MaxNetworkUsage {
		return fmt.Errorf("network usage would exceed limit of %d bytes", rl.limitConfig.MaxNetworkUsage)
	}

	if currentUsage.ExecutionTime+additionalUsage.ExecutionTime > rl.limitConfig.MaxExecutionTime {
		return fmt.Errorf("execution time would exceed limit of %v", rl.limitConfig.MaxExecutionTime)
	}

	if currentUsage.RequestCount+additionalUsage.RequestCount > rl.limitConfig.MaxRequestCount {
		return fmt.Errorf("request count would exceed limit of %d", rl.limitConfig.MaxRequestCount)
	}

	return nil
}

// UpdateResourceUsage updates resource usage for a user
func (rl *ResourceLimiter) UpdateResourceUsage(userID string, usage *ResourceUsage) {
	currentUsage, exists := rl.currentUsage[userID]
	if !exists {
		currentUsage = &ResourceUsage{}
		rl.currentUsage[userID] = currentUsage
	}

	currentUsage.CPUUsage += usage.CPUUsage
	currentUsage.MemoryUsage += usage.MemoryUsage
	currentUsage.NetworkUsage += usage.NetworkUsage
	currentUsage.ExecutionTime += usage.ExecutionTime
	currentUsage.RequestCount += usage.RequestCount
}

// containsDangerousPattern checks if code contains dangerous patterns
func containsDangerousPattern(code string) bool {
	// Check for potentially dangerous patterns in code
	dangerousPatterns := []string{
		"eval(", "exec(", "__import__", "compile(",
		"open(", "file(", "input(", "raw_input(",
		"os.", "subprocess.", "sys.", "importlib.",
		"rm ", "mv ", "chmod ", "chown ", "mount ", "umount ",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return true
		}
	}
	return false
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