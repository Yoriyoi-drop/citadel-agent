package utility

import (
	"math/rand"
	"time"

	"citadel-agent/backend/internal/nodes/base"
	"github.com/google/uuid"
)

// SetVariableNode implements variable storage
type SetVariableNode struct {
	*base.BaseNode
}

// VarConfig holds variable configuration
type VarConfig struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// NewSetVariableNode creates set variable node
func NewSetVariableNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "set_variable",
		Name:        "Set Variable",
		Category:    "utility",
		Description: "Set workflow variable",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "settings",
		Color:       "#64748b",
		Inputs: []base.NodeInput{
			{
				ID:          "value",
				Name:        "Value",
				Type:        "any",
				Required:    true,
				Description: "Value to store",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "value",
				Name:        "Value",
				Type:        "any",
				Description: "Stored value",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "name",
				Label:       "Variable Name",
				Description: "Name of the variable",
				Type:        "string",
				Required:    true,
			},
		},
		Tags: []string{"variable", "storage", "utility"},
	}

	return &SetVariableNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute sets variable
func (n *SetVariableNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	var config VarConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	value := inputs["value"]

	// Store in context variables
	ctx.Variables[config.Name] = value

	result := map[string]interface{}{
		"name":  config.Name,
		"value": value,
	}

	ctx.Logger.Info("Variable set", map[string]interface{}{
		"name": config.Name,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// UUIDNode generates UUIDs
type UUIDNode struct {
	*base.BaseNode
}

// NewUUIDNode creates UUID generator node
func NewUUIDNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "uuid_generate",
		Name:        "UUID Generator",
		Category:    "utility",
		Description: "Generate UUID",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "hash",
		Color:       "#64748b",
		Inputs:      []base.NodeInput{},
		Outputs: []base.NodeOutput{
			{
				ID:          "uuid",
				Name:        "UUID",
				Type:        "string",
				Description: "Generated UUID",
			},
		},
		Config: []base.NodeConfig{},
		Tags:   []string{"uuid", "generator", "utility"},
	}

	return &UUIDNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute generates UUID
func (n *UUIDNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	id := uuid.New()

	result := map[string]interface{}{
		"uuid": id.String(),
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// RandomNumberNode generates random numbers
type RandomNumberNode struct {
	*base.BaseNode
}

// RandomConfig holds random configuration
type RandomConfig struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// NewRandomNumberNode creates random number node
func NewRandomNumberNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "random_number",
		Name:        "Random Number",
		Category:    "utility",
		Description: "Generate random number",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "hash",
		Color:       "#64748b",
		Inputs:      []base.NodeInput{},
		Outputs: []base.NodeOutput{
			{
				ID:          "number",
				Name:        "Number",
				Type:        "number",
				Description: "Random number",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "min",
				Label:       "Minimum",
				Description: "Minimum value",
				Type:        "number",
				Required:    false,
				Default:     0,
			},
			{
				Name:        "max",
				Label:       "Maximum",
				Description: "Maximum value",
				Type:        "number",
				Required:    false,
				Default:     100,
			},
		},
		Tags: []string{"random", "number", "utility"},
	}

	return &RandomNumberNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute generates random number
func (n *RandomNumberNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	var config RandomConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Generate random number
	num := rand.Intn(config.Max-config.Min+1) + config.Min

	result := map[string]interface{}{
		"number": num,
		"min":    config.Min,
		"max":    config.Max,
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}

// DateTimeNode gets current date/time
type DateTimeNode struct {
	*base.BaseNode
}

// NewDateTimeNode creates date/time node
func NewDateTimeNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "datetime_now",
		Name:        "Date/Time Now",
		Category:    "utility",
		Description: "Get current date and time",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "calendar",
		Color:       "#64748b",
		Inputs:      []base.NodeInput{},
		Outputs: []base.NodeOutput{
			{
				ID:          "datetime",
				Name:        "DateTime",
				Type:        "object",
				Description: "Current date/time",
			},
		},
		Config: []base.NodeConfig{},
		Tags:   []string{"date", "time", "utility"},
	}

	return &DateTimeNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute gets current date/time
func (n *DateTimeNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	now := time.Now()

	result := map[string]interface{}{
		"timestamp": now.Unix(),
		"iso":       now.Format(time.RFC3339),
		"date":      now.Format("2006-01-02"),
		"time":      now.Format("15:04:05"),
		"year":      now.Year(),
		"month":     int(now.Month()),
		"day":       now.Day(),
		"hour":      now.Hour(),
		"minute":    now.Minute(),
		"second":    now.Second(),
	}

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
