#!/bin/bash

# Citadel Agent Complete Deployment Tool
# This tool provides a complete solution for deploying Citadel Agent to a remote server

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Default values
REMOTE_USER=""
REMOTE_HOST=""
REMOTE_PORT="22"
REMOTE_PATH="/opt/citadel-agent"
LOCAL_PATH="$(pwd)/../.."
EXCLUDE_GIT="true"
ENV_FILE=""
MODE="full"  # full, upload-only, setup-only

# Help function
usage() {
    echo "Citadel Agent Deployment Tool - Complete Server Deployment Solution"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Modes:"
    echo "  --full      Deploy complete installation (default)"
    echo "  --upload    Upload files only"
    echo "  --setup     Setup server only (requires files already present)"
    echo
    echo "Required Parameters:"
    echo "  -u, --user USERNAME      Remote server username"
    echo "  -h, --host HOST          Remote server hostname/IP"
    echo
    echo "Optional Parameters:"  
    echo "  -p, --port PORT          SSH port (default: 22)"
    echo "  -P, --path PATH          Remote installation path (default: /opt/citadel-agent)"
    echo "  -l, --local-path PATH    Local path to Citadel Agent (default: current directory)"
    echo "  -e, --env-file FILE      Local .env file to copy to server"
    echo "  --exclude-git            Exclude .git directory from upload (default: true)"
    echo "  --include-git            Include .git directory in upload"
    echo "  -y, --yes                Skip confirmation prompts"
    echo "  --create-remote-dir      Create remote directory if it doesn't exist"
    echo "  --skip-validation        Skip prerequisite validation"
    echo "  -h, --help               Show this help"
    echo
    echo "Examples:"
    echo "  $0 --full -u admin -h 203.0.113.10"
    echo "  $0 --upload -u ubuntu -h server.com -e ./my-env-file"
    echo "  $0 --setup -u centos -h 192.168.1.100 --create-remote-dir"
    echo
    exit 1
}

# Parse command line options
while [[ $# -gt 0 ]]; do
    case $1 in
        --full)
            MODE="full"
            shift
            ;;
        --upload)
            MODE="upload"
            shift
            ;;
        --setup)
            MODE="setup"
            shift
            ;;
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
        -l|--local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        -e|--env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --exclude-git)
            EXCLUDE_GIT="true"
            shift
            ;;
        --include-git)
            EXCLUDE_GIT="false"
            shift
            ;;
        --create-remote-dir)
            CREATE_REMOTE_DIR="true"
            shift
            ;;
        --skip-validation)
            SKIP_VALIDATION="true"
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

# Banner
echo
echo "################################################################################"
echo "#                                                                              #"
echo "#                    Citadel Agent Deployment Tool                             #"
echo "#                    Complete Self-Hosting Solution                          #"
echo "#                                                                              #"
echo "################################################################################"
echo

# Validate files exist before starting
validate_prerequisites() {
    if [[ "$SKIP_VALIDATION" != "true" ]]; then
        log_info "Validating prerequisites..."
        
        # Check if required files exist
        if [[ ! -d "$LOCAL_PATH" ]]; then
            log_error "Local path does not exist: $LOCAL_PATH"
            exit 1
        fi
        
        if [[ ! -d "$LOCAL_PATH/backend" ]]; then
            log_error "Citadel Agent backend directory not found in $LOCAL_PATH/backend"
            exit 1
        fi
        
        if [[ -n "$ENV_FILE" && ! -f "$ENV_FILE" ]]; then
            log_error "Specified .env file does not exist: $ENV_FILE"
            exit 1
        fi
        
        # Check if required tools are available locally
        if ! command -v ssh &> /dev/null; then
            log_error "ssh command is not available"
            exit 1
        fi
        
        if ! command -v scp &> /dev/null; then
            log_error "scp command is not available"
            exit 1
        fi
        
        log_info "Prerequisites validated successfully"
    fi
}

