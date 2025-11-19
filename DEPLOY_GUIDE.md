# Citadel Agent - One-Click Server Deployment

## Quick Setup Guide

Deploy Citadel Agent to your own server with these simple steps:

### Prerequisites

1. **A server** running:
   - Ubuntu 20.04+ / Debian 11+ / CentOS 8+ / RHEL 8+ 
   - 4GB+ RAM, 2+ CPU cores
   - Public IP address
   - Ports 22 (SSH), 80 (HTTP), 443 (HTTPS) open

2. **Local access** with:
   - SSH client
   - SCP/SFTP client
   - SSH key access to your server (recommended)

### Step 1: Prepare Your Server

SSH into your server and run the setup script:

```bash
# Download the setup script
wget https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/server/setup.sh

# Make it executable
chmod +x setup.sh

# Run setup (this will install Docker, Git, Nginx, firewall, etc.)
./setup.sh
```

**What this does:**
- Updates your system
- Installs Docker, Docker Compose, Git
- Sets up firewall (ports 80, 443, 5001 allowed)
- Clones Citadel Agent
- Configures Nginx reverse proxy
- Makes you part of docker group

**Important**: After setup completes, log out and log back in to use Docker without sudo.

### Step 2: Configure Environment (Optional but Recommended)

Edit the environment file to customize your installation:

```bash
# Navigate to Citadel Agent directory
cd ~/citadel-agent

# Edit environment file
nano .env
```

**Important to change:**
- `JWT_SECRET`: Generate a strong, unique secret (at least 32 characters)
- `DB_PASSWORD`: Change from default
- Any other sensitive settings

### Step 3: Start the Services

```bash
# Start all services with health checks
./scripts/start.sh

# Check status
./scripts/status.sh

# View logs if needed
docker-compose -f docker/docker-compose.yml logs -f
```

### Step 4: Configure Domain and SSL

1. **Point your domain** to your server's IP address

2. **Obtain SSL certificate** with Certbot:

```bash
# Test Certbot first
sudo certbot --nginx -d yourdomain.com

# For multiple domains
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

3. **Update Nginx configuration** (if needed):
   - Edit `/etc/nginx/sites-available/citadel-agent`
   - Change `server_name _;` to `server_name yourdomain.com;`
   - Reload: `sudo systemctl reload nginx`

### Step 5: Access Your Instance

- **Dashboard**: `https://yourdomain.com`
- **API Health**: `https://yourdomain.com/health`
- **Direct API**: `http://your-server-ip:5001`

## Management Commands

```bash
cd ~/citadel-agent

# Check status
./scripts/status.sh

# Stop services
./scripts/stop.sh

# Start services
./scripts/start.sh  

# View all logs
docker-compose -f docker/docker-compose.yml logs -f

# View specific service logs
docker-compose -f docker/docker-compose.yml logs -f api
docker-compose -f docker/docker-compose.yml logs -f worker
docker-compose -f docker/docker-compose.yml logs -f scheduler
```

## Deployment via Local Machine

Alternatively, deploy from your local machine to any server:

```bash
# On your local machine
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent/server

# Deploy to remote server
./deploy.sh -u your-server-username -h your-server-ip

# Or with specific options
./deploy.sh -u admin -h 203.0.113.10 -e /path/to/your/.env -y
```

## Security Best Practices

1. **Change default passwords** in `.env` file
2. **Enable firewall** (done automatically by setup script)
3. **Use SSL certificates** from Let's Encrypt
4. **Keep your system updated**
5. **Use SSH keys** instead of passwords
6. **Regular backups** of your data

## Troubleshooting

### Services Not Starting
```bash
# Check status
./scripts/status.sh

# Check logs
docker-compose -f docker/docker-compose.yml logs

# Check specific service
docker-compose -f docker/docker-compose.yml logs api
```

### SSL Issues
```bash
# Check certificate status
sudo certbot certificates

# Test renewal
sudo certbot renew --dry-run
```

### Port Issues
```bash
# Check what's using port 5001
sudo lsof -i :5001

# Check firewall
sudo ufw status
```

## Backups and Updates

### Create Backup
```bash
# Full backup
cd ~/citadel-agent
./server/manage.sh backup

# Database only backup
./server/manage.sh backup-db
```

### Update to Latest Version
```bash
cd ~/citadel-agent
./server/manage.sh update
```

## Support

- **Documentation**: Check SELF_HOSTING.md for full details
- **GitHub**: [Create an issue](https://github.com/citadel-agent/citadel-agent/issues)
- **Status**: Check service health at `https://yourdomain.com/health`

---

**Note**: This software is under active development. Always backup before updating and check the CHANGELOG for breaking changes.