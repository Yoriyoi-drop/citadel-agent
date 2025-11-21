package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// RegisterRoutes registers authentication routes
func (s *AuthService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/login", s.LoginHandler)
	mux.HandleFunc("/auth/logout", s.LogoutHandler)
	mux.HandleFunc("/auth/oauth/github", s.GithubLogin)
	mux.HandleFunc("/auth/oauth/github/callback", s.GithubCallback)
	mux.HandleFunc("/auth/oauth/google", s.GoogleLogin)
	mux.HandleFunc("/auth/oauth/google/callback", s.GoogleCallback)
	mux.HandleFunc("/auth/device", s.DeviceCodeInitHandler)
	mux.HandleFunc("/auth/device/verify", s.DeviceCodeVerifyHandler)
	mux.HandleFunc("/auth/token/refresh", s.RefreshTokenHandler)
	mux.HandleFunc("/auth/me", s.MeHandler)
}

// LoginHandler handles local email/password login
func (s *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.Login(w, r)
}

// LogoutHandler handles user logout
func (s *AuthService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Clear the access token cookie
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	
	// In production, invalidate refresh token in database
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// DeviceCodeInitHandler handles device code initiation
func (s *AuthService) DeviceCodeInitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Check if request is from CLI by checking user agent or specific header
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(userAgent, "Citadel-Agent-CLI") {
		s.DeviceCodeInit(w, r)
		return
	}
	
	http.Error(w, "Device code flow is only available for CLI", http.StatusForbidden)
}

// DeviceCodeVerifyHandler handles device code verification
func (s *AuthService) DeviceCodeVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Check if request is from CLI
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(userAgent, "Citadel-Agent-CLI") {
		s.DeviceCodeVerify(w, r)
		return
	}
	
	http.Error(w, "Device code flow is only available for CLI", http.StatusForbidden)
}

// RefreshTokenHandler handles JWT refresh token
func (s *AuthService) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract refresh token from authorization header or request body
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusBadRequest)
		return
	}
	
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}
	
	refreshToken := authHeader[7:]
	
	// Validate refresh token and generate new access token
	// In production, check if refresh token exists in database and is not revoked
	
	// For now, we just return an error as the implementation would be more complex
	http.Error(w, "Refresh token not implemented in this example", http.StatusNotImplemented)
}

// MeHandler returns authenticated user info
func (s *AuthService) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract and validate access token from cookie or authorization header
	cookie, err := r.Cookie("access_token")
	if err != nil {
		// Try authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}
		
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		
		accessToken := authHeader[7:]
		
		// Validate JWT token and extract user info
		claims, err := s.validateToken(accessToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Return user info
		user := map[string]interface{}{
			"id":         claims.UserID,
			"email":      claims.Email,
			"username":   claims.Username,
			"provider":   "local", // This would be determined from the user record
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
	
	// Validate JWT token from cookie
	claims, err := s.validateToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	
	// Return user info
	user := map[string]interface{}{
		"id":         claims.UserID,
		"email":      claims.Email,
		"username":   claims.Username,
		"provider":   "local", // This would be determined from the user record
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// validateToken validates JWT token and returns claims
func (s *AuthService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}