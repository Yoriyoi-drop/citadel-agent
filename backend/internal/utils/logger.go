package utils

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Logger utility functions
var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "citadel-agent: ", log.LstdFlags|log.Lshortfile)
}

// SetupFiberLogger sets up the Fiber logger middleware
func SetupFiberLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip} | ${bytesSent}\n",
	}))
}

// LogInfo logs an informational message
func LogInfo(message string) {
	Logger.Printf("INFO: %s", message)
}

// LogError logs an error message
func LogError(message string, err error) {
	Logger.Printf("ERROR: %s - %v", message, err)
}

// LogDebug logs a debug message
func LogDebug(message string) {
	Logger.Printf("DEBUG: %s", message)
}