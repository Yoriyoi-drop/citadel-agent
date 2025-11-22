package sandbox

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"
)

// RuntimeSandboxManager manages multiple sandbox instances
type RuntimeSandboxManager struct {
	advancedSandbox  *AdvancedSandbox
	containerSandbox *ContainerSandbox
	semaphore        *semaphore.Weighted
	config           *AdvancedSandboxConfig
	activeExecutions int64
}

// NewRuntimeSandboxManager creates a new runtime sandbox manager
func NewRuntimeSandboxManager(config *AdvancedSandboxConfig) *RuntimeSandboxManager {
	// Limit concurrent sandbox executions to prevent resource exhaustion
	sem := semaphore.NewWeighted(10) // Max 10 concurrent executions

	return &RuntimeSandboxManager{
		advancedSandbox:  NewAdvancedSandbox(config),
		containerSandbox: NewContainerSandbox(config),
		semaphore:        sem,
		config:           config,
	}
}

// ExecuteRuntime executes code in an appropriate sandbox based on security requirements
func (rsm *RuntimeSandboxManager) ExecuteRuntime(ctx context.Context, code string, language string, input map[string]interface{}, securityLevel string) (*ExecutionResult, error) {
	// Acquire semaphore to limit concurrent executions
	if err := rsm.semaphore.Acquire(ctx, 1); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   "Failed to acquire execution slot: " + err.Error(),
		}, nil
	}
	defer rsm.semaphore.Release(1)

	// Choose sandbox based on security level
	switch securityLevel {
	case "high", "maximum":
		return rsm.containerSandbox.ExecuteInContainer(ctx, code, language, input)
	case "medium":
		return rsm.advancedSandbox.ExecuteWithSandbox(ctx, code, language, input)
	default: // "low", "basic"
		// For basic security, we might use the original or a simpler sandbox
		return rsm.advancedSandbox.ExecuteWithSandbox(ctx, code, language, input)
	}
}

// ValidateCodeForRuntime validates code before execution
func (rsm *RuntimeSandboxManager) ValidateCodeForRuntime(code string, language string) error {
	// Perform language-specific validation
	switch language {
	case "javascript", "js":
		return rsm.validateJavaScript(code)
	case "python", "py":
		return rsm.validatePython(code)
	case "go":
		return rsm.validateGo(code)
	default:
		// For other languages, check for dangerous patterns
		if containsDangerousPattern(code) {
			return fmt.Errorf("code contains dangerous patterns")
		}
	}

	return nil
}

