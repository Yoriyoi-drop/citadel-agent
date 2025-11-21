// backend/internal/nodes/logic/logic_operations.go
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// LogicOperationType represents the type of logic operation
type LogicOperationType string

const (
	LogicAnd      LogicOperationType = "and"
	LogicOr       LogicOperationType = "or"
	LogicNot      LogicOperationType = "not"
	LogicEqual    LogicOperationType = "equal"
	LogicNotEqual LogicOperationType = "not_equal"
	LogicGreater  LogicOperationType = "greater"
	LogicLess     LogicOperationType = "less"
	LogicGreaterOrEqual LogicOperationType = "greater_or_equal"
	LogicLessOrEqual    LogicOperationType = "less_or_equal"
	LogicContains LogicOperationType = "contains"
	LogicStartsWith LogicOperationType = "starts_with"
	LogicEndsWith LogicOperationType = "ends_with"
	LogicIn       LogicOperationType = "in"
	LogicMatchRegex LogicOperationType = "match_regex"
	LogicIf       LogicOperationType = "if"
	LogicSwitch   LogicOperationType = "switch"
)

// LogicNodeConfig represents the configuration for a logic node
type LogicNodeConfig struct {
	Operation  LogicOperationType `json:"operation"`
	Conditions []LogicCondition   `json:"conditions"`
	LeftValue  interface{}        `json:"left_value"`
	RightValue interface{}        `json:"right_value"`
	CaseValues []interface{}      `json:"case_values"`
	DefaultResult interface{}     `json:"default_result"`
	Regex      string            `json:"regex"`
}

// LogicCondition represents a single condition in a logic operation
type LogicCondition struct {
	LeftValue  interface{}        `json:"left_value"`
	Operator   LogicOperationType `json:"operator"`
	RightValue interface{}        `json:"right_value"`
}

// LogicNode represents a logic operation node
type LogicNode struct {
	config *LogicNodeConfig
}

// NewLogicNode creates a new logic node
func NewLogicNode(config *LogicNodeConfig) *LogicNode {
	return &LogicNode{
		config: config,
	}
}

// Execute executes the logic operation node
func (ln *LogicNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	leftValue := ln.config.LeftValue
	if leftInput, exists := inputs["left_value"]; exists {
		leftValue = leftInput
	}

	rightValue := ln.config.RightValue
	if rightInput, exists := inputs["right_value"]; exists {
		rightValue = rightInput
	}

	// Perform the logic operation based on type
	switch ln.config.Operation {
	case LogicAnd:
		return ln.executeAnd(inputs)
	case LogicOr:
		return ln.executeOr(inputs)
	case LogicNot:
		return ln.executeNot(leftValue)
	case LogicEqual:
		return ln.executeEqual(leftValue, rightValue)
	case LogicNotEqual:
		return ln.executeNotEqual(leftValue, rightValue)
	case LogicGreater:
		return ln.executeGreater(leftValue, rightValue)
	case LogicLess:
		return ln.executeLess(leftValue, rightValue)
	case LogicGreaterOrEqual:
		return ln.executeGreaterOrEqual(leftValue, rightValue)
	case LogicLessOrEqual:
		return ln.executeLessOrEqual(leftValue, rightValue)
	case LogicContains:
		return ln.executeContains(leftValue, rightValue)
	case LogicStartsWith:
		return ln.executeStartsWith(leftValue, rightValue)
	case LogicEndsWith:
		return ln.executeEndsWith(leftValue, rightValue)
	case LogicIn:
		return ln.executeIn(leftValue, ln.config.CaseValues)
	case LogicMatchRegex:
		return ln.executeMatchRegex(leftValue, ln.config.Regex)
	case LogicIf:
		return ln.executeIf(inputs)
	case LogicSwitch:
		return ln.executeSwitch(leftValue, ln.config.CaseValues, ln.config.DefaultResult)
	default:
		return nil, fmt.Errorf("unsupported logic operation: %s", ln.config.Operation)
	}
}

