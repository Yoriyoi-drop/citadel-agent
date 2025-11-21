package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// AIAgentNode represents an AI agent node in the workflow
type AIAgentNode struct {
	Prompt      string                 `json:"prompt"`
	Model       string                 `json:"model"`
	Tools       []string               `json:"tools"`
	Parameters  map[string]interface{} `json:"parameters"`
	Memory      map[string]interface{} `json:"memory"`
	MaxRetries  int                    `json:"max_retries"`
	Temperature float64                `json:"temperature"`
	TopP        float64                `json:"top_p"`
}

// NewAIAgentNode creates a new AI agent node instance
func NewAIAgentNode(config map[string]interface{}) (NodeInstance, error) {
	node := &AIAgentNode{
		MaxRetries:  3,
		Temperature: 0.7,
		TopP:        1.0,
		Parameters:  make(map[string]interface{}),
		Memory:      make(map[string]interface{}),
		Tools:       make([]string, 0),
	}

	if prompt, ok := config["prompt"].(string); ok {
		node.Prompt = prompt
	}

	if model, ok := config["model"].(string); ok {
		node.Model = model
	}

	if tools, ok := config["tools"].([]interface{}); ok {
		for _, t := range tools {
			if toolStr, ok := t.(string); ok {
				node.Tools = append(node.Tools, toolStr)
			}
		}
	}

	if params, ok := config["parameters"].(map[string]interface{}); ok {
		node.Parameters = params
	}

	if memory, ok := config["memory"].(map[string]interface{}); ok {
		node.Memory = memory
	}

	if maxRetries, ok := config["max_retries"].(float64); ok {
		node.MaxRetries = int(maxRetries)
	}

	if temperature, ok := config["temperature"].(float64); ok {
		node.Temperature = temperature
	}

	if topP, ok := config["top_p"].(float64); ok {
		node.TopP = topP
	}

	return node, nil
}

// Execute executes the AI agent node
func (n *AIAgentNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Combine inputs with node parameters
	combinedInputs := make(map[string]interface{})
	for k, v := range n.Parameters {
		combinedInputs[k] = v
	}
	for k, v := range inputs {
		combinedInputs[k] = v
	}

	// Prepare context for the AI agent
	contextData := map[string]interface{}{
		"prompt":      n.Prompt,
		"model":       n.Model,
		"tools":       n.Tools,
		"inputs":      combinedInputs,
		"node_memory": n.Memory,
	}

	// In a real implementation, this would call the AI manager
	// For simulation, return a result
	result := map[string]interface{}{
		"result":     fmt.Sprintf("AI Agent executed with prompt: %s", n.Prompt),
		"model":      n.Model,
		"tools_used": n.Tools,
		"inputs":     combinedInputs,
		"timestamp":  time.Now().Unix(),
		"context":    contextData,
	}

	return result, nil
}

// DataTransformerNode transforms data from one format to another
type DataTransformerNode struct {
	TransformType string                 `json:"transform_type"` // "json_to_xml", "csv_to_json", etc.
	Mapping       map[string]interface{} `json:"mapping"`
	Template      string                 `json:"template"`
}

// NewDataTransformerNode creates a new data transformer node instance
func NewDataTransformerNode(config map[string]interface{}) (NodeInstance, error) {
	node := &DataTransformerNode{
		Mapping: make(map[string]interface{}),
	}

	if transformType, ok := config["transform_type"].(string); ok {
		node.TransformType = transformType
	}

	if mapping, ok := config["mapping"].(map[string]interface{}); ok {
		node.Mapping = mapping
	}

	if template, ok := config["template"].(string); ok {
		node.Template = template
	}

	return node, nil
}

// Execute executes the data transformer node
func (n *DataTransformerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	switch n.TransformType {
	case "json_to_xml", "json_to_csv", "csv_to_json", "xml_to_json":
		// For simulation, just return the input with transformation info
		result = map[string]interface{}{
			"transformed_data": inputs,
			"transform_type":   n.TransformType,
			"mapping_used":     n.Mapping,
			"template_used":    n.Template,
			"timestamp":        time.Now().Unix(),
		}
	case "custom_template":
		// Apply custom template
		result = map[string]interface{}{
			"transformed_data": applyTemplate(inputs, n.Template),
			"transform_type":   n.TransformType,
			"template_used":    n.Template,
			"timestamp":        time.Now().Unix(),
		}
	default:
		// Default: apply mapping
		result = map[string]interface{}{
			"transformed_data": applyMapping(inputs, n.Mapping),
			"transform_type":   "mapping",
			"mapping_used":     n.Mapping,
			"timestamp":        time.Now().Unix(),
		}
	}

	return result, nil
}

