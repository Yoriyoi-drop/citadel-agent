package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// EmailValidatorNode validates email addresses
type EmailValidatorNode struct {
	*base.BaseNode
}

// NewEmailValidatorNode creates email validator node
func NewEmailValidatorNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "email_validator",
		Name:        "Email Validator",
		Category:    "validation",
		Description: "Validate email addresses",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "check-circle",
		Color:       "#22c55e",
		Inputs: []base.NodeInput{
			{
				ID:          "email",
				Name:        "Email",
				Type:        "string",
				Required:    true,
				Description: "Email to validate",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "valid",
				Name:        "Valid",
				Type:        "boolean",
				Description: "Is valid",
			},
			{
				ID:          "error",
				Name:        "Error",
				Type:        "string",
				Description: "Validation error",
			},
		},
		Config: []base.NodeConfig{},
		Tags:   []string{"email", "validation"},
	}

	return &EmailValidatorNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute validates email
func (n *EmailValidatorNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	email, ok := inputs["email"].(string)
	if !ok {
		err := fmt.Errorf("email is required")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Validate email
	_, err := mail.ParseAddress(email)
	valid := err == nil

	result := map[string]interface{}{
		"valid": valid,
		"email": email,
	}

	if !valid {
		result["error"] = err.Error()
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// URLValidatorNode validates URLs
type URLValidatorNode struct {
	*base.BaseNode
}

// NewURLValidatorNode creates URL validator node
func NewURLValidatorNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "url_validator",
		Name:        "URL Validator",
		Category:    "validation",
		Description: "Validate URLs",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "check-circle",
		Color:       "#22c55e",
		Inputs: []base.NodeInput{
			{
				ID:          "url",
				Name:        "URL",
				Type:        "string",
				Required:    true,
				Description: "URL to validate",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "valid",
				Name:        "Valid",
				Type:        "boolean",
				Description: "Is valid",
			},
		},
		Config: []base.NodeConfig{},
		Tags:   []string{"url", "validation"},
	}

	return &URLValidatorNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute validates URL
func (n *URLValidatorNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	urlStr, ok := inputs["url"].(string)
	if !ok {
		err := fmt.Errorf("url is required")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	valid := err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""

	result := map[string]interface{}{
		"valid": valid,
		"url":   urlStr,
	}

	if valid {
		result["scheme"] = parsedURL.Scheme
		result["host"] = parsedURL.Host
		result["path"] = parsedURL.Path
	} else if err != nil {
		result["error"] = err.Error()
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// RegexValidatorNode validates with regex
type RegexValidatorNode struct {
	*base.BaseNode
}

// RegexConfig holds regex configuration
type RegexConfig struct {
	Pattern string `json:"pattern"`
}

// NewRegexValidatorNode creates regex validator node
func NewRegexValidatorNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "regex_validator",
		Name:        "Regex Validator",
		Category:    "validation",
		Description: "Validate with regex pattern",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "check-circle",
		Color:       "#22c55e",
		Inputs: []base.NodeInput{
			{
				ID:          "value",
				Name:        "Value",
				Type:        "string",
				Required:    true,
				Description: "Value to validate",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "valid",
				Name:        "Valid",
				Type:        "boolean",
				Description: "Matches pattern",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "pattern",
				Label:       "Regex Pattern",
				Description: "Regular expression pattern",
				Type:        "string",
				Required:    true,
			},
		},
		Tags: []string{"regex", "validation", "pattern"},
	}

	return &RegexValidatorNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute validates with regex
func (n *RegexValidatorNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config RegexConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	value, ok := inputs["value"].(string)
	if !ok {
		err := fmt.Errorf("value is required")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Compile regex
	re, err := regexp.Compile(config.Pattern)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Match
	valid := re.MatchString(value)

	result := map[string]interface{}{
		"valid":   valid,
		"value":   value,
		"pattern": config.Pattern,
	}

	if valid {
		matches := re.FindStringSubmatch(value)
		if len(matches) > 0 {
			result["matches"] = matches
		}
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
