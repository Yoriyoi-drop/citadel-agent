package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Build from environment variables
		dbHost := getEnvOrDefault("DB_HOST", "localhost")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
		dbName := getEnvOrDefault("DB_NAME", "citadel_agent")
		
		dbURL = "postgresql://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
	}

	// Connect to database
	pool, err := pgxpool.New(
		context.Background(),
		dbURL,
	)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	fmt.Println("Running migrations...")

	// In a real implementation, you would run actual database migrations here
	// For now, this is a placeholder that just confirms connection
	err = runMigrations(pool)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	fmt.Println("Migrations completed successfully!")
}

// runMigrations runs actual database migrations
func runMigrations(pool *pgxpool.Pool) error {
	// Execute migration queries
	queries := []string{
		// Create users table if not exists
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Create workflows table if not exists
		`CREATE TABLE IF NOT EXISTS workflows (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			definition JSONB,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Create nodes table if not exists
		`CREATE TABLE IF NOT EXISTS nodes (
			id SERIAL PRIMARY KEY,
			workflow_id INTEGER REFERENCES workflows(id) ON DELETE CASCADE,
			type VARCHAR(255) NOT NULL,
			config JSONB,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		// Create executions table if not exists
		`CREATE TABLE IF NOT EXISTS executions (
			id SERIAL PRIMARY KEY,
			workflow_id INTEGER REFERENCES workflows(id) ON DELETE CASCADE,
			status VARCHAR(50) DEFAULT 'running',
			result JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := pool.Exec(
			context.Background(),
			query,
		)
		if err != nil {
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}