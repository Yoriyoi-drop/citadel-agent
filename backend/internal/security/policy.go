// backend/internal/security/policy.go
package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/models"
)

// SecurityPolicy defines the security rules for the system
type SecurityPolicy struct {
	AllowedHosts           []string
	BlockedIPs             []*net.IPNet
	AllowedIPs             []*net.IPNet
	RateLimits             map[string]*RateLimit
	ContentSecurityPolicy  *CSPConfig
	NetworkFilter          *NetworkFilter
	AuditLoggingEnabled    bool
	MaxRequestSize         int64
	MaxUploadSize          int64
	SessionTimeout         time.Duration
	PasswordPolicy         *PasswordPolicy
	APIKeyPolicy           *APIKeyPolicy
}

// RateLimit defines rate limiting configuration
type RateLimit struct {
	Requests int
	Window   time.Duration
	Message  string
}

// CSPConfig defines Content Security Policy configuration
type CSPConfig struct {
	DefaultSrc   []string
	ScriptSrc    []string
	StyleSrc     []string
	ImgSrc       []string
	ConnectSrc   []string
	FontSrc      []string
	ObjectSrc    []string
	MediaSrc     []string
	FrameSrc     []string
	Sandbox      []string
	ReportURI    string
}

// NetworkFilter handles network access controls
type NetworkFilter struct {
	BlockedHosts []string
	BlockedIPs   []*net.IPNet
	AllowedHosts []string
	AllowedIPs   []*net.IPNet
	MaxRedirects int
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength    int
	MaxLength    int
	RequireUpper bool
	RequireLower bool
	RequireDigit bool
	RequireSpecial bool
	MaxAge       time.Duration
}

// APIKeyPolicy defines API key policies
type APIKeyPolicy struct {
	MaxKeysPerUser int
	DefaultExpiry  time.Duration
	MaxExpiry      time.Duration
	AllowedPrefixes []string
}

// NewDefaultSecurityPolicy creates a default security policy
func NewDefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		AllowedHosts: []string{
			"api.github.com",
			"api.openai.com", 
			"api.sendgrid.com",
			"hooks.slack.com",
			"graph.facebook.com",
		},
		BlockedIPs: []*net.IPNet{},
		AllowedIPs: []*net.IPNet{},
		RateLimits: map[string]*RateLimit{
			"auth": {
				Requests: 5,
				Window:   15 * time.Minute,
				Message:  "Too many authentication attempts",
			},
			"api": {
				Requests: 1000,
				Window:   time.Hour,
				Message:  "Rate limit exceeded",
			},
		},
		ContentSecurityPolicy: &CSPConfig{
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'", "'unsafe-inline'"}, // In production, avoid 'unsafe-inline'
			StyleSrc:   []string{"'self'", "'unsafe-inline'", "https://fonts.googleapis.com"},
			ImgSrc:     []string{"'self'", "data:", "https:"},
			ConnectSrc: []string{"'self'", "https://api.citadel-agent.com"},
			FontSrc:    []string{"'self'", "https://fonts.gstatic.com"},
		},
		NetworkFilter: &NetworkFilter{
			BlockedHosts: []string{
				"localhost",
				"127.0.0.1",
				"0.0.0.0",
				"::1",
			},
			BlockedIPs:   []*net.IPNet{},
			AllowedHosts: []string{},
			AllowedIPs:   []*net.IPNet{},
			MaxRedirects: 5,
		},
		AuditLoggingEnabled: true,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		MaxUploadSize:       50 * 1024 * 1024, // 50MB
		SessionTimeout:      24 * time.Hour,
		PasswordPolicy: &PasswordPolicy{
			MinLength:    8,
			MaxLength:    128,
			RequireUpper: true,
			RequireLower: true,
			RequireDigit: true,
			RequireSpecial: true,
			MaxAge:       90 * 24 * time.Hour, // 90 days
		},
		APIKeyPolicy: &APIKeyPolicy{
			MaxKeysPerUser: 10,
			DefaultExpiry:  30 * 24 * time.Hour, // 30 days
			MaxExpiry:      365 * 24 * time.Hour, // 1 year
			AllowedPrefixes: []string{"sk_", "pk_", "ak_"},
		},
	}
}

// ValidateRequest validates an incoming request against security policies
func (sp *SecurityPolicy) ValidateRequest(reqURL, method string, headers map[string]string) error {
	// Validate URL
	if err := sp.validateURL(reqURL); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	// Validate headers
	if err := sp.validateHeaders(headers); err != nil {
		return fmt.Errorf("Header validation failed: %w", err)
	}

	// Check for suspicious patterns
	if sp.containsSuspiciousPatterns(reqURL) {
		return fmt.Errorf("request contains suspicious patterns")
	}

	return nil
}

