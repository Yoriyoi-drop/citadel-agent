package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/auth"
	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo *repositories.UserRepository
	repositoryFactory *repositories.RepositoryFactory
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	repositoryFactory := repositories.NewRepositoryFactory(db)

	return &UserService{
		userRepo: repositoryFactory.GetUserRepository(),
		repositoryFactory: repositoryFactory,
	}
}

// CreateUser creates a new user with validation and password hashing
func (s *UserService) CreateUser(user *models.User) error {
	// Validate input
	if user.Email == "" {
		return errors.New("email is required")
	}
	if user.Username == "" {
		return errors.New("username is required")
	}
	if user.Password == "" {
		return errors.New("password is required")
	}
	
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(user.Email)
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}
	
	// Check username uniqueness
	existingUser, _ = s.userRepo.GetByUsername(user.Username)
	if existingUser != nil {
		return errors.New("user with this username already exists")
	}
	
	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)
	
	// Set defaults and timestamps
	if user.Role == "" {
		user.Role = "user"
	}
	if user.Status == "" {
		user.Status = "active"
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	return s.userRepo.Create(user)
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*models.User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}
	
	return s.userRepo.GetByID(id)
}

// GetUserByEmail retrieves a user by email (without password)
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	// Don't return the password hash
	user.Password = ""
	return user, nil
}

// GetUserByUsername retrieves a user by username (without password)
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	// Don't return the password hash
	user.Password = ""
	return user, nil
}

// UpdateUser updates a user with validation
func (s *UserService) UpdateUser(user *models.User) error {
	// Validate input
	if user.ID == "" {
		return errors.New("user ID is required")
	}
	
	// Check if user exists
	existing, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Update allowed fields (excluding password for now)
	existing.FirstName = user.FirstName
	existing.LastName = user.LastName
	existing.Email = user.Email
	existing.Username = user.Username
	existing.Role = user.Role
	existing.Status = user.Status
	existing.UpdatedAt = time.Now()
	
	return s.userRepo.Update(existing)
}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(userID, newPassword string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if newPassword == "" {
		return errors.New("new password is required")
	}
	
	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	existing.Password = string(hashedPassword)
	existing.UpdatedAt = time.Now()
	
	return s.userRepo.Update(existing)
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}
	
	return s.userRepo.Delete(id)
}

// AuthenticateUser authenticates a user by email and password
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}
	
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}
	
	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	// Don't return the password hash
	user.Password = ""
	return user, nil
}

// GenerateUserToken generates a JWT token for a user
func (s *UserService) GenerateUserToken(user *models.User, secret string, expiry int) (string, error) {
	if user == nil {
		return "", errors.New("user cannot be nil")
	}
	
	return auth.GenerateToken(user.ID, user.Email, secret, expiry)
}

// GetAllUsers retrieves all users with optional pagination
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAll()
}

// GetAllUsersWithPagination retrieves all users with pagination
func (s *UserService) GetAllUsersWithPagination(offset, limit int) ([]*models.User, error) {
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, errors.New("invalid pagination parameters")
	}
	
	return s.userRepo.GetAllWithPagination(offset, limit)
}

// GetUsersByRole retrieves users by role
func (s *UserService) GetUsersByRole(role string) ([]*models.User, error) {
	if role == "" {
		return nil, errors.New("role is required")
	}
	
	return s.userRepo.GetByRole(role)
}

// GetUsersByStatus retrieves users by status
func (s *UserService) GetUsersByStatus(status string) ([]*models.User, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}
	
	return s.userRepo.GetByStatus(status)
}

// CountUsers counts all users
func (s *UserService) CountUsers() (int64, error) {
	return s.userRepo.Count()
}

// CountUsersByRole counts users by role
func (s *UserService) CountUsersByRole(role string) (int64, error) {
	if role == "" {
		return 0, errors.New("role is required")
	}
	
	return s.userRepo.CountByRole(role)
}

// CountUsersByStatus counts users by status
func (s *UserService) CountUsersByStatus(status string) (int64, error) {
	if status == "" {
		return 0, errors.New("status is required")
	}
	
	return s.userRepo.CountByStatus(status)
}

// SearchUsersByName searches users by first name or last name
func (s *UserService) SearchUsersByName(name string) ([]*models.User, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	
	return s.userRepo.SearchByName(name)
}

// SearchUsersByEmail searches users by email
func (s *UserService) SearchUsersByEmail(email string) ([]*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	
	return s.userRepo.SearchByEmail(email)
}

// ActivateUser sets user status to active
func (s *UserService) ActivateUser(id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}
	
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	user.Status = "active"
	user.UpdatedAt = time.Now()
	
	return s.userRepo.Update(user)
}

// DeactivateUser sets user status to inactive
func (s *UserService) DeactivateUser(id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}
	
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	user.Status = "inactive"
	user.UpdatedAt = time.Now()
	
	return s.userRepo.Update(user)
}

// CheckPermission checks if a user has a specific permission
func (s *UserService) CheckPermission(userID string, permission auth.Permission) (bool, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	return auth.HasPermission(auth.Role(user.Role), permission), nil
}