#!/bin/bash

# Citadel Agent Code Upload Script
# Simplifies uploading Citadel Agent code to a remote server using SCP

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
    echo "Usage: $0 [OPTIONS]"
    echo "Uploads Citadel Agent source code to a remote server"
    echo
    echo "Required flags:"
    echo "  -u, --user USERNAME      Remote server username"
    echo "  -h, --host HOST          Remote server hostname/IP"
    echo
    echo "Optional flags:"
    echo "  -p, --port PORT          SSH port (default: 22)"
    echo "  -P, --path PATH          Remote installation path (default: /opt/citadel-agent-code)"
    echo "  -l, --local-path PATH    Local path to Citadel Agent source (default: current directory)"
    echo "  --exclude-git            Exclude .git directory from upload (default: true)"
    echo "  --include-git            Include .git directory in upload"
    echo "  -s, --skip-validation    Skip pre-upload validation"
    echo "  -k, --ssh-key PATH       SSH private key to use for authentication"
    echo "  -y, --yes                Skip confirmation prompts"
    echo "  -h, --help               Show this help message"
    echo
    echo "Examples:"
    echo "  $0 -u deploy -h myserver.com"
    echo "  $0 -u root -h 192.168.1.100 -P /home/deploy/citadel-agent"
    echo "  $0 -u ubuntu -h my-vm.compute.amazonaws.com --include-git -y"
    exit 1
}

# Default values
REMOTE_USER=""
REMOTE_HOST=""
REMOTE_PORT="22"
REMOTE_PATH="/opt/citadel-agent-code"
LOCAL_PATH="$(pwd)"
EXCLUDE_GIT="true"
SSH_KEY=""
SKIP_VALIDATION="false"
ASSUME_YES="false"

# Parse command line options
TEMP=$(getopt -o u:h:p:P:l:k:sy --long user:,host:,port:,path:,local-path:,exclude-git,include-git,skip-validation,ssh-key:,yes,help -n "$0" -- "$@")
eval set -- "$TEMP"

while true; do
    case "$1" in
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
        --exclude-git)
            EXCLUDE_GIT="true"
            shift
            ;;
        --include-git)
            EXCLUDE_GIT="false"
            shift
            ;;
        -s|--skip-validation)
            SKIP_VALIDATION="true"
            shift
            ;;
        -k|--ssh-key)
            SSH_KEY="-i $2"
            shift 2
            ;;
        -y|--yes)
            ASSUME_YES="true"
            shift
            ;;
        -h|--help)
            usage
            ;;
        --)
            shift
            break
            ;;
        *)
            log_error "Internal error!"
            exit 1
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
elif [[ ! -f "$LOCAL_PATH/backend/go.mod" && ! -d "$LOCAL_PATH/backend" ]]; then
    log_error "Citadel Agent source code not found in $LOCAL_PATH. Please ensure you're pointing to the correct directory."
    exit 1
fi

# Confirmation prompt
if [[ "$ASSUME_YES" == "false" ]]; then
    echo
    echo "Uploading Citadel Agent from:"
    echo "  Local: $LOCAL_PATH"
    echo "To remote server:"
    echo "  User: $REMOTE_USER"
    echo "  Host: $REMOTE_HOST:$REMOTE_PORT"
    echo "  Path: $REMOTE_PATH"
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

# Validate SSH connection
validate_connection() {
    if [[ "$SKIP_VALIDATION" == "true" ]]; then
        return 0
    fi
    
    log_info "Validating SSH connection to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT..."
    
    # Test SSH connection
    ssh $SSH_KEY -o ConnectTimeout=10 -p $REMOTE_PORT -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST 'exit 0' || {
        log_error "SSH connection failed. Please verify:"
        log_error "  - Hostname/IP address is correct"
        log_error "  - Server is reachable"
        log_error "  - SSH service is running"
        log_error "  - You have SSH access to the server"
        exit 1
    }
    
    log_info "SSH connection validated successfully"
}

# Create remote directory
create_remote_dir() {
    log_info "Creating remote directory: $REMOTE_PATH"
    
    ssh $SSH_KEY -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        mkdir -p '$REMOTE_PATH'
        chmod 755 '$REMOTE_PATH'
    "
}

