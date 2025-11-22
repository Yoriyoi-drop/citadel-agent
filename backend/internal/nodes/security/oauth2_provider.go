// backend/internal/nodes/security/oauth2_provider.go
package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// OAuth2OperationType represents the type of OAuth2 operation
type OAuth2OperationType string

const (
	OAuth2Authorize    OAuth2OperationType = "authorize"
	OAuth2Token        OAuth2OperationType = "token"
	OAuth2Refresh      OAuth2OperationType = "refresh"
	OAuth2UserInfo     OAuth2OperationType = "user_info"
	OAuth2Revoke       OAuth2OperationType = "revoke"
	OAuth2Introspect   OAuth2OperationType = "introspect"
)

// OAuth2GrantType represents the type of OAuth2 grant
type OAuth2GrantType string

const (
	GrantTypeAuthorizationCode OAuth2GrantType = "authorization_code"
	GrantTypeImplicit          OAuth2GrantType = "implicit"
	GrantTypeResourceOwner     OAuth2GrantType = "password"
	GrantTypeClientCredentials OAuth2GrantType = "client_credentials"
	GrantTypeRefreshToken      OAuth2GrantType = "refresh_token"
)

// OAuth2ResponseType represents the OAuth2 response type
type OAuth2ResponseType string

const (
	ResponseTypeCode     OAuth2ResponseType = "code"
	ResponseTypeToken    OAuth2ResponseType = "token"
	ResponseTypeIdToken  OAuth2ResponseType = "id_token"
	ResponseTypeCodeToken OAuth2ResponseType = "code token"
	ResponseTypeCodeIdToken OAuth2ResponseType = "code id_token"
	ResponseTypeNone     OAuth2ResponseType = "none"
)

// OAuth2Scope represents an OAuth2 scope
type OAuth2Scope string

const (
	ScopeOpenID   OAuth2Scope = "openid"
	ScopeProfile  OAuth2Scope = "profile"
	ScopeEmail    OAuth2Scope = "email"
	ScopeAddress  OAuth2Scope = "address"
	ScopePhone    OAuth2Scope = "phone"
)

// OAuth2Client represents an OAuth2 client
type OAuth2Client struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ClientSecret    string   `json:"client_secret"`
	RedirectURIs    []string `json:"redirect_uris"`
	GrantTypes      []string `json:"grant_types"`
	ResponseTypes   []string `json:"response_types"`
	Scope           []string `json:"scope"`
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"` // "client_secret_basic", "client_secret_post", "none"
	Enabled         bool     `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OAuth2AuthCode represents an OAuth2 authorization code
type OAuth2AuthCode struct {
	Code        string    `json:"code"`
	ClientID    string    `json:"client_id"`
	RedirectURI string    `json:"redirect_uri"`
	Scopes      []string  `json:"scopes"`
	UserID      string    `json:"user_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// OAuth2AccessToken represents an OAuth2 access token
