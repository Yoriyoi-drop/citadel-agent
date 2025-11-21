package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// OAuthProvider represents different OAuth providers
type OAuthProvider string

const (
	GitHub OAuthProvider = "github"
	Google OAuthProvider = "google"
)

// DeviceCodeRequest represents request for device code
type DeviceCodeRequest struct {
	Provider OAuthProvider `json:"provider"`
}

// DeviceCodeResponse represents response for device code
type DeviceCodeResponse struct {
	UserCode        string `json:"user_code"`
	DeviceCode      string `json:"device_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceVerifyRequest represents request for device code verification
type DeviceVerifyRequest struct {
	Provider   OAuthProvider `json:"provider"`
	DeviceCode string        `json:"device_code"`
}

// TokenResponse represents JWT token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	Provider     string    `json:"provider" db:"provider"` // github, google, local
	ProviderID   string    `json:"provider_id" db:"provider_id"`
	AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at" db:"last_login_at"`
}

// DeviceSession represents a device flow session
type DeviceSession struct {
	DeviceCode    string    `json:"device_code"`
	UserCode      string    `json:"user_code"`
	Provider      string    `json:"provider"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"` // pending, approved, expired
	AccessToken   string    `json:"access_token,omitempty"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
	ExpiresIn     int       `json:"expires_in"`
	Verification  time.Time `json:"verification_time,omitempty"`
}

// AuthService handles authentication
type AuthService struct {
	db         *database.DB
	oauth      map[OAuthProvider]*oauth2.Config
	deviceCode map[string]*DeviceSession // In production, use Redis or DB
}

// NewAuthService creates a new auth service
func NewAuthService(db *database.DB) *AuthService {
	authService := &AuthService{
		db:         db,
		oauth:      make(map[OAuthProvider]*oauth2.Config),
		deviceCode: make(map[string]*DeviceSession),
	}

	// Initialize GitHub OAuth config
	authService.oauth[GitHub] = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URI"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Initialize Google OAuth config
	authService.oauth[Google] = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return authService
}

// GenerateState generates a random state parameter for OAuth
func (s *AuthService) GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GithubLogin redirects to GitHub OAuth
func (s *AuthService) GithubLogin(w http.ResponseWriter, r *http.Request) {
	state := s.GenerateState()
	url := s.oauth[GitHub].AuthCodeURL(state, oauth2.AccessTypeOnline)
	
	// Store state in session or database (simplified for this example)
	// In production, use secure session management
	http.Redirect(w, r, url, http.StatusFound)
}

// GoogleLogin redirects to Google OAuth
func (s *AuthService) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := s.GenerateState()
	url := s.oauth[Google].AuthCodeURL(state, oauth2.AccessTypeOnline)
	
	// Store state in session or database
	http.Redirect(w, r, url, http.StatusFound)
}

// GithubCallback handles GitHub OAuth callback
func (s *AuthService) GithubCallback(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from the callback
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := s.oauth[GitHub].Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info from GitHub API
	user, err := s.getGithubUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get GitHub user: %v", err)
		http.Error(w, "Failed to get user info from GitHub", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	dbUser, err := s.createOrUpdateUser(user, string(GitHub))
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWT(dbUser)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Update last login at
	s.updateLastLogin(dbUser.ID)

	// In production, set secure httpOnly cookie
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false for development without HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600, // 1 hour
	}
	http.SetCookie(w, &cookie)

	// Redirect to frontend with success
	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/auth/success?token="+accessToken, http.StatusFound)
}

// GoogleCallback handles Google OAuth callback
func (s *AuthService) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from the callback
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := s.oauth[Google].Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google API
	user, err := s.getGoogleUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get Google user: %v", err)
		http.Error(w, "Failed to get user info from Google", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	dbUser, err := s.createOrUpdateUser(user, string(Google))
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateJWT(dbUser)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Update last login at
	s.updateLastLogin(dbUser.ID)

	// In production, set secure httpOnly cookie
	cookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false for development without HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600, // 1 hour
	}
	http.SetCookie(w, &cookie)

	// Redirect to frontend with success
	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"/auth/success?token="+accessToken, http.StatusFound)
}

// DeviceCodeInit initiates device authorization flow
func (s *AuthService) DeviceCodeInit(w http.ResponseWriter, r *http.Request) {
	var req DeviceCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate provider
	if req.Provider != GitHub && req.Provider != Google {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Generate codes
	deviceCode := s.generateRandomString(32)
	userCode := s.generateUserCode()

	// Create device session
	session := &DeviceSession{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Provider:   string(req.Provider),
		ExpiresAt:  time.Now().Add(10 * time.Minute), // 10 minutes expiry
		Status:     "pending",
	}

	// Store session (in production, use Redis or database)
	s.deviceCode[deviceCode] = session

	// Return device code and user instructions
	response := DeviceCodeResponse{
		UserCode:        userCode,
		DeviceCode:      deviceCode,
		VerificationURI: "https://github.com/login/device", // Use provider-specific URL
		ExpiresIn:       600, // 10 minutes
		Interval:        5,   // Poll every 5 seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeviceCodeVerify verifies device authorization
func (s *AuthService) DeviceCodeVerify(w http.ResponseWriter, r *http.Request) {
	var req DeviceVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get stored session
	session, exists := s.deviceCode[req.DeviceCode]
	if !exists {
		http.Error(w, "Invalid device code", http.StatusBadRequest)
		return
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		delete(s.deviceCode, req.DeviceCode)
		http.Error(w, "Device code expired", http.StatusBadRequest)
		return
	}

	// Check if already approved
	if session.Status == "approved" {
		// Return tokens
		response := TokenResponse{
			AccessToken:  session.AccessToken,
			RefreshToken: session.RefreshToken,
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return pending status
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"status": "pending", "message": "Waiting for user to approve"}`)
}

