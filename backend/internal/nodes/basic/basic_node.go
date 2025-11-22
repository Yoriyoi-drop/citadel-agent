// backend/internal/nodes/basic/basic_node.go
package basic

import (
	"context"
	"fmt"
	"strings"

	"github.com/citadel-agent/backend/internal/nodes/utils"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// BasicOperationType represents the type of basic operation
type BasicOperationType string

const (
	BasicOpConstant    BasicOperationType = "constant"
	BasicOpPassthrough BasicOperationType = "passthrough"
	BasicOpDelay       BasicOperationType = "delay"
	BasicOpCounter     BasicOperationType = "counter"
	BasicOpCondition   BasicOperationType = "condition"
	BasicOpLoop        BasicOperationType = "loop"
	BasicOpSwitch      BasicOperationType = "switch"
	BasicOpMath        BasicOperationType = "math"
)

// BasicConfig represents the configuration for a basic node
type BasicConfig struct {
	Operation  BasicOperationType   `json:"operation"`
	Value      interface{}          `json:"value"`
	Condition  string               `json:"condition"`
	Delay      time.Duration        `json:"delay"`
	Counter    int                  `json:"counter"`
	MaxLoops   int                  `json:"max_loops"`
	MathOp     string               `json:"math_operation"`
	MathValues []interface{}        `json:"math_values"`
	Outputs    map[string]interface{} `json:"outputs"`
}

// BasicNode represents a basic node
type BasicNode struct {
	config  *BasicConfig
	counter int
}

// NewBasicNode creates a new basic node
func NewBasicNode(config *BasicConfig) *BasicNode {
	if config.Delay == 0 {
		config.Delay = 100 * time.Millisecond // Default delay
	}
	if config.MaxLoops == 0 {
		config.MaxLoops = 10 // Default max loops
	}

	return &BasicNode{
		config:  config,
		counter: config.Counter,
	}
}

// Execute executes the basic operation
func (bn *BasicNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := bn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = BasicOperationType(opStr)
		}
	}

	switch operation {
	case BasicOpConstant:
		return bn.constantOperation(inputs)
	case BasicOpPassthrough:
		return bn.passthroughOperation(inputs)
	case BasicOpDelay:
		return bn.delayOperation(inputs)
	case BasicOpCounter:
		return bn.counterOperation(inputs)
	case BasicOpCondition:
		return bn.conditionOperation(inputs)
	case BasicOpLoop:
		return bn.loopOperation(inputs)
	case BasicOpSwitch:
		return bn.switchOperation(inputs)
	case BasicOpMath:
		return bn.mathOperation(inputs)
	default:
		return bn.passthroughOperation(inputs) // Default to passthrough
	}
}

// constantOperation returns a constant value
func (bn *BasicNode) constantOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	value := bn.config.Value
	if val, exists := inputs["value"]; exists {
		value = val
	}

	return map[string]interface{}{
		"success":   true,
		"result":    value,
		"operation": "constant",
		"timestamp": time.Now().Unix(),
	}, nil
}

// passthroughOperation passes through the input values
func (bn *BasicNode) passthroughOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Add all inputs to result
	for k, v := range inputs {
		result[k] = v
	}

	// Add config values if not already in inputs
	if bn.config.Value != nil && result["value"] == nil {
		result["value"] = bn.config.Value
	}

	// Add operation metadata
	result["success"] = true
	result["operation"] = "passthrough"
	result["timestamp"] = time.Now().Unix()

	return result, nil
}

// delayOperation adds a delay before proceeding
func (bn *BasicNode) delayOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	delay := bn.config.Delay
	if delayInput, exists := inputs["delay_ms"]; exists {
		if delayFloat, ok := delayInput.(float64); ok {
			delay = time.Duration(delayFloat) * time.Millisecond
		}
	}

	// Sleep for the specified duration
	time.Sleep(delay)

	result := map[string]interface{}{
		"success":   true,
		"operation": "delay",
		"delay":     delay.Milliseconds(),
		"timestamp": time.Now().Unix(),
	}

	// Pass through inputs
	for k, v := range inputs {
		if k != "delay_ms" { // Don't include the delay parameter in output
			result[k] = v
		}
	}

	return result, nil
}

