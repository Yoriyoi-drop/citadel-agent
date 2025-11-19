// citadel-agent/backend/internal/repositories/base_repository.go
package repositories

import (
	"gorm.io/gorm"
)

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetDB returns the GORM database instance
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}

// WithTransaction executes the provided function within a transaction
func (r *BaseRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Count counts all records in the table
func (r *BaseRepository) Count(model interface{}) (int64, error) {
	var count int64
	err := r.db.Model(model).Count(&count).Error
	return count, err
}

// Exists checks if a record exists by ID
func (r *BaseRepository) Exists(model interface{}, id string) (bool, error) {
	var count int64
	err := r.db.Model(model).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}