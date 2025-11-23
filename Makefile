# Citadel Agent Makefile
# Provides common commands for development and deployment

.PHONY: help build build-backend build-frontend test test-unit test-integration run run-backend run-frontend run-dev docker-up docker-down

# Show help message
help: ## Show this help
	@echo "Citadel Agent Makefile"
	@echo "======================"
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build the entire application
build: build-backend build-frontend ## Build the entire application

# Build backend
build-backend: ## Build the backend
	@echo "Building backend..."
	@cd backend && go build -o bin/citadel-api cmd/api/main.go
	@cd backend && go build -o bin/citadel-worker cmd/worker/main.go
	@echo "Backend built successfully!"

# Build frontend
build-frontend: ## Build the frontend
	@echo "Building frontend..."
	@cd frontend && npm run build
	@echo "Frontend built successfully!"

# Run tests
test: test-unit test-integration ## Run all tests
	@echo "All tests completed!"

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@cd backend && go test ./... -v

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@cd backend && go test ./... -tags=integration -v

# Run development servers
run: run-backend run-frontend ## Run both backend and frontend in development mode

run-backend: ## Run backend in development mode
	@echo "Starting backend development server..."
	@cd backend && go run cmd/api/main.go

run-frontend: ## Run frontend in development mode
	@echo "Starting frontend development server..."
	@cd frontend && npm run dev

run-dev: ## Run development environment with auto-reload
	@echo "Starting development environment..."
	@cd backend && air
	@cd frontend && npm run dev

# Docker commands
docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up --build

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

# Download AI models
download-models: ## Download AI models
	@echo "Downloading AI models..."
	@./scripts/download-models.sh

# Setup development environment
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Environment file copied. Please update values in .env file"
	@cd frontend && npm install
	@echo "Setup completed!"

# Database migrations
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go up

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@cd backend && go run cmd/migrate/main.go down

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf backend/bin/*
	@rm -rf frontend/dist/*
	@echo "Clean completed!"

# Install dependencies
deps: ## Install all dependencies
	@echo "Installing backend dependencies..."
	@cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Security scan
security-scan: ## Run security scan
	@echo "Running security scan..."
	@gosec ./backend/...
	@npm audit

# Generate documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@./scripts/generate-docs.sh