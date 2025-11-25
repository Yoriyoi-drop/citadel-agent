package flow

import (
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// DelayNode implements delay/wait functionality
type DelayNode struct {
	*base.BaseNode
}

// DelayConfig holds delay configuration
type DelayConfig struct {
	Duration int    `json:"duration"` // milliseconds
	Unit     string `json:"unit"`     // ms, s, m, h
}

// NewDelayNode creates a new delay node
func NewDelayNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "delay",
		Name:        "Delay/Wait",
		Category:    "flow",
		Description: "Add delay before continuing execution",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "clock",
		Color:       "#0ea5e9",
		Inputs: []base.NodeInput{
			{
				ID:          "trigger",
				Name:        "Trigger",
				Type:        "any",
				Required:    false,
				Description: "Trigger the delay",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "output",
				Name:        "Output",
				Type:        "any",
				Description: "Output after delay",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "duration",
				Label:       "Duration",
				Description: "Delay duration",
				Type:        "number",
				Required:    true,
				Default:     1000,
			},
			{
				Name:        "unit",
				Label:       "Unit",
				Description: "Time unit",
				Type:        "select",
				Required:    true,
				Default:     "ms",
				Options: []base.ConfigOption{
					{Label: "Milliseconds", Value: "ms"},
					{Label: "Seconds", Value: "s"},
					{Label: "Minutes", Value: "m"},
					{Label: "Hours", Value: "h"},
				},
			},
		},
		Tags: []string{"delay", "wait", "sleep"},
	}

	return &DelayNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute adds delay
func (n *DelayNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config DelayConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Calculate delay duration
	var duration time.Duration
	switch config.Unit {
	case "ms":
		duration = time.Duration(config.Duration) * time.Millisecond
	case "s":
		duration = time.Duration(config.Duration) * time.Second
	case "m":
		duration = time.Duration(config.Duration) * time.Minute
	case "h":
		duration = time.Duration(config.Duration) * time.Hour
	default:
		err := fmt.Errorf("invalid time unit: %s", config.Unit)
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	ctx.Logger.Info("Starting delay", map[string]interface{}{
		"duration": duration.String(),
	})

	// Wait with context cancellation support
	select {
	case <-time.After(duration):
		// Delay completed
	case <-ctx.Context.Done():
		err := fmt.Errorf("delay cancelled")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	result := map[string]interface{}{
		"delayed_for": duration.String(),
		"data":        inputs,
	}

	ctx.Logger.Info("Delay completed", map[string]interface{}{
		"duration": duration.String(),
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
