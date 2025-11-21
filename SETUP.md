# Citadel Agent Setup Guide

## Overview
This guide provides detailed instructions for setting up and configuring Citadel Agent after installation.

## Initial Configuration

### 1. Environment Setup
Before starting Citadel Agent, you need to configure environment variables:

1. Create a copy of the example environment file:
```bash
cp .env.example .env
```

2. Edit the `.env` file to match your configuration:
```bash
nano .env
```

### 2. Essential Configuration Variables

#### Security Settings
- `JWT_SECRET`: Generate a secure JWT secret (at least 32 characters)
  ```bash
  openssl rand -base64 32
  ```
- `SECURE_COOKIES`: Set to `true` for HTTPS environments
- `CORS_ORIGINS`: List allowed origins for CORS

#### Database Configuration  
- `DATABASE_URL`: PostgreSQL connection string
- `DB_SSL_MODE`: SSL mode for database connections (disable for development)

#### Redis Configuration
- `REDIS_URL`: Redis connection string
- `REDIS_PREFIX`: Prefix for Citadel Agent keys (optional)

#### Application Settings
- `ENVIRONMENT`: Set to `development`, `staging`, or `production`
- `PORT`: API server port (default: 5001)
- `API_RATE_LIMIT`: API rate limiting (requests per minute)
- `SESSION_TIMEOUT`: Session timeout in seconds

## Database Setup

### 1. PostgreSQL Configuration
Citadel Agent requires PostgreSQL version 12 or higher:

1. Install PostgreSQL:
   - Ubuntu: `sudo apt install postgresql postgresql-contrib`
   - CentOS: `sudo yum install postgresql-server postgresql-contrib`
   - macOS: `brew install postgresql`

2. Start the PostgreSQL service:
```bash
# Ubuntu/Debian
sudo systemctl start postgresql
sudo systemctl enable postgresql

# CentOS/RHEL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# macOS
brew services start postgresql
```

3. Create database and user:
```sql
CREATE DATABASE citadel_agent;
CREATE USER citadel_user WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE citadel_agent TO citadel_user;
```

4. Update your `.env` file:
```
DATABASE_URL=postgresql://citadel_user:your-secure-password@localhost:5432/citadel_agent
```

### 2. Run Database Migrations
If using the command-line tool:
```bash
cd backend
go run cmd/migrate/main.go
```

## Redis Setup

### 1. Install Redis
1. Install Redis:
   - Ubuntu: `sudo apt install redis-server`
   - CentOS: `sudo yum install redis`
   - macOS: `brew install redis`

2. Start Redis:
```bash
# Ubuntu/Debian
sudo systemctl start redis-server
sudo systemctl enable redis-server

# CentOS/RHEL
sudo systemctl start redis
sudo systemctl enable redis

# macOS
brew services start redis
```

3. Verify Redis is running:
```bash
redis-cli ping
# Should return PONG
```

## Backend Configuration

### 1. Service Configuration
Each backend service requires specific configuration:

#### API Service
Located at `cmd/api/main.go`
- Handles authentication
- Manages workflow creation and execution
- Provides user management endpoints

#### Worker Service
Located at `cmd/worker/main.go`
- Processes workflow executions
- Executes nodes
- Manages background jobs

#### Scheduler Service
Located at `cmd/scheduler/main.go`
- Manages scheduled workflows
- Handles cron-based executions
- Maintains execution queues

### 2. SSL Configuration
For production environments:

1. Obtain SSL certificates (e.g., from Let's Encrypt)
2. Configure SSL in your reverse proxy (Nginx, Apache)
3. Set `SECURE_COOKIES=true` in your `.env` file

## Frontend Configuration

### 1. Build Configuration
1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. For production builds:
```bash
npm run build
```

### 2. Environment Variables
Create a `.env.production` file for production settings:
```env
REACT_APP_API_URL=https://yourdomain.com/api/v1
REACT_APP_ENVIRONMENT=production
```

## Network Configuration

### 1. Firewall Rules
Configured ports for Citadel Agent:

- **5001**: Primary API endpoint
- **3000**: Frontend application (development)
- **5432**: PostgreSQL database
- **6379**: Redis cache
- **80/443**: Web server ports (when using reverse proxy)

### 2. Reverse Proxy Configuration
Example Nginx configuration:
```nginx
upstream citadel-api {
    server localhost:5001;
}

server {
    listen 80;
    server_name yourdomain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /path/to/your/cert.pem;
    ssl_certificate_key /path/to/your/key.pem;
    
    # API endpoints
    location /api {
        proxy_pass http://citadel-api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Serve frontend files
    location / {
        root /path/to/your/frontend/build;
        try_files $uri $uri/ /index.html;
    }
}
```

## Testing Configuration

### 1. Health Checks
Verify all components are running:

1. API health check:
```bash
curl http://localhost:5001/api/v1/health
```

2. Database connectivity:
```bash
# Check if you can connect to PostgreSQL
psql $DATABASE_URL -c "SELECT version();"
```

3. Redis connectivity:
```bash
redis-cli -u $REDIS_URL ping
```

### 2. Initial User Setup
1. Access the web interface
2. Register a new account or use existing admin credentials
3. Verify admin panel access

## Security Configuration

### 1. Authentication Settings
- Enable Two-Factor Authentication (2FA) for admin accounts
- Set up Single Sign-On (SSO) if needed
- Configure role-based access controls (RBAC)

### 2. Network Security
- Use VPN for internal access
- Restrict API access by IP when possible
- Implement rate limiting at network level

## Performance Tuning

### 1. Database Optimization
- Enable PostgreSQL connection pooling
- Configure appropriate memory settings
- Set up database connection limits

### 2. Application Settings
- Adjust worker pool sizes based on available resources
- Tune timeout settings for long-running workflows
- Configure appropriate logging levels for production

## Monitoring Setup

### 1. Logging Configuration
Ensure proper logging is set up:
- Centralized log aggregation
- Alerting for critical errors
- Log rotation settings

### 2. System Metrics
Consider setting up monitoring with:
- Prometheus and Grafana
- ELK stack (Elasticsearch, Logstash, Kibana)
- Application Performance Monitoring (APM) tool

## Backup and Recovery

### 1. Database Backup
Set up regular PostgreSQL backups:
```bash
pg_dump citadel_agent > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. Configuration Backup
Backup your configuration files:
- `.env` file (store securely)
- Database schema exports
- SSL certificates (if applicable)

---

**Citadel Agent v0.1.0** - Advanced Workflow Automation Platform