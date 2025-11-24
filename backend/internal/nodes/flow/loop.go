package flow

import (
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/nodes/base"
)

// ForEachNode implements iteration over collections
type ForEachNode struct {
	*base.BaseNode
}

// ForEachConfig holds for-each configuration
type ForEachConfig struct {
	BatchSize int `json:"batch_size"` // process items in batches
}

// NewForEachNode creates a new for-each node
func NewForEachNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "for_each",
		Name:        "For Each Loop",
		Category:    "flow",
		Description: "Iterate over collection items",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "repeat",
		Color:       "#0ea5e9",
		Inputs: []base.NodeInput{
			{
				ID:          "items",
				Name:        "Items",
				Type:        "array",
				Required:    true,
				Description: "Collection to iterate",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "item",
				Name:        "Current Item",
				Type:        "any",
				Description: "Current iteration item",
			},
			{
				ID:          "index",
				Name:        "Index",
				Type:        "number",
				Description: "Current index",
			},
			{
				ID:          "results",
				Name:        "Results",
				Type:        "array",
				Description: "All iteration results",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "batch_size",
				Label:       "Batch Size",
				Description: "Process items in batches (0 = all at once)",
				Type:        "number",
				Required:    false,
				Default:     0,
			},
		},
		Tags: []string{"loop", "iteration", "foreach"},
	}

	return &ForEachNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute iterates over items
func (n *ForEachNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config ForEachConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Get items
	items, ok := inputs["items"].([]interface{})
	if !ok {
		err := fmt.Errorf("items must be an array")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Process items
	results := make([]interface{}, 0, len(items))

	for index, item := range items {
		// Check context cancellation
		select {
		case <-ctx.Context.Done():
			err := fmt.Errorf("iteration cancelled")
			return base.CreateErrorResult(err, time.Since(startTime)), err
		default:
		}

		// Process item (in real implementation, this would trigger child nodes)
		results = append(results, map[string]interface{}{
			"item":  item,
			"index": index,
		})

		// Log progress
		if (index+1)%100 == 0 {
			ctx.Logger.Debug("Iteration progress", map[string]interface{}{
				"processed": index + 1,
				"total":     len(items),
			})
		}
	}

	result := map[string]interface{}{
		"results": results,
		"count":   len(results),
	}

	ctx.Logger.Info("For-each completed", map[string]interface{}{
		"items": len(items),
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
