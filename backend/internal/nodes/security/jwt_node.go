package security

import (
	"time"

	"citadel-agent/backend/internal/nodes/base"
	"github.com/golang-jwt/jwt/v5"
)

// JWTSignNode implements JWT signing
type JWTSignNode struct {
	*base.BaseNode
}

// NewJWTSignNode creates a new JWT signing node
func NewJWTSignNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "jwt_sign",
		Name:        "JWT Sign",
		Category:    "security",
		Description: "Sign a JWT token",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "key",
		Color:       "#ef4444",
		Inputs: []base.NodeInput{
			{
				ID:          "payload",
				Name:        "Payload",
				Type:        "object",
				Required:    true,
				Description: "Token claims",
			},
			{
				ID:          "secret",
				Name:        "Secret",
				Type:        "string",
				Required:    true,
				Description: "Signing secret",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "token",
				Name:        "Token",
				Type:        "string",
				Description: "Signed JWT",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "expires_in",
				Label:       "Expires In (seconds)",
				Description: "Token expiration time",
				Type:        "number",
				Required:    false,
				Default:     3600,
			},
		},
		Tags: []string{"security", "jwt", "token"},
	}

	return &JWTSignNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute performs JWT signing
func (n *JWTSignNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config struct {
		ExpiresIn int `json:"expires_in"`
	}
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	payload, ok := inputs["payload"].(map[string]interface{})
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Payload must be an object"}, time.Since(startTime)), nil
	}

	secret, ok := inputs["secret"].(string)
	if !ok {
		return base.CreateErrorResult(&base.ExecutionError{Message: "Secret must be a string"}, time.Since(startTime)), nil
	}

	// Create claims
	claims := jwt.MapClaims{}
	for k, v := range payload {
		claims[k] = v
	}

	// Set expiration if not present
	if _, ok := claims["exp"]; !ok {
		if config.ExpiresIn > 0 {
			claims["exp"] = time.Now().Add(time.Duration(config.ExpiresIn) * time.Second).Unix()
		}
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	return base.CreateSuccessResult(map[string]interface{}{
		"token": signedToken,
	}, time.Since(startTime)), nil
}
