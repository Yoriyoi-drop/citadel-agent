#!/bin/bash

# Citadel Agent Deployment Script
# This script deploys Citadel Agent to a remote server using SCP and SSH

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Default values
REMOTE_USER=""
REMOTE_HOST=""
REMOTE_PORT="22"
REMOTE_PATH="/opt/citadel-agent"
BRANCH="main"

# Usage function
usage() {
    echo "Usage: $0 -u <username> -h <host> [OPTIONS]"
    echo "  -u, --user USERNAME      Remote server username"
    echo "  -h, --host HOST          Remote server hostname/IP"
    echo "  -p, --port PORT          SSH port (default: 22)"
    echo "  -P, --path PATH          Remote installation path (default: /opt/citadel-agent)"
    echo "  -b, --branch BRANCH      Git branch to deploy (default: main)"
    echo "  -e, --env-file FILE      Local .env file to copy to server"
    echo "  --skip-setup             Skip server setup (if already configured)"
    echo "  --with-ssl               Include SSL setup instructions"
    echo "  -y, --yes                Skip confirmation prompts"
    echo "  -h, --help               Show this help"
    exit 1
}

# Parse command line options
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--user)
            REMOTE_USER="$2"
            shift 2
            ;;
        -h|--host)
            REMOTE_HOST="$2"
            shift 2
            ;;
        -p|--port)
            REMOTE_PORT="$2"
            shift 2
            ;;
        -P|--path)
            REMOTE_PATH="$2"
            shift 2
            ;;
        -b|--branch)
            BRANCH="$2"
            shift 2
            ;;
        -e|--env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --skip-setup)
            SKIP_SETUP="true"
            shift
            ;;
        --with-ssl)
            WITH_SSL="true"
            shift
            ;;
        -y|--yes)
            ASSUME_YES="true"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required parameters
if [[ -z "$REMOTE_USER" || -z "$REMOTE_HOST" ]]; then
    log_error "Username and hostname are required!"
    usage
fi

# Validate .env file if provided
if [[ -n "$ENV_FILE" && ! -f "$ENV_FILE" ]]; then
    log_error "Specified .env file does not exist: $ENV_FILE"
    exit 1
fi

# Confirmation prompt
if [[ -z "$ASSUME_YES" ]]; then
    echo
    echo "Deploying Citadel Agent to:"
    echo "  User: $REMOTE_USER"
    echo "  Host: $REMOTE_HOST:$REMOTE_PORT"
    echo "  Path: $REMOTE_PATH"
    echo "  Branch: $BRANCH"
    echo
    read -p "Continue with deployment? (y/N): " -n 1 -r REPLY
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Deployment cancelled."
        exit 0
    fi
fi

# Test SSH connection
test_ssh_connection() {
    log_info "Testing SSH connection to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT..."
    
    if ssh -o ConnectTimeout=10 -p $REMOTE_PORT -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST 'exit 0'; then
        log_info "SSH connection successful"
    else
        log_error "SSH connection failed. Please check host, user, port, and SSH keys."
        exit 1
    fi
}