// executeAnd performs logical AND operation
func (ln *LogicNode) executeAnd(inputs map[string]interface{}) (map[string]interface{}, error) {
	if len(ln.config.Conditions) == 0 {
		return nil, fmt.Errorf("no conditions provided for AND operation")
	}

	for _, condition := range ln.config.Conditions {
		result, err := ln.evaluateCondition(condition, inputs)
		if err != nil {
			return nil, fmt.Errorf("error evaluating condition: %w", err)
		}
		if !result {
			return map[string]interface{}{
				"success": true,
				"result":  false,
				"operation": string(LogicAnd),
			}, nil
		}
	}

	return map[string]interface{}{
		"success": true,
		"result":  true,
		"operation": string(LogicAnd),
	}, nil
}

// executeOr performs logical OR operation
func (ln *LogicNode) executeOr(inputs map[string]interface{}) (map[string]interface{}, error) {
	if len(ln.config.Conditions) == 0 {
		return nil, fmt.Errorf("no conditions provided for OR operation")
	}

	for _, condition := range ln.config.Conditions {
		result, err := ln.evaluateCondition(condition, inputs)
		if err != nil {
			return nil, fmt.Errorf("error evaluating condition: %w", err)
		}
		if result {
			return map[string]interface{}{
				"success": true,
				"result":  true,
				"operation": string(LogicOr),
			}, nil
		}
	}

	return map[string]interface{}{
		"success": true,
		"result":  false,
		"operation": string(LogicOr),
	}, nil
}

// executeNot performs logical NOT operation
func (ln *LogicNode) executeNot(value interface{}) (map[string]interface{}, error) {
	boolValue, err := ln.toBool(value)
	if err != nil {
		return nil, fmt.Errorf("cannot convert value to boolean: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"result":  !boolValue,
		"operation": string(LogicNot),
	}, nil
}

// executeEqual performs equality comparison
func (ln *LogicNode) executeEqual(left, right interface{}) (map[string]interface{}, error) {
	result := ln.compareValues(left, right) == 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicEqual),
		"left":    left,
		"right":   right,
	}, nil
}

// executeNotEqual performs inequality comparison
func (ln *LogicNode) executeNotEqual(left, right interface{}) (map[string]interface{}, error) {
	result := ln.compareValues(left, right) != 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicNotEqual),
		"left":    left,
		"right":   right,
	}, nil
}

// executeGreater performs greater than comparison
func (ln *LogicNode) executeGreater(left, right interface{}) (map[string]interface{}, error) {
	comp := ln.compareValues(left, right)
	result := comp > 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicGreater),
		"left":    left,
		"right":   right,
	}, nil
}

// executeLess performs less than comparison
func (ln *LogicNode) executeLess(left, right interface{}) (map[string]interface{}, error) {
	comp := ln.compareValues(left, right)
	result := comp < 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicLess),
		"left":    left,
		"right":   right,
	}, nil
}

// executeGreaterOrEqual performs greater than or equal comparison
func (ln *LogicNode) executeGreaterOrEqual(left, right interface{}) (map[string]interface{}, error) {
	comp := ln.compareValues(left, right)
	result := comp >= 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicGreaterOrEqual),
		"left":    left,
		"right":   right,
	}, nil
}

// executeLessOrEqual performs less than or equal comparison
func (ln *LogicNode) executeLessOrEqual(left, right interface{}) (map[string]interface{}, error) {
	comp := ln.compareValues(left, right)
	result := comp <= 0
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicLessOrEqual),
		"left":    left,
		"right":   right,
	}, nil
}

// executeContains checks if left value contains right value
func (ln *LogicNode) executeContains(left, right interface{}) (map[string]interface{}, error) {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	
	result := strings.Contains(leftStr, rightStr)
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicContains),
		"left":    left,
		"right":   right,
	}, nil
}

// executeStartsWith checks if left value starts with right value
func (ln *LogicNode) executeStartsWith(left, right interface{}) (map[string]interface{}, error) {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	
	result := strings.HasPrefix(leftStr, rightStr)
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicStartsWith),
		"left":    left,
		"right":   right,
	}, nil
}

// executeEndsWith checks if left value ends with right value
func (ln *LogicNode) executeEndsWith(left, right interface{}) (map[string]interface{}, error) {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	
	result := strings.HasSuffix(leftStr, rightStr)
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicEndsWith),
		"left":    left,
		"right":   right,
	}, nil
}