// validateJavaScript validates JavaScript code for dangerous patterns
func (rsm *RuntimeSandboxManager) validateJavaScript(code string) error {
	// Check for Node.js specific dangerous patterns
	dangerousPatterns := []string{
		"require('child_process')",
		"require('fs')",
		"require('net')",
		"require('dns')",
		"require('cluster')",
		"process.",
		"global.",
		"Buffer.",
		"__proto__",
		"constructor",
		"prototype",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("JavaScript code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// validatePython validates Python code for dangerous patterns
func (rsm *RuntimeSandboxManager) validatePython(code string) error {
	// Check for Python-specific dangerous patterns
	dangerousPatterns := []string{
		"import os",
		"import sys",
		"import subprocess",
		"import socket",
		"import urllib",
		"import requests",
		"import http",
		"import ftplib",
		"import smtplib",
		"eval(",
		"exec(",
		"compile(",
		"__import__(",
		"open(",
		"file(",
		"execfile(",
		"input(",
		"raw_input(",
		"globals(",
		"locals(",
		"vars(",
		"getattr(",
		"setattr(",
		"hasattr(",
		"delattr(",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("Python code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// validateGo validates Go code for dangerous patterns
func (rsm *RuntimeSandboxManager) validateGo(code string) error {
	// Check for Go-specific dangerous patterns
	dangerousPatterns := []string{
		"import \"os\"",
		"import \"os/exec\"",
		"import \"net\"",
		"import \"net/http\"",
		"import \"syscall\"",
		"import \"unsafe\"",
		"import \"plugin\"",
		"import \"os/signal\"",
		"C.",
		"unsafe.",
		"exec.Command",
		"syscall.",
	}

	for _, pattern := range dangerousPatterns {
		if containsIgnoreCase(code, pattern) {
			return fmt.Errorf("Go code contains dangerous pattern: %s", pattern)
		}
	}

	return nil
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

// SecurityPolicy represents a security policy for code execution
type SecurityPolicy struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	AllowedLanguages []string               `json:"allowed_languages"`
	ForbiddenPatterns []string              `json:"forbidden_patterns"`
	MaxExecutionTime time.Duration          `json:"max_execution_time"`
	MaxMemory        int64                  `json:"max_memory"`
	NetworkAccess    bool                   `json:"network_access"`
	FileAccess       bool                   `json:"file_access"`
	AllowedHosts     []string               `json:"allowed_hosts"`
	BlockedPaths     []string               `json:"blocked_paths"`
	RequiredScanning bool                   `json:"required_scanning"`
	EnforceContainer bool                   `json:"enforce_container"`
}

// PolicyEnforcer enforces security policies
type PolicyEnforcer struct {
	policies map[string]*SecurityPolicy
}

// NewPolicyEnforcer creates a new policy enforcer
func NewPolicyEnforcer() *PolicyEnforcer {
	return &PolicyEnforcer{
		policies: make(map[string]*SecurityPolicy),
	}
}

// AddPolicy adds a security policy
func (pe *PolicyEnforcer) AddPolicy(policy *SecurityPolicy) {
	pe.policies[policy.Name] = policy
}

// EnforcePolicy enforces a security policy on code
func (pe *PolicyEnforcer) EnforcePolicy(policyName string, code string, language string) error {
	policy, exists := pe.policies[policyName]
	if !exists {
		return fmt.Errorf("policy %s does not exist", policyName)
	}

	// Check if language is allowed
	allowed := false
	for _, allowedLang := range policy.AllowedLanguages {
		if allowedLang == language {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("language %s is not allowed by policy %s", language, policyName)
	}

	// Check for forbidden patterns
	for _, forbiddenPattern := range policy.ForbiddenPatterns {
		if containsIgnoreCase(code, forbiddenPattern) {
			return fmt.Errorf("code contains forbidden pattern: %s", forbiddenPattern)
		}
	}

	return nil
}

// GetDefaultPolicy returns a default security policy
func GetDefaultPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		Name:             "default",
		Description:      "Default security policy for code execution",
		AllowedLanguages: []string{"javascript", "python", "go"},
		ForbiddenPatterns: []string{
			"eval(", "exec(", "__import__",
			"open(", "file(", "os.", "subprocess.",
			"Function(", "setTimeout(", "setInterval(",
		},
		MaxExecutionTime: 30 * time.Second,
		MaxMemory:        100 * 1024 * 1024, // 100MB
		NetworkAccess:    false,
		FileAccess:       false,
		AllowedHosts:     []string{"api.github.com", "api.openai.com", "httpbin.org"},
		BlockedPaths:     []string{"/etc/", "/proc/", "/sys/", "/root/", "/home/"},
		RequiredScanning: true,
		EnforceContainer: false,
	}
}

// GetHighSecurityPolicy returns a high-security policy
func GetHighSecurityPolicy() *SecurityPolicy {
	policy := GetDefaultPolicy()
	policy.Name = "high_security"
	policy.Description = "High security policy for sensitive code execution"
	policy.MaxExecutionTime = 10 * time.Second
	policy.MaxMemory = 50 * 1024 * 1024 // 50MB
	policy.NetworkAccess = false
	policy.FileAccess = false
	policy.AllowedHosts = []string{} // No network access
	policy.EnforceContainer = true
	policy.RequiredScanning = true
	
	return policy
}

// ResourceMonitor monitors resource usage during execution
type ResourceMonitor struct {
	startTime    time.Time
	maxMemory    int64
	maxExecution time.Duration
	monitoring   bool
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(maxMemory int64, maxExecution time.Duration) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemory:    maxMemory,
		maxExecution: maxExecution,
		monitoring:   false,
	}
}

// StartMonitoring starts resource monitoring
func (rm *ResourceMonitor) StartMonitoring() {
	rm.startTime = time.Now()
	rm.monitoring = true
}

// CheckResources checks if resource limits are exceeded
func (rm *ResourceMonitor) CheckResources() error {
	if !rm.monitoring {
		return nil
	}

	// Check execution time
	if time.Since(rm.startTime) > rm.maxExecution {
		return fmt.Errorf("execution time exceeded limit of %v", rm.maxExecution)
	}

	// In a real implementation, we would check actual memory usage
	// For now, we'll just return nil as a placeholder

	return nil
}

// StopMonitoring stops resource monitoring
func (rm *ResourceMonitor) StopMonitoring() {
	rm.monitoring = false
}