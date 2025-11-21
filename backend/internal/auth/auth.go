package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Hashed password
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"` // 'admin', 'user', 'viewer'
	Status    string    `json:"status"` // 'active', 'inactive', 'suspended'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthService handles authentication and authorization
type AuthService struct {
	db          *pgxpool.Pool
	rbacManager *RBACManager
	jwtSecret   []byte
}

// NewAuthService creates a new auth service
func NewAuthService(db *pgxpool.Pool, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		rbacManager: NewRBACManager(db),
		jwtSecret:   []byte(jwtSecret),
	}
}

// AuthMiddleware provides authentication middleware for protected routes
func (s *AuthService) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header missing",
			})
		}

		// Check if it's a Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Extract token
		tokenString := authHeader[7:]

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		})

		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Check if token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user ID from claims and store in context
			if userID, exists := claims["user_id"]; exists {
				if id, ok := userID.(string); ok {
					c.Locals("user_id", id)
				}
			}

			// Extract role from claims and store in context
			if role, exists := claims["role"]; exists {
				if roleStr, ok := role.(string); ok {
					c.Locals("user_role", roleStr)
				}
			}
		} else {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		return c.Next()
	}
}

// AuthenticateUser authenticates a user with email and password
func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	// Query the user from the database
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, role, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND status = 'active'
	`

	var user User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password, // This is the hashed password
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found or inactive")
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *User) (string, error) {
	// Create claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),                      // Issued at time
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	// Validate input
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate user ID
	userID := uuid.New().String()

	// Insert user into database
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, first_name, last_name, role, status, created_at, updated_at
	`

	var user User
	err = s.db.QueryRow(ctx,
		query,
		userID,
		email,
		string(hashedPassword),
		firstName,
		lastName,
		"viewer", // Default role
		"active",
		time.Now(),
		time.Now(),
	).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Assign default role to the user
	err = s.rbacManager.AssignRole(ctx, user.ID, "viewer", "system")
	if err != nil {
		// Log the error but don't fail the registration
		// The user is created but without the default role
		fmt.Printf("Warning: failed to assign default role to user %s: %v\n", user.ID, err)
	}

	return &user, nil
}

// GetUserPermissions gets all permissions for a user
func (s *AuthService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// First check if user has admin role (admin has all permissions)
	userRoles, err := s.rbacManager.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	var allPermissions []string
	for _, role := range userRoles {
		// If any role is admin, return all permissions (represented by wildcard)
		if role.ID == "admin" || role.Name == "Administrator" {
			return []string{"*:*"}, nil
		}
		allPermissions = append(allPermissions, role.Permissions...)
	}

	return allPermissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *AuthService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	return s.rbacManager.HasPermission(ctx, userID, permission)
}

// GetRBACManager returns the RBAC manager
func (s *AuthService) GetRBACManager() *RBACManager {
	return s.rbacManager
}