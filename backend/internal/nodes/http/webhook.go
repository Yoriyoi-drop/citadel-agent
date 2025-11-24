package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/nodes/base"
)

// WebhookNode implements webhook trigger functionality
type WebhookNode struct {
	*base.BaseNode
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	Path            string `json:"path"`
	Method          string `json:"method"`
	Secret          string `json:"secret"`
	VerifySignature bool   `json:"verify_signature"`
}

// NewWebhookNode creates a new webhook node
func NewWebhookNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "http_webhook",
		Name:        "HTTP Webhook",
		Category:    "http",
		Description: "Receive HTTP webhooks and trigger workflows",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "webhook",
		Color:       "#3b82f6",
		Inputs:      []base.NodeInput{},
		Outputs: []base.NodeOutput{
			{
				ID:          "payload",
				Name:        "Payload",
				Type:        "object",
				Description: "Webhook payload data",
			},
			{
				ID:          "headers",
				Name:        "Headers",
				Type:        "object",
				Description: "Request headers",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "path",
				Label:       "Webhook Path",
				Description: "URL path for webhook (e.g., /webhook/github)",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "method",
				Label:       "HTTP Method",
				Description: "Allowed HTTP method",
				Type:        "select",
				Required:    false,
				Default:     "POST",
				Options: []base.ConfigOption{
					{Label: "POST", Value: "POST"},
					{Label: "GET", Value: "GET"},
					{Label: "PUT", Value: "PUT"},
					{Label: "ANY", Value: "ANY"},
				},
			},
			{
				Name:        "secret",
				Label:       "Webhook Secret",
				Description: "Secret for signature verification",
				Type:        "password",
				Required:    false,
			},
			{
				Name:        "verify_signature",
				Label:       "Verify Signature",
				Description: "Verify webhook signature",
				Type:        "boolean",
				Required:    false,
				Default:     false,
			},
		},
		Tags: []string{"webhook", "trigger", "http"},
	}

	return &WebhookNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute processes webhook request
func (n *WebhookNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config WebhookConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Get request data from inputs
	requestData, ok := inputs["request"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf("invalid request data")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Verify signature if enabled
	if config.VerifySignature && config.Secret != "" {
		if err := n.verifySignature(requestData, config.Secret); err != nil {
			ctx.Logger.Error("Signature verification failed", err, nil)
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
	}

	// Extract payload
	payload := requestData["body"]
	headers := requestData["headers"]

	result := map[string]interface{}{
		"payload": payload,
		"headers": headers,
		"path":    config.Path,
		"method":  requestData["method"],
	}

	ctx.Logger.Info("Webhook received", map[string]interface{}{
		"path": config.Path,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// verifySignature verifies webhook signature
func (n *WebhookNode) verifySignature(requestData map[string]interface{}, secret string) error {
	signature, ok := requestData["signature"].(string)
	if !ok {
		return fmt.Errorf("signature not found in request")
	}

	body, ok := requestData["body"].(string)
	if !ok {
		bodyBytes, _ := json.Marshal(requestData["body"])
		body = string(bodyBytes)
	}

	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if signature != expectedSignature {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
