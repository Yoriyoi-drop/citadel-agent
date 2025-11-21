// backend/internal/repositories/user_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *pgxpool.Pool
}

// User represents a user in the database
type User struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	Password  string                 `json:"password"` // Hashed password
	Role      string                 `json:"role"`     // 'admin', 'user', 'viewer'
	Status    string                 `json:"status"`   // 'active', 'inactive', 'suspended'
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Profile   map[string]interface{} `json:"profile"`
	Preferences map[string]interface{} `json:"preferences"`
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (ur *UserRepository) Create(ctx context.Context, user *User) (*User, error) {
	// Serialize profile and preferences to JSON
	profileJSON, err := json.Marshal(user.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		INSERT INTO users (
			id, email, name, password, role, status, 
			created_at, updated_at, profile, preferences
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, email, name, role, status, created_at, updated_at
	`

	var createdUser User
	err = ur.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Password,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
		profileJSON,
		preferencesJSON,
	).Scan(
		&createdUser.ID,
		&createdUser.Email,
		&createdUser.Name,
		&createdUser.Role,
		&createdUser.Status,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Set the non-returned fields
	createdUser.Profile = user.Profile
	createdUser.Preferences = user.Preferences

	return &createdUser, nil
}

// GetByID retrieves a user by ID
func (ur *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, name, role, status, created_at, updated_at, profile, preferences
		FROM users
		WHERE id = $1
	`

	var user User
	var profileJSON, preferencesJSON []byte

	err := ur.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&profileJSON,
		&preferencesJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Deserialize profile and preferences
	if profileJSON != nil {
		if err := json.Unmarshal(profileJSON, &user.Profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
		}
	}

	if preferencesJSON != nil {
		if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
		}
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, role, status, created_at, updated_at, profile, preferences
		FROM users
		WHERE email = $1
	`

	var user User
	var profileJSON, preferencesJSON []byte

	err := ur.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&profileJSON,
		&preferencesJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Deserialize profile and preferences
	if profileJSON != nil {
		if err := json.Unmarshal(profileJSON, &user.Profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
		}
	}

	if preferencesJSON != nil {
		if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
		}
	}

	return &user, nil
}

// Update updates an existing user
func (ur *UserRepository) Update(ctx context.Context, user *User) (*User, error) {
	// Serialize profile and preferences to JSON
	profileJSON, err := json.Marshal(user.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE users
		SET name = $2, role = $3, status = $4, updated_at = $5, 
		    profile = $6, preferences = $7
		WHERE id = $1
		RETURNING id, email, name, role, status, created_at, updated_at
	`

	var updatedUser User
	err = ur.db.QueryRow(ctx, query,
		user.ID,
		user.Name,
		user.Role,
		user.Status,
		time.Now(),
		profileJSON,
		preferencesJSON,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Email,
		&updatedUser.Name,
		&updatedUser.Role,
		&updatedUser.Status,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Set the non-returned fields
	updatedUser.Profile = user.Profile
	updatedUser.Preferences = user.Preferences

	return &updatedUser, nil
}

// List retrieves a list of users
func (ur *UserRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, role, status, created_at, updated_at, profile, preferences
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := ur.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var profileJSON, preferencesJSON []byte

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&profileJSON,
			&preferencesJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Deserialize profile and preferences
		if profileJSON != nil {
			if err := json.Unmarshal(profileJSON, &user.Profile); err != nil {
				return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
			}
		}

		if preferencesJSON != nil {
			if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
				return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
			}
		}

		users = append(users, &user)
	}

	return users, nil
}

// Delete removes a user by ID
func (ur *UserRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// GetByRole retrieves users by role
func (ur *UserRepository) GetByRole(ctx context.Context, role string) ([]*User, error) {
	query := `
		SELECT id, email, name, role, status, created_at, updated_at, profile, preferences
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
	`

	rows, err := ur.db.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by role: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var profileJSON, preferencesJSON []byte

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&profileJSON,
			&preferencesJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Deserialize profile and preferences
		if profileJSON != nil {
			if err := json.Unmarshal(profileJSON, &user.Profile); err != nil {
				return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
			}
		}

		if preferencesJSON != nil {
			if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
				return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
			}
		}

		users = append(users, &user)
	}

	return users, nil
}

// CountByRole counts users by role
func (ur *UserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	query := "SELECT COUNT(*) FROM users WHERE role = $1"

	var count int64
	err := ur.db.QueryRow(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	return count, nil
}