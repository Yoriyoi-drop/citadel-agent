package transform

import (
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// DataMapperNode implements data field mapping
type DataMapperNode struct {
	*base.BaseNode
}

// MapperConfig holds data mapper configuration
type MapperConfig struct {
	Mappings map[string]string `json:"mappings"` // source_field -> target_field
}

// NewDataMapperNode creates a new data mapper node
func NewDataMapperNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "data_mapper",
		Name:        "Data Mapper",
		Category:    "transform",
		Description: "Map data fields from source to target structure",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "shuffle",
		Color:       "#14b8a6",
		Inputs: []base.NodeInput{
			{
				ID:          "data",
				Name:        "Data",
				Type:        "object",
				Required:    true,
				Description: "Source data to map",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "mapped",
				Name:        "Mapped Data",
				Type:        "object",
				Description: "Mapped data",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "mappings",
				Label:       "Field Mappings",
				Description: "Field mappings (JSON object: source -> target)",
				Type:        "textarea",
				Required:    true,
			},
		},
		Tags: []string{"mapper", "transform", "data"},
	}

	return &DataMapperNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute maps data fields
func (n *DataMapperNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config MapperConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Get source data
	sourceData, ok := inputs["data"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf("invalid source data")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Map fields
	mappedData := make(map[string]interface{})
	for sourceField, targetField := range config.Mappings {
		if value, exists := sourceData[sourceField]; exists {
			mappedData[targetField] = value
		}
	}

	result := map[string]interface{}{
		"mapped": mappedData,
		"count":  len(mappedData),
	}

	ctx.Logger.Info("Data mapped successfully", map[string]interface{}{
		"fields": len(mappedData),
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