type OAuth2AccessToken struct {
	Token       string    `json:"token"`
	ClientID    string    `json:"client_id"`
	UserID      string    `json:"user_id"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	TokenType   string    `json:"token_type"`
}

// OAuth2RefreshToken represents an OAuth2 refresh token
type OAuth2RefreshToken struct {
	Token       string    `json:"token"`
	ClientID    string    `json:"client_id"`
	UserID      string    `json:"user_id"`
	Scopes      []string  `json:"scopes"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// OAuth2Config represents the configuration for an OAuth2 provider node
type OAuth2Config struct {
	Operation       OAuth2OperationType `json:"operation"`
	Issuer          string              `json:"issuer"`
	AuthCodeExpiry  time.Duration       `json:"auth_code_expiry"`
	AccessTokenExpiry time.Duration     `json:"access_token_expiry"`
	RefreshTokenExpiry time.Duration     `json:"refresh_token_expiry"`
	EnablePKCE      bool                `json:"enable_pkce"`
	EnableLogging   bool                `json:"enable_logging"`
	Clients         []OAuth2Client      `json:"clients"`
	SigningKey      string              `json:"signing_key"`
	EncryptionKey   string              `json:"encryption_key"`
}

// OAuth2ProviderNode represents an OAuth2 provider node
type OAuth2ProviderNode struct {
	config *OAuth2Config
	authCodes map[string]*OAuth2AuthCode
	accessTokens map[string]*OAuth2AccessToken
	refreshTokens map[string]*OAuth2RefreshToken
}

// NewOAuth2ProviderNode creates a new OAuth2 provider node
func NewOAuth2ProviderNode(config *OAuth2Config) *OAuth2ProviderNode {
	if config.AuthCodeExpiry == 0 {
		config.AuthCodeExpiry = 10 * time.Minute // 10 minutes default
	}

	if config.AccessTokenExpiry == 0 {
		config.AccessTokenExpiry = 1 * time.Hour // 1 hour default
	}

	if config.RefreshTokenExpiry == 0 {
		config.RefreshTokenExpiry = 30 * 24 * time.Hour // 30 days default
	}

	return &OAuth2ProviderNode{
		config: config,
		authCodes: make(map[string]*OAuth2AuthCode),
		accessTokens: make(map[string]*OAuth2AccessToken),
		refreshTokens: make(map[string]*OAuth2RefreshToken),
	}
}

// Execute executes the OAuth2 operation
func (opn *OAuth2ProviderNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	operation := opn.config.Operation
	if op, exists := inputs["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = OAuth2OperationType(opStr)
		}
	}

	// Perform the OAuth2 operation based on type
	switch operation {
	case OAuth2Authorize:
		return opn.authorize(inputs)
	case OAuth2Token:
		return opn.token(inputs)
	case OAuth2Refresh:
		return opn.refresh(inputs)
	case OAuth2UserInfo:
		return opn.userInfo(inputs)
	case OAuth2Revoke:
		return opn.revoke(inputs)
	case OAuth2Introspect:
		return opn.introspect(inputs)
	default:
		return nil, fmt.Errorf("unsupported OAuth2 operation: %s", operation)
	}
}

// authorize handles the authorization flow
func (opn *OAuth2ProviderNode) authorize(inputs map[string]interface{}) (map[string]interface{}, error) {
	clientID := getStringValue(inputs["client_id"])
	redirectURI := getStringValue(inputs["redirect_uri"])
	responseType := getStringValue(inputs["response_type"])
	scope := getStringValue(inputs["scope"])
	state := getStringValue(inputs["state"])
	codeChallenge := getStringValue(inputs["code_challenge"])
	codeChallengeMethod := getStringValue(inputs["code_challenge_method"])

	// Validate client ID
	client := opn.findClient(clientID)
	if client == nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	// Validate redirect URI
	isValidURI := false
	for _, validURI := range client.RedirectURIs {
		if validURI == redirectURI {
			isValidURI = true
			break
		}
	}

	if !isValidURI {
		return nil, fmt.Errorf("invalid redirect URI: %s", redirectURI)
	}

	// Validate response type
	rt := OAuth2ResponseType(responseType)
	if rt != ResponseTypeCode && rt != ResponseTypeToken && rt != ResponseTypeIdToken {
		return nil, fmt.Errorf("unsupported response type: %s", responseType)
	}

	// Validate PKCE if required
	if opn.config.EnablePKCE && codeChallenge == "" {
		return nil, fmt.Errorf("PKCE code challenge required")
	}

	// Generate authorization code
	authCode := opn.generateAuthCode()

	scopes := strings.Split(scope, " ")
	if scopes[0] == "" {
		scopes = []string{}
	}

	authCodeObj := &OAuth2AuthCode{
		Code:        authCode,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		UserID:      getStringValue(inputs["user_id"]), // In real system, this would come from user authentication
		ExpiresAt:   time.Now().Add(opn.config.AuthCodeExpiry),
		CreatedAt:   time.Now(),
	}

	opn.authCodes[authCode] = authCodeObj

	// For authorization code flow, return the code
	result := map[string]interface{}{
		"success": true,
		"code":    authCode,
		"state":   state,
		"operation": string(OAuth2Authorize),
		"timestamp": time.Now().Unix(),
	}

	// If response type includes token (implicit flow), also return access token
	if rt == ResponseTypeToken || rt == ResponseTypeCodeToken || rt == ResponseTypeCodeIdToken {
		accessToken, err := opn.createAccessToken(clientID, getStringValue(inputs["user_id"]), scopes)
		if err != nil {
			return nil, fmt.Errorf("failed to create access token: %w", err)
		}
		
		result["access_token"] = accessToken.Token
		result["token_type"] = "Bearer"
		result["expires_in"] = int(opn.config.AccessTokenExpiry.Seconds())
	}

	return result, nil
}

