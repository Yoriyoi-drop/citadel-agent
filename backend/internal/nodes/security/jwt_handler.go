// backend/internal/nodes/security/jwt_handler.go
package security

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// JWTOperationType represents the type of JWT operation
type JWTOperationType string

const (
	JWTCreate    JWTOperationType = "create"
	JWTValidate  JWTOperationType = "validate"
	JWTRefresh   JWTOperationType = "refresh"
	JWTDecode    JWTOperationType = "decode"
	JWTRevoke    JWTOperationType = "revoke"
)

// JWTAlgorithm represents the JWT signing algorithm
type JWTAlgorithm string

const (
	AlgorithmHS256 JWTAlgorithm = "HS256"
	AlgorithmHS384 JWTAlgorithm = "HS384"
	AlgorithmHS512 JWTAlgorithm = "HS512"
	AlgorithmRS256 JWTAlgorithm = "RS256"
	AlgorithmRS384 JWTAlgorithm = "RS384"
	AlgorithmRS512 JWTAlgorithm = "RS512"
	AlgorithmES256 JWTAlgorithm = "ES256"
	AlgorithmES384 JWTAlgorithm = "ES384"
	AlgorithmES512 JWTAlgorithm = "ES512"
)

// JWTConfig represents the configuration for a JWT handler node
type JWTConfig struct {
	Operation       JWTOperationType `json:"operation"`
	Algorithm       JWTAlgorithm     `json:"algorithm"`
	SecretKey       string           `json:"secret_key"`
	PublicKey       string           `json:"public_key"`
	PrivateKey      string           `json:"private_key"`
	Issuer          string           `json:"issuer"`
	Audience        []string         `json:"audience"`
	Subject         string           `json:"subject"`
	DefaultExpiry   time.Duration    `json:"default_expiry"`
	RefreshExpiry   time.Duration    `json:"refresh_expiry"`
	EnableLogging   bool             `json:"enable_logging"`
	EnableBlacklist bool             `json:"enable_blacklist"`
	Blacklist       map[string]bool  `json:"-"` // In-memory blacklist
}

// JWTHandlerNode represents a JWT handler node
type JWTHandlerNode struct {
	config *JWTConfig
}

// NewJWTHandlerNode creates a new JWT handler node
func NewJWTHandlerNode(config *JWTConfig) *JWTHandlerNode {
	if config.Algorithm == "" {
		config.Algorithm = AlgorithmHS256
	}

	if config.DefaultExpiry == 0 {
		config.DefaultExpiry = 15 * time.Minute // 15 minutes default
	}

	if config.RefreshExpiry == 0 {
		config.RefreshExpiry = 24 * time.Hour // 24 hours default
	}

	if config.Blacklist == nil {
		config.Blacklist = make(map[string]bool)
	}

	return &JWTHandlerNode{
		config: config,
	}
}

// Execute executes the JWT operation
func (jhn *JWTHandlerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := jhn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = JWTOperationType(opStr)
		}
	}

	// Perform the JWT operation based on type
	switch operation {
	case JWTCreate:
		return jhn.createJWT(inputs)
	case JWTValidate:
		return jhn.validateJWT(inputs)
	case JWTRefresh:
		return jhn.refreshJWT(inputs)
	case JWTDecode:
		return jhn.decodeJWT(inputs)
	case JWTRevoke:
		return jhn.revokeJWT(inputs)
	default:
		return nil, fmt.Errorf("unsupported JWT operation: %s", operation)
	}
}

