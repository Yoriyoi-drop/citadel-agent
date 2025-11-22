package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/plugins"
	"github.com/hashicorp/go-plugin"
	"golang.org/x/crypto/bcrypt"
)

// SecurityOperationType represents the type of security operation
type SecurityOperationType string

const (
	SecurityOpHash          SecurityOperationType = "hash"
	SecurityOpEncrypt       SecurityOperationType = "encrypt"
	SecurityOpDecrypt       SecurityOperationType = "decrypt"
	SecurityOpSign          SecurityOperationType = "sign"
	SecurityOpVerify        SecurityOperationType = "verify"
	SecurityOpValidate      SecurityOperationType = "validate"
	SecurityOpMask          SecurityOperationType = "mask"
	SecurityOpGenerateToken SecurityOperationType = "generate_token"
	SecurityOpValidateToken SecurityOperationType = "validate_token"
)

// SecurityAlgorithm represents the cryptographic algorithm to use
type SecurityAlgorithm string

const (
	AlgorithmSHA256   SecurityAlgorithm = "sha256"
	AlgorithmSHA512   SecurityAlgorithm = "sha512"
	AlgorithmAES256   SecurityAlgorithm = "aes256"
	AlgorithmHMACSHA256 SecurityAlgorithm = "hmac_sha256"
	AlgorithmBCrypt   SecurityAlgorithm = "bcrypt"
)

// SecurityPlugin implements the NodePlugin interface for security operations
type SecurityPlugin struct{}

// Execute implements the node execution logic
func (s *SecurityPlugin) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Extract operation parameters
	operation := getStringValue(inputs["operation"], string(SecurityOpHash))
	algorithm := getStringValue(inputs["algorithm"], string(AlgorithmSHA256))
	secretKey := getStringValue(inputs["secret_key"], "")
	data := getStringValue(inputs["data"], "")

	switch SecurityOperationType(operation) {
	case SecurityOpHash:
		return s.hashData(data, SecurityAlgorithm(algorithm), secretKey)
	case SecurityOpEncrypt:
		return s.encryptData(data, secretKey, SecurityAlgorithm(algorithm))
	case SecurityOpDecrypt:
		return s.decryptData(data, secretKey, SecurityAlgorithm(algorithm))
	case SecurityOpSign:
		return s.signData(data, secretKey, SecurityAlgorithm(algorithm))
	case SecurityOpVerify:
		signature := getStringValue(inputs["signature"], "")
		return s.verifyData(data, signature, secretKey, SecurityAlgorithm(algorithm))
	default:
		return nil, fmt.Errorf("unsupported security operation: %s", operation)
	}
}

