// citadel-agent/backend/internal/repositories/user_repository.go
package repositories

import (
	"citadel-agent/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository handles user database operations
type UserRepository struct {
	BaseRepository
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: *NewBaseRepository(db),
	}
}

// Create creates a new user with password hashing
func (r *UserRepository) Create(user *models.User) error {
	// Hash the password before storing
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	
	return r.BaseRepository.db.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.BaseRepository.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.BaseRepository.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.BaseRepository.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.BaseRepository.db.Save(user).Error
}

// Delete soft deletes a user by ID
func (r *UserRepository) Delete(id string) error {
	return r.BaseRepository.db.Delete(&models.User{}, "id = ?", id).Error
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllWithPagination retrieves all users with pagination
func (r *UserRepository) GetAllWithPagination(offset, limit int) ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetByRole retrieves users by role
func (r *UserRepository) GetByRole(role string) ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Where("role = ?", role).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetByStatus retrieves users by status
func (r *UserRepository) GetByStatus(status string) ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Where("status = ?", status).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Count counts all users
func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountByRole counts users by role
func (r *UserRepository) CountByRole(role string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

// CountByStatus counts users by status
func (r *UserRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.BaseRepository.db.Model(&models.User{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// SearchByName searches users by first name or last name (case-insensitive partial match)
func (r *UserRepository) SearchByName(name string) ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Where("LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?)", "%"+name+"%", "%"+name+"%").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// SearchByEmail searches users by email (case-insensitive partial match)
func (r *UserRepository) SearchByEmail(email string) ([]*models.User, error) {
	var users []*models.User
	err := r.BaseRepository.db.Where("LOWER(email) LIKE LOWER(?)", "%"+email+"%").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// AuthenticateUser authenticates a user by email and password
func (r *UserRepository) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := r.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	
	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}
	
	// Don't return the password hash to the caller
	user.Password = ""
	return user, nil
}

// UpdatePassword updates a user's password (with password hashing)
func (r *UserRepository) UpdatePassword(userID, newPassword string) error {
	user, err := r.GetByID(userID)
	if err != nil {
		return err
	}
	
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	return r.BaseRepository.db.Model(user).Update("password", string(hashedPassword)).Error
}

// ActivateUser activates a user account
func (r *UserRepository) ActivateUser(userID string) error {
	return r.BaseRepository.db.Model(&models.User{}).Where("id = ?", userID).Update("status", "active").Error
}

// DeactivateUser deactivates a user account
func (r *UserRepository) DeactivateUser(userID string) error {
	return r.BaseRepository.db.Model(&models.User{}).Where("id = ?", userID).Update("status", "inactive").Error
}