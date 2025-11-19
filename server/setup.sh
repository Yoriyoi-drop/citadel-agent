#!/bin/bash

# Citadel Agent Server Setup Script
# This script installs all necessary dependencies on a fresh Ubuntu/Debian server

set -e

echo "🚀 Starting Citadel Agent Server Setup..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
    log_error "This script should NOT be run as root. Run as a regular user with sudo access."
    exit 1
fi

# Check if user has sudo access
if ! sudo -v &>/dev/null; then
    log_error "This script requires sudo access but user doesn't have it."
    exit 1
fi

# Detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        log_error "Cannot detect OS"
        exit 1
    fi
    
    if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
        DISTRO="ubuntu"
    elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
        DISTRO="centos"
    else
        log_error "Unsupported OS: $OS. Currently supports Ubuntu/Debian/CentOS/RHEL."
        exit 1
    fi
    
    log_info "Detected OS: $OS ($DISTRO)"
}

# Update system packages
update_system() {
    log_info "Updating system packages..."
    
    if [ "$DISTRO" = "ubuntu" ]; then
        sudo apt update && sudo apt upgrade -y
    elif [ "$DISTRO" = "centos" ]; then
        sudo yum update -y
    fi
    
    log_info "System updated successfully"
}

# Install Docker
install_docker() {
    log_info "Installing Docker..."
    
    if [ "$DISTRO" = "ubuntu" ]; then
        # Remove old versions
        sudo apt remove docker docker-engine docker.io containerd runc 2>/dev/null || true
        
        # Install prerequisites
        sudo apt update
        sudo apt install -y ca-certificates curl gnupg lsb-release
        
        # Add Docker's official GPG key
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        
        # Set up the repository
        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
          $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        
        # Update package index and install Docker
        sudo apt update
        sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    elif [ "$DISTRO" = "centos" ]; then
        sudo yum install -y yum-utils
        sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        sudo yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        sudo systemctl start docker
        sudo systemctl enable docker
    fi
    
    # Add current user to docker group
    sudo usermod -aG docker $USER
    
    log_info "Docker installed and user added to docker group"
    log_warn "Please log out and log back in to use Docker without sudo"
}

# Install Git
install_git() {
    log_info "Installing Git..."
    
    if [ "$DISTRO" = "ubuntu" ]; then
        sudo apt install -y git
    elif [ "$DISTRO" = "centos" ]; then
        sudo yum install -y git
    fi
    
    log_info "Git installed successfully"
}

# Install essential tools
install_tools() {
    log_info "Installing essential tools..."
    
    if [ "$DISTRO" = "ubuntu" ]; then
        sudo apt install -y wget curl nginx certbot python3-certbot-nginx ufw fail2ban
    elif [ "$DISTRO" = "centos" ]; then
        sudo yum install -y wget curl nginx certbot python3-certbot-nginx firewalld
    fi
    
    log_info "Essential tools installed"
}

# Setup firewall
setup_firewall() {
    log_info "Setting up firewall..."
    
    if [ "$DISTRO" = "ubuntu" ]; then
        sudo ufw allow OpenSSH
        sudo ufw allow 'Nginx Full'
        sudo ufw allow 5001/tcp  # Citadel Agent API
        sudo ufw --force enable
    elif [ "$DISTRO" = "centos" ]; then
        sudo firewall-cmd --permanent --add-service=ssh
        sudo firewall-cmd --permanent --add-service=http
        sudo firewall-cmd --permanent --add-service=https
        sudo firewall-cmd --permanent --add-port=5001/tcp
        sudo firewall-cmd --reload
    fi
    
    log_info "Firewall configured"
}

# Clone Citadel Agent repository
clone_repository() {
    log_info "Cloning Citadel Agent repository..."
    
    if [ -d "~/citadel-agent" ]; then
        log_warn "~/citadel-agent directory already exists. Backing up and re-cloning..."
        mv ~/citadel-agent ~/citadel-agent-backup-$(date +%Y%m%d_%H%M%S)
    fi
    
    git clone https://github.com/citadel-agent/citadel-agent.git ~/citadel-agent
    cd ~/citadel-agent
    
    log_info "Repository cloned successfully"
}

# Setup Citadel Agent
setup_citadel() {
    log_info "Setting up Citadel Agent..."
    
    cd ~/citadel-agent
    
    # Create environment file
    if [ ! -f ".env" ]; then
        cp .env.example .env
        
        # Generate a secure JWT secret
        JWT_SECRET=$(openssl rand -hex 32)
        sed -i "s/YOUR_SUPER_SECRET_JWT_KEY_HERE_CHANGE_IN_PRODUCTION/$JWT_SECRET/" .env
        
        log_info "Generated new JWT secret and saved to .env"
    fi
    
    # Make scripts executable
    chmod +x ./scripts/start.sh ./scripts/stop.sh ./scripts/status.sh
    
    log_info "Citadel Agent setup complete"
}

# Setup Nginx reverse proxy
setup_nginx() {
    log_info "Setting up Nginx reverse proxy..."
    
    # Create Nginx configuration
    sudo tee /etc/nginx/sites-available/citadel-agent > /dev/null <<EOF
server {
    listen 80;
    server_name _; # Change this to your domain

    # Redirect all HTTP traffic to HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name _; # Change this to your domain

    # SSL Certificate - Update paths when you get certificates
    # ssl_certificate /path/to/your/certificate.crt;
    # ssl_certificate_key /path/to/your/private.key;
    
    # For testing, you can use these temporary certificates
    ssl_certificate /etc/ssl/certs/ssl-cert-snakeoil.pem;
    ssl_certificate_key /etc/ssl/private/ssl-cert-snakeoil.key;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

    location / {
        proxy_pass http://localhost:5001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
EOF

    # Enable the site
    sudo ln -sf /etc/nginx/sites-available/citadel-agent /etc/nginx/sites-enabled/
    sudo rm -f /etc/nginx/sites-enabled/default
    
    # Test Nginx configuration
    sudo nginx -t
    
    # Reload Nginx
    sudo systemctl reload nginx
    
    log_info "Nginx reverse proxy configured"
}

# Print setup completion message
print_completion_message() {
    log_info "🎉 Citadel Agent Server Setup Complete!"
    echo
    echo "📋 Next Steps:"
    echo "1. ${YELLOW}Log out and log back in${NC} to use Docker without sudo"
    echo "2. Update the server_name in /etc/nginx/sites-available/citadel-agent with your domain"
    echo "3. Obtain SSL certificates: sudo certbot --nginx -d yourdomain.com"
    echo "4. Start Citadel Agent: cd ~/citadel-agent && ./scripts/start.sh"
    echo "5. Check status: cd ~/citadel-agent && ./scripts/status.sh"
    echo
    echo "🔗 Access your Citadel Agent at: https://yourdomain.com"
    echo "🔧 API Health Check: https://yourdomain.com/health"
    echo
    echo "ℹ️  For SSL certificate setup (after domain points to your server):"
    echo "   sudo certbot --nginx -d yourdomain.com"
    echo
    echo "🔒 Security Note: Change the JWT_SECRET in .env file to a strong password"
    echo "   JWT_SECRET should be at least 32 characters and highly random"
}

# Main function
main() {
    detect_os
    update_system
    install_docker
    install_git
    install_tools
    setup_firewall
    clone_repository
    setup_citadel
    setup_nginx
    print_completion_message
}

# Run main function
main