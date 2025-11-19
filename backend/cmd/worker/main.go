package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"citadel-agent/backend/internal/config"
	"citadel-agent/backend/internal/database"
	"citadel-agent/backend/internal/engine"
	"citadel-agent/backend/internal/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize services
	executionService := services.NewExecutionService(db.GormDB)

	// Initialize workflow engine
	nodeRegistry := engine.NewNodeRegistry()
	executor := engine.NewExecutor()
	runner := engine.NewRunner(executor)

	// Create worker
	worker := services.NewWorker(executionService, runner, nodeRegistry)

	// Create context that is cancelled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Fatal("Worker failed to start:", err)
		}
	}()

	log.Println("Worker started successfully")

	// Wait for interrupt signal to gracefully shutdown the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	
	// Create context with timeout for graceful shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, 30*time.Second)
	defer cancelShutdown()

	if err := worker.Stop(shutdownCtx); err != nil {
		log.Printf("Error during worker shutdown: %v", err)
	}

	log.Println("Worker stopped")
}