// counterOperation maintains and increments a counter
func (bn *BasicNode) counterOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	// Increment the counter
	bn.counter++

	result := map[string]interface{}{
		"success":   true,
		"counter":   bn.counter,
		"operation": "counter",
		"timestamp": time.Now().Unix(),
	}

	// Pass through inputs
	for k, v := range inputs {
		result[k] = v
	}

	return result, nil
}

// conditionOperation evaluates a condition and returns result
func (bn *BasicNode) conditionOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	// For this implementation, we'll evaluate simple conditions
	// In a real implementation, a more sophisticated expression evaluator would be used
	
	leftValue := inputs["left"]
	rightValue := inputs["right"]
	operator := utils.GetStringValue(inputs["operator"], "==")

	if leftValue == nil || rightValue == nil {
		leftValue = bn.config.Value
		rightValue = inputs["value"]
		operator = utils.GetStringValue(inputs["condition"], utils.GetStringValue(bn.config.Condition, "=="))
	}

	var resultValue bool
	switch operator {
	case "==", "eq":
		resultValue = fmt.Sprintf("%v", leftValue) == fmt.Sprintf("%v", rightValue)
	case "!=", "ne":
		resultValue = fmt.Sprintf("%v", leftValue) != fmt.Sprintf("%v", rightValue)
	case ">", "gt":
		left, leftOk := toFloat64(leftValue)
		right, rightOk := toFloat64(rightValue)
		resultValue = leftOk && rightOk && left > right
	case "<", "lt":
		left, leftOk := toFloat64(leftValue)
		right, rightOk := toFloat64(rightValue)
		resultValue = leftOk && rightOk && left < right
	case ">=", "gte":
		left, leftOk := toFloat64(leftValue)
		right, rightOk := toFloat64(rightValue)
		resultValue = leftOk && rightOk && left >= right
	case "<=", "lte":
		left, leftOk := toFloat64(leftValue)
		right, rightOk := toFloat64(rightValue)
		resultValue = leftOk && rightOk && left <= right
	case "contains":
		leftStr := fmt.Sprintf("%v", leftValue)
		rightStr := fmt.Sprintf("%v", rightValue)
		resultValue = contains(leftStr, rightStr)
	default:
		resultValue = false
	}

	result := map[string]interface{}{
		"success":   true,
		"result":    resultValue,
		"operation": "condition",
		"left":      leftValue,
		"right":     rightValue,
		"operator":  operator,
		"timestamp": time.Now().Unix(),
	}

	// Pass through inputs
	for k, v := range inputs {
		if k != "left" && k != "right" && k != "operator" && k != "condition" {
			result[k] = v
		}
	}

	return result, nil
}

// loopOperation simulates a loop operation
func (bn *BasicNode) loopOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	maxLoops := bn.config.MaxLoops
	if maxInput, exists := inputs["max_loops"]; exists {
		if maxFloat, ok := maxInput.(float64); ok {
			maxLoops = int(maxFloat)
		}
	}

	currentIteration := 0
	if currentInput, exists := inputs["current_iteration"]; exists {
		if currentFloat, ok := currentInput.(float64); ok {
			currentIteration = int(currentFloat)
		}
	}

	// Increment iteration
	currentIteration++

	continueLoop := currentIteration < maxLoops
	result := map[string]interface{}{
		"success":          true,
		"current_iteration": currentIteration,
		"max_loops":        maxLoops,
		"continue_loop":    continueLoop,
		"operation":        "loop",
		"timestamp":        time.Now().Unix(),
	}

	// Pass through inputs
	for k, v := range inputs {
		result[k] = v
	}

	return result, nil
}

// switchOperation implements a switch/case operation
func (bn *BasicNode) switchOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	inputValue := inputs["value"]
	if inputValue == nil {
		inputValue = bn.config.Value
	}

	// Get the outputs map (cases)
	outputs := bn.config.Outputs
	if outputs == nil {
		outputs = make(map[string]interface{})
	}
	
	// If 'cases' is provided in inputs, use that instead
	if casesInput, exists := inputs["cases"]; exists {
		if casesMap, ok := casesInput.(map[string]interface{}); ok {
			outputs = casesMap
		}
	}

	// Find the matching case
	var resultValue interface{}
	keyStr := fmt.Sprintf("%v", inputValue)
	if val, exists := outputs[keyStr]; exists {
		resultValue = val
	} else if val, exists := outputs["default"]; exists {
		// Use default case if available
		resultValue = val
	} else {
		// No match and no default
		resultValue = nil
	}

	result := map[string]interface{}{
		"success":   true,
		"input":     inputValue,
		"result":    resultValue,
		"operation": "switch",
		"timestamp": time.Now().Unix(),
	}

	// Pass through inputs
	for k, v := range inputs {
		result[k] = v
	}

	return result, nil
}

