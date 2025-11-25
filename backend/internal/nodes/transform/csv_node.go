package transform

import (
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// CSVParserNode implements CSV parsing
type CSVParserNode struct {
	*base.BaseNode
	transformer *CSVTransformer
}

// NewCSVParserNode creates a new CSV parser node
func NewCSVParserNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "csv_parser",
		Name:        "CSV Parser",
		Category:    "transform",
		Description: "Parse or stringify CSV data",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "table",
		Color:       "#10b981",
		Inputs: []base.NodeInput{
			{
				ID:          "input",
				Name:        "Input",
				Type:        "any",
				Required:    true,
				Description: "Input data (string or array of objects)",
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
					{Label: "Parse (CSV to JSON)", Value: "parse"},
					{Label: "Stringify (JSON to CSV)", Value: "stringify"},
				},
			},
			{
				Name:        "delimiter",
				Label:       "Delimiter",
				Description: "CSV delimiter",
				Type:        "string",
				Required:    true,
				Default:     ",",
			},
			{
				Name:        "has_header",
				Label:       "Has Header",
				Description: "First row is header",
				Type:        "boolean",
				Required:    true,
				Default:     true,
			},
		},
		Tags: []string{"csv", "transform", "parser"},
	}

	return &CSVParserNode{
		BaseNode: base.NewBaseNode(metadata),
		transformer: NewCSVTransformer(CSVConfig{
			Delimiter: ',',
			HasHeader: true,
		}),
	}
}

// Execute performs CSV transformation
func (n *CSVParserNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config struct {
		Operation string `json:"operation"`
		Delimiter string `json:"delimiter"`
		HasHeader bool   `json:"has_header"`
	}
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Update transformer config
	delimiter := ','
	if len(config.Delimiter) > 0 {
		delimiter = rune(config.Delimiter[0])
	}
	n.transformer = NewCSVTransformer(CSVConfig{
		Delimiter: delimiter,
		HasHeader: config.HasHeader,
	})

	input := inputs["input"]
	var result interface{}
	var err error

	switch config.Operation {
	case "parse":
		if str, ok := input.(string); ok {
			result, err = n.transformer.Parse(str)
		} else {
			err = fmt.Errorf("input must be a string for parse operation")
		}

	case "stringify":
		if arr, ok := input.([]interface{}); ok {
			// Convert []interface{} to []map[string]interface{}
			var maps []map[string]interface{}
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					maps = append(maps, m)
				}
			}
			result, err = n.transformer.Stringify(maps)
		} else if maps, ok := input.([]map[string]interface{}); ok {
			result, err = n.transformer.Stringify(maps)
		} else {
			err = &base.ExecutionError{Message: "Input must be an array of objects for stringify operation"}
		}

	default:
		err = &base.ExecutionError{Message: "Unknown operation: " + config.Operation}
	}

	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	return base.CreateSuccessResult(map[string]interface{}{
		"output": result,
	}, time.Since(startTime)), nil
}
