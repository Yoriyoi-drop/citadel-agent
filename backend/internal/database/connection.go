// backend/internal/database/connection.go
package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a connection to the database
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run auto-migration for all models
	// In a real implementation, you'd want to run migrations properly
	// err = db.AutoMigrate(
	// 	&auth.User{},
	// 	&auth.Role{},
	// 	&auth.Permission{},
	// 	&auth.APIKey{},
	// 	&ai.Agent{},
	// 	// Add other models here
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to migrate database: %w", err)
	// }

	log.Println("Database connected successfully")
	
	return db, nil
}