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

### 2. Backup Strategies and Scheduling

For production environments, implement regular backup schedules:

#### Daily Automated Backups
```bash
# Add to crontab for daily backups at 2 AM
sudo crontab -e
# Add this line for daily full backups:
0 2 * * * cd /opt/citadel-agent && ./server/manage.sh backup

# Or for database-only backups (faster):
0 2 * * * cd /opt/citadel-agent && ./server/manage.sh backup-db
```

#### Backup Storage Locations
- **Local**: `/tmp/` directory on your server (default)
- **Remote**: Use rsync or scp to push backups to external storage
- **Cloud**: Amazon S3, Google Cloud Storage, or other cloud providers

### 3. Backup Verification

Always verify your backups to ensure they can be restored:

```bash
# List backup files with details
ls -la /tmp/citadel-backup-*.tar.gz

# Check backup contents without extracting
tar -tvf /tmp/citadel-backup-20231201-120000.tar.gz

# Verify database backup integrity
# (restore a copy to a test database)
```

### 4. Restore from Backup

```bash
# Restore from backup file
./server/manage.sh restore /path/to/backup.tar.gz

# Example
./server/manage.sh restore /tmp/citadel-full-backup-20231201-120000.tar.gz

# Important: Services will be stopped during restore process
# Make sure to schedule restores during maintenance windows
```

### 5. Manual Backup Commands

```bash
# Database backup
cd /opt/citadel-agent
docker-compose -f docker/docker-compose.yml exec postgres pg_dump -U postgres -d citadel_agent > backup.sql

# Configuration backup
tar -czf config-backup.tar.gz .env docker/

# Custom backup with compression
tar -czf /tmp/citadel-custom-backup-$(date +%Y%m%d-%H%M%S).tar.gz .env docker/ data/
```

### 6. Backup Retention Policies

Implement a retention policy to manage storage space:

```bash
# Keep daily backups for 7 days, weekly for 4 weeks, monthly for 12 months
# This can be implemented with a custom script

# Example cleanup script
find /tmp -name "citadel-backup-*.tar.gz" -mtime +7 -delete  # Delete backups older than 7 days
find /tmp -name "citadel-weekly-backup-*.tar.gz" -mtime +30 -delete  # Delete weekly backups older than 30 days
```

### 7. Remote Backup Storage

For enhanced security, store backups on remote systems:

```bash
# Copy backup to remote server
scp /tmp/citadel-backup-*.tar.gz user@remote-server:/path/to/backup-storage/

# Or using rsync with compression
rsync -avz /tmp/citadel-backup-*.tar.gz user@remote-server:/path/to/backup-storage/

# AWS S3 example (requires AWS CLI setup)
aws s3 cp /tmp/citadel-backup-*.tar.gz s3://your-backup-bucket/citadel-agent/

# Google Cloud Storage example (requires gsutil setup)
gsutil cp /tmp/citadel-backup-*.tar.gz gs://your-backup-bucket/citadel-agent/
```

### 8. Disaster Recovery Plan

1. Ensure backup files are stored offsite or in cloud storage
2. Test restore procedures regularly (at least monthly)
3. Document the recovery process for your team
4. Keep encryption keys and credentials secure and separate from backups
5. Have a secondary server ready to restore to in case of primary server failure

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

### 3. Update Strategies and Best Practices

#### Pre-Update Checklist
- [ ] Verify backup is recent and functional: `./server/manage.sh backup`
- [ ] Check current system status: `./server/manage.sh status`
- [ ] Review release notes for breaking changes
- [ ] Schedule update during low-traffic periods
- [ ] Inform users of planned maintenance window

#### Update Scheduling
For production environments, implement regular update schedules:

```bash
# Weekly updates (Sundays at 3 AM)
sudo crontab -e
# Add for automatic updates:
0 3 * * 0 cd /opt/citadel-agent && ./server/manage.sh update

# Or with manual backup before updates:
0 3 * * 0 cd /opt/citadel-agent && ./server/manage.sh backup && ./server/manage.sh update
```

#### Version Pinning
To maintain stability in production environments:

```bash
# Check current version
git describe --tags

# Update to specific version (not just latest)
git fetch
git checkout v1.2.3  # Replace with desired version

# Or update to latest release tag
git fetch --tags
git checkout $(git describe --tags $(git rev-list --tags --max-count=1))
```

#### Rollback Procedures
If an update causes issues:

```bash
# Stop services
./scripts/stop.sh

# Restore from backup if needed
./server/manage.sh restore /path/to/previous-backup.tar.gz

# Or rollback to previous version
git fetch
git checkout HEAD~1  # Go back one commit
# Or checkout specific previous version tag

# Pull old Docker images if needed
docker-compose -f docker/docker-compose.yml pull

# Start services
./scripts/start.sh
```

### 4. Update Verification

```bash
# Check service status
./server/manage.sh status

# Test API health
curl -k https://yourdomain.com/health

# Verify version
git describe --tags

# Check all services are running
docker-compose -f docker/docker-compose.yml ps
```

### 5. Update Notifications

Set up monitoring for new releases:

