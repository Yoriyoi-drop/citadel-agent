// backend/internal/nodes/security/api_key_manager.go
package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// APIKeyType represents the type of API key
type APIKeyType string

const (
	APIKeyBearer APIKeyType = "bearer"
	APIKeyBasic  APIKeyType = "basic"
	APIKeyCustom APIKeyType = "custom"
)

// APIKeyScope represents the scope of an API key
type APIKeyScope struct {
	Resource string `json:"resource"`
	Actions  []string `json:"actions"`
}

// APIKey represents an API key
type APIKey struct {
	ID          string       `json:"id"`
	Key         string       `json:"key"`
	HashedKey   string       `json:"hashed_key"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Scopes      []APIKeyScope `json:"scopes"`
	CreatedAt   time.Time    `json:"created_at"`
	ExpiresAt   *time.Time   `json:"expires_at"`
	Revoked     bool         `json:"revoked"`
	RevokedAt   *time.Time   `json:"revoked_at"`
}

// APIKeyManagerOperation represents the operation to perform
type APIKeyManagerOperation string

const (
	APIKeyOpCreate  APIKeyManagerOperation = "create"
	APIKeyOpValidate APIKeyManagerOperation = "validate"
	APIKeyOpRevoke  APIKeyManagerOperation = "revoke"
	APIKeyOpRotate  APIKeyManagerOperation = "rotate"
	APIKeyOpList    APIKeyManagerOperation = "list"
)

// APIKeyManagerConfig represents the configuration for an API key manager node
type APIKeyManagerConfig struct {
	DefaultExpiryDuration time.Duration `json:"default_expiry_duration"`
	MaxValidKeys          int           `json:"max_valid_keys"`
	AllowedScopes         []APIKeyScope  `json:"allowed_scopes"`
	StorageBackend        string        `json:"storage_backend"` // "memory", "database", etc.
	Operation             APIKeyManagerOperation `json:"operation"`
}

// APIKeyManagerNode represents an API key manager node
type APIKeyManagerNode struct {
	config *APIKeyManagerConfig
	keys   map[string]*APIKey // In-memory storage for demo purposes
}

// NewAPIKeyManagerNode creates a new API key manager node
func NewAPIKeyManagerNode(config *APIKeyManagerConfig) *APIKeyManagerNode {
	if config.DefaultExpiryDuration == 0 {
		config.DefaultExpiryDuration = 30 * 24 * time.Hour // 30 days
	}

	if config.MaxValidKeys == 0 {
		config.MaxValidKeys = 1000
	}

	return &APIKeyManagerNode{
		config: config,
		keys:   make(map[string]*APIKey),
	}
}

// Execute executes the API key manager operation
func (akmn *APIKeyManagerNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := akmn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = APIKeyManagerOperation(opStr)
		}
	}

	switch operation {
	case APIKeyOpCreate:
		return akmn.createAPIKey(inputs)
	case APIKeyOpValidate:
		return akmn.validateAPIKey(inputs)
	case APIKeyOpRevoke:
		return akmn.revokeAPIKey(inputs)
	case APIKeyOpRotate:
		return akmn.rotateAPIKey(inputs)
	case APIKeyOpList:
		return akmn.listAPIKeys(inputs)
	default:
		return nil, fmt.Errorf("unsupported API key operation: %s", operation)
	}
}

// createAPIKey creates a new API key
func (akmn *APIKeyManagerNode) createAPIKey(inputs map[string]interface{}) (map[string]interface{}, error) {
	// Get parameters
	name := getStringValue(inputs["name"])
	description := getStringValue(inputs["description"])
	
	// Get scopes from input or config
	var scopes []APIKeyScope
	if scopesInput, exists := inputs["scopes"]; exists {
		if scopesSlice, ok := scopesInput.([]interface{}); ok {
			scopes = make([]APIKeyScope, len(scopesSlice))
			for i, scopeInterface := range scopesSlice {
				if scopeMap, ok := scopeInterface.(map[string]interface{}); ok {
					var actions []string
					if actionsInput, exists := scopeMap["actions"]; exists {
						if actionsSlice, ok := actionsInput.([]interface{}); ok {
							actions = make([]string, len(actionsSlice))
							for j, action := range actionsSlice {
								actions[j] = getStringValue(action)
							}
						}
					}
					scopes[i] = APIKeyScope{
						Resource: getStringValue(scopeMap["resource"]),
						Actions:  actions,
					}
				}
			}
		}
	}
	
	// Get expiry from input or use default
	var expiresAt *time.Time
	if expiryInput, exists := inputs["expiry_seconds"]; exists {
		if expiryFloat, ok := expiryInput.(float64); ok {
			exp := time.Now().Add(time.Duration(expiryFloat) * time.Second)
			expireAt := exp
			expiresAt = &expireAt
		}
	} else {
		exp := time.Now().Add(akmn.config.DefaultExpiryDuration)
		expiresAt = &exp
	}

	// Generate API key
	apiKey, hashedKey, err := akmn.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create API key object
	id := generateID()
	newKey := &APIKey{
		ID:          id,
		Key:         apiKey,
		HashedKey:   hashedKey,
		Name:        name,
		Description: description,
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Revoked:     false,
	}

	// Store in memory (in real implementation, store in database)
	akmn.keys[id] = newKey

	return map[string]interface{}{
		"success":      true,
		"api_key":      newKey.Key, // This is the only time the API key is returned
		"key_id":       newKey.ID,
		"name":         newKey.Name,
		"description":  newKey.Description,
		"scopes":       newKey.Scopes,
		"created_at":   newKey.CreatedAt.Unix(),
		"expires_at":   newKey.ExpiresAt.Unix(),
		"operation":    "create",
		"timestamp":    time.Now().Unix(),
	}, nil
}

// validateAPIKey validates an API key
func (akmn *APIKeyManagerNode) validateAPIKey(inputs map[string]interface{}) (map[string]interface{}, error) {
	apiKey := getStringValue(inputs["api_key"])
	resource := getStringValue(inputs["resource"])
	action := getStringValue(inputs["action"])

	if apiKey == "" {
		return map[string]interface{}{
			"success":   false,
			"valid":     false,
			"reason":    "API key is required",
			"operation": "validate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Find the API key by hashing and comparing
	hashedKey := hashAPIKey(apiKey)
	var foundKey *APIKey

	for _, key := range akmn.keys {
		if key.HashedKey == hashedKey {
			foundKey = key
			break
		}
	}

	if foundKey == nil {
		return map[string]interface{}{
			"success":   false,
			"valid":     false,
			"reason":    "Invalid API key",
			"operation": "validate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Check if revoked
	if foundKey.Revoked {
		return map[string]interface{}{
			"success":   false,
			"valid":     false,
			"reason":    "API key has been revoked",
			"key_id":    foundKey.ID,
			"operation": "validate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Check if expired
	if foundKey.ExpiresAt != nil && time.Now().After(*foundKey.ExpiresAt) {
		return map[string]interface{}{
			"success":   false,
			"valid":     false,
			"reason":    "API key has expired",
			"key_id":    foundKey.ID,
			"expires_at": foundKey.ExpiresAt.Unix(),
			"operation": "validate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Check scopes if resource and action specified
	if resource != "" && action != "" {
		hasAccess := false
		for _, scope := range foundKey.Scopes {
			if scope.Resource == resource {
				// Check if action is allowed
				if len(scope.Actions) == 0 {
					hasAccess = true
					break
				}
				for _, allowedAction := range scope.Actions {
					if allowedAction == action {
						hasAccess = true
						break
					}
				}
			}
			if hasAccess {
				break
			}
		}

		if !hasAccess {
			return map[string]interface{}{
				"success":   false,
				"valid":     false,
				"reason":    "API key does not have permission for this resource/action",
				"key_id":    foundKey.ID,
				"resource":  resource,
				"action":    action,
				"operation": "validate",
				"timestamp": time.Now().Unix(),
			}, nil
		}
	}

	return map[string]interface{}{
		"success":   true,
		"valid":     true,
		"key_id":    foundKey.ID,
		"name":      foundKey.Name,
		"scopes":    foundKey.Scopes,
		"created_at": foundKey.CreatedAt.Unix(),
		"expires_at": foundKey.ExpiresAt.Unix(),
		"operation": "validate",
		"timestamp": time.Now().Unix(),
	}, nil
}

// revokeAPIKey revokes an API key
func (akmn *APIKeyManagerNode) revokeAPIKey(inputs map[string]interface{}) (map[string]interface{}, error) {
	keyID := getStringValue(inputs["key_id"])

	if keyID == "" {
		return map[string]interface{}{
			"success":   false,
			"reason":    "Key ID is required",
			"operation": "revoke",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	key, exists := akmn.keys[keyID]
	if !exists {
		return map[string]interface{}{
			"success":   false,
			"reason":    "API key not found",
			"key_id":    keyID,
			"operation": "revoke",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Revoke the key
	now := time.Now()
	key.Revoked = true
	key.RevokedAt = &now

	return map[string]interface{}{
		"success":   true,
		"key_id":    keyID,
		"name":      key.Name,
		"revoked_at": now.Unix(),
		"operation": "revoke",
		"timestamp": time.Now().Unix(),
	}, nil
}

// rotateAPIKey generates a new API key replacing an existing one
func (akmn *APIKeyManagerNode) rotateAPIKey(inputs map[string]interface{}) (map[string]interface{}, error) {
	keyID := getStringValue(inputs["key_id"])

	if keyID == "" {
		return map[string]interface{}{
			"success":   false,
			"reason":    "Key ID is required",
			"operation": "rotate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	oldKey, exists := akmn.keys[keyID]
	if !exists {
		return map[string]interface{}{
			"success":   false,
			"reason":    "API key not found",
			"key_id":    keyID,
			"operation": "rotate",
			"timestamp": time.Now().Unix(),
		}, nil
	}

	// Generate new API key
	newAPIKey, newHashedKey, err := akmn.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new API key: %w", err)
	}

	// Update the key with new values
	newID := generateID()
	newKey := &APIKey{
		ID:          newID,
		Key:         newAPIKey,
		HashedKey:   newHashedKey,
		Name:        oldKey.Name,
		Description: oldKey.Description,
		Scopes:      oldKey.Scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   oldKey.ExpiresAt,
		Revoked:     false,
	}

	// Add new key and revoke old key
	delete(akmn.keys, keyID) // Remove old key
	akmn.keys[newID] = newKey

	// Revoke the old key
	now := time.Now()
	oldKey.Revoked = true
	oldKey.RevokedAt = &now

	return map[string]interface{}{
		"success":      true,
		"new_api_key":  newKey.Key, // Return the new key
		"old_key_id":   keyID,
		"new_key_id":   newID,
		"name":         newKey.Name,
		"description":  newKey.Description,
		"scopes":       newKey.Scopes,
		"created_at":   newKey.CreatedAt.Unix(),
		"expires_at":   newKey.ExpiresAt.Unix(),
		"operation":    "rotate",
		"timestamp":    time.Now().Unix(),
	}, nil
}

// listAPIKeys lists all available API keys (without revealing the actual keys)
func (akmn *APIKeyManagerNode) listAPIKeys(inputs map[string]interface{}) (map[string]interface{}, error) {
	var keysInfo []map[string]interface{}

	for id, key := range akmn.keys {
		keyInfo := map[string]interface{}{
			"id":           id,
			"name":         key.Name,
			"description":  key.Description,
			"scopes":       key.Scopes,
			"created_at":   key.CreatedAt.Unix(),
			"revoked":      key.Revoked,
			"revoked_at":   getUnixTimeOrNil(key.RevokedAt),
			"expires_at":   getUnixTimeOrNil(key.ExpiresAt),
		}
		keysInfo = append(keysInfo, keyInfo)
	}

	return map[string]interface{}{
		"success":   true,
		"keys":      keysInfo,
		"total":     len(keysInfo),
		"operation": "list",
		"timestamp": time.Now().Unix(),
	}, nil
}

// generateAPIKey generates a new random API key and its hash
func (akmn *APIKeyManagerNode) generateAPIKey() (string, string, error) {
	// Generate random bytes
	bytes := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Convert to hex string (in real implementation, you might want to use base64)
	apiKey := hex.EncodeToString(bytes)

	// Hash the key for storage
	hashedKey := hashAPIKey(apiKey)

	return apiKey, hashedKey, nil
}

// hashAPIKey hashes an API key using SHA-256
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// generateID generates a unique ID for API keys
func generateID() string {
	bytes := make([]byte, 16) // 128-bit ID
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getUnixTimeOrNil converts time pointer to unix timestamp or nil
func getUnixTimeOrNil(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Unix()
}

// APIKeyManagerNodeFromConfig creates a new API key manager node from a configuration map
func APIKeyManagerNodeFromConfig(config map[string]interface{}) (interfaces.NodeInstance, error) {
	var defaultExpiryDuration float64
	if expiry, exists := config["default_expiry_duration_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			defaultExpiryDuration = expiryFloat
		}
	}

	var maxValidKeys float64
	if max, exists := config["max_valid_keys"]; exists {
		if maxFloat, ok := max.(float64); ok {
			maxValidKeys = maxFloat
		}
	}

	var allowedScopes []APIKeyScope
	if scopesInput, exists := config["allowed_scopes"]; exists {
		if scopesSlice, ok := scopesInput.([]interface{}); ok {
			allowedScopes = make([]APIKeyScope, len(scopesSlice))
			for i, scopeInterface := range scopesSlice {
				if scopeMap, ok := scopeInterface.(map[string]interface{}); ok {
					var actions []string
					if actionsInput, exists := scopeMap["actions"]; exists {
						if actionsSlice, ok := actionsInput.([]interface{}); ok {
							actions = make([]string, len(actionsSlice))
							for j, action := range actionsSlice {
								actions[j] = getStringValue(action)
							}
						}
					}
					allowedScopes[i] = APIKeyScope{
						Resource: getStringValue(scopeMap["resource"]),
						Actions:  actions,
					}
				}
			}
		}
	}

	var storageBackend string
	if backend, exists := config["storage_backend"]; exists {
		if backendStr, ok := backend.(string); ok {
			storageBackend = backendStr
		}
	}

	var operation APIKeyManagerOperation
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = APIKeyManagerOperation(opStr)
		}
	}

	nodeConfig := &APIKeyManagerConfig{
		DefaultExpiryDuration: time.Duration(defaultExpiryDuration) * time.Second,
		MaxValidKeys:          int(maxValidKeys),
		AllowedScopes:         allowedScopes,
		StorageBackend:        storageBackend,
		Operation:             operation,
	}

	return NewAPIKeyManagerNode(nodeConfig), nil
}

// RegisterAPIKeyManagerNode registers the API key manager node type with the engine
func RegisterAPIKeyManagerNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("api_key_manager", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return APIKeyManagerNodeFromConfig(config)
	})
}