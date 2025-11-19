# Citadel Agent - Complete Automation Platform

Citadel Agent is a powerful workflow automation platform designed to handle complex systems with 200+ built-in nodes, enterprise security, and cloud-native scalability. It's built as a more modern, faster, and lighter alternative to n8n.

## 🏗️ System Architecture

The Citadel Agent platform consists of:

- **Backend Services**: Go-based microservices (API, Worker, Scheduler)
- **CLI Tool**: Node.js package for easy installation and management
- **Frontend**: React-based workflow editor (coming soon)
- **Database**: PostgreSQL for data persistence
- **Cache**: Redis for sessions and job queues

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Git
- Node.js v16+ (for CLI management)

### Installation

#### Option 1: Using CLI (Recommended)
```bash
# Install the CLI globally
npm install -g @citadel-agent/cli

# Install Citadel Agent
citadel install

# Start all services
citadel start
```

#### Option 2: From Source
```bash
# Clone the repository
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent

# Copy environment file
cp .env.example .env

# Start all services with proper startup order and health checks
./scripts/start.sh
```

### Access Services
- **API**: [http://localhost:5001](http://localhost:5001)
- **Health Check**: [http://localhost:5001/health](http://localhost:5001/health)

### Useful Scripts
- `./scripts/start.sh` - Start all services with health checks and dependency management
- `./scripts/stop.sh` - Stop all services cleanly
- `./scripts/status.sh` - Check the status of all services

## 🛠️ Technology Stack

### Backend Services
- **Go**: High-performance backend services
- **Fiber**: Web framework for API service
- **GORM**: Database ORM
- **Golang JWT**: Authentication and authorization

### Deployment
- **Docker/Docker Compose**: Container orchestration
- **PostgreSQL**: Relational database
- **Redis**: Cache and job queues

### Infrastructure
- **Health Checks**: Built-in service health monitoring
- **Startup Dependencies**: Proper service initialization order
- **Resource Management**: Optimized container resource usage

## 🧩 Core Components

### API Service
- REST API for workflow management
- Authentication and user management
- Real-time workflow monitoring

### Worker Service
- Isolated execution of workflow nodes
- Resource limiting per workflow
- Timeout enforcement

### Scheduler Service
- Cron-based workflow scheduling
- Event-driven triggers
- Time-based execution

## 🔐 Security Features

- **Service Isolation**: Each service runs in separate containers
- **Database Security**: Parameterized queries and connection pooling
- **Authentication**: JWT-based authentication
- **Resource Limits**: Container-based resource limiting

## 📊 Key Features

- **Foundation Engine**: Robust workflow execution with dependency resolution
- **200+ Built-in Nodes**: Comprehensive automation capabilities
- **Enterprise Security**: Node sandboxing, RBAC, audit logging
- **High Performance**: Optimized for speed and scalability
- **Real-time Monitoring**: WebSocket support for live updates
- **Extensible Architecture**: Plugin system and custom nodes
- **Cloud-Native**: Kubernetes-ready deployment

## 🤝 Management & Operations

### Service Management
- **Health Checks**: Comprehensive service health monitoring
- **Startup Dependencies**: Ensures services start in correct order
- **Resource Monitoring**: Built-in resource usage tracking
- **Logging**: Structured logging across all services

### Development
- **Hot Reloading**: Development mode with auto-restart
- **Docker Compose Dev**: Separate development environment
- **Local Testing**: Easy local development setup

## 🚢 Production Deployment

For production deployments, refer to:
- `SETUP.md` - Complete production setup guide
- `TROUBLESHOOTING.md` - Common issues and solutions
- `docker/compose.prod.yml` - Production configuration

## 📚 Documentation

- [Setup Guide](SETUP.md) - Complete installation and configuration
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions
- [Architecture](DOCS.md) - Detailed technical architecture
- [API Documentation](docs/api.md) - REST API specification

## 🤝 Contributing

We welcome contributions! Please see our contributing guide for more details.

## 📄 License

Apache 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Note**: This platform is actively under development. For the latest updates, check our GitHub repository.