// applyMapping applies a field mapping to the input data
func applyMapping(input map[string]interface{}, mapping map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for newKey, valueSpec := range mapping {
		switch spec := valueSpec.(type) {
		case string:
			// Direct mapping: newKey -> oldKey
			if val, exists := input[spec]; exists {
				result[newKey] = val
			} else {
				result[newKey] = spec // Use as literal value
			}
		case map[string]interface{}:
			// Nested mapping
			if nestedInput, exists := input[newKey]; exists {
				if nestedMap, ok := nestedInput.(map[string]interface{}); ok {
					result[newKey] = applyMapping(nestedMap, spec)
				}
			}
		default:
			result[newKey] = spec
		}
	}
	
	return result
}

// applyTemplate applies a template to the input data
func applyTemplate(input map[string]interface{}, template string) interface{} {
	// In a real implementation, this would parse and apply the template
	// For now, just return the input
	return input
}

// NotificationNode sends notifications through various channels
type NotificationNode struct {
	Channel      string                 `json:"channel"`      // "email", "slack", "webhook", etc.
	Destination  string                 `json:"destination"`  // email address, webhook URL, etc.
	Template     string                 `json:"template"`
	Subject      string                 `json:"subject"`
	Attachments  []string               `json:"attachments"`
	ChannelConfig map[string]interface{} `json:"channel_config"`
}

// NewNotificationNode creates a new notification node instance
func NewNotificationNode(config map[string]interface{}) (NodeInstance, error) {
	node := &NotificationNode{
		ChannelConfig: make(map[string]interface{}),
		Attachments:   make([]string, 0),
	}

	if channel, ok := config["channel"].(string); ok {
		node.Channel = channel
	}

	if destination, ok := config["destination"].(string); ok {
		node.Destination = destination
	}

	if template, ok := config["template"].(string); ok {
		node.Template = template
	}

	if subject, ok := config["subject"].(string); ok {
		node.Subject = subject
	}

	if attachments, ok := config["attachments"].([]interface{}); ok {
		for _, att := range attachments {
			if attStr, ok := att.(string); ok {
				node.Attachments = append(node.Attachments, attStr)
			}
		}
	}

	if channelConfig, ok := config["channel_config"].(map[string]interface{}); ok {
		node.ChannelConfig = channelConfig
	}

	return node, nil
}

// Execute executes the notification node
func (n *NotificationNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// In a real implementation, this would send the notification
	// For simulation, return what would be sent
	
	messageContent := n.Template
	if n.Template == "" {
		// If no template, use the inputs as message content
		if inputsJSON, err := json.Marshal(inputs); err == nil {
			messageContent = string(inputsJSON)
		}
	}

	result := map[string]interface{}{
		"notification_sent": true,
		"channel":          n.Channel,
		"destination":      n.Destination,
		"subject":          n.Subject,
		"message":          messageContent,
		"attachments":      n.Attachments,
		"channel_config":   n.ChannelConfig,
		"input_data":       inputs,
		"timestamp":        time.Now().Unix(),
	}

	return result, nil
}

// LoopNode executes a sub-workflow in a loop
type LoopNode struct {
	ItemsSource   string                 `json:"items_source"`    // Field in inputs that contains the loop items
	SubWorkflow   string                 `json:"sub_workflow"`    // ID of the sub-workflow to execute
	MaxIterations int                    `json:"max_iterations"`  // Maximum number of iterations
	Condition     string                 `json:"condition"`       // Condition to continue the loop
	Parallel      bool                   `json:"parallel"`        // Whether to execute iterations in parallel
	ResultsKey    string                 `json:"results_key"`     // Key to store results in output
	Timeout       int                    `json:"timeout"`         // Timeout per iteration in seconds
	IterationVars map[string]interface{} `json:"iteration_vars"`  // Variables passed to each iteration
}

// NewLoopNode creates a new loop node instance
func NewLoopNode(config map[string]interface{}) (NodeInstance, error) {
	node := &LoopNode{
		MaxIterations: 100,
		Parallel:      false,
		ResultsKey:    "loop_results",
		Timeout:       30,
		IterationVars: make(map[string]interface{}),
	}

	if itemsSource, ok := config["items_source"].(string); ok {
		node.ItemsSource = itemsSource
	}

	if subWorkflow, ok := config["sub_workflow"].(string); ok {
		node.SubWorkflow = subWorkflow
	}

	if maxIterations, ok := config["max_iterations"].(float64); ok {
		node.MaxIterations = int(maxIterations)
	}

	if condition, ok := config["condition"].(string); ok {
		node.Condition = condition
	}

	if parallel, ok := config["parallel"].(bool); ok {
		node.Parallel = parallel
	}

	if resultsKey, ok := config["results_key"].(string); ok {
		node.ResultsKey = resultsKey
	}

	if timeout, ok := config["timeout"].(float64); ok {
		node.Timeout = int(timeout)
	}

	if iterationVars, ok := config["iteration_vars"].(map[string]interface{}); ok {
		node.IterationVars = iterationVars
	}

	return node, nil
}

