package utility

import (
	"context"
	"fmt"
	"strings"
	"time"

	"citadel-agent/backend/internal/interfaces"
)

// IfElseNode implements a node that evaluates conditions and returns different outputs
type IfElseNode struct {
	id           string
	nodeType     string
	condition    string
	leftValue    interface{}
	operator     string
	rightValue   interface{}
	trueResult   interface{}
	falseResult  interface{}
	config       map[string]interface{}
}

// Initialize sets up the if/else node with configuration
func (ie *IfElseNode) Initialize(config map[string]interface{}) error {
	ie.config = config

	if condition, ok := config["condition"]; ok {
		if cond, ok := condition.(string); ok {
			ie.condition = cond
		} else {
			return fmt.Errorf("condition must be a string")
		}
	}

	if leftValue, ok := config["left_value"]; ok {
		ie.leftValue = leftValue
	}

	if operator, ok := config["operator"]; ok {
		if op, ok := operator.(string); ok {
			ie.operator = op
		} else {
			return fmt.Errorf("operator must be a string")
		}
	}

	if rightValue, ok := config["right_value"]; ok {
		ie.rightValue = rightValue
	}

	if trueResult, ok := config["true_result"]; ok {
		ie.trueResult = trueResult
	}

	if falseResult, ok := config["false_result"]; ok {
		ie.falseResult = falseResult
	}

	return nil
}

// Execute evaluates the condition and returns appropriate result
func (ie *IfElseNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// If condition is provided as a string, evaluate it
	// Otherwise, use the structured condition
	var conditionResult bool

	if ie.condition != "" {
		// For now, we'll support simple template replacement in condition
		// A more robust solution would involve a proper expression evaluator
		conditionEvaluated := ie.replaceTemplateVariables(ie.condition, inputs)

		// For this implementation, we'll check if the condition is truthy when evaluated
		conditionResult = conditionEvaluated != "" && conditionEvaluated != "false" && conditionEvaluated != "0"
	} else {
		// Use structured condition with left operand, operator, and right operand
		conditionResult = ie.evaluateStructuredCondition(inputs)
	}

	// Prepare the result based on the condition
	result := make(map[string]interface{})
	result["condition_result"] = conditionResult
	result["input_data"] = inputs

	if conditionResult {
		result["result"] = ie.trueResult
		result["branch"] = "true"
	} else {
		result["result"] = ie.falseResult
		result["branch"] = "false"
	}

	return result, nil
}

// evaluateStructuredCondition evaluates the structured condition
func (ie *IfElseNode) evaluateStructuredCondition(inputData map[string]interface{}) bool {
	// Replace template variables in operands if they're strings
	var leftVal, rightVal interface{}

	if leftStr, isStr := ie.leftValue.(string); isStr {
		leftVal = ie.replaceTemplateVariables(leftStr, inputData)
	} else {
		leftVal = ie.leftValue
	}

	if rightStr, isStr := ie.rightValue.(string); isStr {
		rightVal = ie.replaceTemplateVariables(rightStr, inputData)
	} else {
		rightVal = ie.rightValue
	}

	// Evaluate based on operator
	switch ie.operator {
	case "==", "eq":
		return ie.compareValues(leftVal, rightVal) == 0
	case "!=", "ne":
		return ie.compareValues(leftVal, rightVal) != 0
	case ">", "gt":
		return ie.compareValues(leftVal, rightVal) > 0
	case "<", "lt":
		return ie.compareValues(leftVal, rightVal) < 0
	case ">=", "ge":
		return ie.compareValues(leftVal, rightVal) >= 0
	case "<=", "le":
		return ie.compareValues(leftVal, rightVal) <= 0
	case "contains":
		return ie.contains(leftVal, rightVal)
	case "starts_with":
		return ie.startsWith(leftVal, rightVal)
	case "ends_with":
		return ie.endsWith(leftVal, rightVal)
	case "is_empty":
		return ie.isEmpty(leftVal)
	case "is_null":
		return leftVal == nil
	default:
		// Default to false for unknown operators
		return false
	}
}

// compareValues compares two values and returns -1, 0, or 1
func (ie *IfElseNode) compareValues(a, b interface{}) int {
	// Convert both values to strings for comparison if they're not numbers
	// In a production system, you'd want more sophisticated type handling

	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Try to convert to float64 for numeric comparison
	var aFloat, bFloat float64
	var aOk, bOk bool

	if aNum, isNum := a.(float64); isNum {
		aFloat, aOk = aNum, true
	} else if aNum, isNum := a.(int); isNum {
		aFloat, aOk = float64(aNum), true
	} else {
		// Not a number, convert to string
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	if bNum, isNum := b.(float64); isNum {
		bFloat, bOk = bNum, true
	} else if bNum, isNum := b.(int); isNum {
		bFloat, bOk = float64(bNum), true
	} else {
		// Not a number, convert to string
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	if !aOk || !bOk {
		// At least one wasn't a number, fall back to string comparison
		aStr := fmt.Sprintf("%v", a)
		bStr := fmt.Sprintf("%v", b)
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

// contains checks if left value contains right value
func (ie *IfElseNode) contains(left, right interface{}) bool {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return containsSubstring(leftStr, rightStr)
}

// startsWith checks if left value starts with right value
func (ie *IfElseNode) startsWith(left, right interface{}) bool {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return len(leftStr) >= len(rightStr) && leftStr[:len(rightStr)] == rightStr
}

// endsWith checks if left value ends with right value
func (ie *IfElseNode) endsWith(left, right interface{}) bool {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return len(leftStr) >= len(rightStr) && leftStr[len(leftStr)-len(rightStr):] == rightStr
}

// isEmpty checks if the value is empty
func (ie *IfElseNode) isEmpty(val interface{}) bool {
	if val == nil {
		return true
	}
	str := fmt.Sprintf("%v", val)
	return str == ""
}

// replaceTemplateVariables replaces template variables in the string with values from input data
func (ie *IfElseNode) replaceTemplateVariables(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	return result
}

// GetType returns the type of the node
func (ie *IfElseNode) GetType() string {
	return ie.nodeType
}

// GetID returns the unique identifier for this node instance
func (ie *IfElseNode) GetID() string {
	return ie.id
}

// NewIfElseNode creates a new if/else node constructor for the registry
func NewIfElseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	node := &IfElseNode{
		id:       fmt.Sprintf("ifelse_%d", time.Now().UnixNano()),
		nodeType: "if_else",
	}

	if err := node.Initialize(config); err != nil {
		return nil, err
	}

	return node, nil
}

// containsSubstring checks if the main string contains the substring
func containsSubstring(main, sub string) bool {
	return len(sub) == 0 || len(main) >= len(sub) &&
		(main == sub ||
			containsSubstring(main[1:], sub) ||
			(main[:len(sub)] == sub))
}