// token handles the token exchange flow
func (opn *OAuth2ProviderNode) token(inputs map[string]interface{}) (map[string]interface{}, error) {
	grantType := getStringValue(inputs["grant_type"])
	clientID := getStringValue(inputs["client_id"])
	clientSecret := getStringValue(inputs["client_secret"])
	code := getStringValue(inputs["code"])
	redirectURI := getStringValue(inputs["redirect_uri"])
	refreshToken := getStringValue(inputs["refresh_token"])
	username := getStringValue(inputs["username"])
	password := getStringValue(inputs["password"])
	codeVerifier := getStringValue(inputs["code_verifier"])

	// Validate client credentials
	client := opn.findClient(clientID)
	if client == nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	// Check client secret for confidential clients
	if client.TokenEndpointAuthMethod != "none" {
		if clientSecret == "" || clientSecret != client.ClientSecret {
			return nil, fmt.Errorf("invalid client secret")
		}
	}

	switch OAuth2GrantType(grantType) {
	case GrantTypeAuthorizationCode:
		return opn.handleAuthCodeGrant(code, clientID, redirectURI, codeVerifier)
	case GrantTypeResourceOwner:
		return opn.handleResourceOwnerGrant(username, password, clientID)
	case GrantTypeClientCredentials:
		return opn.handleClientCredentialsGrant(clientID)
	case GrantTypeRefreshToken:
		return opn.handleRefreshTokenGrant(refreshToken, clientID)
	default:
		return nil, fmt.Errorf("unsupported grant type: %s", grantType)
	}
}