// Execute executes the loop node
func (n *LoopNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get the items to iterate over
	var items []interface{}
	
	if n.ItemsSource != "" {
		if itemsVal, exists := inputs[n.ItemsSource]; exists {
			switch itemsData := itemsVal.(type) {
			case []interface{}:
				items = itemsData
			case []map[string]interface{}:
				// Convert to []interface{}
				for _, item := range itemsData {
					items = append(items, item)
				}
			default:
				// If it's a single item, wrap in array
				items = []interface{}{itemsVal}
			}
		}
	} else {
		// If no specific items_source, use the entire inputs as a single item
		items = []interface{}{inputs}
	}

	// Limit the number of iterations
	if len(items) > n.MaxIterations {
		items = items[:n.MaxIterations]
	}

	// Execute the loop
	var results []interface{}
	
	if n.Parallel {
		// Execute iterations in parallel
		resultsChan := make(chan interface{}, len(items))
		errChan := make(chan error, len(items))
		
		for i, item := range items {
			go func(index int, itemData interface{}) {
				iterationCtx, cancel := context.WithTimeout(ctx, time.Duration(n.Timeout)*time.Second)
				defer cancel()
				
				iterationResult, err := n.executeIteration(iterationCtx, itemData, index, inputs)
				if err != nil {
					errChan <- fmt.Errorf("iteration %d failed: %w", index, err)
					return
				}
				
				resultsChan <- iterationResult
			}(i, item)
		}
		
		// Collect results
		for i := 0; i < len(items); i++ {
			select {
			case result := <-resultsChan:
				results = append(results, result)
			case err := <-errChan:
				return nil, err
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	} else {
		// Execute iterations sequentially
		for i, item := range items {
			iterationCtx, cancel := context.WithTimeout(ctx, time.Duration(n.Timeout)*time.Second)
			
			iterationResult, err := n.executeIteration(iterationCtx, item, i, inputs)
			cancel() // Cancel the timeout context
			
			if err != nil {
				return nil, fmt.Errorf("loop iteration %d failed: %w", i, err)
			}
			
			results = append(results, iterationResult)
			
			// Check if we should continue based on condition
			if n.Condition != "" && !evaluateCondition(n.Condition, iterationResult) {
				break
			}
		}
	}

	// Return results
	output := make(map[string]interface{})
	output[n.ResultsKey] = results
	output["total_iterations"] = len(results)
	output["timestamp"] = time.Now().Unix()
	
	// Add original inputs to output
	for k, v := range inputs {
		if k != n.ItemsSource { // Don't overwrite the items source
			output[k] = v
		}
	}

	return output, nil
}

// executeIteration executes a single iteration of the loop
func (n *LoopNode) executeIteration(ctx context.Context, item interface{}, index int, originalInputs map[string]interface{}) (map[string]interface{}, error) {
	// Prepare inputs for this iteration
	iterationInputs := make(map[string]interface{})
	
	// Add the current item as 'item' and 'current_item'
	iterationInputs["item"] = item
	iterationInputs["current_item"] = item
	iterationInputs["index"] = index
	iterationInputs["original_inputs"] = originalInputs
	
	// Add any iteration variables
	for k, v := range n.IterationVars {
		iterationInputs[k] = v
	}
	
	// Add other original inputs
	for k, v := range originalInputs {
		if k != n.ItemsSource { // Don't overwrite the items source
			iterationInputs[k] = v
		}
	}
	
	// Simulate executing the sub-workflow
	result := map[string]interface{}{
		"iteration_index": index,
		"item_data":       item,
		"processed_inputs": iterationInputs,
		"sub_workflow_id": n.SubWorkflow,
		"result":          fmt.Sprintf("Processed item %d in loop", index),
		"timestamp":       time.Now().Unix(),
	}
	
	return result, nil
}

// evaluateCondition evaluates a simple condition (for demonstration)
func evaluateCondition(condition string, data interface{}) bool {
	// In a real implementation, this would evaluate the condition expression
	// For now, just return true
	return true
}

// ErrorHandlerNode handles errors from other nodes
type ErrorHandlerNode struct {
	ErrorType     string                 `json:"error_type"`      // "all", "specific_error", "timeout", etc.
	HandlerAction string                 `json:"handler_action"`  // "retry", "skip", "fail_workflow", "custom"
	FallbackValue interface{}            `json:"fallback_value"`
	RetryCount    int                    `json:"retry_count"`
	RetryDelay    int                    `json:"retry_delay"`     // in seconds
	CustomHandler string                 `json:"custom_handler"`  // ID of custom handler workflow
	LogError      bool                   `json:"log_error"`
	NextNode      string                 `json:"next_node"`       // Which node to go to after handling
	HandlerConfig map[string]interface{} `json:"handler_config"`
}

// NewErrorHandlerNode creates a new error handler node instance
func NewErrorHandlerNode(config map[string]interface{}) (NodeInstance, error) {
	node := &ErrorHandlerNode{
		ErrorType:     "all",
		HandlerAction: "retry",
		RetryCount:    3,
		RetryDelay:    1,
		LogError:      true,
		HandlerConfig: make(map[string]interface{}),
	}

	if errorType, ok := config["error_type"].(string); ok {
		node.ErrorType = errorType
	}

	if handlerAction, ok := config["handler_action"].(string); ok {
		node.HandlerAction = handlerAction
	}

	if fallbackValue, exists := config["fallback_value"]; exists {
		node.FallbackValue = fallbackValue
	}

	if retryCount, ok := config["retry_count"].(float64); ok {
		node.RetryCount = int(retryCount)
	}

	if retryDelay, ok := config["retry_delay"].(float64); ok {
		node.RetryDelay = int(retryDelay)
	}

	if customHandler, ok := config["custom_handler"].(string); ok {
		node.CustomHandler = customHandler
	}

	if logError, ok := config["log_error"].(bool); ok {
		node.LogError = logError
	}

	if nextNode, ok := config["next_node"].(string); ok {
		node.NextNode = nextNode
	}

	if handlerConfig, ok := config["handler_config"].(map[string]interface{}); ok {
		node.HandlerConfig = handlerConfig
	}

	return node, nil
}

// Execute executes the error handler node
func (n *ErrorHandlerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get error information from inputs
	var errorInfo map[string]interface{}
	
	if errData, exists := inputs["error_info"]; exists {
		if errMap, ok := errData.(map[string]interface{}); ok {
			errorInfo = errMap
		}
	} else {
		// If no error_info provided, treat as a normal operation
		// This would happen if the error handler is being executed as part of normal flow
		result := map[string]interface{}{
			"handler_status":  "executed",
			"handler_action":  n.HandlerAction,
			"error_type":      n.ErrorType,
			"timestamp":       time.Now().Unix(),
			"original_inputs": inputs,
		}
		return result, nil
	}

	// Handle the error based on configuration
	result := map[string]interface{}{
		"error_handled":   true,
		"error_type":      n.ErrorType,
		"handler_action":  n.HandlerAction,
		"error_info":      errorInfo,
		"timestamp":       time.Now().Unix(),
		"original_inputs": inputs,
	}

	// Log error if configured
	if n.LogError {
		result["log_entry"] = map[string]interface{}{
			"severity":    "error",
			"timestamp":   time.Now().Unix(),
			"error_data":  errorInfo,
			"handler":     n.HandlerAction,
			"workflow_id": inputs["workflow_id"],
		}
	}

	// Process handler action
	switch n.HandlerAction {
	case "retry":
		result["retry_attempt"] = 1
		result["retry_config"] = map[string]interface{}{
			"max_retries": n.RetryCount,
			"delay_secs":  n.RetryDelay,
		}
	case "skip":
		result["action_taken"] = "skipped"
		if n.FallbackValue != nil {
			result["fallback_value"] = n.FallbackValue
		}
	case "fail_workflow":
		result["action_taken"] = "workflow_failed"
		result["should_fail"] = true
	case "custom":
		result["action_taken"] = "custom_handler_called"
		result["custom_handler_id"] = n.CustomHandler
	case "continue":
		result["action_taken"] = "continued"
	default:
		result["action_taken"] = "default_retry"
		result["retry_config"] = map[string]interface{}{
			"max_retries": n.RetryCount,
			"delay_secs":  n.RetryDelay,
		}
	}

	// Add next node info if specified
	if n.NextNode != "" {
		result["next_node"] = n.NextNode
	}

	return result, nil
}

// NodeInstance interface for all node types
type NodeInstance interface {
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}