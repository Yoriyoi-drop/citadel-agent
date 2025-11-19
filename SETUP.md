# Citadel Agent - Setup Guide

## Prerequisites

### System Requirements
- **Operating System**: Linux, macOS, or Windows with WSL2
- **Docker**: Version 20.10+ with Docker Compose
- **Memory**: Minimum 4GB RAM (8GB+ recommended)
- **Storage**: Minimum 2GB free space
- **CPU**: Modern processor with virtualization support

### Required Software
1. **Docker Desktop** or **Docker Engine** with Compose plugin
2. **Git** (for cloning the repository)
3. **Terminal/shell access**

## Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent
```

### 2. Configure Environment

Copy the sample environment file and customize as needed:
```bash
cp .env.example .env
# Edit .env to adjust configurations if needed
vi .env  # or use your preferred editor
```

### 3. Start the Services
```bash
# Start all services (recommended for first time)
./scripts/start.sh
```

### 4. Verify Installation
Open your browser and navigate to:
- **API Health**: [http://localhost:5001/health](http://localhost:5001/health)
- **API Documentation**: [http://localhost:5001](http://localhost:5001)

## Manual Installation

### Start Individual Services
If you prefer to start services individually:

```bash
# Start database services only
docker-compose -f docker/docker-compose.yml up postgres redis -d

# Start API first to initialize database
docker-compose -f docker/docker-compose.yml up api -d

# Then start worker and scheduler
docker-compose -f docker/docker-compose.yml up worker scheduler -d
```

### Check Service Status
```bash
# View all running services
./scripts/status.sh

# View detailed logs
docker-compose -f docker/docker-compose.yml logs -f
```

## Configuration Options

### Environment Variables
Edit the `.env` file to customize your installation:

```
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=citadel_agent

# API Server
SERVER_PORT=5001

# Security (change these for production!)
JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
JWT_EXPIRY=86400

# Redis Configuration
REDIS_ADDR=localhost:6379
```

### Scaling Services
You can scale individual services as needed:

```bash
# Scale worker service to 3 instances
docker-compose -f docker/docker-compose.yml up -d --scale worker=3

# Scale scheduler service to 2 instances  
docker-compose -f docker/docker-compose.yml up -d --scale scheduler=2
```

## Service Architecture

The Citadel Agent consists of several services:

### Core Services
- **API Service** (`api`): Handles REST API requests, workflow management, and user authentication
- **Worker Service** (`worker`): Executes workflow nodes in isolated environments
- **Scheduler Service** (`scheduler`): Manages scheduled workflows and triggers

### Supporting Services
- **PostgreSQL** (`postgres`): Main database for workflows, nodes, and executions
- **Redis** (`redis`): Caching, sessions, and job queues

### Ports Used
- **5001**: API Service (adjustable via SERVER_PORT in .env)
- **5432**: PostgreSQL Database
- **6379**: Redis

## Production Setup

### Security Recommendations
1. **Change Default Credentials**: Update passwords in `.env` file
2. **Enable SSL/HTTPS**: Use a reverse proxy (Nginx) with SSL termination
3. **Secure JWT Secret**: Use a strong, randomly generated JWT secret
4. **Network Isolation**: Restrict database access to internal network only

### SSL Setup
For production environments, set up SSL using Nginx:

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;

    location / {
        proxy_pass http://localhost:5001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Backup Strategy
Regular backups are essential for production:

```bash
# Backup PostgreSQL database
docker-compose -f docker/docker-compose.yml exec postgres pg_dump -U postgres citadel_agent > backup.sql

# Backup Redis data (if persistence is enabled)
docker-compose -f docker/docker-compose.yml exec redis redis-cli BGSAVE
```

## Monitoring and Maintenance

### Service Health
Monitor service health with:
```bash
# Check service status
./scripts/status.sh

# Monitor logs continuously
docker-compose -f docker/docker-compose.yml logs -f --tail=100

# Check resource usage
docker stats
```

### Performance Tuning
#### PostgreSQL
- Adjust shared_buffers (typically 25% of available RAM)
- Tune effective_cache_size (typically 50-75% of available RAM)
- Configure connection pooling

#### Redis
- Set appropriate maxmemory limits
- Choose appropriate eviction policies
- Enable persistence if needed

### Update Process
To update to a newer version:

```bash
# Stop current services
./scripts/stop.sh

# Pull latest changes
git pull origin main

# Pull latest Docker images
docker-compose -f docker/docker-compose.yml pull

# Start services
./scripts/start.sh
```

## Troubleshooting

Refer to the [TROUBLESHOOTING.md](TROUBLESHOOTING.md) file for common issues and solutions.

## Development Setup

For development, use the development compose file:
```bash
# Start with hot-reloading enabled
docker-compose -f docker/docker-compose.dev.yml up --build
```

### Running Tests
```bash
# Run backend tests
cd backend && go test ./...

# Run specific service tests
cd backend && go test ./internal/engine/... -v
```

## Uninstalling

To completely remove Citadel Agent:

```bash
# Stop all services
./scripts/stop.sh

# Remove all containers, networks, and volumes (data will be lost!)
docker-compose -f docker/docker-compose.yml down -v

# Remove Docker images (optional)
docker rmi $(docker images "citadel-*" -q)

# Clean up system
docker system prune -a
```

## Support

- **Documentation**: [https://citadel-agent.com/docs](https://citadel-agent.com/docs)
- **Issues**: [GitHub Issues](https://github.com/citadel-agent/citadel-agent/issues)
- **Community**: [Join our Discord](https://discord.gg/citadel-agent) (if available)

---

**Note**: This software is under active development. For best results, always use the latest stable release.