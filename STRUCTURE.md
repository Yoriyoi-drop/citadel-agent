# Citadel Agent - Project Structure

## Overview
Citadel Agent adalah platform otomasi workflow modern dengan kemampuan AI agent, multi-language runtime, dan sandboxing keamanan lanjutan. Repository ini mengikuti struktur standar Go dengan beberapa modifikasi untuk menyesuaikan kebutuhan microservices.

## Directory Structure

```
citadel-agent/
├── backend/                 # Go backend services
│   ├── cmd/                # Main applications
│   │   ├── api/            # API service entry point
│   │   ├── worker/         # Worker service entry point
│   │   ├── scheduler/      # Scheduler service entry point
│   │   ├── migrate/        # Database migration tool
│   │   └── seed/           # Database seeding tool
│   ├── internal/           # Internal application code
│   │   ├── api/            # API layer (handlers, middleware)
│   │   ├── auth/           # Authentication & authorization
│   │   ├── ai/             # AI agent functionality
│   │   ├── engine/         # Workflow engine core
│   │   ├── runtimes/       # Multi-language runtime support
│   │   ├── database/       # Database connection & utilities
│   │   ├── nodes/          # Node types and implementations
│   │   └── interfaces/     # Interface definitions
│   ├── pkg/                # Shared libraries (if any)
│   ├── test/               # Test files
│   │   ├── unit/           # Unit tests
│   │   ├── integration/    # Integration tests
│   │   └── e2e/            # End-to-end tests
│   ├── go.mod             # Go module definition
│   └── go.sum             # Go module checksums
├── frontend/               # React frontend application
│   ├── src/               # Source code
│   ├── public/            # Static assets
│   ├── package.json       # Node.js dependencies
│   └── ...
├── docker/                 # Docker configurations
│   ├── api.Dockerfile     # API service Dockerfile
│   ├── worker.Dockerfile  # Worker service Dockerfile
│   └── scheduler.Dockerfile # Scheduler service Dockerfile
├── docs/                   # Documentation
│   ├── api/               # API documentation
│   ├── architecture/      # System architecture docs
│   ├── guides/            # User guides
│   └── nodes/             # Node documentation
├── scripts/                # Utility scripts
├── configs/                # Configuration files
├── examples/               # Example configurations and workflows
├── services/               # Service-specific configurations
├── storage/                # Storage-related configurations
├── sdk/                    # Software Development Kit
├── packages/               # Shared packages (if using monorepo approach)
├── plugins/                # Plugin system
├── tests/                  # Top-level tests (if needed)
├── .github/                # GitHub configurations
│   └── workflows/         # GitHub Actions workflows
├── .dockerignore          # Docker ignore file
├── .editorconfig          # Editor configuration
├── .gitignore             # Git ignore file
├── docker-compose.yml     # Development Docker Compose
├── docker-compose.prod.yml # Production Docker Compose
├── Makefile               # Make commands for common tasks
├── README.md              # Main project documentation
├── INSTALL.md             # Installation instructions
├── SETUP.md               # Setup instructions
├── DEPLOY_GUIDE.md        # Deployment guide
├── .env.example           # Environment variables example
└── ...
```

## Services

### 1. API Service (`/backend/cmd/api`)
- **Purpose**: REST API server handling authentication, workflow CRUD, and user management
- **Framework**: Fiber
- **Port**: 5001

### 2. Worker Service (`/backend/cmd/worker`)
- **Purpose**: Executes workflow nodes and processes tasks
- **Features**: Runtime isolation, resource management, task queue processing

### 3. Scheduler Service (`/backend/cmd/scheduler`)
- **Purpose**: Manages scheduled workflows and time-based triggers
- **Features**: Cron scheduling, event scheduling, workflow triggers

## Development Guidelines

### Go Code Standards
- Use `gofmt` for code formatting
- Write comprehensive unit tests
- Follow Go naming conventions
- Document exported functions/types

### Git Workflow
- Use feature branches for new features
- Follow conventional commits
- Write meaningful commit messages
- Submit pull requests for code review

### Docker Images
- Use multi-stage builds for smaller images
- Use specific version tags instead of `latest`
- Include security scanning in CI