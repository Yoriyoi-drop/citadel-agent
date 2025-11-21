# Citadel Agent Makefile
# Simplifies common development tasks

.PHONY: help build test run clean docker-build docker-up docker-down

# Show help message
help:
	@echo "Citadel Agent Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build             - Build all services"
	@echo "  make test              - Run all tests"
	@echo "  make run-api           - Run API service"
	@echo "  make run-worker        - Run Worker service"
	@echo "  make run-scheduler     - Run Scheduler service"
	@echo "  make run-all           - Run all services with Docker Compose"
	@echo "  make docker-build      - Build Docker images"
	@echo "  make docker-up         - Start services with Docker Compose"
	@echo "  make docker-down       - Stop services with Docker Compose"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make lint              - Run linters"
	@echo "  make migrate           - Run database migrations"
	@echo "  make seed              - Seed database with initial data"
	@echo ""

# Build all services
build:
	@echo "Building Citadel Agent services..."
	@cd backend && go build -o ../bin/api cmd/api/main.go
	@cd backend && go build -o ../bin/worker cmd/worker/main.go
	@cd backend && go build -o ../bin/scheduler cmd/scheduler/main.go
	@echo "Build completed successfully"

# Run tests
test:
	@echo "Running tests..."
	@cd backend && go test ./... -v

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./backend/...

# Run API service
run-api:
	@echo "Starting API service..."
	@cd backend && go run cmd/api/main.go

# Run Worker service
run-worker:
	@echo "Starting Worker service..."
	@cd backend && go run cmd/worker/main.go

# Run Scheduler service
run-scheduler:
	@echo "Starting Scheduler service..."
	@cd backend && go run cmd/scheduler/main.go

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	@docker-compose -f docker-compose.yml build

# Start services with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose -f docker-compose.yml up -d

# Stop services with Docker Compose
docker-down:
	@echo "Stopping services with Docker Compose..."
	@docker-compose -f docker-compose.yml down

# Run all services
run-all: docker-up

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go

# Seed database with initial data
seed:
	@echo "Seeding database..."
	@cd backend && go run cmd/seed/main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@cd backend && go mod tidy
	@cd frontend && npm install