// executeIn checks if left value is in the case values
func (ln *LogicNode) executeIn(left interface{}, caseValues []interface{}) (map[string]interface{}, error) {
	for _, value := range caseValues {
		if ln.compareValues(left, value) == 0 {
			return map[string]interface{}{
				"success": true,
				"result":  true,
				"operation": string(LogicIn),
				"value":   left,
				"list":    caseValues,
			}, nil
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"result":  false,
		"operation": string(LogicIn),
		"value":   left,
		"list":    caseValues,
	}, nil
}

// executeMatchRegex checks if left value matches the regex
func (ln *LogicNode) executeMatchRegex(left interface{}, regex string) (map[string]interface{}, error) {
	// In a real implementation, we would use regex matching
	// For now, we'll return a mock result
	leftStr := fmt.Sprintf("%v", left)
	
	// Simple mock implementation - in reality you'd use the regexp package
	// and validate the regex pattern
	result := strings.Contains(leftStr, regex) || regex == ".*" // Simple mock
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicMatchRegex),
		"value":   left,
		"regex":   regex,
	}, nil
}

// executeIf implements if-then-else logic
func (ln *LogicNode) executeIf(inputs map[string]interface{}) (map[string]interface{}, error) {
	condition, exists := inputs["condition"]
	if !exists {
		return nil, fmt.Errorf("condition is required for if operation")
	}
	
	conditionBool, err := ln.toBool(condition)
	if err != nil {
		return nil, fmt.Errorf("cannot convert condition to boolean: %w", err)
	}
	
	var result interface{}
	if conditionBool {
		result = inputs["then"]
	} else {
		result = inputs["else"]
		if result == nil {
			result = inputs["else_value"]
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"result":  result,
		"operation": string(LogicIf),
		"condition": condition,
		"condition_result": conditionBool,
	}, nil
}

// executeSwitch implements switch-case logic
func (ln *LogicNode) executeSwitch(input interface{}, caseValues []interface{}, defaultResult interface{}) (map[string]interface{}, error) {
	for i, caseVal := range caseValues {
		if ln.compareValues(input, caseVal) == 0 {
			// Return the corresponding result
			// In a real implementation, we would have an array of results
			// For now, we'll just return the matched value as a simple implementation
			return map[string]interface{}{
				"success": true,
				"result":  caseVal,
				"operation": string(LogicSwitch),
				"input":   input,
				"matched": caseVal,
				"case_index": i,
			}, nil
		}
	}
	
	return map[string]interface{}{
		"success": true,
		"result":  defaultResult,
		"operation": string(LogicSwitch),
		"input":   input,
		"matched": "default",
		"default": defaultResult,
	}, nil
}

// evaluateCondition evaluates a single condition
func (ln *LogicNode) evaluateCondition(condition LogicCondition, inputs map[string]interface{}) (bool, error) {
	// Resolve values from inputs if they're keys
	leftValue := condition.LeftValue
	if leftStr, ok := condition.LeftValue.(string); ok && strings.HasPrefix(leftStr, "input.") {
		// Extract the input key (e.g., "input.field" -> "field")
		inputKey := strings.TrimPrefix(leftStr, "input.")
		if inputValue, exists := inputs[inputKey]; exists {
			leftValue = inputValue
		}
	}

	rightValue := condition.RightValue
	if rightStr, ok := condition.RightValue.(string); ok && strings.HasPrefix(rightStr, "input.") {
		inputKey := strings.TrimPrefix(rightStr, "input.")
		if inputValue, exists := inputs[inputKey]; exists {
			rightValue = inputValue
		}
	}

	switch condition.Operator {
	case LogicEqual:
		return ln.compareValues(leftValue, rightValue) == 0, nil
	case LogicNotEqual:
		return ln.compareValues(leftValue, rightValue) != 0, nil
	case LogicGreater:
		return ln.compareValues(leftValue, rightValue) > 0, nil
	case LogicLess:
		return ln.compareValues(leftValue, rightValue) < 0, nil
	case LogicGreaterOrEqual:
		return ln.compareValues(leftValue, rightValue) >= 0, nil
	case LogicLessOrEqual:
		return ln.compareValues(leftValue, rightValue) <= 0, nil
	case LogicContains:
		leftStr := fmt.Sprintf("%v", leftValue)
		rightStr := fmt.Sprintf("%v", rightValue)
		return strings.Contains(leftStr, rightStr), nil
	case LogicStartsWith:
		leftStr := fmt.Sprintf("%v", leftValue)
		rightStr := fmt.Sprintf("%v", rightValue)
		return strings.HasPrefix(leftStr, rightStr), nil
	case LogicEndsWith:
		leftStr := fmt.Sprintf("%v", leftValue)
		rightStr := fmt.Sprintf("%v", rightValue)
		return strings.HasSuffix(leftStr, rightStr), nil
	default:
		return false, fmt.Errorf("unsupported condition operator: %s", condition.Operator)
	}
}

// compareValues compares two values of potentially different types
func (ln *LogicNode) compareValues(left, right interface{}) int {
	// Convert both values to strings for comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	// Try to convert to numbers if possible
	leftFloat, leftErr := strconv.ParseFloat(leftStr, 64)
	rightFloat, rightErr := strconv.ParseFloat(rightStr, 64)

	// If both are numbers, compare numerically
	if leftErr == nil && rightErr == nil {
		if leftFloat < rightFloat {
			return -1
		} else if leftFloat > rightFloat {
			return 1
		} else {
			return 0
		}
	}

	// Otherwise, compare as strings
	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	} else {
		return 0
	}
}