# Create archive of the code
create_archive() {
    log_info "Creating code archive..."
    
    # Determine the base directory name
    BASE_NAME=$(basename "$LOCAL_PATH")
    
    # Create temporary archive name
    TEMP_ARCHIVE="/tmp/citadel-agent-source-$(date +%Y%m%d-%H%M%S).tar.gz"
    
    # Build the tar command
    if [[ "$EXCLUDE_GIT" == "true" ]]; then
        # Exclude git directory
        tar --exclude='.git' --exclude='node_modules' --exclude='*.log' --exclude='tmp/' --exclude='dist/' --exclude='*.swp' --exclude='.DS_Store' \
            -czf "$TEMP_ARCHIVE" -C "$(dirname "$LOCAL_PATH")" "$BASE_NAME"
    else
        # Include everything
        tar --exclude='node_modules' --exclude='*.log' --exclude='tmp/' --exclude='dist/' --exclude='*.swp' --exclude='.DS_Store' \
            -czf "$TEMP_ARCHIVE" -C "$(dirname "$LOCAL_PATH")" "$BASE_NAME"
    fi
    
    ARCHIVE_SIZE=$(du -h "$TEMP_ARCHIVE" | cut -f1)
    log_info "Archive created: $TEMP_ARCHIVE (Size: $ARCHIVE_SIZE)"
    
    echo "$TEMP_ARCHIVE"  # Output for use in upload step
}

# Upload archive to remote server
upload_archive() {
    local archive_path=$1
    log_info "Uploading archive to remote server..."
    
    scp $SSH_KEY -P $REMOTE_PORT "$archive_path" "$REMOTE_USER@$REMOTE_HOST:/tmp/"
    
    log_info "Archive uploaded to remote server"
}

# Extract archive on remote server
extract_on_remote() {
    local archive_filename=$(basename "$1")
    
    log_info "Extracting archive on remote server..."
    
    ssh $SSH_KEY -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        cd '$REMOTE_PATH'
        
        # Backup existing installation if present
        if [ -f .env ]; then
            echo 'Backing up existing .env file...'
            cp .env .env.backup.\$(date +%Y%m%d_%H%M%S) 2>/dev/null || echo 'No existing .env to backup'
        fi
        
        # Extract the new code
        echo 'Extracting new code...'
        tar -xzf '/tmp/$archive_filename' --strip-components=1
        
        # Restore previous .env if it existed and new one doesn't exist
        for backup in .env.backup.*; do
            if [ -f \"\$backup\" ] && [ ! -f .env ]; then
                echo 'Restoring previous .env file...'
                mv \"\$backup\" .env
                break
            fi
        done
        
        # Remove the archive
        rm '/tmp/$archive_filename'
        
        # Make scripts executable
        find . -name '*.sh' -exec chmod +x {} \; 2>/dev/null || echo 'No scripts to modify'
        
        echo 'Code extraction completed at: $REMOTE_PATH'
    "
}

# Validate uploaded code
validate_upload() {
    if [[ "$SKIP_VALIDATION" == "true" ]]; then
        return 0
    fi
    
    log_info "Validating uploaded code on remote server..."
    
    ssh $SSH_KEY -p $REMOTE_PORT $REMOTE_USER@$REMOTE_HOST "
        if [ -f '$REMOTE_PATH/backend/go.mod' ]; then
            echo '✓ go.mod file found'
        else
            echo '✗ go.mod file not found'
            exit 1
        fi
        
        if [ -d '$REMOTE_PATH/docker' ]; then
            echo '✓ docker directory found'
        else
            echo '✗ docker directory not found'
            exit 1
        fi
        
        if [ -f '$REMOTE_PATH/scripts/start.sh' ]; then
            echo '✓ start.sh script found'
        else
            echo '✗ start.sh script not found'
            exit 1
        fi
        
        echo '✓ Upload validation passed'
    "
}

# Main execution
main() {
    log_info "🚀 Starting Citadel Agent code upload to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT"

    validate_connection
    create_remote_dir

    # Create archive and get path
    ARCHIVE_PATH=$(create_archive)

    # Upload and extract
    upload_archive "$ARCHIVE_PATH"
    extract_on_remote "$ARCHIVE_PATH"
    validate_upload

    # Clean up local archive
    rm -f "$ARCHIVE_PATH"

    log_info "🎉 Citadel Agent code successfully uploaded to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"
    echo
    log_info "Next steps on the remote server:"
    echo "  1. cd $REMOTE_PATH"
    echo "  2. Copy .env.example to .env and update configuration as needed"
    echo "  3. Run ./scripts/start.sh to start the services"
    echo
    log_info "For more details, see the deployment documentation in the uploaded code"
}

# Run main
main