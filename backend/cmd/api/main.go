// backend/cmd/api/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/citadel-agent/backend/config"
	"github.com/citadel-agent/backend/internal/app"
)

func main() {
	// Load configuration
	cfg := config.DefaultEngineConfig()

	// Override with environment variables if needed
	if port := os.Getenv("PORT"); port != "" {
		// Parse port from env if needed
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.DatabaseConfig.Host = dbHost
	}

	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		// Parse port from env if needed
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.DatabaseConfig.Name = dbName
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.DatabaseConfig.Username = dbUser
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.DatabaseConfig.Password = dbPassword
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.SecurityConfig.JWTSecret = jwtSecret
	}

	// Create app instance
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	go func() {
		if err := application.Start(); err != nil {
			log.Fatalf("Failed to start application: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %s, shutting down gracefully...", sig)

	// Shutdown the application
	if err := application.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	} else {
		log.Println("Application shut down successfully")
	}
}