package flow

import (
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/nodes/base"
)

// IfElseNode implements conditional branching
type IfElseNode struct {
	*base.BaseNode
}

// IfElseConfig holds if/else configuration
type IfElseConfig struct {
	Condition string      `json:"condition"` // expression to evaluate
	Operator  string      `json:"operator"`  // ==, !=, >, <, >=, <=, contains
	Value1    interface{} `json:"value1"`
	Value2    interface{} `json:"value2"`
}

// NewIfElseNode creates a new if/else node
func NewIfElseNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "if_else",
		Name:        "If/Else Condition",
		Category:    "flow",
		Description: "Conditional branching based on condition",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "git-branch",
		Color:       "#0ea5e9",
		Inputs: []base.NodeInput{
			{
				ID:          "data",
				Name:        "Data",
				Type:        "any",
				Required:    false,
				Description: "Input data",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "true",
				Name:        "True",
				Type:        "any",
				Description: "Output if condition is true",
			},
			{
				ID:          "false",
				Name:        "False",
				Type:        "any",
				Description: "Output if condition is false",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "operator",
				Label:       "Operator",
				Description: "Comparison operator",
				Type:        "select",
				Required:    true,
				Options: []base.ConfigOption{
					{Label: "Equal (==)", Value: "=="},
					{Label: "Not Equal (!=)", Value: "!="},
					{Label: "Greater Than (>)", Value: ">"},
					{Label: "Less Than (<)", Value: "<"},
					{Label: "Greater or Equal (>=)", Value: ">="},
					{Label: "Less or Equal (<=)", Value: "<="},
					{Label: "Contains", Value: "contains"},
				},
			},
			{
				Name:        "value1",
				Label:       "Value 1",
				Description: "First value to compare",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "value2",
				Label:       "Value 2",
				Description: "Second value to compare",
				Type:        "string",
				Required:    true,
			},
		},
		Tags: []string{"condition", "if", "else", "logic"},
	}

	return &IfElseNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute evaluates condition and routes accordingly
func (n *IfElseNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config IfElseConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Evaluate condition
	conditionResult, err := n.evaluateCondition(config)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Route based on condition
	var outputPort string
	if conditionResult {
		outputPort = "true"
	} else {
		outputPort = "false"
	}

	result := map[string]interface{}{
		"condition": conditionResult,
		"output":    outputPort,
		"data":      inputs["data"],
	}

	ctx.Logger.Info("Condition evaluated", map[string]interface{}{
		"result": conditionResult,
		"port":   outputPort,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// evaluateCondition evaluates the condition
func (n *IfElseNode) evaluateCondition(config IfElseConfig) (bool, error) {
	switch config.Operator {
	case "==":
		return fmt.Sprint(config.Value1) == fmt.Sprint(config.Value2), nil
	case "!=":
		return fmt.Sprint(config.Value1) != fmt.Sprint(config.Value2), nil
	case ">":
		v1, ok1 := config.Value1.(float64)
		v2, ok2 := config.Value2.(float64)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("values must be numbers for > operator")
		}
		return v1 > v2, nil
	case "<":
		v1, ok1 := config.Value1.(float64)
		v2, ok2 := config.Value2.(float64)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("values must be numbers for < operator")
		}
		return v1 < v2, nil
	case ">=":
		v1, ok1 := config.Value1.(float64)
		v2, ok2 := config.Value2.(float64)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("values must be numbers for >= operator")
		}
		return v1 >= v2, nil
	case "<=":
		v1, ok1 := config.Value1.(float64)
		v2, ok2 := config.Value2.(float64)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("values must be numbers for <= operator")
		}
		return v1 <= v2, nil
	case "contains":
		return contains(fmt.Sprint(config.Value1), fmt.Sprint(config.Value2)), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", config.Operator)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