// handleAuthCodeGrant handles authorization code grant
func (opn *OAuth2ProviderNode) handleAuthCodeGrant(code, clientID, redirectURI, codeVerifier string) (map[string]interface{}, error) {
	authCode, exists := opn.authCodes[code]
	if !exists {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Check if code has expired
	if time.Now().After(authCode.ExpiresAt) {
		delete(opn.authCodes, code)
		return nil, fmt.Errorf("authorization code has expired")
	}

	// Check client ID
	if authCode.ClientID != clientID {
		return nil, fmt.Errorf("invalid client ID for authorization code")
	}

	// Check redirect URI
	if authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect URI mismatch")
	}

	// If PKCE was used, verify code verifier
	// (In a real implementation, we would have stored and compared the code challenge)

	// Delete the auth code as it's now consumed
	delete(opn.authCodes, code)

	// Create access token
	accessToken, err := opn.createAccessToken(clientID, authCode.UserID, authCode.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create refresh token
	refreshToken, err := opn.createRefreshToken(clientID, authCode.UserID, authCode.Scopes, accessToken.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"access_token":  accessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(opn.config.AccessTokenExpiry.Seconds()),
		"refresh_token": refreshToken.Token,
		"scope":         strings.Join(authCode.Scopes, " "),
		"operation":     string(OAuth2Token),
		"timestamp":     time.Now().Unix(),
	}, nil
}

// handleResourceOwnerGrant handles resource owner password credentials grant
func (opn *OAuth2ProviderNode) handleResourceOwnerGrant(username, password, clientID string) (map[string]interface{}, error) {
	// In a real implementation, this would validate user credentials
	// For simulation, we'll just proceed and create tokens
	
	// Create access token
	accessToken, err := opn.createAccessToken(clientID, username, []string{}) // Empty scopes for now
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create refresh token
	refreshToken, err := opn.createRefreshToken(clientID, username, []string{}, accessToken.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"access_token":  accessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(opn.config.AccessTokenExpiry.Seconds()),
		"refresh_token": refreshToken.Token,
		"operation":     string(OAuth2Token),
		"timestamp":     time.Now().Unix(),
	}, nil
}

// handleClientCredentialsGrant handles client credentials grant
func (opn *OAuth2ProviderNode) handleClientCredentialsGrant(clientID string) (map[string]interface{}, error) {
	// Client credentials grant doesn't include a user ID, typically used for service-to-service
	accessToken, err := opn.createAccessToken(clientID, "service", []string{}) // Empty scopes for now
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"access_token": accessToken.Token,
		"token_type":   "Bearer",
		"expires_in":   int(opn.config.AccessTokenExpiry.Seconds()),
		"operation":    string(OAuth2Token),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// handleRefreshTokenGrant handles refresh token grant
func (opn *OAuth2ProviderNode) handleRefreshTokenGrant(refreshTokenStr, clientID string) (map[string]interface{}, error) {
	// Find refresh token
	refreshToken, exists := opn.refreshTokens[refreshTokenStr]
	if !exists {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if refresh token has expired
	if time.Now().After(refreshToken.ExpiresAt) {
		delete(opn.refreshTokens, refreshTokenStr)
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Check client ID
	if refreshToken.ClientID != clientID {
		return nil, fmt.Errorf("invalid client ID for refresh token")
	}

	// Create new access token
	newAccessToken, err := opn.createAccessToken(clientID, refreshToken.UserID, refreshToken.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Optionally create new refresh token (rotate refresh tokens)
	delete(opn.refreshTokens, refreshTokenStr)
	newRefreshToken, err := opn.createRefreshToken(clientID, refreshToken.UserID, refreshToken.Scopes, newAccessToken.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"access_token":  newAccessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(opn.config.AccessTokenExpiry.Seconds()),
		"refresh_token": newRefreshToken.Token,
		"scope":         strings.Join(refreshToken.Scopes, " "),
		"operation":     string(OAuth2Token),
		"timestamp":     time.Now().Unix(),
	}, nil
}

// refresh handles token refresh
func (opn *OAuth2ProviderNode) refresh(inputs map[string]interface{}) (map[string]interface{}, error) {
	refreshToken := getStringValue(inputs["refresh_token"])
	clientID := getStringValue(inputs["client_id"])
	clientSecret := getStringValue(inputs["client_secret"])

	// This is essentially the same as the refresh token grant
	return opn.handleRefreshTokenGrant(refreshToken, clientID)
}

// userInfo returns user information for a valid access token
func (opn *OAuth2ProviderNode) userInfo(inputs map[string]interface{}) (map[string]interface{}, error) {
	accessTokenStr := getStringValue(inputs["access_token"])
	
	accessToken, exists := opn.accessTokens[accessTokenStr]
	if !exists {
		return nil, fmt.Errorf("invalid access token")
	}

	// Check if access token has expired
	if time.Now().After(accessToken.ExpiresAt) {
		delete(opn.accessTokens, accessTokenStr)
		return nil, fmt.Errorf("access token has expired")
	}

	// Return user information (in a real system, this would come from user database)
	userInfo := map[string]interface{}{
		"sub":          accessToken.UserID,
		"client_id":    accessToken.ClientID,
		"scopes":       accessToken.Scopes,
		"exp":          accessToken.ExpiresAt.Unix(),
		"iat":          accessToken.CreatedAt.Unix(),
	}

	return map[string]interface{}{
		"success":     true,
		"user_info":   userInfo,
		"operation":   string(OAuth2UserInfo),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// revoke revokes a token
func (opn *OAuth2ProviderNode) revoke(inputs map[string]interface{}) (map[string]interface{}, error) {
	token := getStringValue(inputs["token"])
	tokenTypeHint := getStringValue(inputs["token_type_hint"]) // "access_token", "refresh_token"

	revoked := false
	
	// Try to delete as access token
	if _, exists := opn.accessTokens[token]; exists {
		delete(opn.accessTokens, token)
		revoked = true
	}
	
	// Or try to delete as refresh token
	if _, exists := opn.refreshTokens[token]; exists {
		delete(opn.refreshTokens, token)
		revoked = true
	}

	// Or try to delete as auth code
	if _, exists := opn.authCodes[token]; exists {
		delete(opn.authCodes, token)
		revoked = true
	}

	return map[string]interface{}{
		"success":     true,
		"revoked":     revoked,
		"operation":   string(OAuth2Revoke),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// introspect introspects a token to get its information
func (opn *OAuth2ProviderNode) introspect(inputs map[string]interface{}) (map[string]interface{}, error) {
	token := getStringValue(inputs["token"])
	
	// Check if it's an access token
	if accessToken, exists := opn.accessTokens[token]; exists {
		active := !time.Now().After(accessToken.ExpiresAt)
		
		return map[string]interface{}{
			"success":   true,
			"active":    active,
			"token_type": accessToken.TokenType,
			"client_id": accessToken.ClientID,
			"username":  accessToken.UserID,
			"scope":     strings.Join(accessToken.Scopes, " "),
			"exp":       accessToken.ExpiresAt.Unix(),
			"iat":       accessToken.CreatedAt.Unix(),
			"operation": string(OAuth2Introspect),
			"timestamp": time.Now().Unix(),
		}, nil
	}
	
	// Check if it's a refresh token
	if refreshToken, exists := opn.refreshTokens[token]; exists {
		active := !time.Now().After(refreshToken.ExpiresAt)
		
		return map[string]interface{}{
			"success":   true,
			"active":    active,
			"token_type": "refresh_token",
			"client_id": refreshToken.ClientID,
			"username":  refreshToken.UserID,
			"scope":     strings.Join(refreshToken.Scopes, " "),
			"exp":       refreshToken.ExpiresAt.Unix(),
			"iat":       refreshToken.CreatedAt.Unix(),
			"operation": string(OAuth2Introspect),
			"timestamp": time.Now().Unix(),
		}, nil
	}
	
	// Token not found
	return map[string]interface{}{
		"success":   true,
		"active":    false,
		"operation": string(OAuth2Introspect),
		"timestamp": time.Now().Unix(),
	}, nil
}

// createAccessToken creates a new access token
func (opn *OAuth2ProviderNode) createAccessToken(clientID, userID string, scopes []string) (*OAuth2AccessToken, error) {
	token := opn.generateToken()
	
	accessToken := &OAuth2AccessToken{
		Token:       token,
		ClientID:    clientID,
		UserID:      userID,
		Scopes:      scopes,
		ExpiresAt:   time.Now().Add(opn.config.AccessTokenExpiry),
		CreatedAt:   time.Now(),
		TokenType:   "Bearer",
	}
	
	opn.accessTokens[token] = accessToken
	
	return accessToken, nil
}

// createRefreshToken creates a new refresh token
func (opn *OAuth2ProviderNode) createRefreshToken(clientID, userID string, scopes []string, accessToken string) (*OAuth2RefreshToken, error) {
	token := opn.generateToken()
	
	refreshToken := &OAuth2RefreshToken{
		Token:       token,
		ClientID:    clientID,
		UserID:      userID,
		Scopes:      scopes,
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(opn.config.RefreshTokenExpiry),
		CreatedAt:   time.Now(),
	}
	
	opn.refreshTokens[token] = refreshToken
	
	return refreshToken, nil
}

// generateAuthCode generates a new authorization code
func (opn *OAuth2ProviderNode) generateAuthCode() string {
	return opn.generateToken()
}

// generateToken generates a random token
func (opn *OAuth2ProviderNode) generateToken() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	
	token := base64.URLEncoding.EncodeToString(randomBytes)
	token = strings.ReplaceAll(token, "=", "") // Remove padding
	token = strings.ReplaceAll(token, "+", "") // Avoid URL encoding issues
	token = strings.ReplaceAll(token, "/", "") // Avoid URL encoding issues
	
	return token[:32] // Truncate to 32 characters
}

// findClient finds a client by ID
func (opn *OAuth2ProviderNode) findClient(clientID string) *OAuth2Client {
	for _, client := range opn.config.Clients {
		if client.ID == clientID {
			return &client
		}
	}
	
	return nil
}

// OAuth2ProviderNodeFromConfig creates a new OAuth2 provider node from a configuration map
func OAuth2ProviderNodeFromConfig(config map[string]interface{}) (engine.NodeInstance, error) {
	var operation OAuth2OperationType
	if op, exists := config["operation"]; exists {
		if opStr, ok := op.(string); ok {
			operation = OAuth2OperationType(opStr)
		}
	}

	var issuer string
	if iss, exists := config["issuer"]; exists {
		if issStr, ok := iss.(string); ok {
			issuer = issStr
		}
	}

	var authCodeExpiry float64
	if expiry, exists := config["auth_code_expiry_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			authCodeExpiry = expiryFloat
		}
	}

	var accessTokenExpiry float64
	if expiry, exists := config["access_token_expiry_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			accessTokenExpiry = expiryFloat
		}
	}

	var refreshTokenExpiry float64
	if expiry, exists := config["refresh_token_expiry_seconds"]; exists {
		if expiryFloat, ok := expiry.(float64); ok {
			refreshTokenExpiry = expiryFloat
		}
	}

	var enablePKCE bool
	if pkce, exists := config["enable_pkce"]; exists {
		enablePKCE = getBoolValue(pkce)
	}

	var enableLogging bool
	if logging, exists := config["enable_logging"]; exists {
		enableLogging = getBoolValue(logging)
	}

	var clients []OAuth2Client
	if clientsSlice, exists := config["clients"]; exists {
		if clientsInterface, ok := clientsSlice.([]interface{}); ok {
			clients = make([]OAuth2Client, len(clientsInterface))
			for i, clientInterface := range clientsInterface {
				if clientMap, ok := clientInterface.(map[string]interface{}); ok {
					var redirectURIs []string
					if uris, exists := clientMap["redirect_uris"]; exists {
						if urisSlice, ok := uris.([]interface{}); ok {
							redirectURIs = make([]string, len(urisSlice))
							for j, uri := range urisSlice {
								redirectURIs[j] = getStringValue(uri)
							}
						}
					}

					var grantTypes []string
					if types, exists := clientMap["grant_types"]; exists {
						if typesSlice, ok := types.([]interface{}); ok {
							grantTypes = make([]string, len(typesSlice))
							for j, t := range typesSlice {
								grantTypes[j] = getStringValue(t)
							}
						}
					}

					var responseTypes []string
					if types, exists := clientMap["response_types"]; exists {
						if typesSlice, ok := types.([]interface{}); ok {
							responseTypes = make([]string, len(typesSlice))
							for j, t := range typesSlice {
								responseTypes[j] = getStringValue(t)
							}
						}
					}

					var scope []string
					if scopes, exists := clientMap["scope"]; exists {
						if scopesSlice, ok := scopes.([]interface{}); ok {
							scope = make([]string, len(scopesSlice))
							for j, s := range scopesSlice {
								scope[j] = getStringValue(s)
							}
						}
					}

					var createdAt time.Time
					if created, exists := clientMap["created_at"]; exists {
						if createdStr, ok := created.(string); ok {
							createdAt, _ = time.Parse(time.RFC3339, createdStr)
						}
					}

					var updatedAt time.Time
					if updated, exists := clientMap["updated_at"]; exists {
						if updatedStr, ok := updated.(string); ok {
							updatedAt, _ = time.Parse(time.RFC3339, updatedStr)
						}
					}

					clients[i] = OAuth2Client{
						ID:                      getStringValue(clientMap["id"]),
						Name:                    getStringValue(clientMap["name"]),
						Description:             getStringValue(clientMap["description"]),
						ClientSecret:            getStringValue(clientMap["client_secret"]),
						RedirectURIs:            redirectURIs,
						GrantTypes:              grantTypes,
						ResponseTypes:           responseTypes,
						Scope:                   scope,
						TokenEndpointAuthMethod: getStringValue(clientMap["token_endpoint_auth_method"]),
						Enabled:                 getBoolValue(clientMap["enabled"]),
						CreatedAt:               createdAt,
						UpdatedAt:               updatedAt,
					}
				}
			}
		}
	}

	var signingKey string
	if key, exists := config["signing_key"]; exists {
		if keyStr, ok := key.(string); ok {
			signingKey = keyStr
		}
	}

	var encryptionKey string
	if key, exists := config["encryption_key"]; exists {
		if keyStr, ok := key.(string); ok {
			encryptionKey = keyStr
		}
	}

	nodeConfig := &OAuth2Config{
		Operation:       operation,
		Issuer:          issuer,
		AuthCodeExpiry:  time.Duration(authCodeExpiry) * time.Second,
		AccessTokenExpiry: time.Duration(accessTokenExpiry) * time.Second,
		RefreshTokenExpiry: time.Duration(refreshTokenExpiry) * time.Second,
		EnablePKCE:      enablePKCE,
		EnableLogging:   enableLogging,
		Clients:         clients,
		SigningKey:      signingKey,
		EncryptionKey:   encryptionKey,
	}

	return NewOAuth2ProviderNode(nodeConfig), nil
}

// RegisterOAuth2ProviderNode registers the OAuth2 provider node type with the engine
func RegisterOAuth2ProviderNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("oauth2_provider", func(config map[string]interface{}) (engine.NodeInstance, error) {
		return OAuth2ProviderNodeFromConfig(config)
	})
}