// validateURL validates a URL against security policies
func (sp *SecurityPolicy) validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if the host is in the blocked list
	host := strings.ToLower(parsedURL.Hostname())
	for _, blockedHost := range sp.NetworkFilter.BlockedHosts {
		if strings.Contains(host, strings.ToLower(blockedHost)) {
			return fmt.Errorf("host %s is blocked", host)
		}
	}

	// Check allowed hosts if specified
	if len(sp.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range sp.AllowedHosts {
			if strings.Contains(host, strings.ToLower(allowedHost)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("host %s is not in allowed list", host)
		}
	}

	// Check IP restrictions
	ips, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		// If we can't resolve the hostname, check if it's an IP directly
		host, _, _ := net.SplitHostPort(parsedURL.Host)
		if host == "" {
			host = parsedURL.Host
		}
		
		if ip := net.ParseIP(host); ip != nil {
			if sp.isIPBlocked(ip) {
				return fmt.Errorf("IP %s is blocked", ip.String())
			}
		}
		return nil // Can't resolve, so we can't check IP
	}

	for _, ip := range ips {
		if sp.isIPBlocked(ip) {
			return fmt.Errorf("IP %s is blocked", ip.String())
		}
	}

	return nil
}

// validateHeaders validates request headers
func (sp *SecurityPolicy) validateHeaders(headers map[string]string) error {
	// Check for common attack headers
	attackHeaders := []string{
		"x-forwarded-for",
		"x-real-ip",
		"x-client-ip",
		"x-forwarded-host",
		"x-original-host",
	}

	for header, value := range headers {
		if containsIgnoreCaseSlice(attackHeaders, header) {
			// Validate the value for suspicious patterns
			if sp.containsSuspiciousPatterns(value) {
				return fmt.Errorf("suspicious value in header %s", header)
			}
		}
	}

	// Check content length
	if contentLength, exists := headers["content-length"]; exists {
		// This would normally be validated in the server, but we'll check here too
	}

	return nil
}