# Check if remote server is configured for Citadel Agent
check_remote_setup() {
    log_info "Checking if server is pre-configured for Citadel Agent..."
    
    setup_status=$(ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        if command -v docker >/dev/null 2>&1 && command -v docker-compose >/dev/null 2>&1 && docker ps >/dev/null 2>&1; then
            echo 'configured'
        else
            echo 'not_configured'
        fi
    ")
    
    if [[ "$setup_status" == "not_configured" ]]; then
        if [[ "$SKIP_SETUP" != "true" ]]; then
            log_warn "Remote server is not configured for Citadel Agent."
            if [[ -z "$ASSUME_YES" ]]; then
                read -p "Run server setup script remotely? (y/N): " -n 1 -r REPLY
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    run_remote_setup
                else
                    log_error "Server must be configured before deployment. Exiting."
                    exit 1
                fi
            else
                log_info "Running remote setup automatically..."
                run_remote_setup
            fi
        else
            log_warn "Skipping setup check as requested."
        fi
    else
        log_info "Remote server appears to be configured."
    fi
}

# Run setup script on remote server
run_remote_setup() {
    log_info "Uploading and running setup script on remote server..."
    
    # Upload the setup script
    scp -P $REMOTE_PORT /home/whale-d/fajar/citadel-agent/server/setup.sh $REMOTE_USER@$REMOTE_HOST:/tmp/citadel_setup.sh
    
    # Run the setup script remotely
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        chmod +x /tmp/citadel_setup.sh
        echo 'Running Citadel Agent server setup...'
        /tmp/citadel_setup.sh
        echo 'Setup completed. You may need to log out and back in to use Docker without sudo.'
    "
    
    log_info "Remote setup completed."
    
    # Ask user to log back in
    log_warn "Please log into your server and verify Docker works without sudo before continuing."
    if [[ -z "$ASSUME_YES" ]]; then
        read -p "Press Enter to continue after verifying Docker access..." 
    fi
}

# Create deployment package locally
create_deployment_package() {
    log_info "Creating deployment package..."
    
    # Create a temporary directory for the deployment package
    DEPLOY_DIR=$(mktemp -d)
    
    # Copy the entire project to the deployment directory
    rsync -av \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='*.log' \
        --exclude='tmp/' \
        --exclude='dist/' \
        --exclude='.DS_Store' \
        /home/whale-d/fajar/citadel-agent/ "$DEPLOY_DIR/citadel-agent/"
    
    # Create the deployment archive
    DEPLOY_ARCHIVE="/tmp/citadel-agent-deploy-$(date +%Y%m%d_%H%M%S).tar.gz"
    tar -czf "$DEPLOY_ARCHIVE" -C "$(dirname $DEPLOY_DIR)" "$(basename $DEPLOY_DIR)/citadel-agent"
    
    # Clean up temporary directory
    rm -rf $DEPLOY_DIR
    
    log_info "Deployment package created: $DEPLOY_ARCHIVE"
}

# Upload deployment package to remote server
upload_deployment_package() {
    log_info "Uploading deployment package to remote server..."
    
    # Upload the archive to remote server
    scp -P $REMOTE_PORT "$DEPLOY_ARCHIVE" "$REMOTE_USER@$REMOTE_HOST:/tmp/"
    
    # Extract on remote server
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        # Create target directory
        mkdir -p $REMOTE_PATH
        
        # Extract the archive
        tar -xzf '$DEPLOY_ARCHIVE' -C $REMOTE_PATH --strip-components=1
        
        # Make scripts executable
        chmod +x $REMOTE_PATH/scripts/*.sh
        
        # Remove the archive
        rm '$DEPLOY_ARCHIVE'
        
        echo 'Deployment package uploaded and extracted to $REMOTE_PATH'
    "
    
    log_info "Deployment package uploaded and extracted."
}

# Upload environment file if provided
upload_env_file() {
    if [[ -n "$ENV_FILE" ]]; then
        log_info "Uploading .env file..."
        
        # Upload the .env file
        scp -P $REMOTE_PORT "$ENV_FILE" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/.env"
        
        # Update the .env file on remote server to use the uploaded version
        ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
            cd $REMOTE_PATH
            if [ ! -f .env ]; then
                cp .env.example .env
                echo 'Created .env file from example.'
            fi
            echo 'Environment file uploaded.'
        "
        
        log_info ".env file uploaded."
    else
        # Check if .env exists on remote server
        env_exists=$(ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
            if [ -f $REMOTE_PATH/.env ]; then
                echo 'exists'
            else
                echo 'missing'
            fi
        ")
        
        if [[ "$env_exists" == "missing" ]]; then
            log_warn ".env file does not exist on remote server. Creating from example..."
            ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
                cd $REMOTE_PATH
                cp .env.example .env
                echo 'Created .env file from example. Please update with your configuration.'
            "
        else
            log_info ".env file already exists on remote server."
        fi
    fi
}

# Start Citadel Agent services
start_services() {
    log_info "Starting Citadel Agent services..."
    
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd $REMOTE_PATH
        ./scripts/start.sh
    "
    
    log_info "Services started. Checking service status..."
    
    # Wait a moment for services to start
    sleep 10
    
    # Check status
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd $REMOTE_PATH
        ./scripts/status.sh
    "
}

# Display access information
display_access_info() {
    log_info "🎉 Citadel Agent deployment completed!"
    echo
    echo "📋 Access Information:"
    echo "   URL: https://$REMOTE_HOST (if SSL configured)"
    echo "   URL: http://$REMOTE_HOST:5001 (direct API access)"
    echo "   Health: https://$REMOTE_HOST/health (if SSL) or http://$REMOTE_HOST:5001/health (direct)"
    echo
    echo "🔧 Management Commands (on remote server):"
    echo "   cd $REMOTE_PATH"
    echo "   ./scripts/status.sh    # Check service status"
    echo "   ./scripts/stop.sh      # Stop services"
    echo "   ./scripts/start.sh     # Start services"
    echo
    echo "💡 SSL Setup:"
    if [[ "$WITH_SSL" == "true" ]]; then
        echo "   1. Point your domain to this server"
        echo "   2. Update server_name in Nginx config"
        echo "   3. Run: sudo certbot --nginx -d yourdomain.com"
    else
        echo "   For SSL, run: sudo certbot --nginx -d yourdomain.com"
        echo "   Then update server_name in Nginx configuration."
    fi
    echo
    echo "🔐 Security Notes:"
    echo "   1. Change JWT_SECRET in .env file to a strong password"
    echo "   2. Update firewall rules to restrict access as needed"
    echo "   3. Regularly update your system and Docker images"
    echo
    echo "⚠️  Troubleshooting: Check logs with: docker-compose -f docker/docker-compose.yml logs -f"
}

# Main deployment function
main() {
    log_info "🚀 Starting Citadel Agent deployment to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT"
    
    test_ssh_connection
    check_remote_setup
    create_deployment_package
    upload_deployment_package
    upload_env_file
    start_services
    display_access_info
    
    log_info "Deployment completed successfully!"
}

# Run main function
main