// createJWT creates a new JWT
func (jhn *JWTHandlerNode) createJWT(inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get payload from inputs
	payloadData := make(map[string]interface{})
	if payload, exists := inputs["payload"]; exists {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			payloadData = payloadMap
		}
	}

	// Get expiration from inputs or use default
	var expiry time.Duration
	if exp, exists := inputs["expiry_seconds"]; exists {
		if expFloat, ok := exp.(float64); ok {
			expiry = time.Duration(expFloat) * time.Second
		} else {
			expiry = jhn.config.DefaultExpiry
		}
	} else {
		expiry = jhn.config.DefaultExpiry
	}

	// Get issuer from inputs or config
	issuer := jhn.config.Issuer
	if iss, exists := inputs["issuer"]; exists {
		if issStr, ok := iss.(string); ok {
			issuer = issStr
		}
	}

	// Get subject from inputs or config
	subject := jhn.config.Subject
	if sub, exists := inputs["subject"]; exists {
		if subStr, ok := sub.(string); ok {
			subject = subStr
		}
	}

	// Get audience from inputs or config
	audience := jhn.config.Audience
	if aud, exists := inputs["audience"]; exists {
		if audSlice, ok := aud.([]interface{}); ok {
			audience = make([]string, len(audSlice))
			for i, v := range audSlice {
				audience[i] = getStringValue(v)
			}
		}
	}

	// Set standard claims
	payloadData["iat"] = time.Now().Unix()
	payloadData["exp"] = time.Now().Add(expiry).Unix()
	payloadData["nbf"] = time.Now().Unix() - 10 // Allow 10 second clock skew
	if issuer != "" {
		payloadData["iss"] = issuer
	}
	if subject != "" {
		payloadData["sub"] = subject
	}
	if len(audience) > 0 {
		if len(audience) == 1 {
			payloadData["aud"] = audience[0]
		} else {
			payloadData["aud"] = audience
		}
	}

	// Create JWT header
	header := map[string]interface{}{
		"typ": "JWT",
		"alg": string(jhn.config.Algorithm),
	}

	// Encode header and payload
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadJSON, err := json.Marshal(payloadData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	signature, err := jhn.sign(strings.Join([]string{headerEncoded, payloadEncoded}, "."), jhn.config.SecretKey, jhn.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Construct JWT
	jwt := strings.Join([]string{headerEncoded, payloadEncoded, signature}, ".")

	return map[string]interface{}{
		"success":      true,
		"jwt":          jwt,
		"payload":      payloadData,
		"algorithm":    string(jhn.config.Algorithm),
		"expiry":       time.Now().Add(expiry).Unix(),
		"operation":    string(JWTCreate),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// validateJWT validates a JWT
func (jhn *JWTHandlerNode) validateJWT(inputs map[string]interface{}) (map[string]interface{}, error) {
	jwt := getStringValue(inputs["jwt"])
	if jwt == "" {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "JWT is required",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Check if token is blacklisted
	if jhn.config.EnableBlacklist && jhn.config.Blacklist[jwt] {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "JWT has been revoked",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Split JWT into parts
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid JWT format",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signature := parts[2]

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid header encoding",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid header JSON",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Extract algorithm from header
	algStr, ok := header["alg"].(string)
	if !ok {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Missing algorithm in header",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid payload encoding",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid payload JSON",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Verify signature
	message := strings.Join([]string{headerEncoded, payloadEncoded}, ".")
	if valid, err := jhn.verifySignature(message, signature, algStr, jhn.config.SecretKey, jhn.config.PublicKey); err != nil || !valid {
		return map[string]interface{}{
			"success":      false,
			"valid":        false,
			"error":        "Invalid signature",
			"operation":    string(JWTValidate),
			"timestamp":    time.Now().Unix(),
		}, nil
	}

	// Check expiration
	if exp, ok := payload["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return map[string]interface{}{
				"success":      false,
				"valid":        false,
				"error":        "JWT has expired",
				"operation":    string(JWTValidate),
				"timestamp":    time.Now().Unix(),
			}, nil
		}
	}

	// Check not before
	if nbf, ok := payload["nbf"].(float64); ok {
		if time.Now().Unix() < int64(nbf) {
			return map[string]interface{}{
				"success":      false,
				"valid":        false,
				"error":        "JWT is not yet valid",
				"operation":    string(JWTValidate),
				"timestamp":    time.Now().Unix(),
			}, nil
		}
	}

	return map[string]interface{}{
		"success":      true,
		"valid":        true,
		"payload":      payload,
		"algorithm":    algStr,
		"operation":    string(JWTValidate),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// refreshJWT refreshes a JWT token
func (jhn *JWTHandlerNode) refreshJWT(inputs map[string]interface{}) (map[string]interface{}, error) {
	// Refresh is similar to create but with new expiration
	// In a real implementation, you'd typically need a refresh token
	// For this implementation, we'll just recreate the token with new expiry

	refreshToken := getStringValue(inputs["refresh_token"])
	if refreshToken != "" {
		// Validate refresh token
		refreshResult, err := jhn.validateJWT(map[string]interface{}{"jwt": refreshToken})
		if err != nil || !getBoolValue(refreshResult["valid"]) {
			return map[string]interface{}{
				"success":      false,
				"error":        "Invalid refresh token",
				"operation":    string(JWTRefresh),
				"timestamp":    time.Now().Unix(),
			}, nil
		}
	}

	// For this implementation, we'll just create a new token based on the old one
	// In a real system, you'd have a separate refresh token mechanism
	return jhn.createJWT(inputs)
}

// decodeJWT decodes a JWT without validating it
func (jhn *JWTHandlerNode) decodeJWT(inputs map[string]interface{}) (map[string]interface{}, error) {
	jwt := getStringValue(inputs["jwt"])
	if jwt == "" {
		return nil, fmt.Errorf("JWT is required")
	}

	// Split JWT into parts
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("invalid header JSON: %w", err)
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload JSON: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"header":       header,
		"payload":      payload,
		"operation":    string(JWTDecode),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// revokeJWT revokes a JWT token
func (jhn *JWTHandlerNode) revokeJWT(inputs map[string]interface{}) (map[string]interface{}, error) {
	jwt := getStringValue(inputs["jwt"])
	if jwt == "" {
		return nil, fmt.Errorf("JWT is required for revocation")
	}

	if jhn.config.EnableBlacklist {
		jhn.config.Blacklist[jwt] = true
	}

	return map[string]interface{}{
		"success":      true,
		"jwt_revoked":  true,
		"operation":    string(JWTRevoke),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// sign signs a JWT with the configured algorithm
func (jhn *JWTHandlerNode) sign(message, secretKey, privateKey string) (string, error) {
	switch jhn.config.Algorithm {
	case AlgorithmHS256, AlgorithmHS384, AlgorithmHS512:
		// HMAC signing
		h := sha256.New()
		h.Write([]byte(secretKey))
		key := h.Sum(nil)

		h2 := hmacSHA256([]byte(message), key)
		signature := base64.RawURLEncoding.EncodeToString(h2)
		return signature, nil

	case AlgorithmRS256, AlgorithmRS384, AlgorithmRS512:
		// RSA signing
		// Parse private key
		block, _ := pem.Decode([]byte(privateKey))
		if block == nil {
			return "", fmt.Errorf("failed to decode private key")
		}

		privateKeyParsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaPrivateKey, ok := privateKeyParsed.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("not an RSA private key")
		}

		h := sha256.New()
		h.Write([]byte(message))
		digest := h.Sum(nil)

		signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, digest)
		if err != nil {
			return "", fmt.Errorf("failed to sign message: %w", err)
		}

		signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)
		return signatureEncoded, nil

	default:
		return "", fmt.Errorf("unsupported algorithm: %s", jhn.config.Algorithm)
	}
}

// verifySignature verifies a JWT signature
func (jhn *JWTHandlerNode) verifySignature(message, signature, algorithm, secretKey, publicKey string) (bool, error) {
	switch JWTAlgorithm(algorithm) {
	case AlgorithmHS256, AlgorithmHS384, AlgorithmHS512:
		// HMAC verification
		h := sha256.New()
		h.Write([]byte(secretKey))
		key := h.Sum(nil)

		expectedSignature, err := base64.RawURLEncoding.DecodeString(signature)
		if err != nil {
			return false, fmt.Errorf("invalid signature encoding: %w", err)
		}

		actualSignature := hmacSHA256([]byte(message), key)
		return hmac.Equal(expectedSignature, actualSignature), nil

	case AlgorithmRS256, AlgorithmRS384, AlgorithmRS512:
		// RSA verification
		// Parse public key
		block, _ := pem.Decode([]byte(publicKey))
		if block == nil {
			return false, fmt.Errorf("failed to decode public key")
		}

		publicKeyParsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return false, fmt.Errorf("failed to parse public key: %w", err)
		}

		rsaPublicKey, ok := publicKeyParsed.(*rsa.PublicKey)
		if !ok {
			return false, fmt.Errorf("not an RSA public key")
		}

		signatureBytes, err := base64.RawURLEncoding.DecodeString(signature)
		if err != nil {
			return false, fmt.Errorf("invalid signature encoding: %w", err)
		}

		h := sha256.New()
		h.Write([]byte(message))
		digest := h.Sum(nil)

		err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, digest, signatureBytes)
		return err == nil, nil

	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// hmacSHA256 computes an HMAC-SHA256
func hmacSHA256(message, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

// JWTHandlerNodeFromConfig creates a new JWT handler node from a configuration map
func JWTHandlerNodeFromConfig(config map[string]interface{}) (engine.NodeInstance, error) {
	var operation JWTOperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = JWTOperationType(opStr)
		}
	}

	var algorithm JWTAlgorithm
	if algo, exists := config["algorithm"]; exists {
		if algoStr, ok := algo.(string); ok {
			algorithm = JWTAlgorithm(algoStr)
		}
	}

	var secretKey string
	if key, exists := config["secret_key"]; exists {
		if keyStr, ok := key.(string); ok {
			secretKey = keyStr
		}
	}

	var publicKey string
	if key, exists := config["public_key"]; exists {
		if keyStr, ok := key.(string); ok {
			publicKey = keyStr
		}
	}

	var privateKey string
	if key, exists := config["private_key"]; exists {
		if keyStr, ok := key.(string); ok {
			privateKey = keyStr
		}
	}

	var issuer string
	if iss, exists := config["issuer"]; exists {
		if issStr, ok := iss.(string); ok {
			issuer = issStr
		}
	}

	var audience []string
	if aud, exists := config["audience"]; exists {
		if audSlice, ok := aud.([]interface{}); ok {
			audience = make([]string, len(audSlice))
			for i, v := range audSlice {
				audience[i] = getStringValue(v)
			}
		}
	}

	var subject string
	if sub, exists := config["subject"]; exists {
		if subStr, ok := sub.(string); ok {
			subject = subStr
		}
	}

	var defaultExpiry float64
	if expiry, exists := config["default_expiry_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			defaultExpiry = expiryFloat
		}
	}

	var refreshExpiry float64
	if expiry, exists := config["refresh_expiry_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			refreshExpiry = expiryFloat
		}
	}

	var enableLogging bool
	if logging, exists := config["enable_logging"]; exists {
		enableLogging = getBoolValue(logging)
	}

	var enableBlacklist bool
	if blacklist, exists := config["enable_blacklist"]; exists {
		enableBlacklist = getBoolValue(blacklist)
	}

	nodeConfig := &JWTConfig{
		Operation:       operation,
		Algorithm:       algorithm,
		SecretKey:       secretKey,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		Issuer:          issuer,
		Audience:        audience,
		Subject:         subject,
		DefaultExpiry:   time.Duration(defaultExpiry) * time.Second,
		RefreshExpiry:   time.Duration(refreshExpiry) * time.Second,
		EnableLogging:   enableLogging,
		EnableBlacklist: enableBlacklist,
		Blacklist:       make(map[string]bool),
	}

	return NewJWTHandlerNode(nodeConfig), nil
}

// RegisterJWTHandlerNode registers the JWT handler node type with the engine
func RegisterJWTHandlerNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("jwt_handler", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return JWTHandlerNodeFromConfig(config)
	})
}