# Test SSH connection
test_ssh_connection() {
    log_info "Testing SSH connection to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT..."
    
    if ssh -o ConnectTimeout=10 -p $REMOTE_PORT -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST 'exit 0' 2>/dev/null; then
        log_success "SSH connection successful"
    else
        log_error "SSH connection failed. Please check host, user, port, and SSH keys."
        exit 1
    fi
}

# Upload files to server
upload_files() {
    log_info "Starting file upload process..."
    
    # Create local archive
    log_info "Creating local archive..."
    BASE_NAME=$(basename "$LOCAL_PATH")
    ARCHIVE_NAME="/tmp/citadel-agent-deploy-$(date +%Y%m%d_%H%M%S).tar.gz"
    
    TAR_CMD="tar -czf '$ARCHIVE_NAME' -C '$(dirname "$LOCAL_PATH")' '$BASE_NAME'"
    if [[ "$EXCLUDE_GIT" == "true" ]]; then
        TAR_CMD="tar --exclude='$BASE_NAME/.git' -czf '$ARCHIVE_NAME' -C '$(dirname "$LOCAL_PATH")' '$BASE_NAME'"
    fi
    
    eval "$TAR_CMD"
    ARCHIVE_SIZE=$(du -h "$ARCHIVE_NAME" | cut -f1)
    log_info "Archive created: $ARCHIVE_NAME (Size: $ARCHIVE_SIZE)"
    
    # Create remote directory if needed
    if [[ "$CREATE_REMOTE_DIR" == "true" ]]; then
        log_info "Creating remote directory: $REMOTE_PATH"
        ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "mkdir -p '$REMOTE_PATH'"
    fi
    
    # Upload archive
    log_info "Uploading archive to remote server..."
    scp -P $REMOTE_PORT "$ARCHIVE_NAME" "$REMOTE_USER@$REMOTE_HOST:/tmp/"
    
    # Extract on remote server
    log_info "Extracting files on remote server..."
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH' &&
        rm -rf * .* 2>/dev/null || true  # Clean directory, ignoring errors
        mkdir -p '$REMOTE_PATH' &&
        tar -xzf '$ARCHIVE_NAME' --strip-components=1 -C '$REMOTE_PATH' &&
        chmod +x '$REMOTE_PATH'/scripts/*.sh 2>/dev/null || echo 'No scripts to make executable' &&
        rm '$ARCHIVE_NAME' &&
        echo 'Files uploaded and extracted successfully'
    "
    
    # Upload .env file if specified
    if [[ -n "$ENV_FILE" ]]; then
        log_info "Uploading .env file..."
        scp -P $REMOTE_PORT "$ENV_FILE" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/.env"
    elif [[ ! -f "$REMOTE_PATH/.env" ]]; then
        # Create .env from example if it doesn't exist
        ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
            if [ ! -f '$REMOTE_PATH/.env' ]; then
                cp '$REMOTE_PATH/.env.example' '$REMOTE_PATH/.env' 2>/dev/null || echo 'Creating basic .env file'
                # Generate secure JWT secret
                NEW_SECRET=\$(openssl rand -hex 32)
                sed -i \"s/YOUR_SUPER_SECRET_JWT_KEY_HERE_CHANGE_IN_PRODUCTION/\$NEW_SECRET/\" '$REMOTE_PATH/.env' 2>/dev/null || echo 'Could not set JWT secret'
            fi
        "
    fi
    
    log_success "Files uploaded successfully!"
}

# Run server setup
run_server_setup() {
    log_info "Running server setup on $REMOTE_USER@$REMOTE_HOST..."
    
    # Upload and run setup script
    log_info "Uploading setup script to remote server..."
    scp -P $REMOTE_PORT /home/whale-d/fajar/citadel-agent/server/setup.sh "$REMOTE_USER@$REMOTE_HOST:/tmp/setup_citadel.sh"
    
    log_info "Running server setup (this may take several minutes)..."
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        chmod +x /tmp/setup_citadel.sh &&
        echo 'Starting Citadel Agent server setup...' &&
        /tmp/setup_citadel.sh &&
        echo 'Server setup completed.' &&
        rm /tmp/setup_citadel.sh
    "
    
    log_warn "Setup completed! You may need to log out and log back in to use Docker without sudo."
    if [[ -z "$ASSUME_YES" ]]; then
        read -p "Press Enter after logging back in to continue with service startup..." 
    fi
    
    log_success "Server setup completed!"
}

# Finalize installation
finalize_installation() {
    log_info "Finalizing installation on remote server..."
    
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH' &&
        # Set proper permissions
        chmod +x scripts/*.sh &&
        # Generate new JWT secret if using default
        if grep -q 'your-super-secret-jwt-key-here-change-in-production' .env; then
            NEW_SECRET=\$(openssl rand -hex 32)
            sed -i \"s/your-super-secret-jwt-key-here-change-in-production/\$NEW_SECRET/\" .env
            echo 'Generated new JWT secret'
        fi &&
        echo 'Installation finalized'
    "
    
    log_info "Starting services..."
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH' &&
        ./scripts/start.sh
    "
    
    log_info "Waiting for services to start (allowing time for initialization)..."
    sleep 30
    
    log_info "Checking service status..."
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH' &&
        ./scripts/status.sh
    "
    
    log_success "Citadel Agent is now running!"
}

# Display access information
display_access_info() {
    echo
    echo "################################################################################"
    echo "#                                                                              #"
    echo "#                    🎉 CITADEL AGENT DEPLOYMENT COMPLETE 🎉                 #"
    echo "#                                                                              #"
    echo "################################################################################"
    echo
    echo "📋 Access Information:"
    echo "   🌐 Web Interface: https://$REMOTE_HOST"
    echo "   📡 API Endpoint: https://$REMOTE_HOST (port 443) or http://$REMOTE_HOST:5001"
    echo "   🛠️  Health Check: https://$REMOTE_HOST/health"
    echo "   📂 Installation: $REMOTE_PATH"
    echo
    echo "🔧 Management Commands (on remote server):"
    echo "   cd $REMOTE_PATH"
    echo "   ./scripts/status.sh    # Check service status"
    echo "   ./scripts/start.sh     # Start services"
    echo "   ./scripts/stop.sh      # Stop services"
    echo "   ./scripts/restart.sh   # Restart services"
    echo
    echo "🔐 Security Notes:"
    echo "   1. Change JWT_SECRET in $REMOTE_PATH/.env to a strong value"
    echo "   2. Configure SSL with Let's Encrypt: sudo certbot --nginx -d $REMOTE_HOST"
    echo "   3. Update firewall rules as needed"
    echo "   4. Regularly update your system and application"
    echo
    echo "📈 Monitoring:"
    echo "   - Check status: $REMOTE_PATH/scripts/status.sh"
    echo "   - View logs: docker-compose -f $REMOTE_PATH/docker/docker-compose.yml logs -f"
    echo
    echo "💡 SSL Setup (if using a domain):"
    echo "   1. Point your domain DNS to this server"
    echo "   2. Run: sudo certbot --nginx -d yourdomain.com"
    echo "   3. Update Nginx configuration as needed"
    echo
    echo "⚠️  Important: Remember to secure your server and change default passwords."
    echo
    echo "################################################################################"
}

# Main function based on mode
case $MODE in
    "full")
        log_info "Running full deployment (upload + setup)..."
        validate_prerequisites
        test_ssh_connection
        upload_files
        run_server_setup
        finalize_installation
        display_access_info
        ;;
    "upload")
        log_info "Running upload-only deployment..."
        validate_prerequisites
        test_ssh_connection
        upload_files
        log_success "Upload completed! Files are in $REMOTE_PATH on $REMOTE_HOST"
        ;;
    "setup")
        log_info "Running server setup only..."
        validate_prerequisites
        test_ssh_connection
        run_server_setup
        log_success "Server setup completed! You can now start services manually."
        ;;
    *)
        log_error "Invalid mode: $MODE"
        usage
        ;;
esac

log_success "Deployment $MODE completed successfully!"