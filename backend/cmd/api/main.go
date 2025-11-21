package main

import (
	"log"
	"os"

	"github.com/citadel-agent/backend/internal/api"
	"github.com/citadel-agent/backend/internal/auth"
	"github.com/citadel-agent/backend/internal/ai"
	"github.com/citadel-agent/backend/internal/runtimes"
	"github.com/citadel-agent/backend/internal/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection
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

	dbPool, err := pgxpool.New(
		// context.Background(),
		dbURL,
	)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbPool.Close()

	// Initialize services
	authService := auth.NewAuthService(dbPool)
	aiService := ai.NewAIService()
	runtimeMgr := runtimes.NewMultiRuntimeManager()
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor(nodeRegistry)
	runner := engine.NewRunner(executor)

	// Create API server
	server := api.NewServer(
		dbPool,
		authService,
		aiService,
		nodeRegistry,
		executor,
		runner,
		runtimeMgr,
	)

	// Get port from environment or use default
	port := getEnvOrDefault("SERVER_PORT", "5001")

	// Start the server
	log.Printf("Starting Citadel Agent API server on port %s", port)
	if err := server.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}