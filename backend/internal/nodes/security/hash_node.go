package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// HashSHA256Node implements SHA256 hashing
type HashSHA256Node struct {
	*base.BaseNode
}

// NewHashSHA256Node creates a new SHA256 hash node
func NewHashSHA256Node() base.Node {
	metadata := base.NodeMetadata{
		ID:          "hash_sha256",
		Name:        "SHA256 Hash",
		Category:    "security",
		Description: "Hash data using SHA256",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "hash",
		Color:       "#ef4444",
		Inputs: []base.NodeInput{
			{
				ID:          "data",
				Name:        "Data",
				Type:        "string",
				Required:    true,
				Description: "Data to hash",
			},
			{
				ID:          "secret",
				Name:        "Secret",
				Type:        "string",
				Required:    false,
				Description: "HMAC Secret (optional)",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "hash",
				Name:        "Hash",
				Type:        "string",
				Description: "Hashed data (hex)",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "encoding",
				Label:       "Encoding",
				Description: "Output encoding",
				Type:        "select",
				Required:    true,
				Default:     "hex",
				Options: []base.ConfigOption{
					{Label: "Hex", Value: "hex"},
					{Label: "Base64", Value: "base64"},
				},
			},
		},
		Tags: []string{"security", "hash", "sha256", "hmac"},
	}

	return &HashSHA256Node{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute performs hashing
func (n *HashSHA256Node) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	data, ok := inputs["data"].(string)
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Data must be a string"}, time.Since(startTime)), nil
	}

	secret, _ := inputs["secret"].(string)

	var result string

	if secret != "" {
		// HMAC
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(data))
		result = hex.EncodeToString(h.Sum(nil))
	} else {
		// SHA256
		hash := sha256.Sum256([]byte(data))
		result = hex.EncodeToString(hash[:])
	}

	return base.CreateSuccessResult(map[string]interface{}{
		"hash": result,
	}, time.Since(startTime)), nil
}