// mathOperation performs basic math operations
func (bn *BasicNode) mathOperation(inputs map[string]interface{}) (map[string]interface{}, error) {
	values := bn.config.MathValues
	if vals, exists := inputs["values"]; exists {
		if valsSlice, ok := vals.([]interface{}); ok {
			values = valsSlice
		}
	} else if len(values) == 0 {
		// If no values provided in config, use input values
		for _, k := range []string{"a", "b", "c", "d"} {
			if v, exists := inputs[k]; exists {
				values = append(values, v)
			}
		}
	}

	if len(values) < 2 {
		return nil, fmt.Errorf("math operation requires at least 2 values")
	}

	operation := bn.config.MathOp
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = opStr
		}
	}

	if operation == "" {
		operation = "add"
	}

	// Convert first two values to float64
	a, err := toFloat64(values[0])
	if err != nil {
		return nil, fmt.Errorf("first value must be a number: %w", err)
	}
	
	b, err := toFloat64(values[1])
	if err != nil {
		return nil, fmt.Errorf("second value must be a number: %w", err)
	}

	var result float64
	switch operation {
	case "add", "+":
		result = a + b
	case "subtract", "-":
		result = a - b
	case "multiply", "*":
		result = a * b
	case "divide", "/":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	case "modulo", "%":
		if b == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		result = float64(int(a) % int(b))
	default:
		return nil, fmt.Errorf("unsupported math operation: %s", operation)
	}

	return map[string]interface{}{
		"success":   true,
		"result":    result,
		"operation": "math",
		"input_a":   a,
		"input_b":   b,
		"math_op":   operation,
		"timestamp": time.Now().Unix(),
	}, nil
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// toFloat64 converts an interface value to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		var result float64
		_, err := fmt.Sscanf(v, "%f", &result)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to float64", v)
		}
		return result, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}


// BasicNodeFromConfig creates a new basic node from a configuration map
func BasicNodeFromConfig(config map[string]interface{}) (engine.NodeInstance, error) {
	var operation BasicOperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = BasicOperationType(opStr)
		}
	}

	var value interface{}
	if val, exists := config["value"]; exists {
		value = val
	}

	var condition string
	if cond, exists := config["condition"]; exists {
		if condStr, ok := cond.(string); ok {
			condition = condStr
		}
	}

	var delay float64
	if d, exists := config["delay_ms"]; exists {
		if dFloat, ok := d.(float64); ok {
			delay = dFloat
		}
	}

	var counter float64
	if c, exists := config["counter"]; exists {
		if cFloat, ok := c.(float64); ok {
			counter = cFloat
		}
	}

	var maxLoops float64
	if max, exists := config["max_loops"]; exists {
		if maxFloat, ok := max.(float64); ok {
			maxLoops = maxFloat
		}
	}

	var mathOp string
	if op, exists := config["math_operation"]; exists {
		if opStr, ok := op.(string); ok {
			mathOp = opStr
		}
	}

	var mathValues []interface{}
	if vals, exists := config["math_values"]; exists {
		if valsSlice, ok := vals.([]interface{}); ok {
			mathValues = valsSlice
		}
	}

	var outputs map[string]interface{}
	if outs, exists := config["outputs"]; exists {
		if outsMap, ok := outs.(map[string]interface{}); ok {
			outputs = outsMap
		}
	}

	nodeConfig := &BasicConfig{
		Operation:  operation,
		Value:      value,
		Condition:  condition,
		Delay:      time.Duration(delay) * time.Millisecond,
		Counter:    int(counter),
		MaxLoops:   int(maxLoops),
		MathOp:     mathOp,
		MathValues: mathValues,
		Outputs:    outputs,
	}

	return NewBasicNode(nodeConfig), nil
}

// RegisterBasicNode registers the basic node type with the engine
func RegisterBasicNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("basic", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return BasicNodeFromConfig(config)
	})
}