package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/citadel-agent/backend/internal/api"
	"github.com/citadel-agent/backend/internal/startup"
	"github.com/citadel-agent/backend/internal/temporal"
	"github.com/citadel-agent/backend/internal/plugins"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

func main() {
	// Initialize base engine
	baseEngine := engine.NewEngine(&engine.Config{
		// Logger and Storage need to be configured in a real implementation
		Parallelism: 10,
	})

	// Initialize plugin manager
	pluginManager := plugins.NewNodeManager()

	// Initialize Temporal client
	temporalConfig := temporal.GetDefaultConfig()
	temporalConfig.Address = getEnv("TEMPORAL_ADDRESS", "localhost:7233")
	temporalConfig.Namespace = getEnv("TEMPORAL_NAMESPACE", "default")

	temporalClient, err := temporal.NewTemporalClient(&temporal.Config{
		Address:   temporalConfig.Address,
		Namespace: temporalConfig.Namespace,
	})
	if err != nil {
		log.Fatal("Failed to create Temporal client:", err)
	}

	// Initialize Temporal workflow service
	workflowService := temporal.NewTemporalWorkflowService(temporalClient, baseEngine)

	// Register node types for compatibility
	workflowService.RegisterNodeTypes()

	// Create API server
	serverConfig := api.GetDefaultConfig()
	server := api.NewServer(workflowService, pluginManager, baseEngine, serverConfig)

	// Check if auto-open browser is enabled
	autoOpen := getEnv("AUTO_OPEN_BROWSER", "true") == "true"

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Open browser automatically if enabled (in a separate goroutine to not block)
	if autoOpen {
		go func() {
			startup.WaitForServer(serverConfig.Port, 10*time.Second)
		}()
	}

	log.Printf("🚀 Citadel Agent API server started on http://localhost:%s", serverConfig.Port)
	if autoOpen {
		log.Println("🌐 Browser will open automatically in a few seconds...")
	} else {
		log.Println("💡 AUTO_OPEN_BROWSER is disabled. Start manually at: http://localhost:" + serverConfig.Port)
	}
	log.Println("Press Ctrl+C to stop.")

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Gracefully shutdown the server
	if err := server.Shutdown(); err != nil {
		log.Fatal("Error during server shutdown:", err)
	}

	log.Println("Server stopped")
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}