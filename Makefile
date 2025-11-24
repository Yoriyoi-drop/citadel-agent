# Citadel Agent Makefile

# Build the backend application
.PHONY: build-backend
build-backend:
	cd backend && go build -o citadel-api main.go

# Run the backend application
.PHONY: run-backend
run-backend:
	cd backend && go run main.go

# Run the backend in development mode with live reload (requires air: https://github.com/cosmtrek/air)
.PHONY: dev-backend
dev-backend:
	cd backend && air

# Install backend dependencies
.PHONY: deps-backend
deps-backend:
	cd backend && go mod tidy

# Run backend tests
.PHONY: test-backend
test-backend:
	cd backend && go test ./... -v

# Run backend tests with coverage
.PHONY: test-backend-coverage
test-backend-coverage:
	cd backend && go test ./... -v -coverprofile=coverage.out && go tool cover -html=coverage.out

# Build the frontend application
.PHONY: build-frontend
build-frontend:
	cd frontend && npm install && npm run build

# Run the frontend development server
.PHONY: dev-frontend
dev-frontend:
	cd frontend && npm install && npm run dev

# Run frontend tests
.PHONY: test-frontend
test-frontend:
	cd frontend && npm run test

# Run all tests
.PHONY: test
test: test-backend

# Run the full application with docker-compose
.PHONY: up
up:
	docker-compose up

# Run the full application in detached mode
.PHONY: up-detached
up-detached:
	docker-compose up -d

# Stop the application
.PHONY: down
down:
	docker-compose down

# Reset the environment (stop, remove volumes, start)
.PHONY: reset
reset:
	docker-compose down -v
	docker-compose up

# Run linter on backend
.PHONY: lint-backend
lint-backend:
	cd backend && golangci-lint run

# Create database migrations
.PHONY: migrate-up
migrate-up:
	# Add your migration command here (e.g., migrate -path configs/db -database postgres://user:pass@localhost/db up)

# Create initial database schema
.PHONY: schema-create
schema-create:
	# Add your schema creation command here

# Run security scan on the project
.PHONY: security-scan
security-scan:
	# Add security scanning commands (e.g., go run github.com/securego/gosec/v2/cmd/gosec@latest ./...)

# Generate documentation
.PHONY: docs
docs:
	# Add documentation generation commands

# Install all dependencies
.PHONY: setup
setup: deps-backend
	cd frontend && npm install

# Run all services in development mode
.PHONY: dev
dev: setup up-detached
	# Run frontend in dev mode in the background
	# This requires the services to be running via docker-compose
	@echo "Services are running. Start frontend with: cd frontend && npm run dev"