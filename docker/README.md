# Docker Configuration

This directory contains Docker configurations for Citadel Agent services.

## Files

- `api.Dockerfile` - Production Dockerfile for API service
- `api-dev.Dockerfile` - Development Dockerfile for API service with hot reload
- `worker.Dockerfile` - Production Dockerfile for Worker service
- `worker-dev.Dockerfile` - Development Dockerfile for Worker service
- `scheduler.Dockerfile` - Production Dockerfile for Scheduler service
- `scheduler-dev.Dockerfile` - Development Dockerfile for Scheduler service

## Usage

### Development
```bash
# Start all services in development mode with hot reload
docker-compose -f docker-compose.yml -f docker-compose.override.yml up
```

### Production
```bash
# Build and start all services in production mode
docker-compose -f docker-compose.yml up -d
```

### Individual Service
```bash
# Build just the API service
docker build -f docker/api.Dockerfile -t citadel-agent-api .

# Build just the Worker service
docker build -f docker/worker.Dockerfile -t citadel-agent-worker .

# Build just the Scheduler service
docker build -f docker/scheduler.Dockerfile -t citadel-agent-scheduler .
```

## Multi-stage Builds

All Dockerfiles use multi-stage builds to:
- Keep final images small
- Separate build dependencies from runtime dependencies
- Improve security by minimizing attack surface