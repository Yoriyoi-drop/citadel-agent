package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	fmt.Println("Seeding database...")

	// Seed the database
	err = seedDatabase(pool)
	if err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	fmt.Println("Database seeding completed successfully!")
}

// seedDatabase adds initial data to the database
func seedDatabase(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// Create default admin user
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // Default for development only
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Set defaults for admin user if environment variables are not set
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@citadel-agent.com"
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminFirstName := os.Getenv("ADMIN_FIRST_NAME")
	if adminFirstName == "" {
		adminFirstName = "Admin"
	}

	adminLastName := os.Getenv("ADMIN_LAST_NAME")
	if adminLastName == "" {
		adminLastName = "User"
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, username, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO NOTHING`,
		adminEmail,
		adminUsername,
		string(hashedPassword),
		adminFirstName,
		adminLastName,
	)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create sample workflow
	_, err = pool.Exec(ctx, `
		INSERT INTO workflows (name, description, definition, status) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO NOTHING`,
		"Sample Workflow",
		"A basic workflow for testing",
		`{"nodes": []}`,
		"active",
	)
	if err != nil {
		return fmt.Errorf("failed to create sample workflow: %w", err)
	}

	// Add more seed data as needed
	fmt.Println("Created admin user: admin@citadel-agent.com with password 'admin123'")
	fmt.Println("Created sample workflow: 'Sample Workflow'")

	return nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}