// containsSuspiciousPatterns checks if text contains suspicious patterns
func (sp *SecurityPolicy) containsSuspiciousPatterns(text string) bool {
	suspiciousPatterns := []string{
		"../",           // Directory traversal
		"../../../",     // More directory traversal
		"0x",            // Hexadecimal
		"eval",          // Code evaluation
		"exec",          // Code execution
		"system",        // System command execution
		"shell_exec",    // Shell execution
		"proc_open",     // Process opening
		"passthru",      // PHP command execution
		"<script",       // XSS attempt
		"javascript:",   // JavaScript attempt
		"vbscript:",     // VBScript attempt
		"onerror",       // Event handler attempt
		"onload",        // Event handler attempt
		"document.cookie", // Cookie access attempt
		"window.location", // Location access attempt
	}

	textLower := strings.ToLower(text)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// isIPBlocked checks if an IP is blocked
func (sp *SecurityPolicy) isIPBlocked(ip net.IP) bool {
	for _, blockedIP := range sp.NetworkFilter.BlockedIPs {
		if blockedIP.Contains(ip) {
			return true
		}
	}

	// If allowed IPs are specified, check if IP is in the allowed list
	if len(sp.NetworkFilter.AllowedIPs) > 0 {
		for _, allowedIP := range sp.NetworkFilter.AllowedIPs {
			if allowedIP.Contains(ip) {
				return false // IP is allowed
			}
		}
		return true // IP is not in allowed list
	}

	return false
}

// ValidateUser validates a user against security policies
func (sp *SecurityPolicy) ValidateUser(user *models.User) error {
	// Validate password if provided
	if user.PasswordHash != "" {
		// Additional validation would happen here
	}

	// Validate email format (basic check)
	if !isValidEmail(user.Email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateAPIKey validates an API key against policies
func (sp *SecurityPolicy) ValidateAPIKey(apiKey *models.APIKey) error {
	// Check if key has expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return fmt.Errorf("API key has expired")
	}

	// Check if key is for a valid user
	if apiKey.UserID == "" {
		return fmt.Errorf("API key has no associated user")
	}

	// Validate key prefix
	validPrefix := false
	for _, prefix := range sp.APIKeyPolicy.AllowedPrefixes {
		if strings.HasPrefix(apiKey.Prefix, prefix) {
			validPrefix = true
			break
		}
	}

	if !validPrefix {
		return fmt.Errorf("API key prefix is invalid")
	}

	return nil
}

// ValidatePassword validates a password against policy
func (sp *SecurityPolicy) ValidatePassword(password string) error {
	policy := sp.PasswordPolicy

	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters long", policy.MinLength)
	}

	if len(password) > policy.MaxLength {
		return fmt.Errorf("password must be no more than %d characters long", policy.MaxLength)
	}

	if policy.RequireUpper && !containsUpper(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if policy.RequireLower && !containsLower(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if policy.RequireDigit && !containsDigit(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	if policy.RequireSpecial && !containsSpecial(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateWorkflow validates a workflow against security policies
func (sp *SecurityPolicy) ValidateWorkflow(workflow *models.Workflow) error {
	// Check workflow name for suspicious patterns
	if sp.containsSuspiciousPatterns(workflow.Name) {
		return fmt.Errorf("workflow name contains suspicious patterns")
	}

	// Check workflow description for suspicious patterns
	if workflow.Description != "" && sp.containsSuspiciousPatterns(workflow.Description) {
		return fmt.Errorf("workflow description contains suspicious patterns")
	}

	// Validate each node in the workflow
	for _, node := range workflow.Nodes {
		if err := sp.validateNode(node); err != nil {
			return fmt.Errorf("node validation failed: %w", err)
		}
	}

	// Validate connections
	for _, conn := range workflow.Connections {
		if err := sp.validateConnection(conn); err != nil {
			return fmt.Errorf("connection validation failed: %w", err)
		}
	}

	return nil
}

// validateNode validates a single node
func (sp *SecurityPolicy) validateNode(node *models.Node) error {
	// Check node name for suspicious patterns
	if sp.containsSuspiciousPatterns(node.Name) {
		return fmt.Errorf("node name contains suspicious patterns")
	}

	// Check node config for suspicious patterns
	if node.Config != nil {
		configStr := fmt.Sprintf("%v", node.Config)
		if sp.containsSuspiciousPatterns(configStr) {
			return fmt.Errorf("node config contains suspicious patterns")
		}
	}

	return nil
}

// validateConnection validates a connection between nodes
func (sp *SecurityPolicy) validateConnection(conn *models.Connection) error {
	// Check for suspicious patterns in connection properties
	if sp.containsSuspiciousPatterns(conn.SourceNodeID) {
		return fmt.Errorf("connection source contains suspicious patterns")
	}

	if sp.containsSuspiciousPatterns(conn.TargetNodeID) {
		return fmt.Errorf("connection target contains suspicious patterns")
	}

	if conn.Condition != "" && sp.containsSuspiciousPatterns(conn.Condition) {
		return fmt.Errorf("connection condition contains suspicious patterns")
	}

	return nil
}

// CheckRateLimit checks if the request exceeds rate limits
func (sp *SecurityPolicy) CheckRateLimit(key, endpoint string) error {
	// In a real implementation, this would check against a rate limiting store
	// For now, we'll just return nil to indicate the check passed
	return nil
}

// GetCSPHeader returns the Content Security Policy header
func (sp *SecurityPolicy) GetCSPHeader() string {
	if sp.ContentSecurityPolicy == nil {
		return ""
	}

	csp := sp.ContentSecurityPolicy
	parts := []string{}

	if len(csp.DefaultSrc) > 0 {
		parts = append(parts, fmt.Sprintf("default-src %s", joinQuoted(csp.DefaultSrc)))
	}
	if len(csp.ScriptSrc) > 0 {
		parts = append(parts, fmt.Sprintf("script-src %s", joinQuoted(csp.ScriptSrc)))
	}
	if len(csp.StyleSrc) > 0 {
		parts = append(parts, fmt.Sprintf("style-src %s", joinQuoted(csp.StyleSrc)))
	}
	if len(csp.ImgSrc) > 0 {
		parts = append(parts, fmt.Sprintf("img-src %s", joinQuoted(csp.ImgSrc)))
	}
	if len(csp.ConnectSrc) > 0 {
		parts = append(parts, fmt.Sprintf("connect-src %s", joinQuoted(csp.ConnectSrc)))
	}
	if len(csp.FontSrc) > 0 {
		parts = append(parts, fmt.Sprintf("font-src %s", joinQuoted(csp.FontSrc)))
	}
	if len(csp.ObjectSrc) > 0 {
		parts = append(parts, fmt.Sprintf("object-src %s", joinQuoted(csp.ObjectSrc)))
	}
	if len(csp.MediaSrc) > 0 {
		parts = append(parts, fmt.Sprintf("media-src %s", joinQuoted(csp.MediaSrc)))
	}
	if len(csp.FrameSrc) > 0 {
		parts = append(parts, fmt.Sprintf("frame-src %s", joinQuoted(csp.FrameSrc)))
	}
	if len(csp.Sandbox) > 0 {
		parts = append(parts, fmt.Sprintf("sandbox %s", joinQuoted(csp.Sandbox)))
	}
	if csp.ReportURI != "" {
		parts = append(parts, fmt.Sprintf("report-uri %s", csp.ReportURI))
	}

	return strings.Join(parts, "; ")
}

// Helper functions
func isValidEmail(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func containsUpper(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return true
		}
	}
	return false
}

func containsIgnoreCaseSlice(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func joinQuoted(items []string) string {
	result := make([]string, len(items))
	for i, item := range items {
		if strings.HasPrefix(item, "'") || strings.HasPrefix(item, "http") {
			result[i] = item
		} else {
			result[i] = fmt.Sprintf("'%s'", item)
		}
	}
	return strings.Join(result, " ")
}