```bash
# Check for updates manually
git fetch
git tag -l | sort -V | tail -n 5  # Show latest 5 tags

# Or use GitHub API to check for releases
curl -s https://api.github.com/repos/citadel-agent/citadel-agent/releases/latest | grep tag_name
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

# Check disk space
df -h

# Monitor API health
curl -k https://yourdomain.com/health
```

### 2. Monitoring Tools and Metrics

#### System Resource Monitoring
```bash
# Real-time monitoring with htop
htop

# Disk usage
du -sh /opt/citadel-agent/

# Docker container resource usage
docker stats

# Network monitoring
iftop  # Install with: apt install iftop
```

#### Log Monitoring
```bash
# Follow logs in real-time
./server/manage.sh logs -f

# Monitor specific service logs
./server/manage.sh logs api -f

# Check error logs specifically
./server/manage.sh logs | grep -i error | tail -n 20

# Monitor access logs if needed
sudo tail -f /var/log/nginx/access.log
```

#### Performance Metrics
- **CPU Usage**: Should remain under 80% during normal operations
- **Memory Usage**: Monitor for memory leaks - containers should not grow indefinitely
- **Disk Space**: Keep at least 20% free space available
- **Response Times**: Monitor API response times for degradation

### 3. Weekly Maintenance

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Clean Docker resources
./server/manage.sh cleanup

# Rotate logs if needed
sudo journalctl --rotate
sudo journalctl --vacuum-time=7d

# Check for SSL certificate renewal
sudo certbot renew --dry-run

# Verify backup integrity
./server/manage.sh backup
# Verify the backup file was created successfully
ls -la /tmp/citadel-backup-*.tar.gz
```

### 4. Monthly Tasks

- Review security logs
- Update SSL certificates if needed: `sudo certbot renew --dry-run`
- Check storage space: `df -h`
- Test backup restoration process
- Review application logs for recurring issues
- Update monitoring configuration if needed

### 5. Advanced Monitoring Setup

#### Setting up monitoring with external tools:

```bash
# For Prometheus monitoring, you can expose metrics endpoints
# Add to your .env file:
METRICS_ENABLED=true
METRICS_PORT=9090

# Configure Prometheus to scrape from your server
```

#### Setting up automated alerts:

```bash
# Create a monitoring script that checks service health
# Save as /opt/citadel-agent/scripts/health-check.sh
#!/bin/bash
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" https://yourdomain.com/health)

if [ $HEALTH_CHECK -ne 200 ]; then
    echo "CRITICAL: Service health check failed with code $HEALTH_CHECK"
    # Add notification command (e.g., send email, Slack message)
    # curl -X POST -H 'Content-type: application/json' --data '{"text":"Citadel Agent health check failed!"}' $SLACK_WEBHOOK_URL
else
    echo "OK: Service is healthy"
fi
```

Add to crontab for regular monitoring:
```bash
# Check health every 5 minutes
*/5 * * * * /opt/citadel-agent/scripts/health-check.sh >> /var/log/citadel-health.log 2>&1
```

### 6. Performance Optimization

#### Database Optimization
```bash
# Check database performance
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d citadel_agent -c "SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del FROM pg_stat_user_tables;"

# Clean up database if needed
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d citadel_agent -c "VACUUM ANALYZE;"
```

#### Cleanup Tasks
```bash
# Remove unused Docker images
docker image prune -f

# Remove unused Docker volumes
docker volume prune -f

# Check for orphaned containers
docker container prune -f
```

## Support

### 1. Issue Reporting Guidelines

When requesting support, please include the following information:

#### System Information
- Operating system and version
- Docker and Docker Compose versions
- Citadel Agent version
- Installation method (automated script or manual)

#### Issue Details
- Specific error messages (copy and paste exact text)
- Steps to reproduce the issue
- Expected behavior vs. actual behavior
- Screenshots if relevant (for UI issues)

#### Diagnostic Information
```bash
# Collect system information
docker --version
docker-compose --version
git -C /opt/citadel-agent describe --tags
./server/manage.sh config
./server/manage.sh status
./server/manage.sh logs --since 1h
```

### 2. Self-Help Resources

#### Documentation
- [Citadel Agent Documentation](https://citadel-agent.com/docs)
- [API Documentation](https://citadel-agent.com/docs/api)
- [Troubleshooting Guide](#troubleshooting) (in this document)
- [Configuration Guide](https://citadel-agent.com/docs/configuration)

#### Community Resources
- [GitHub Discussions](https://github.com/citadel-agent/citadel-agent/discussions)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/citadel-agent)
- [Official Wiki](https://github.com/citadel-agent/citadel-agent/wiki)

### 3. Support Channels

For issues with the self-hosted version:

- **GitHub Issues**: [Citadel Agent Issues](https://github.com/citadel-agent/citadel-agent/issues) (Bug reports and feature requests)
- **Documentation**: [Self-Hosting Docs](https://citadel-agent.com/docs/self-hosting)
- **Community**: [Discord Community](https://discord.gg/citadel-agent) (if available)
- **Commercial Support**: [Contact Information](https://citadel-agent.com/support) (for enterprise users)

### 4. Diagnostic Commands

Before creating a support ticket, run these diagnostic commands:

```bash
# Check system resources
./server/manage.sh monitor

