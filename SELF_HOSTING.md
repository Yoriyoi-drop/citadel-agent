# Citadel Agent - Self-Hosting Guide

## Table of Contents
1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Quick Deployment](#quick-deployment)
4. [Manual Setup](#manual-setup)
5. [Server Management](#server-management)
6. [SSL Configuration](#ssl-configuration)
7. [Security Best Practices](#security-best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Backup and Restore](#backup-and-restore)
10. [Updates](#updates)

## Overview

Citadel Agent can be self-hosted on your own infrastructure. This guide provides comprehensive instructions for deploying, managing, and maintaining your own Citadel Agent instance.

## Prerequisites

### Server Requirements
- **Operating System**: Ubuntu 20.04 LTS or later, Debian 11+, CentOS 8+ or RHEL 8+
- **RAM**: Minimum 4GB (8GB+ recommended)
- **Storage**: Minimum 10GB free space
- **CPU**: 2+ cores recommended
- **Network**: Public IP with ports 22 (SSH), 80 (HTTP), 443 (HTTPS), and 5001 (API) accessible

### Local Requirements (for deployment)
- **SSH Access**: To your server with sudo privileges
- **SCP and SSH**: Available on your local machine
- **Git**: Installed locally (for deployment script)

## Quick Deployment

The fastest way to deploy Citadel Agent is using our automated deployment script.

### 1. Prepare Your Server

Your server should meet the requirements above and have SSH access enabled.

### 2. Run the Deployment Script

```bash
# Download the deployment script
curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/server/deploy.sh
chmod +x deploy.sh

# Deploy to your server
./deploy.sh -u your-username -h your-server-ip
```

**Options:**
- `-u, --user`: Username for SSH access to your server
- `-h, --host`: Server IP address or hostname
- `-p, --port`: SSH port (default: 22)
- `-P, --path`: Remote installation path (default: /opt/citadel-agent)
- `-e, --env-file`: Local .env file to copy to server
- `--skip-setup`: Skip server setup if already configured
- `--with-ssl`: Include SSL setup instructions
- `-y, --yes`: Skip confirmation prompts

### 3. Example Deployment Command

```bash
# Deploy with custom port and environment file
./deploy.sh -u admin -h 203.0.113.10 -p 2222 -e /path/to/my-env-file -y

# Deploy with SSL setup
./deploy.sh -u ubuntu -h your-server.com --with-ssl -y
```

## Manual Setup

If you prefer to set up manually, follow these steps:

### 1. Server Preparation (Run on your server)

```bash
# SSH into your server
ssh your-username@your-server-ip

# Download and run setup script
curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/server/setup.sh
chmod +x setup.sh
sudo ./setup.sh
```

### 2. After Setup (on your server)

```bash
# Log out and back in to get Docker permissions
logout
ssh your-username@your-server-ip

# Navigate to Citadel Agent
cd ~/citadel-agent

# Customize your .env file
nano .env

# Start the services
./scripts/start.sh

# Check status
./scripts/status.sh
```

## Server Management

Once deployed, you can manage your Citadel Agent instance using the management script.

### 1. Download Management Script

```bash
# On your server
cd /opt/citadel-agent

# Or if you need it locally
curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/server/manage.sh
chmod +x manage.sh
```

### 2. Management Commands

```bash
# Check service status
./server/manage.sh status

# Start services
./server/manage.sh start

# Stop services
./server/manage.sh stop

# Restart services
./server/manage.sh restart

# View logs
./server/manage.sh logs            # All services
./server/manage.sh logs api        # API service only
./server/manage.sh logs worker     # Worker service only

# Update to latest version
./server/manage.sh update

# Create backup
./server/manage.sh backup

# Database backup only
./server/manage.sh backup-db

# View system resources
./server/manage.sh monitor

# Show configuration
./server/manage.sh config

# Clean up Docker resources
./server/manage.sh cleanup
```

## SSL Configuration

### 1. Using the Management Script (Recommended)

```bash
# Setup SSL for your domain
./server/manage.sh ssl yourdomain.com

# Example
./server/manage.sh ssl app.mycompany.com
```

### 2. Manual SSL Setup

```bash
# On your server
sudo certbot --nginx -d yourdomain.com

# Or with specific Nginx configuration
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

### 3. SSL Configuration File

The setup script creates an Nginx configuration at `/etc/nginx/sites-available/citadel-agent`. You'll need to update the server_name directive:

```nginx
server {
    listen 80;
    server_name yourdomain.com;  # Update this line
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;  # Update this line
    
    # SSL certificates managed by Certbot
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    
    # ... rest of configuration
}
```

After updating, reload Nginx:
```bash
sudo nginx -t  # Test configuration
sudo systemctl reload nginx  # Reload
```

## Security Best Practices

### 1. Environment Security

Update your `.env` file with secure values:

```env
# Generate a strong JWT secret (at least 32 chars)
JWT_SECRET=your-very-long-random-string-at-least-32-characters-for-production

# Change default database credentials
DB_USER=citadel_user
DB_PASSWORD=your-very-secure-password-here

# Set environment to production
ENVIRONMENT=production
```

### 2. Firewall Configuration

The setup script configures UFW firewall. Verify it's running:

```bash
# Check firewall status
sudo ufw status

# Allow only necessary ports
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw allow 5001/tcp  # If you need direct API access
sudo ufw --force enable
```

### 3. SSH Security

Secure your SSH access:

```bash
# Disable root login
sudo nano /etc/ssh/sshd_config
# Set: PermitRootLogin no
# Set: PasswordAuthentication no (if using SSH keys)

# Restart SSH
sudo systemctl restart sshd
```

### 4. Regular Updates

Keep your system and Citadel Agent updated:

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y  # Ubuntu/Debian
# Or
sudo yum update -y  # CentOS/RHEL

# Update Citadel Agent
cd /opt/citadel-agent
./server/manage.sh update
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Service Not Starting

If services fail to start:

```bash
# Check service status
./server/manage.sh status

# View detailed logs
./server/manage.sh logs

# Check Docker status
docker ps
docker-compose -f docker/docker-compose.yml ps

# Check logs for a specific service
./server/manage.sh logs api
```

#### 2. Database Connection Issues

```bash
# Check if database containers are running
./server/manage.sh status

# Check database logs
./server/manage.sh logs postgres
./server/manage.sh logs redis

# Test database connection
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -c "\dt"
```

#### 3. SSL Issues

```bash
# Check Nginx configuration
sudo nginx -t

# Check SSL certificate status
sudo certbot certificates

# Renew certificate if needed
sudo certbot renew
```

#### 4. Port Conflicts

```bash
# Check what's using port 5001
sudo lsof -i :5001

# Check all listening ports
sudo netstat -tlnp
```

### Diagnostic Commands

```bash
# Comprehensive system check
./server/manage.sh monitor

# Configuration check
./server/manage.sh config

# View all logs with timestamps
cd /opt/citadel-agent
docker-compose -f docker/docker-compose.yml logs --since 1h
```

## Backup and Restore

### 1. Create Backups

```bash
# Full backup (includes config, database dump)
./server/manage.sh backup

# Database backup only
./server/manage.sh backup-db

# The backup files will be created in /tmp/ on your server
```

### 2. Restore from Backup

```bash
# Restore from backup file
./server/manage.sh restore /path/to/backup.tar.gz

# Example
./server/manage.sh restore /tmp/citadel-full-backup-20231201-120000.tar.gz
```

### 3. Manual Backup Commands

```bash
# Database backup
cd /opt/citadel-agent
docker-compose -f docker/docker-compose.yml exec postgres pg_dump -U postgres -d citadel_agent > backup.sql

# Configuration backup
tar -czf config-backup.tar.gz .env docker/
```

## Updates

### 1. Automatic Update

```bash
cd /opt/citadel-agent
./server/manage.sh update
```

### 2. Manual Update

```bash
# Stop services
./scripts/stop.sh

# Backup first
./server/manage.sh backup

# Update codebase
git fetch
git pull origin main

# Update Docker images
docker-compose -f docker/docker-compose.yml pull

# Start services
./scripts/start.sh
```

### 3. Update Verification

```bash
# Check service status
./server/manage.sh status

# Test API health
curl -k https://yourdomain.com/health
```

## Monitoring and Maintenance

### 1. Daily Checks

```bash
# Check service status
./server/manage.sh status

# Check system resources
./server/manage.sh monitor

# Check logs for errors
./server/manage.sh logs | grep -i error
```

### 2. Weekly Maintenance

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Clean Docker resources
./server/manage.sh cleanup

# Rotate logs if needed
sudo journalctl --rotate
sudo journalctl --vacuum-time=7d
```

### 3. Monthly Tasks

- Review security logs
- Update SSL certificates if needed: `sudo certbot renew --dry-run`
- Check storage space: `df -h`
- Test backup restoration process

## Support

For issues with the self-hosted version:

- **GitHub Issues**: [Citadel Agent Issues](https://github.com/citadel-agent/citadel-agent/issues)
- **Documentation**: [Self-Hosting Docs](https://citadel-agent.com/docs/self-hosting)
- **Community**: [Discord Community](https://discord.gg/citadel-agent) (if available)

## Migration from Demo/Cloud Version

If migrating from a demo or cloud version:

1. Export your workflows from the current instance
2. Create a backup using the backup commands
3. Deploy to your server following the quick deployment guide
4. Import your workflows through the API or UI
5. Update any domain references in your workflows

---

**Important Notes:**
- Always backup before making changes
- Test SSL and access after configuration
- Keep your server and Citadel Agent updated
- Monitor disk space regularly
- Secure your server access credentials

**Security Warning**: For production use, ensure you follow all security best practices outlined in this document.