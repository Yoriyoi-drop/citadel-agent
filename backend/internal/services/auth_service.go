// backend/internal/services/auth_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication and authorization
type AuthService struct {
	userRepo      *repositories.UserRepository
	apiKeyRepo    *repositories.APIKeyRepository
	teamRepo      *repositories.TeamRepository
	jwtSecret     string
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

// AuthResult represents the result of an authentication operation
type AuthResult struct {
	User       *models.User
	Token      string
	Refresh    string
	TokenType  string
	ExpiresIn  int64
	TeamID     *string
	TeamRole   *string
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo *repositories.UserRepository,
	apiKeyRepo *repositories.APIKeyRepository,
	teamRepo *repositories.TeamRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		apiKeyRepo:    apiKeyRepo,
		teamRepo:      teamRepo,
		jwtSecret:     jwtSecret,
		tokenExpiry:   24 * time.Hour,      // 24 hours
		refreshExpiry: 30 * 24 * time.Hour, // 30 days
	}
}

// RegisterUser registers a new user
func (as *AuthService) RegisterUser(ctx context.Context, email, password, name string) (*AuthResult, error) {
	// Check if user already exists
	existingUser, err := as.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	user := &models.User{
		Email:         email,
		Name:          name,
		PasswordHash:  string(hashedPassword),
		Role:          models.UserRoleViewer,
		Status:        models.UserStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Profile:       models.UserProfile{},
		Preferences:   make(map[string]interface{}),
	}

	createdUser, err := as.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create initial team for the user
	team := &models.Team{
		Name:        fmt.Sprintf("%s's Team", name),
		Description: fmt.Sprintf("Personal team for %s", name),
		OwnerID:     createdUser.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Settings:    make(map[string]interface{}),
	}

	createdTeam, err := as.teamRepo.Create(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to create user team: %w", err)
	}

	// Add user to their team as owner
	member := &models.TeamMember{
		TeamID:    createdTeam.ID,
		UserID:    createdUser.ID,
		Role:      models.UserRoleAdmin,
		JoinedAt:  time.Now(),
		IsActive:  true,
		UpdatedAt: time.Now(),
	}

	_, err = as.teamRepo.AddMember(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to team: %w", err)
	}

	// Generate tokens
	token, refresh, err := as.generateTokens(createdUser, &createdTeam.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	authResult := &AuthResult{
		User:       createdUser,
		Token:      token,
		Refresh:    refresh,
		TokenType:  "Bearer",
		ExpiresIn:  int64(as.tokenExpiry.Seconds()),
		TeamID:     &createdTeam.ID,
		TeamRole:   &member.Role,
	}

	return authResult, nil
}

// LoginUser authenticates a user and returns tokens
func (as *AuthService) LoginUser(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := as.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Status == models.UserStatusSuspended {
		return nil, fmt.Errorf("account suspended")
	}

	if user.Status == models.UserStatusInactive {
		return nil, fmt.Errorf("account inactive")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	user.LastLoginAt = time.Now()
	_, err = as.userRepo.Update(ctx, user.ID, user)
	if err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Get user's team membership
	teamMembers, err := as.teamRepo.GetUserMemberships(ctx, user.ID)
	if err != nil {
		// User might not be in a team yet
		teamMembers = []*models.TeamMember{}
	}

	var teamID *string
	var teamRole *string

	if len(teamMembers) > 0 {
		// Use the first active team
		for _, member := range teamMembers {
			if member.IsActive {
				teamID = &member.TeamID
				teamRole = &member.Role
				break
			}
		}
	}

	// Generate tokens
	token, refresh, err := as.generateTokens(user, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	authResult := &AuthResult{
		User:       user,
		Token:      token,
		Refresh:    refresh,
		TokenType:  "Bearer",
		ExpiresIn:  int64(as.tokenExpiry.Seconds()),
		TeamID:     teamID,
		TeamRole:   teamRole,
	}

	return authResult, nil
}

// ValidateToken validates a JWT token
func (as *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	// In a real implementation, this would decode and validate the JWT token
	// For now, we'll just return an error as the implementation requires JWT package
	return nil, fmt.Errorf("JWT validation not implemented yet")
}

// RefreshToken refreshes an access token using a refresh token
func (as *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// In a real implementation, this would validate the refresh token
	// and generate new access tokens
	// For now, we'll return an error as the implementation requires JWT package
	return nil, fmt.Errorf("token refresh not implemented yet")
}

// ChangePassword changes a user's password
func (as *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := as.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("incorrect old password")
	}

	// Hash new password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.PasswordHash = string(hashedNewPassword)
	user.UpdatedAt = time.Now()

	_, err = as.userRepo.Update(ctx, userID, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key for a user
func (as *AuthService) CreateAPIKey(ctx context.Context, userID, teamID *string, name string, permissions []string, expiresAt *time.Time) (*models.APIKey, error) {
	// In a real system, we would generate a secure API key
	// For now, we'll create a placeholder implementation
	apiKey := &models.APIKey{
		Name:        name,
		UserID:      *userID,
		TeamID:      teamID,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		LastUsedAt:  nil,
	}

	// Generate a secure key (placeholder)
	apiKey.Prefix = "sk_1234567890" // This should be generated securely in real implementation
	apiKey.KeyHash = "hashed_key_placeholder" // This should be a real hash

	createdKey, err := as.apiKeyRepo.Create(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return createdKey, nil
}

// ValidateAPIKey validates an API key and its permissions
func (as *AuthService) ValidateAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	// Extract key prefix for lookup
	if len(key) < 10 {
		return nil, fmt.Errorf("invalid key format")
	}

	prefix := key[:10]
	apiKey, err := as.apiKeyRepo.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if apiKey.Status == models.APIKeyStatusInactive {
		return nil, fmt.Errorf("API key is inactive")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	return apiKey, nil
}

// RevokeAPIKey revokes an API key
func (as *AuthService) RevokeAPIKey(ctx context.Context, keyID string) error {
	return as.apiKeyRepo.Delete(ctx, keyID)
}

// GetCurrentUser gets the current user by ID
func (as *AuthService) GetCurrentUser(ctx context.Context, userID string) (*models.User, error) {
	return as.userRepo.GetByID(ctx, userID)
}

// UpdateUserProfile updates a user's profile
func (as *AuthService) UpdateUserProfile(ctx context.Context, userID string, profile models.UserProfile) (*models.User, error) {
	user, err := as.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	user.Profile = profile
	user.UpdatedAt = time.Now()

	updatedUser, err := as.userRepo.Update(ctx, userID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return updatedUser, nil
}

// UpdateUserPreferences updates a user's preferences
func (as *AuthService) UpdateUserPreferences(ctx context.Context, userID string, preferences map[string]interface{}) (*models.User, error) {
	user, err := as.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Merge new preferences with existing ones
	for key, value := range preferences {
		user.Preferences[key] = value
	}
	user.UpdatedAt = time.Now()

	updatedUser, err := as.userRepo.Update(ctx, userID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}

	return updatedUser, nil
}

// generateTokens generates access and refresh tokens for a user
func (as *AuthService) generateTokens(user *models.User, teamID *string) (string, string, error) {
	// In a real implementation, this would generate actual JWT tokens
	// For now, we'll return placeholder tokens
	accessToken := "access_token_placeholder"
	refreshToken := "refresh_token_placeholder"
	
	return accessToken, refreshToken, nil
}

// GetUserByAPIKey gets user information based on API key
func (as *AuthService) GetUserByAPIKey(ctx context.Context, apiKey *models.APIKey) (*models.User, error) {
	user, err := as.userRepo.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for API key: %w", err)
	}

	return user, nil
}

// CheckPermission checks if a user has a specific permission
func (as *AuthService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	user, err := as.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Admins have all permissions
	if user.Role == models.UserRoleAdmin {
		return true, nil
	}

	// Check if user has the specific permission
	// In a real system, this would check against role-based permissions
	// For now, we'll implement a basic check
	return true, nil
}