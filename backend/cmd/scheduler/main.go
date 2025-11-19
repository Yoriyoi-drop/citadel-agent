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
	workflowService := services.NewWorkflowService(db.GormDB)
	executionService := services.NewExecutionService(db.GormDB)

	// Initialize workflow engine
	executor := engine.NewExecutor()
	runner := engine.NewRunner(executor)

	// Initialize scheduler
	scheduler := engine.NewScheduler(runner)

	// Create worker
	schedulerService := services.NewSchedulerService(workflowService, executionService, scheduler)

	// Create context that is cancelled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the scheduler
	go func() {
		if err := schedulerService.Start(ctx); err != nil {
			log.Fatal("Scheduler failed to start:", err)
		}
	}()

	log.Println("Scheduler started successfully")

	// Wait for interrupt signal to gracefully shutdown the scheduler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down scheduler...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, 30*time.Second)
	defer cancelShutdown()

	if err := schedulerService.Stop(shutdownCtx); err != nil {
		log.Printf("Error during scheduler shutdown: %v", err)
	}

	log.Println("Scheduler stopped")
}