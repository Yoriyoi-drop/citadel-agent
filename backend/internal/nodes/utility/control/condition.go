package control

import (
	"fmt"
	"reflect"
)

// ConditionOperator represents the type of comparison
type ConditionOperator string

const (
	OpEqual              ConditionOperator = "=="
	OpNotEqual           ConditionOperator = "!="
	OpGreaterThan        ConditionOperator = ">"
	OpLessThan           ConditionOperator = "<"
	OpGreaterThanOrEqual ConditionOperator = ">="
	OpLessThanOrEqual    ConditionOperator = "<="
	OpContains           ConditionOperator = "contains"
)

// ConditionEvaluator evaluates conditions
type ConditionEvaluator struct{}

// Evaluate evaluates a condition
func (e *ConditionEvaluator) Evaluate(value1 interface{}, operator ConditionOperator, value2 interface{}) (bool, error) {
	switch operator {
	case OpEqual:
		return reflect.DeepEqual(value1, value2), nil
	case OpNotEqual:
		return !reflect.DeepEqual(value1, value2), nil
	// Note: Implementing other operators requires type assertions and is more complex in Go
	// For brevity, we'll handle strings and numbers for basic cases
	default:
		return false, fmt.Errorf("operator %s not fully implemented", operator)
	}
}
