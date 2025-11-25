package transform

import (
	"encoding/xml"
	"fmt"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// XMLParserNode implements XML parsing
type XMLParserNode struct {
	*base.BaseNode
}

// XMLConfig holds XML parser configuration
type XMLConfig struct {
	Data string `json:"data"`
}

// NewXMLParserNode creates a new XML parser node
func NewXMLParserNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "xml_parse",
		Name:        "XML Parser",
		Category:    "transform",
		Description: "Parse XML data into JSON",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "code",
		Color:       "#14b8a6",
		Inputs: []base.NodeInput{
			{
				ID:          "data",
				Name:        "XML Data",
				Type:        "string",
				Required:    true,
				Description: "XML data to parse",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "result",
				Name:        "Result",
				Type:        "object",
				Description: "Parsed XML as JSON",
			},
		},
		Config: []base.NodeConfig{},
		Tags:   []string{"xml", "parse", "transform"},
	}

	return &XMLParserNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute parses XML data
func (n *XMLParserNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Get XML data from inputs
	xmlData, ok := inputs["data"].(string)
	if !ok {
		err := fmt.Errorf("XML data is required")
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Parse XML into generic map
	var result map[string]interface{}
	if err := xml.Unmarshal([]byte(xmlData), &result); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	output := map[string]interface{}{
		"result": result,
	}

	ctx.Logger.Info("XML parsed successfully", nil)

	return base.CreateSuccessResult(output, time.Since(startTime)), nil
}
