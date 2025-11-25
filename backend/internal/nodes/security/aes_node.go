package security

import (
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// AESEncryptNodeV2 implements AES encryption (New System)
type AESEncryptNodeV2 struct {
	*base.BaseNode
}

// NewAESEncryptNode creates a new AES encryption node
func NewAESEncryptNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "aes_encrypt",
		Name:        "AES Encrypt",
		Category:    "security",
		Description: "Encrypt/Decrypt data using AES",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "lock",
		Color:       "#ef4444",
		Inputs: []base.NodeInput{
			{
				ID:          "data",
				Name:        "Data",
				Type:        "string",
				Required:    true,
				Description: "Data to process",
			},
			{
				ID:          "key",
				Name:        "Key",
				Type:        "string",
				Required:    true,
				Description: "Encryption key",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "result",
				Name:        "Result",
				Type:        "string",
				Description: "Processed data",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "mode",
				Label:       "Mode",
				Description: "Operation mode",
				Type:        "select",
				Required:    true,
				Default:     "encrypt",
				Options: []base.ConfigOption{
					{Label: "Encrypt", Value: "encrypt"},
					{Label: "Decrypt", Value: "decrypt"},
				},
			},
			{
				Name:        "algorithm",
				Label:       "Algorithm",
				Description: "AES Algorithm",
				Type:        "select",
				Required:    true,
				Default:     "aes256",
				Options: []base.ConfigOption{
					{Label: "AES-256", Value: "aes256"},
					{Label: "AES-192", Value: "aes192"},
					{Label: "AES-128", Value: "aes128"},
				},
			},
		},
		Tags: []string{"security", "encryption", "aes"},
	}

	return &AESEncryptNodeV2{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute performs AES encryption/decryption
func (n *AESEncryptNodeV2) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config struct {
		Mode      string `json:"mode"`
		Algorithm string `json:"algorithm"`
	}
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	data, ok := inputs["data"].(string)
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Data must be a string"}, time.Since(startTime)), nil
	}

	key, ok := inputs["key"].(string)
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Key must be a string"}, time.Since(startTime)), nil
	}

	// Use existing EncryptionNode logic by instantiating it
	// This is a bit inefficient but reuses logic.
	// Ideally we should refactor EncryptionNode to be stateless or separate logic.
	// For now, we'll just instantiate it.

	encNode, err := NewEncryptionNode(map[string]interface{}{
		"mode":      config.Mode,
		"algorithm": config.Algorithm,
		"key":       key,
	})
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Cast to concrete type to access internal methods if needed, or just use Execute interface if it matched
	// But EncryptionNode.Execute signature is different (context.Context vs *base.ExecutionContext)
	// We need to call internal methods encryptData/decryptData which are exported? No, they are private.
	// But Execute is public.

	// Let's use the public Execute method of EncryptionNode
	// It expects inputs map.

	encInputs := map[string]interface{}{
		"data": data,
		"key":  key,
	}

	// EncryptionNode expects config to be set in struct, but also allows overrides in inputs.
	// We set config in NewEncryptionNode.

	// We need to adapt context
	// EncryptionNode.Execute takes context.Context

	// Wait, EncryptionNode.Execute returns map[string]interface{}

	// We need to cast encNode to *EncryptionNode to call Execute
	en, ok := encNode.(*EncryptionNode)
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Failed to cast encryption node"}, time.Since(startTime)), nil
	}

	resultMap, err := en.Execute(ctx.Context, encInputs)
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Extract result
	result := resultMap["result"]

	return base.CreateSuccessResult(map[string]interface{}{
		"result": result,
	}, time.Since(startTime)), nil
}
