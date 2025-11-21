# Citadel Agent Installation Guide

## Overview
This guide explains how to install and set up Citadel Agent on your system.

## Prerequisites
Before installing Citadel Agent, ensure your system meets the following requirements:

- Operating System: Linux (Ubuntu 20.04+, CentOS 8+), macOS 10.15+, or Windows 10+
- Memory: Minimum 4GB RAM (8GB+ recommended)
- Storage: Minimum 10GB free space
- Go: Version 1.19 or higher
- Node.js: Version 16 or higher
- Docker: Version 20.10 or higher (optional but recommended)
- Docker Compose: Version 2.0 or higher (optional but recommended)
- Git: Version 2.20 or higher

## Installation Methods

### Method 1: Quick Install (Recommended)
For users who want to quickly get Citadel Agent up and running:

1. Download the installation script:
```bash
curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/install.sh
```

2. Make the script executable:
```bash
chmod +x install.sh
```

3. Run the installer:
```bash
./install.sh
```

4. Follow the prompts to select your installation options:
   - Choose between development and production setup
   - Configure database settings
   - Set up security parameters

### Method 2: Manual Installation
For users who prefer to install components manually:

#### Backend Installation
1. Clone the repository:
```bash
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent/backend
```

2. Install Go dependencies:
```bash
go mod tidy
```

3. Build the backend services:
```bash
go build -o ../bin/api cmd/api/main.go
go build -o ../bin/worker cmd/worker/main.go  
go build -o ../bin/scheduler cmd/scheduler/main.go
```

#### Frontend Installation
1. Navigate to the frontend directory:
```bash
cd ../frontend
```

2. Install Node.js dependencies:
```bash
npm install
```

3. Build the frontend (for production):
```bash
npm run build
```

## Configuration

### Environment Variables
Create a `.env` file in the root directory and configure the following variables:

```env
# JWT Secret (change to a secure value)
JWT_SECRET=your-super-secret-jwt-key-here-at-least-32-characters-for-production

# Database configuration
DATABASE_URL=postgresql://postgres:password@localhost:5432/citadel_agent

# Redis configuration
REDIS_URL=redis://localhost:6379

# Application configuration
ENVIRONMENT=production
PORT=5001
API_RATE_LIMIT=1000
SESSION_TIMEOUT=86400

# Security settings
SECURE_COOKIES=true
CORS_ORIGINS=https://yourdomain.com
```

### Database Setup
Citadel Agent uses PostgreSQL as its primary database. Set up the database:

1. Create a PostgreSQL database named `citadel_agent`
2. Create a user with appropriate permissions
3. Update the `DATABASE_URL` in your `.env` file

### Running the Application

#### Development Mode
To run in development mode with hot reloading:

```bash
# Backend
cd backend
go run cmd/api/main.go

# Frontend in another terminal
cd frontend
npm run dev
```

#### Production Mode
To run in production mode:

##### Using Docker Compose
```bash
docker-compose -f docker/compose/docker-compose.prod.yml up -d
```

##### Using Built Binaries
```bash
# Backend services
./bin/api
./bin/worker
./bin/scheduler

# Serve frontend files
# Configure your web server (Nginx, Apache) to serve frontend/build/
```

## Post-Installation Steps

1. **Configure SSL/TLS**: Set up SSL certificates for production
2. **Set up Firewall**: Restrict access to necessary ports only
3. **Configure Backups**: Set up regular database and application backups
4. **Review Security Settings**: Verify all security configurations
5. **Create Admin User**: Set up your first administrator account
6. **Test Installation**: Verify all components are working correctly

## Troubleshooting

### Common Issues

#### Database Connection Issues
- Ensure PostgreSQL is running and accessible
- Verify the `DATABASE_URL` in your `.env` file
- Check database permissions

#### Frontend Build Issues
- Ensure Node.js and npm are properly installed
- Run `npm install` again if dependencies are missing
- Check available disk space

#### Port Conflicts
- Modify the `PORT` variable in `.env` if 5001 is in use
- Check for conflicting services

## Support

For additional help with installation:
- Check our [documentation](https://citadel-agent.com/docs)
- Join our [community forum](https://community.citadel-agent.com)
- Contact support at [support@citadel-agent.com](mailto:support@citadel-agent.com)

---

**Citadel Agent v0.1.0** - Advanced Workflow Automation Platform