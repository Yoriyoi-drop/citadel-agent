package transform

import (
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// JSONParserNode implements JSON parsing
type JSONParserNode struct {
	*base.BaseNode
	transformer *JSONTransformer
}

// NewJSONParserNode creates a new JSON parser node
func NewJSONParserNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "json_parser",
		Name:        "JSON Parser",
		Category:    "transform",
		Description: "Parse, stringify, or manipulate JSON",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "code",
		Color:       "#f59e0b",
		Inputs: []base.NodeInput{
			{
				ID:          "input",
				Name:        "Input",
				Type:        "any",
				Required:    true,
				Description: "Input data (string or object)",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "output",
				Name:        "Output",
				Type:        "any",
				Description: "Transformed data",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "operation",
				Label:       "Operation",
				Description: "Transformation operation",
				Type:        "select",
				Required:    true,
				Default:     "parse",
				Options: []base.ConfigOption{
					{Label: "Parse (String to JSON)", Value: "parse"},
					{Label: "Stringify (JSON to String)", Value: "stringify"},
					{Label: "Minify", Value: "minify"},
					{Label: "Pretty Print", Value: "pretty"},
				},
			},
		},
		Tags: []string{"json", "transform", "parser"},
	}

	return &JSONParserNode{
		BaseNode:    base.NewBaseNode(metadata),
		transformer: NewJSONTransformer(),
	}
}

// Execute performs JSON transformation
func (n *JSONParserNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config struct {
		Operation string `json:"operation"`
	}
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	input := inputs["input"]
	var result interface{}
	var err error

	switch config.Operation {
	case "parse":
		if str, ok := input.(string); ok {
			result, err = n.transformer.Parse(str)
		} else {
			// Already an object, pass through
			result = input
		}

	case "stringify":
		result, err = n.transformer.Stringify(input)

	case "minify":
		if str, ok := input.(string); ok {
			result, err = n.transformer.Minify(str)
		} else {
			// Convert to string first
			str, _ := n.transformer.Stringify(input)
			result, err = n.transformer.Minify(str)
		}

	case "pretty":
		if str, ok := input.(string); ok {
			result, err = n.transformer.PrettyPrint(str)
		} else {
			// Convert to string first
			str, _ := n.transformer.Stringify(input)
			result, err = n.transformer.PrettyPrint(str)
		}

	default:
		err = fmt.Errorf("unknown operation: %s", config.Operation)
	}

	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	return base.CreateSuccessResult(map[string]interface{}{
		"output": result,
	}, time.Since(startTime)), nil
}
