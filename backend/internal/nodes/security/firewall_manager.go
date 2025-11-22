// backend/internal/nodes/security/firewall_manager.go
package security

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// FirewallRuleType represents the type of firewall rule
type FirewallRuleType string

const (
	RuleTypeAllow FirewallRuleType = "allow"
	RuleTypeBlock FirewallRuleType = "block"
	RuleTypeLog   FirewallRuleType = "log"
)

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string           `json:"id"`
	Type        FirewallRuleType `json:"type"`
	SourceIP    string           `json:"source_ip"`
	Destination string           `json:"destination"`
	Port        int              `json:"port"`
	Protocol    string           `json:"protocol"`
	Enabled     bool             `json:"enabled"`
	Description string           `json:"description"`
}

// FirewallManagerConfig represents the configuration for a firewall manager node
type FirewallManagerConfig struct {
	Rules            []FirewallRule `json:"rules"`
	WhitelistIPs     []string       `json:"whitelist_ips"`
	BlacklistIPs     []string       `json:"blacklist_ips"`
	DefaultAction    string         `json:"default_action"` // "allow" or "block"
	MaxConnections   int            `json:"max_connections"`
	TimeoutDuration  time.Duration  `json:"timeout_duration"`
	EnableLogging    bool           `json:"enable_logging"`
	EnableRateLimiting bool         `json:"enable_rate_limiting"`
	RateLimitWindow  time.Duration  `json:"rate_limit_window"`
	RateLimitMax     int            `json:"rate_limit_max"`
}

// FirewallManagerNode represents a firewall manager node
type FirewallManagerNode struct {
	config *FirewallManagerConfig
}

// NewFirewallManagerNode creates a new firewall manager node
func NewFirewallManagerNode(config *FirewallManagerConfig) *FirewallManagerNode {
	if config.TimeoutDuration == 0 {
		config.TimeoutDuration = 30 * time.Second
	}

	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = 1 * time.Minute
	}

	if config.RateLimitMax == 0 {
		config.RateLimitMax = 100
	}

	return &FirewallManagerNode{
		config: config,
	}
}

