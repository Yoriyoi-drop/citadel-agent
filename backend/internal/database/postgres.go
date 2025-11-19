package database

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"citadel-agent/backend/internal/config"
)

// DB holds the database connection
type DB struct {
	GormDB *gorm.DB
	Redis  *redis.Client
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*DB, error) {
	var dsn string
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	}

	// Connect to PostgreSQL
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Default Redis address
		Password: "",               // No password by default
		DB:       0,                // Use default DB
	})

	// Test Redis connection
	if err := redisClient.Ping(nil).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		// Continue without Redis instead of failing
	} else {
		log.Println("Successfully connected to Redis")
	}

	return &DB{
		GormDB: gormDB,
		Redis:  redisClient,
	}, nil
}

// Close closes the database connections
func (db *DB) Close() error {
	// Close PostgreSQL connection
	sqlDB, err := db.GormDB.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}

	// Close Redis connection
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}