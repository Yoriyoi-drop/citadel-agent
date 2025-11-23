package http

import (
	"encoding/base64"
)

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeAPIKey AuthType = "apikey"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     AuthType
	Username string
	Password string
	Token    string
	APIKey   string
	KeyName  string // Header name for API Key
	InHeader bool   // Whether API Key is in header or query param
}

// ApplyAuth applies authentication headers to the request headers
func ApplyAuth(headers map[string]string, config AuthConfig) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}

	switch config.Type {
	case AuthTypeBasic:
		auth := config.Username + ":" + config.Password
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		headers["Authorization"] = "Basic " + encoded
	case AuthTypeBearer:
		headers["Authorization"] = "Bearer " + config.Token
	case AuthTypeAPIKey:
		if config.InHeader {
			keyName := config.KeyName
			if keyName == "" {
				keyName = "X-API-Key"
			}
			headers[keyName] = config.APIKey
		}
		// Query param handling would be done in the URL construction
	}

	return headers
}
