package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// UUIDGeneratorNodeConfig represents the configuration for a UUIDGenerator node
type UUIDGeneratorNodeConfig struct {
	Version      int    `json:"version"`     // 1, 3, 4, or 5 (default is 4)
	NameSpace    string `json:"namespace,omitempty"` // for versions 3 and 5
	Name         string `json:"name,omitempty"`      // for versions 3 and 5
	Count        int    `json:"count,omitempty"`     // number of UUIDs to generate (default 1)
	OutputFormat string `json:"output_format,omitempty"` // "string" or "object" (default "string")
}

// UUIDGeneratorNode generates UUIDs using google/uuid
type UUIDGeneratorNode struct {
	config UUIDGeneratorNodeConfig
}

// NewUUIDGeneratorNode creates a new UUIDGenerator node with the given configuration
func NewUUIDGeneratorNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Extract config values
	version := 4 // Default to version 4 (random)
	if v, exists := config["version"]; exists {
		if vFloat, ok := v.(float64); ok {
			version = int(vFloat)
		}
	}

	nameSpace := getStringValue(config["namespace"], "")
	name := getStringValue(config["name"], "")

	count := 1
	if c, exists := config["count"]; exists {
		if cFloat, ok := c.(float64); ok {
			count = int(cFloat)
		}
	}

	outputFormat := getStringValue(config["output_format"], "string")

	// Validate version
	if version != 1 && version != 3 && version != 4 && version != 5 {
		return nil, fmt.Errorf("invalid UUID version: %d, only 1, 3, 4, or 5 are supported", version)
	}

	// Set defaults
	if count <= 0 {
		count = 1
	}

	if outputFormat == "" {
		outputFormat = "string"
	}

	uuidConfig := UUIDGeneratorNodeConfig{
		Version:      version,
		NameSpace:    nameSpace,
		Name:         name,
		Count:        count,
		OutputFormat: outputFormat,
	}

	return &UUIDGeneratorNode{
		config: uuidConfig,
	}, nil
}

// Execute implements the NodeInstance interface
func (u *UUIDGeneratorNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var uuids []interface{}

	// Override config with input values if provided
	version := u.config.Version
	if inputVersion, exists := input["version"]; exists {
		if inputVersionFloat, ok := inputVersion.(float64); ok {
			version = int(inputVersionFloat)
		}
	}

	count := u.config.Count
	if inputCount, exists := input["count"]; exists {
		if inputCountFloat, ok := inputCount.(float64); ok {
			count = int(inputCountFloat)
		}
	}

	namespaceStr := u.config.NameSpace
	if inputNamespace, exists := input["namespace"]; exists {
		if inputNamespaceStr, ok := inputNamespace.(string); ok {
			namespaceStr = inputNamespaceStr
		}
	}

	name := u.config.Name
	if inputName, exists := input["name"]; exists {
		if inputNameStr, ok := inputName.(string); ok {
			name = inputNameStr
		}
	}

	outputFormat := u.config.OutputFormat
	if inputFormat, exists := input["output_format"]; exists {
		if inputFormatStr, ok := inputFormat.(string); ok {
			outputFormat = inputFormatStr
		}
	}

	// Generate UUIDs
	for i := 0; i < count; i++ {
		var generatedUUID uuid.UUID

		switch version {
		case 1:
			generatedUUID = uuid.NewUUID()
		case 3:
			// Use namespace and name to generate UUID
			ns := getNamespace(namespaceStr)
			generatedUUID = uuid.NewMD5(ns, []byte(name))
		case 4:
			generatedUUID = uuid.New()
		case 5:
			// Use namespace and name to generate UUID
			ns := getNamespace(namespaceStr)
			generatedUUID = uuid.NewSHA1(ns, []byte(name))
		default:
			// Default to version 4
			generatedUUID = uuid.New()
		}

		// Format output based on configuration
		if outputFormat == "object" {
			uuidObj := map[string]interface{}{
				"uuid":      generatedUUID.String(),
				"version":   version,
				"timestamp": time.Now().Unix(),
			}
			uuids = append(uuids, uuidObj)
		} else {
			uuids = append(uuids, generatedUUID.String())
		}
	}

	var resultData interface{}
	if count == 1 {
		// If only one UUID was requested, return it directly
		resultData = uuids[0]
	} else {
		// If multiple UUIDs were requested, return as array
		resultData = uuids
	}

	return map[string]interface{}{
		"success": true,
		"generated_uuids": resultData,
		"count":          count,
		"version":        version,
		"timestamp":      time.Now().Unix(),
	}, nil
}

// getNamespace returns the UUID namespace based on the namespace string
func getNamespace(nsStr string) uuid.UUID {
	switch nsStr {
	case "dns":
		return uuid.NameSpaceDNS
	case "url":
		return uuid.NameSpaceURL
	case "oid":
		return uuid.NameSpaceOID
	case "x500":
		return uuid.NameSpaceX500
	default:
		// If invalid namespace string, use DNS as default or try to parse as UUID
		if parsedUUID, err := uuid.Parse(nsStr); err == nil {
			return parsedUUID
		}
		return uuid.NameSpaceDNS // default to DNS namespace
	}
}

// getStringValue safely extracts a string value
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}