// generateRandomString generates a random string of given length
func (s *AuthService) generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateUserCode generates a user-friendly code in format XXXX-YYYY
func (s *AuthService) generateUserCode() string {
	b := make([]byte, 4) // 4 bytes = 8 hex chars
	rand.Read(b)
	code := strings.ToUpper(hex.EncodeToString(b))
	
	// Format as XXXX-YYYY for readability
	if len(code) >= 8 {
		return code[:4] + "-" + code[4:8]
	}
	return code
}

// getGithubUser retrieves user information from GitHub API
func (s *AuthService) getGithubUser(accessToken string) (*User, error) {
	// This would make an HTTP request to GitHub API
	// For example: GET https://api.github.com/user with Authorization: Bearer {token}
	
	// Mock implementation
	user := &User{
		ID:        s.generateRandomString(16),
		Email:     "githubuser@example.com",
		Username:  "GithubUser",
		Provider:  "github",
		ProviderID: "github_user_id",
		AvatarURL: "https://avatars.githubusercontent.com/u/123456789",
	}
	
	return user, nil
}

// getGoogleUser retrieves user information from Google API
func (s *AuthService) getGoogleUser(accessToken string) (*User, error) {
	// This would make an HTTP request to Google API
	// For example: GET https://www.googleapis.com/oauth2/v2/userinfo with Authorization: Bearer {token}
	
	// Mock implementation
	user := &User{
		ID:        s.generateRandomString(16),
		Email:     "googleuser@example.com",
		Username:  "GoogleUser",
		Provider:  "google",
		ProviderID: "google_user_id",
		AvatarURL: "https://lh3.googleusercontent.com/a-/123456789",
	}
	
	return user, nil
}

// createOrUpdateUser creates or updates a user in the database
func (s *AuthService) createOrUpdateUser(user *User, provider string) (*User, error) {
	// In a real implementation, this would query the database
	// to check if the user already exists and update or create accordingly
	
	// Mock implementation
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	return user, nil
}

// generateJWT creates JWT tokens for the user
func (s *AuthService) generateJWT(user *User) (string, string, error) {
	// Create access token (short-lived)
	accessTokenExp := time.Now().Add(1 * time.Hour)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "citadel-agent",
		},
	})

	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	// Create refresh token (long-lived, stored in DB)
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour) // 7 days
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "citadel-agent",
		},
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// updateLastLogin updates the last login time for a user
func (s *AuthService) updateLastLogin(userID string) {
	// In a real implementation, this would update the user's last login time in the database
	log.Printf("User %s logged in at %v", userID, time.Now())
}

// Login handles local email/password authentication
func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	// Implementation for local email/password authentication
	// This would validate credentials and generate JWT tokens
}