// hashData performs hashing operation
func (s *SecurityPlugin) hashData(data string, algorithm SecurityAlgorithm, secretKey string) (map[string]interface{}, error) {
	var hashResult string
	var err error

	switch algorithm {
	case AlgorithmSHA256:
		h := sha256.New()
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmSHA512:
		h := sha512.New()
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmHMACSHA256:
		if secretKey == "" {
			return nil, fmt.Errorf("secret key is required for HMAC")
		}
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(data))
		hashResult = hex.EncodeToString(h.Sum(nil))
	case AlgorithmBCrypt:
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("bcrypt hash failed: %w", err)
		}
		hashResult = string(hashBytes)
	default:
		return nil, fmt.Errorf("unsupported hashing algorithm: %s", algorithm)
	}

	return map[string]interface{}{
		"success":    true,
		"result":     hashResult,
		"algorithm":  string(algorithm),
		"operation":  string(SecurityOpHash),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// encryptData performs encryption operation
func (s *SecurityPlugin) encryptData(data, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	switch algorithm {
	case AlgorithmAES256:
		return s.aes256Encrypt(data, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}
}

// aes256Encrypt performs AES-256 encryption
func (s *SecurityPlugin) aes256Encrypt(plaintext, key string) (map[string]interface{}, error) {
	// Create cipher
	block, err := aes.NewCipher([]byte(createKey(key, 32)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Pad plaintext
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	paddedText := []byte(plaintext)
	for i := 0; i < padding; i++ {
		paddedText = append(paddedText, byte(padding))
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Combine IV and ciphertext
	result := append(iv, ciphertext...)

	return map[string]interface{}{
		"success":        true,
		"encrypted_data": base64.StdEncoding.EncodeToString(result),
		"algorithm":      string(AlgorithmAES256),
		"operation":      string(SecurityOpEncrypt),
		"iv":             base64.StdEncoding.EncodeToString(iv),
		"input_size":     len(plaintext),
		"output_size":    len(result),
		"timestamp":      time.Now().Unix(),
	}, nil
}

// decryptData performs decryption operation
func (s *SecurityPlugin) decryptData(encryptedDataBase64, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	switch algorithm {
	case AlgorithmAES256:
		return s.aes256Decrypt(encryptedDataBase64, key)
	default:
		return nil, fmt.Errorf("unsupported decryption algorithm: %s", algorithm)
	}
}

// aes256Decrypt performs AES-256 decryption
func (s *SecurityPlugin) aes256Decrypt(encryptedDataBase64 string, key string) (map[string]interface{}, error) {
	// Decode the base64 data
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Extract IV (first 16 bytes) and ciphertext
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Create cipher
	block, err := aes.NewCipher([]byte(createKey(key, 32)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) {
		return nil, fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padding]

	return map[string]interface{}{
		"success":       true,
		"decrypted_data": string(plaintext),
		"algorithm":     string(AlgorithmAES256),
		"operation":     string(SecurityOpDecrypt),
		"input_size":    len(encryptedData),
		"output_size":   len(plaintext),
		"timestamp":     time.Now().Unix(),
	}, nil
}

// signData signs data with specified algorithm
func (s *SecurityPlugin) signData(data, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required for signing")
	}

	var signature string

	switch algorithm {
	case AlgorithmHMACSHA256:
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(data))
		signature = hex.EncodeToString(h.Sum(nil))
	default:
		return nil, fmt.Errorf("unsupported signing algorithm: %s", algorithm)
	}

	return map[string]interface{}{
		"success":   true,
		"signature": signature,
		"algorithm": string(algorithm),
		"operation": string(SecurityOpSign),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// verifyData verifies a signature
func (s *SecurityPlugin) verifyData(data, signature, key string, algorithm SecurityAlgorithm) (map[string]interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required for verification")
	}

	if signature == "" {
		return nil, fmt.Errorf("signature is required")
	}

	var expectedSig string
	var err error

	switch algorithm {
	case AlgorithmHMACSHA256:
		h := hmac.New(sha256.New, []byte(key))
		h.Write([]byte(data))
		expectedSig = hex.EncodeToString(h.Sum(nil))
	default:
		return nil, fmt.Errorf("unsupported verification algorithm: %s", algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to compute signature: %w", err)
	}

	verified := hmac.Equal([]byte(signature), []byte(expectedSig))

	return map[string]interface{}{
		"success":   true,
		"verified":  verified,
		"algorithm": string(algorithm),
		"operation": string(SecurityOpVerify),
		"input_size": len(data),
		"timestamp":  time.Now().Unix(),
	}, nil
}

// GetConfigSchema returns the JSON schema for configuration
func (s *SecurityPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type": "string",
				"title": "Security Operation",
				"enum": []string{"hash", "encrypt", "decrypt", "sign", "verify"},
				"default": "hash",
			},
			"algorithm": map[string]interface{}{
				"type": "string",
				"title": "Algorithm",
				"enum": []string{"sha256", "sha512", "aes256", "hmac_sha256", "bcrypt"},
				"default": "sha256",
			},
			"secret_key": map[string]interface{}{
				"type": "string",
				"title": "Secret Key",
				"description": "Key for encryption/signing",
			},
			"data": map[string]interface{}{
				"type": "string",
				"title": "Input Data",
				"description": "Data to process",
			},
		},
		"required": []string{"operation", "data"},
	}
}

// GetMetadata returns metadata about the plugin
func (s *SecurityPlugin) GetMetadata() plugins.NodeMetadata {
	return plugins.NodeMetadata{
		ID:          "security_operation",
		Name:        "Security Operation",
		Description: "Performs various security operations like hashing, encryption, signing, etc.",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Category:    "security",
	}
}

// Helper functions
func getStringValue(v interface{}, defaultValue string) string {
	if v == nil {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return defaultValue
}

// createKey creates a key of specified length from input
func createKey(key string, length int) []byte {
	result := make([]byte, length)
	copy(result, []byte(key))

	// If key is shorter than required length, extend it
	if len(key) < length {
		for i := len(key); i < length; i++ {
			result[i] = 0 // Fill with zeros
		}
	}

	return result
}

// Handshake is the magic handshake configuration
var handshakeConfig = plugins.Handshake

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"node": &plugins.NodePluginImpl{Impl: &SecurityPlugin{}},
		},
	})
}