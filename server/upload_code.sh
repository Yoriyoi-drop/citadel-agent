#!/bin/bash

# Citadel Agent Code Upload Script
# This script packages your local Citadel Agent code and uploads it to a remote server

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

usage() {
    echo "Usage: $0 -u <username> -h <host> [OPTIONS]"
    echo "  -u, --user USERNAME      Remote server username"
    echo "  -h, --host HOST          Remote server hostname/IP"
    echo "  -p, --port PORT          SSH port (default: 22)"
    echo "  -P, --path PATH          Remote installation path (default: /opt/citadel-agent)"
    echo "  -l, --local-path PATH    Local path to Citadel Agent (default: current directory)"
    echo "  -e, --env-file FILE      Local .env file to copy to server"
    echo "  --exclude-git            Exclude .git directory from upload (default: true)"
    echo "  --include-git            Include .git directory in upload"
    echo "  -y, --yes                Skip confirmation prompts"
    echo "  --create-remote-dir      Create remote directory if it doesn't exist"
    echo "  -h, --help               Show this help"
    exit 1
}

# Default values
REMOTE_USER=""
REMOTE_HOST=""
REMOTE_PORT="22"
REMOTE_PATH="/opt/citadel-agent"
LOCAL_PATH="$(pwd)/../.."
EXCLUDE_GIT="true"
ENV_FILE=""
CREATE_REMOTE_DIR="false"

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

# Validate local path exists
if [[ ! -d "$LOCAL_PATH" ]]; then
    log_error "Local path does not exist: $LOCAL_PATH"
    exit 1
fi

# Validate local path contains Citadel Agent
if [[ ! -f "$LOCAL_PATH/backend/main.go" && ! -f "$LOCAL_PATH/cmd/api/main.go" ]]; then
    log_warn "Main Go files not found in $LOCAL_PATH. Checking if it's the correct directory..."
    if [[ ! -d "$LOCAL_PATH/backend" && ! -d "$LOCAL_PATH/docker" ]]; then
        log_error "Citadel Agent directory structure not detected in $LOCAL_PATH. Please check your local path."
        exit 1
    fi
fi

# Validate .env file if provided
if [[ -n "$ENV_FILE" && ! -f "$ENV_FILE" ]]; then
    log_error "Specified .env file does not exist: $ENV_FILE"
    exit 1
fi

# Confirmation prompt
if [[ -z "$ASSUME_YES" ]]; then
    echo
    echo "Preparing to upload Citadel Agent from:"
    echo "  Local: $LOCAL_PATH"
    echo "To remote server:"
    echo "  User: $REMOTE_USER"
    echo "  Host: $REMOTE_HOST:$REMOTE_PORT"
    echo "  Path: $REMOTE_PATH"
    if [[ -n "$ENV_FILE" ]]; then
        echo "  Env file: $ENV_FILE"
    fi
    if [[ "$EXCLUDE_GIT" == "true" ]]; then
        echo "  Excluding: .git directory"
    else
        echo "  Including: .git directory"
    fi
    echo
    read -p "Continue with upload? (y/N): " -n 1 -r REPLY
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Upload cancelled."
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