// Execute executes the firewall manager operation
func (fmn *FirewallManagerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get target IP from inputs
	targetIP := ""
	if ip, exists := inputs["target_ip"]; exists {
		if ipStr, ok := ip.(string); ok {
			targetIP = ipStr
		}
	}

	// Get target port from inputs
	targetPort := -1
	if port, exists := inputs["target_port"]; exists {
		if portFloat, ok := port.(float64); ok {
			targetPort = int(portFloat)
		}
	}

	// Get protocol from inputs
	protocol := ""
	if proto, exists := inputs["protocol"]; exists {
		if protoStr, ok := proto.(string); ok {
			protocol = protoStr
		}
	}

	// Check if the target IP is in the blacklist
	if fmn.isIPBlacklisted(targetIP) {
		return map[string]interface{}{
			"success":      false,
			"blocked":      true,
			"reason":       "IP is blacklisted",
			"target_ip":    targetIP,
			"target_port":  targetPort,
			"protocol":     protocol,
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Check if the target IP is in the whitelist
	if fmn.isIPWhitelisted(targetIP) {
		return map[string]interface{}{
			"success":      true,
			"allowed":      true,
			"reason":       "IP is whitelisted",
			"target_ip":    targetIP,
			"target_port":  targetPort,
			"protocol":     protocol,
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Process rules
	action := fmn.processRules(targetIP, targetPort, protocol)

	// If no specific rule matched, use default action
	if action == "" {
		action = fmn.config.DefaultAction
		if action == "" {
			action = "allow"
		}
	}

	// Determine if allowed based on action
	allowed := action == "allow"
	reason := fmt.Sprintf("Action based on %s", action)

	return map[string]interface{}{
		"success":      allowed,
		"allowed":      allowed,
		"blocked":      !allowed,
		"reason":       reason,
		"target_ip":    targetIP,
		"target_port":  targetPort,
		"protocol":     protocol,
		"timestamp":    time.Now().Unix(),
	}, nil
}

// isIPWhitelisted checks if an IP is in the whitelist
func (fmn *FirewallManagerNode) isIPWhitelisted(ip string) bool {
	for _, whitelistIP := range fmn.config.WhitelistIPs {
		if matchIP(ip, whitelistIP) {
			return true
		}
	}
	return false
}

// isIPBlacklisted checks if an IP is in the blacklist
func (fmn *FirewallManagerNode) isIPBlacklisted(ip string) bool {
	for _, blacklistIP := range fmn.config.BlacklistIPs {
		if matchIP(ip, blacklistIP) {
			return true
		}
	}
	return false
}

// processRules processes firewall rules to determine action
func (fmn *FirewallManagerNode) processRules(targetIP string, targetPort int, protocol string) string {
	for _, rule := range fmn.config.Rules {
		if !rule.Enabled {
			continue
		}

		// Check IP match
		ipMatch := matchIP(targetIP, rule.SourceIP)
		if !ipMatch && rule.SourceIP != "" {
			continue
		}

		// Check port match
		portMatch := rule.Port == 0 || targetPort == -1 || rule.Port == targetPort
		if !portMatch {
			continue
		}

		// Check protocol match
		protocolMatch := rule.Protocol == "" || strings.EqualFold(rule.Protocol, protocol)
		if !protocolMatch {
			continue
		}

		// Rule matched
		if rule.Type == RuleTypeAllow {
			return "allow"
		} else if rule.Type == RuleTypeBlock {
			return "block"
		}
	}

	// No rule matched
	return ""
}

// matchIP checks if an IP matches a rule IP or CIDR
func matchIP(targetIP, ruleIP string) bool {
	if targetIP == ruleIP {
		return true
	}

	// Check if ruleIP is a CIDR
	if _, ipnet, err := net.ParseCIDR(ruleIP); err == nil {
		if ipnet.Contains(net.ParseIP(targetIP)) {
			return true
		}
	}

	// Check for wildcard patterns
	if strings.Contains(ruleIP, "*") {
		pattern := strings.ReplaceAll(ruleIP, "*", ".*")
		matched, _ := regexp.MatchString("^"+pattern+"$", targetIP)
		return matched
	}

	return false
}

// FirewallManagerNodeFromConfig creates a new firewall manager node from a configuration map
func FirewallManagerNodeFromConfig(config map[string]interface{}) (interfaces.NodeInstance, error) {
	var rules []FirewallRule
	if rulesSlice, exists := config["rules"]; exists {
		if rulesInterface, ok := rulesSlice.([]interface{}); ok {
			rules = make([]FirewallRule, len(rulesInterface))
			for i, ruleInterface := range rulesInterface {
				if ruleMap, ok := ruleInterface.(map[string]interface{}); ok {
					rules[i] = FirewallRule{
						ID:          getStringValue(ruleMap["id"]),
						Type:        FirewallRuleType(getStringValue(ruleMap["type"])),
						SourceIP:    getStringValue(ruleMap["source_ip"]),
						Destination: getStringValue(ruleMap["destination"]),
						Port:        int(getFloat64Value(ruleMap["port"])),
						Protocol:    getStringValue(ruleMap["protocol"]),
						Enabled:     getBoolValue(ruleMap["enabled"]),
						Description: getStringValue(ruleMap["description"]),
					}
				}
			}
		}
	}

	var whitelistIPs []string
	if ips, exists := config["whitelist_ips"]; exists {
		if ipsSlice, ok := ips.([]interface{}); ok {
			whitelistIPs = make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				whitelistIPs[i] = getStringValue(ip)
			}
		}
	}

	var blacklistIPs []string
	if ips, exists := config["blacklist_ips"]; exists {
		if ipsSlice, ok := ips.([]interface{}); ok {
			blacklistIPs = make([]string, len(ipsSlice))
			for i, ip := range ipsSlice {
				blacklistIPs[i] = getStringValue(ip)
			}
		}
	}

	var defaultAction string
	if action, exists := config["default_action"]; exists {
		if actionStr, ok := action.(string); ok {
			defaultAction = actionStr
		}
	}

	var maxConnections int
	if max, exists := config["max_connections"]; exists {
		if maxFloat, ok := max.(float64); ok {
			maxConnections = int(maxFloat)
		}
	}

	var timeoutDuration float64
	if dur, exists := config["timeout_duration_seconds"]; exists {
		if durFloat, ok := dur.(float64); ok {
			timeoutDuration = durFloat
		}
	}

	var enableLogging bool
	if logging, exists := config["enable_logging"]; exists {
		enableLogging = getBoolValue(logging)
	}

	var enableRateLimiting bool
	if rateLimit, exists := config["enable_rate_limiting"]; exists {
		enableRateLimiting = getBoolValue(rateLimit)
	}

	var rateLimitWindow float64
	if window, exists := config["rate_limit_window_seconds"]; exists {
		if windowFloat, ok := window.(float64); ok {
			rateLimitWindow = windowFloat
		}
	}

	var rateLimitMax int
	if max, exists := config["rate_limit_max"]; exists {
		if maxFloat, ok := max.(float64); ok {
			rateLimitMax = int(maxFloat)
		}
	}

	nodeConfig := &FirewallManagerConfig{
		Rules:             rules,
		WhitelistIPs:      whitelistIPs,
		BlacklistIPs:      blacklistIPs,
		DefaultAction:     defaultAction,
		MaxConnections:    maxConnections,
		TimeoutDuration:   time.Duration(timeoutDuration) * time.Second,
		EnableLogging:     enableLogging,
		EnableRateLimiting: enableRateLimiting,
		RateLimitWindow:   time.Duration(rateLimitWindow) * time.Second,
		RateLimitMax:      rateLimitMax,
	}

	return NewFirewallManagerNode(nodeConfig), nil
}

// getStringValue safely gets a string value from interface{}
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// getBoolValue safely gets a boolean value from interface{}
func getBoolValue(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		b, _ := strconv.ParseBool(s)
		return b
	}
	if f, ok := v.(float64); ok {
		return f != 0
	}
	return false
}

// getFloat64Value safely gets a float64 value from interface{}
func getFloat64Value(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if s, ok := v.(string); ok {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	if b, ok := v.(bool); ok {
		if b {
			return 1.0
		}
		return 0.0
	}
	return 0.0
}

// RegisterFirewallManagerNode registers the firewall manager node type with the engine
func RegisterFirewallManagerNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("firewall_manager", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return FirewallManagerNodeFromConfig(config)
	})
}