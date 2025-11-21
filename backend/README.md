# Citadel Agent Backend

The backend of Citadel Agent is built with Go and provides the core services for workflow automation, AI agents, and multi-language runtime execution.

## Architecture

The backend follows a microservices architecture with three main services:

1. **API Service** - Handles HTTP requests, authentication, and business logic
2. **Worker Service** - Executes workflow nodes in isolated environments
3. **Scheduler Service** - Manages scheduled workflows and event triggers

## Project Structure

```
backend/
├── cmd/                 # Main applications
│   ├── api/            # API service entry point
│   ├── worker/         # Worker service entry point
│   ├── scheduler/      # Scheduler service entry point
│   ├── migrate/        # Database migration tool
│   └── seed/           # Database seeding tool
├── internal/           # Internal application code
│   ├── api/            # API handlers and middleware
│   ├── auth/           # Authentication & authorization
│   ├── ai/             # AI agent functionality
│   ├── engine/         # Workflow engine core
│   ├── runtimes/       # Multi-language runtime support
│   ├── database/       # Database utilities
│   └── nodes/          # Node types and implementations
├── test/               # Test files
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
└── go.work            # Go workspace file (if using workspaces)
```

## Services

### API Service
The API service provides REST endpoints for:
- User authentication and management
- Workflow creation, retrieval, update, and deletion
- Node management and execution
- AI agent management
- System monitoring and health checks

### Worker Service
The worker service handles:
- Execution of workflow nodes
- Isolation of code execution in secure environments
- Resource management and monitoring
- Task queue processing
- Error handling and retry mechanisms

### Scheduler Service
The scheduler service manages:
- Cron-based workflow execution
- Event-driven triggers
- Time-based tasks
- Workflow execution scheduling

## Development

### Prerequisites
- Go 1.21 or higher
- PostgreSQL database
- Redis server

### Setup
1. Clone the repository
2. Navigate to the backend directory: `cd backend`
3. Install dependencies: `go mod tidy`
4. Set up environment variables (see `.env.example` in root)
5. Run database migrations: `go run cmd/migrate/main.go`
6. Start services (see main README)

### Running Services
```bash
# API Service
go run cmd/api/main.go

# Worker Service
go run cmd/worker/main.go

# Scheduler Service
go run cmd/scheduler/main.go

# Database Migrations
go run cmd/migrate/main.go

# Database Seeding
go run cmd/seed/main.go
```

## Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/api/...
```

## Building
```bash
# Build API service
go build -o bin/api cmd/api/main.go

# Build Worker service
go build -o bin/worker cmd/worker/main.go

# Build Scheduler service
go build -o bin/scheduler cmd/scheduler/main.go

# Build all services
make build  # Using the Makefile in root
```