# Check and create remote directory if needed
check_and_create_remote_dir() {
    if [[ "$CREATE_REMOTE_DIR" == "true" ]]; then
        log_info "Creating remote directory $REMOTE_PATH if it doesn't exist..."
        
        ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
            if [ ! -d '$REMOTE_PATH' ]; then
                mkdir -p '$REMOTE_PATH'
                echo 'Created directory: $REMOTE_PATH'
            else
                echo 'Directory already exists: $REMOTE_PATH'
            fi
        "
    else
        # Check if remote directory exists
        dir_exists=$(ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
            if [ -d '$REMOTE_PATH' ]; then
                echo 'exists'
            else
                echo 'missing'
            fi
        " 2>/dev/null || echo "missing")
        
        if [[ "$dir_exists" == "missing" ]]; then
            log_error "Remote directory does not exist: $REMOTE_PATH"
            log_info "Use --create-remote-dir option to create it automatically"
            exit 1
        fi
    fi
}

# Create local archive
create_archive() {
    log_info "Creating local archive of Citadel Agent..."
    
    # Determine the base directory name
    BASE_NAME=$(basename "$LOCAL_PATH")
    ARCHIVE_NAME="/tmp/citadel-agent-upload-$(date +%Y%m%d_%H%M%S).tar.gz"
    
    # Build the tar command with exclusions
    TAR_CMD="tar -czf '$ARCHIVE_NAME' -C '$(dirname "$LOCAL_PATH")' '$BASE_NAME'"
    
    if [[ "$EXCLUDE_GIT" == "true" ]]; then
        # Add exclude for .git directory
        TAR_CMD="tar --exclude='$BASE_NAME/.git' -czf '$ARCHIVE_NAME' -C '$(dirname "$LOCAL_PATH")' '$BASE_NAME'"
    fi
    
    eval "$TAR_CMD"
    
    ARCHIVE_SIZE=$(du -h "$ARCHIVE_NAME" | cut -f1)
    log_info "Archive created: $ARCHIVE_NAME (Size: $ARCHIVE_SIZE)"
}

# Upload archive to remote server
upload_archive() {
    log_info "Uploading archive to remote server..."
    
    scp -P $REMOTE_PORT "$ARCHIVE_NAME" "$REMOTE_USER@$REMOTE_HOST:/tmp/"
    
    log_info "Archive uploaded successfully."
}

# Extract archive on remote server
extract_archive() {
    log_info "Extracting archive on remote server..."
    
    ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH'
        
        # Backup existing installation if present
        if [ -f .env ]; then
            echo 'Backing up existing .env file...'
            cp .env .env.backup.$(date +%Y%m%d_%H%M%S) 2>/dev/null || echo 'No existing .env to backup'
        fi
        
        # Remove existing files except backups
        echo 'Removing existing files...'
        find . -mindepth 1 -maxdepth 1 -not -name '.*' -not -name 'citadel-agent-upload*' -exec rm -rf {} + 2>/dev/null || true
        
        # Extract the new archive
        echo 'Extracting new files...'
        tar -xzf '$ARCHIVE_NAME' --strip-components=1 -C .
        
        # Make scripts executable
        chmod +x scripts/*.sh 2>/dev/null || echo 'No scripts directory to make executable'
        
        # If a .env file was backed up, preserve it
        for backup in .env.backup.*; do
            if [ -f \"\$backup\" ]; then
                echo 'Restoring previous .env file...'
                mv \"\$backup\" .env
                break
            fi
        done
        
        # Remove the archive
        rm '$ARCHIVE_NAME'
        
        echo 'Upload and extraction completed: $REMOTE_PATH'
    "
    
    log_info "Archive extracted to $REMOTE_PATH on remote server."
}

# Upload .env file if provided
upload_env_file() {
    if [[ -n "$ENV_FILE" ]]; then
        log_info "Uploading .env file..."
        
        # Upload the .env file
        scp -P $REMOTE_PORT "$ENV_FILE" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/.env"
        
        log_info ".env file uploaded to $REMOTE_PATH/.env"
    fi
}

# Verify upload
verify_upload() {
    log_info "Verifying upload on remote server..."
    
    file_count=$(ssh -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH'
        find . -maxdepth 2 -type f | wc -l
    ")
    
    if [[ "$file_count" -gt 10 ]]; then
        log_info "Upload verified successfully ($file_count files found)"
    else
        log_warn "Few files found after upload ($file_count files). Verification may be needed."
    fi
}

# Display next steps
display_next_steps() {
    log_info "🎉 Upload completed!"
    echo
    echo "📋 Next steps on remote server ($REMOTE_HOST):"
    echo "1. SSH to your server: ssh $REMOTE_USER@$REMOTE_HOST"
    echo "2. Navigate to the directory: cd $REMOTE_PATH"
    echo "3. Ensure scripts are executable: chmod +x scripts/*.sh"
    echo "4. If this is a fresh install, copy environment: cp .env.example .env"
    echo "5. Edit .env as needed: nano .env"
    echo "6. Start services: ./scripts/start.sh"
    echo
    echo "💡 Tip: Run setup script if this is the first time:"
    echo "   curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/server/setup.sh"
    echo "   chmod +x setup.sh && ./setup.sh"
}

# Main function
main() {
    log_info "🚀 Starting Citadel Agent code upload to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT"
    
    test_ssh_connection
    check_and_create_remote_dir
    create_archive
    upload_archive
    extract_archive
    upload_env_file
    verify_upload
    display_next_steps
    
    log_info "Upload process completed! Check the next steps above to finalize the installation."
}

# Run main function
main