// backend/internal/auth/service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService provides authentication and authorization services
type AuthService struct {
	DB         *gorm.DB
	JWTSecret  string
	TokenExpiry time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		DB:          db,
		JWTSecret:   jwtSecret,
		TokenExpiry: 24 * time.Hour, // Default 24 hours
	}
}

// RegisterUser creates a new user
func (s *AuthService) RegisterUser(ctx context.Context, email, username, password, firstName, lastName string) (*User, error) {
	// Check if user already exists
	var existingUser User
	if err := s.DB.Where("email = ? OR username = ?", email, username).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email or username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:     email,
		Username:  username,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	if err := s.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// AuthenticateUser authenticates a user and returns a JWT token
func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (string, *User, error) {
	var user User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid credentials")
		}
		return "", nil, err
	}

	if !user.IsActive {
		return "", nil, errors.New("user account is deactivated")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Update last login
	user.LastLogin = &time.Now()
	s.DB.Save(&user)

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		var user User
		if err := s.DB.First(&user, uint(userID)).Error; err != nil {
			return nil, err
		}

		return &user, nil
	}

	return nil, errors.New("invalid token")
}

// generateJWT creates a JWT token for the given user
func (s *AuthService) generateJWT(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"username":  user.Username,
		"exp":       time.Now().Add(s.TokenExpiry).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.JWTSecret))
}

// HasPermission checks if a user has a specific permission
func (s *AuthService) HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	var user User
	if err := s.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return false, err
	}

	// Check if user has the required permission
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Resource == resource && permission.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}

// AssignRole assigns a role to a user
func (s *AuthService) AssignRole(ctx context.Context, userID, roleID uint) error {
	var user User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var role Role
	if err := s.DB.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := s.DB.Model(&user).Association("Roles").Append(&role); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RevokeRole removes a role from a user
func (s *AuthService) RevokeRole(ctx context.Context, userID, roleID uint) error {
	var user User
	if err := s.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var role Role
	if err := s.DB.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Remove role from user
	if err := s.DB.Model(&user).Association("Roles").Delete(&role); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key for a user
func (s *AuthService) CreateAPIKey(ctx context.Context, userID uint, name string) (string, *APIKey, error) {
	// Generate API key (we'll use a simple approach, in production use crypto/rand)
	userKey := fmt.Sprintf("citadel_%d_%d_%s", userID, time.Now().Unix(), name)
	
	// Hash the key for storage
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(userKey), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	apiKey := &APIKey{
		UserID:  userID,
		Name:    name,
		KeyHash: string(hashedKey),
	}

	if err := s.DB.Create(apiKey).Error; err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return userKey, apiKey, nil
}

// ValidateAPIKey validates an API key and returns user info
func (s *AuthService) ValidateAPIKey(ctx context.Context, key string) (*User, error) {
	// Hash the provided key to compare with stored hash
	var apiKey APIKey
	if err := s.DB.Where("key_hash = ?", key).First(&apiKey).Error; err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	if !apiKey.IsActive {
		return nil, errors.New("API key is deactivated")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New("API key has expired")
	}

	var user User
	if err := s.DB.First(&user, apiKey.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}