# Check configuration
./server/manage.sh config

# Check service status
./server/manage.sh status

# View recent logs
./server/manage.sh logs --since 1h

# Check Docker containers
docker ps -a

# Check disk space
df -h

# Test database connectivity
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d citadel_agent -c "SELECT 1;"
```

### 5. Security Issues

For security-related issues or vulnerabilities:
- Do not report through public channels
- Email security directly: security@citadel-agent.com
- Include detailed information about the vulnerability
- Follow responsible disclosure practices
- Allow time for fixes before public disclosure

## Migration from Demo/Cloud Version

Use this guide when migrating from a demo, cloud, or another Citadel Agent instance to your self-hosted instance.

### 1. Pre-Migration Checklist

- [ ] Verify self-hosted instance is running and accessible
- [ ] Ensure sufficient storage space on the new server
- [ ] Test backup and restore functionality on a non-production instance
- [ ] Plan a maintenance window for the migration
- [ ] Inform users about planned downtime
- [ ] Create a full backup of the source instance
- [ ] Document all active workflows and configurations

### 2. Data Export from Source Instance

#### Backup Current Instance
```bash
# If source is also self-hosted, create a backup
./server/manage.sh backup

# Or export data via API
curl -X GET "https://source-instance.com/api/v1/export" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -o exported-data.json
```

#### Export Workflows and Configurations
```bash
# Using API to export all workflows
curl -X GET "https://source-instance.com/api/v1/workflows" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -o workflows.json

# Export triggers and integrations
curl -X GET "https://source-instance.com/api/v1/triggers" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -o triggers.json
```

### 3. Prepare Self-Hosted Instance

#### Verify Configuration
- Update your `.env` settings to reflect your production requirements
- Ensure SSL certificates are properly configured
- Verify database connection settings
- Check storage paths and permissions

#### Test Instance
```bash
# Verify all services are running
./server/manage.sh status

# Test API connectivity
curl https://yourdomain.com/health

# Verify database access
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d citadel_agent -c "SELECT version();"
```

### 4. Data Migration

#### Option A: Using Backup/Restore (Recommended)
```bash
# Transfer backup file to your server
scp backup-file.tar.gz user@your-server:/tmp/

# Restore on your instance
./server/manage.sh restore /tmp/backup-file.tar.gz
```

#### Option B: API-Based Migration
```bash
# Import workflows via API
curl -X POST "https://yourdomain.com/api/v1/workflows/import" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d @workflows.json

# Update domain references in workflows
# This may require manual editing depending on how domain-specific data is stored
```

### 5. Post-Migration Verification

#### Data Verification
```bash
# Check that workflows exist in the system
curl -X GET "https://yourdomain.com/api/v1/workflows" \
  -H "Authorization: Bearer YOUR_API_TOKEN"

# Verify database records
docker-compose -f docker/docker-compose.yml exec postgres psql -U postgres -d citadel_agent -c "SELECT COUNT(*) FROM workflows;"
```

#### Functional Testing
- Test each migrated workflow manually
- Verify triggers are working correctly
- Check that notifications and webhooks are functioning
- Validate user access and permissions

### 6. Domain and URL Updates

#### Update Domain References
If migrating from a different domain, you may need to update domain-specific configurations:

```bash
# Search for domain references in configuration
grep -r "olddomain.com" /opt/citadel-agent/

# Update any hardcoded domain references in workflows or configurations
sed -i 's/olddomain.com/newdomain.com/g' /path/to/config/files
```

### 7. DNS and Traffic Switch

#### DNS Updates
- Update your DNS records to point to your new server's IP address
- Wait for DNS propagation (typically 24-48 hours, but may be faster)
- Monitor for any issues during transition

#### Gradual Cutover
- Consider running both instances simultaneously during transition
- Monitor logs for any issues
- Perform testing on the new instance
- Redirect traffic gradually if possible

### 8. Post-Migration Tasks

#### Cleanup
- Archive the old instance after confirming successful migration
- Update any external integrations to use the new domain
- Update documentation with new configuration details
- Remove any temporary migration configurations

#### Verification Checklist
- [ ] All workflows are functional
- [ ] Data integrity verified
- [ ] Users can access the system
- [ ] API endpoints working correctly
- [ ] Notifications and triggers functioning
- [ ] Security settings properly configured
- [ ] Monitoring and alerts configured

### 9. Rollback Plan

If issues occur during migration:

1. Document the issue and stop further migration steps
2. Switch traffic back to the original instance
3. Restore from the backup created before migration if needed
4. Investigate and resolve issues on a test instance
5. Repeat migration when resolved

---

**Important Notes:**
- Always backup before making changes
- Test SSL and access after configuration
- Keep your server and Citadel Agent updated
- Monitor disk space regularly
- Secure your server access credentials

**Security Warning**: For production use, ensure you follow all security best practices outlined in this document.