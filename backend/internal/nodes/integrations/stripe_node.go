// backend/internal/nodes/integrations/stripe_node.go
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// StripeNodeConfig represents the configuration for a Stripe node
type StripeNodeConfig struct {
	APIKey     string `json:"stripe_api_key"`
	Endpoint   string `json:"endpoint"` // Customers, charges, payments, etc.
	Method     string `json:"method"`   // GET, POST, PUT, DELETE
	Parameters map[string]interface{} `json:"parameters"` // Specific parameters for the endpoint
	APIVersion string `json:"api_version"` // Stripe API version
}

// StripeNode represents a Stripe integration node
type StripeNode struct {
	config *StripeNodeConfig
}

// NewStripeNode creates a new Stripe node
func NewStripeNode(config *StripeNodeConfig) *StripeNode {
	if config.APIVersion == "" {
		config.APIVersion = "2023-10-16" // Latest API version as of late 2023
	}
	if config.Method == "" {
		config.Method = "GET"
	}

	return &StripeNode{
		config: config,
	}
}

// Execute executes the Stripe operation
func (sn *StripeNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	apiKey := sn.config.APIKey
	if key, exists := inputs["stripe_api_key"]; exists {
		if keyStr, ok := key.(string); ok {
			apiKey = keyStr
		}
	}

	endpoint := sn.config.Endpoint
	if ep, exists := inputs["endpoint"]; exists {
		if epStr, ok := ep.(string); ok {
			endpoint = epStr
		}
	}

	method := sn.config.Method
	if m, exists := inputs["method"]; exists {
		if mStr, ok := m.(string); ok {
			method = mStr
		}
	}

	// Validate required fields
	if apiKey == "" {
		return nil, fmt.Errorf("Stripe API key is required")
	}

	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Prepare the API URL
	apiURL := fmt.Sprintf("https://api.stripe.com/v1/%s", strings.TrimPrefix(endpoint, "/"))

	// Prepare headers
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/x-www-form-urlencoded",
		"Stripe-Version": sn.config.APIVersion,
	}

	// Prepare parameters (for POST/PUT requests)
	var bodyReader io.Reader
	if method == "POST" || method == "PUT" {
		params := make(map[string]interface{})
		
		// Start with configured parameters
		for k, v := range sn.config.Parameters {
			params[k] = v
		}
		
		// Override with inputs
		for k, v := range inputs {
			if k != "stripe_api_key" && k != "endpoint" && k != "method" {
				params[k] = v
			}
		}
		
		// Convert parameters to form-encoded format
		formData := url.Values{}
		for k, v := range params {
			formData.Set(k, fmt.Sprintf("%v", v))
		}
		
		bodyReader = strings.NewReader(formData.Encode())
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Stripe request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		// If JSON parsing fails, return the raw response
		apiResp = map[string]interface{}{
			"raw_response": string(responseBody),
		}
	}

	// Check if the request was successful
	// Stripe returns 2xx codes for success, 4xx/5xx for errors
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	result := map[string]interface{}{
		"success":     success,
		"status_code": resp.StatusCode,
		"response":    apiResp,
		"method":      method,
		"endpoint":    endpoint,
		"timestamp":   time.Now().Unix(),
		"api_version": sn.config.APIVersion,
	}

	if !success {
		result["error"] = fmt.Sprintf("Stripe API returned status %d", resp.StatusCode)
		if errorMsg, exists := apiResp["error"]; exists {
			result["error"] = fmt.Sprintf("Stripe API error: %v", errorMsg)
		}
	}

	return result, nil
}

// CreateCustomer creates a new customer in Stripe
func (sn *StripeNode) CreateCustomer(ctx context.Context, customerData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": "customers",
		"method":   "POST",
	}
	
	for k, v := range customerData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// CreatePaymentIntent creates a payment intent in Stripe
func (sn *StripeNode) CreatePaymentIntent(ctx context.Context, paymentData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": "payment_intents",
		"method":   "POST",
	}
	
	for k, v := range paymentData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// GetCustomer retrieves a customer from Stripe
func (sn *StripeNode) GetCustomer(ctx context.Context, customerID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("customers/%s", customerID),
		"method":   "GET",
	}
	
	return sn.Execute(ctx, params)
}

// UpdateCustomer updates a customer in Stripe
func (sn *StripeNode) UpdateCustomer(ctx context.Context, customerID string, updateData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("customers/%s", customerID),
		"method":   "POST",
	}
	
	for k, v := range updateData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// DeleteCustomer deletes a customer from Stripe
func (sn *StripeNode) DeleteCustomer(ctx context.Context, customerID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("customers/%s", customerID),
		"method":   "DELETE",
	}
	
	return sn.Execute(ctx, params)
}

// CreateCharge creates a charge in Stripe
func (sn *StripeNode) CreateCharge(ctx context.Context, chargeData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": "charges",
		"method":   "POST",
	}
	
	for k, v := range chargeData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// GetCharge retrieves a charge from Stripe
func (sn *StripeNode) GetCharge(ctx context.Context, chargeID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("charges/%s", chargeID),
		"method":   "GET",
	}
	
	return sn.Execute(ctx, params)
}

// CaptureCharge captures a charge in Stripe
func (sn *StripeNode) CaptureCharge(ctx context.Context, chargeID string, captureData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("charges/%s/capture", chargeID),
		"method":   "POST",
	}
	
	for k, v := range captureData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// RefundCharge refunds a charge in Stripe
func (sn *StripeNode) RefundCharge(ctx context.Context, chargeID string, refundData map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"endpoint": fmt.Sprintf("refunds"),
		"method":   "POST",
		"charge":   chargeID,
	}
	
	for k, v := range refundData {
		params[k] = v
	}
	
	return sn.Execute(ctx, params)
}

// RegisterStripeNode registers the Stripe node type with the engine
func RegisterStripeNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("stripe_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var apiKey string
		if key, exists := config["stripe_api_key"]; exists {
			if keyStr, ok := key.(string); ok {
				apiKey = keyStr
			}
		}

		var endpoint string
		if ep, exists := config["endpoint"]; exists {
			if epStr, ok := ep.(string); ok {
				endpoint = epStr
			}
		}

		var method string
		if m, exists := config["method"]; exists {
			if mStr, ok := m.(string); ok {
				method = mStr
			}
		}

		var apiVersion string
		if version, exists := config["api_version"]; exists {
			if versionStr, ok := version.(string); ok {
				apiVersion = versionStr
			}
		}

		var parameters map[string]interface{}
		if params, exists := config["parameters"]; exists {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				parameters = paramsMap
			}
		}

		nodeConfig := &StripeNodeConfig{
			APIKey:     apiKey,
			Endpoint:   endpoint,
			Method:     method,
			Parameters: parameters,
			APIVersion: apiVersion,
		}

		return NewStripeNode(nodeConfig), nil
	})
}