// toBool converts an interface{} value to boolean
func (ln *LogicNode) toBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		switch s {
		case "true", "1", "yes", "on", "y", "t":
			return true, nil
		case "false", "0", "no", "off", "n", "f", "":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert string '%s' to boolean", v)
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Int() != 0, nil
	case float32, float64:
		return v != 0.0, nil
	case []interface{}:
		return len(v) > 0, nil
	case map[string]interface{}:
		return len(v) > 0, nil
	default:
		// For any other type, consider it true if not nil
		return value != nil, nil
	}
}

// RegisterLogicNode registers the logic node type with the engine
func RegisterLogicNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("logic_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var operation LogicOperationType
		if opVal, exists := config["operation"]; exists {
			if opStr, ok := opVal.(string); ok {
				operation = LogicOperationType(opStr)
			}
		}

		var leftValue interface{}
		if leftVal, exists := config["left_value"]; exists {
			leftValue = leftVal
		}

		var rightValue interface{}
		if rightVal, exists := config["right_value"]; exists {
			rightValue = rightVal
		}

		var defaultResult interface{}
		if defaultVal, exists := config["default_result"]; exists {
			defaultResult = defaultVal
		}

		var regex string
		if regexVal, exists := config["regex"]; exists {
			if regexStr, ok := regexVal.(string); ok {
				regex = regexStr
			}
		}

		var conditions []LogicCondition
		if condVal, exists := config["conditions"]; exists {
			if condSlice, ok := condVal.([]interface{}); ok {
				for _, cond := range condSlice {
					if condMap, ok := cond.(map[string]interface{}); ok {
						var condition LogicCondition
						
						if left, exists := condMap["left_value"]; exists {
							condition.LeftValue = left
						}
						if op, exists := condMap["operator"]; exists {
							if opStr, ok := op.(string); ok {
								condition.Operator = LogicOperationType(opStr)
							}
						}
						if right, exists := condMap["right_value"]; exists {
							condition.RightValue = right
						}
						
						conditions = append(conditions, condition)
					}
				}
			}
		}

		var caseValues []interface{}
		if casesVal, exists := config["case_values"]; exists {
			if casesSlice, ok := casesVal.([]interface{}); ok {
				caseValues = casesSlice
			}
		}

		nodeConfig := &LogicNodeConfig{
			Operation:  operation,
			LeftValue:  leftValue,
			RightValue: rightValue,
			Conditions: conditions,
			CaseValues: caseValues,
			DefaultResult: defaultResult,
			Regex:      regex,
		}

		return NewLogicNode(nodeConfig), nil
	})
}