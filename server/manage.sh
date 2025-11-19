#!/bin/bash

# Citadel Agent Server Management Script
# This script provides commands for managing Citadel Agent after deployment

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

# Default paths
INSTALL_PATH="/opt/citadel-agent"
SERVICE_NAME="citadel-agent"

# Usage function
usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo
    echo "Commands:"
    echo "  status                     Check service status"
    echo "  start                      Start all services"
    echo "  stop                       Stop all services"
    echo "  restart                    Restart all services"
    echo "  logs [service]             View service logs (api, worker, scheduler, or all)"
    echo "  update                     Update to latest version from repo"
    echo "  backup                     Create backup of data and configs"
    echo "  restore <backup-file>      Restore from backup file"
    echo "  cleanup                    Clean up old Docker images and containers"
    echo "  backup-db                  Backup database only"
    echo "  monitor                    Monitor system resources"
    echo "  config                     Show configuration information"
    echo "  ssl [domain]               Setup SSL certificates with Certbot"
    echo "  help                       Show this help message"
    echo
    echo "Options:"
    echo "  -p, --path PATH            Installation path (default: /opt/citadel-agent)"
    echo "  -f, --force                Force operation without confirmation"
    echo
    exit 1
}

# Check if running on the server where citadel-agent is installed
check_installation() {
    if [ ! -d "$INSTALL_PATH" ]; then
        log_error "Citadel Agent not found at $INSTALL_PATH"
        log_info "Check if the installation path is correct or run the setup script first."
        exit 1
    fi
    
    if [ ! -f "$INSTALL_PATH/docker/docker-compose.yml" ]; then
        log_error "Docker Compose file not found at $INSTALL_PATH/docker/docker-compose.yml"
        exit 1
    fi
}

# Get service status
get_status() {
    log_info "Getting service status for Citadel Agent..."
    
    cd $INSTALL_PATH
    ./scripts/status.sh
    
    # Also show system resource usage
    echo
    log_info "System Resources:"
    if command -v docker &> /dev/null; then
        echo "Docker container stats:"
        docker stats --no-stream | head -15
        echo
    fi
    
    # Show disk usage
    echo "Disk usage for installation:"
    du -sh $INSTALL_PATH
    echo
}

# Start services
start_services() {
    log_info "Starting Citadel Agent services..."
    
    cd $INSTALL_PATH
    
    if [ -z "$FORCE" ]; then
        read -p "Start all Citadel Agent services? [y/N]: " -n 1 -r REPLY
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled."
            exit 0
        fi
    fi
    
    ./scripts/start.sh
    log_info "Services started successfully!"
}

# Stop services
stop_services() {
    log_info "Stopping Citadel Agent services..."
    
    cd $INSTALL_PATH
    
    if [ -z "$FORCE" ]; then
        read -p "Stop all Citadel Agent services? [y/N]: " -n 1 -r REPLY
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled."
            exit 0
        fi
    fi
    
    ./scripts/stop.sh
    log_info "Services stopped successfully!"
}

# Restart services
restart_services() {
    log_info "Restarting Citadel Agent services..."
    
    stop_services
    sleep 5
    start_services
}

# View logs
view_logs() {
    local service=${1:-"all"}
    
    cd $INSTALL_PATH
    
    case $service in
        "api")
            log_info "Viewing API service logs..."
            docker-compose -f docker/docker-compose.yml logs -f api
            ;;
        "worker")
            log_info "Viewing Worker service logs..."
            docker-compose -f docker/docker-compose.yml logs -f worker
            ;;
        "scheduler")
            log_info "Viewing Scheduler service logs..."
            docker-compose -f docker/docker-compose.yml logs -f scheduler
            ;;
        "postgres")
            log_info "Viewing PostgreSQL service logs..."
            docker-compose -f docker/docker-compose.yml logs -f postgres
            ;;
        "redis")
            log_info "Viewing Redis service logs..."
            docker-compose -f docker/docker-compose.yml logs -f redis
            ;;
        "all")
            log_info "Viewing all service logs..."
            docker-compose -f docker/docker-compose.yml logs -f
            ;;
        *)
            log_error "Invalid service: $service. Available: api, worker, scheduler, postgres, redis, all"
            exit 1
            ;;
    esac
}

# Update to latest version
update_version() {
    log_info "Updating Citadel Agent to latest version..."
    
    if [ -z "$FORCE" ]; then
        read -p "This will update the codebase. Proceed? [y/N]: " -n 1 -r REPLY
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled."
            exit 0
        fi
    fi
    
    # Stop services first
    stop_services
    
    cd $INSTALL_PATH
    
    # Backup current version
    log_info "Creating backup of current version..."
    local backup_name="citadel-agent-backup-$(date +%Y%m%d-%H%M%S)"
    cp -r . "../$backup_name"
    log_info "Backup created: ../$backup_name"
    
    # Fetch latest changes
    git fetch
    git pull origin main
    
    # Pull latest Docker images
    docker-compose -f docker/docker-compose.yml pull
    
    # Restart services
    start_services
    
    log_info "Update completed successfully!"
    
    if [ -z "$FORCE" ]; then
        echo
        read -p "Remove backup? Backup is at ../$backup_name [y/N]: " -n 1 -r REPLY
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "../$backup_name"
            log_info "Backup removed."
        else
            log_info "Backup retained at: ../$backup_name"
        fi
    fi
}

# Create backup
create_backup() {
    log_info "Creating backup of Citadel Agent..."
    
    cd $INSTALL_PATH
    
    # Create backup directory
    local backup_dir="/tmp/citadel-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"
    
    # Backup configuration files
    log_info "Backing up configuration files..."
    cp -r "$INSTALL_PATH/.env" "$backup_dir/env.bak" 2>/dev/null || log_warn "No .env file found"
    cp -r "$INSTALL_PATH/docker/" "$backup_dir/docker-config/" 2>/dev/null || log_warn "No docker config found"
    
    # Attempt database backup
    log_info "Creating database backup..."
    if docker-compose -f docker/docker-compose.yml ps postgres >/dev/null 2>&1; then
        # Create backup in the container first
        docker-compose -f docker/docker-compose.yml exec postgres pg_dump -U postgres -d citadel_agent > "$backup_dir/db-backup.sql" 2>/dev/null || {
            log_warn "Database backup failed - PostgreSQL may not be running"
        }
    else
        log_warn "PostgreSQL container not found, skipping database backup"
    fi
    
    # Create the final backup archive
    local final_backup="/tmp/citadel-full-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    tar -czf "$final_backup" -C "/tmp" "$(basename $backup_dir)"
    
    # Clean up temporary directory
    rm -rf "$backup_dir"
    
    log_info "Backup created: $final_backup"
    log_info "Size: $(du -h "$final_backup" | cut -f1)"
}

# Restore from backup
restore_from_backup() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Backup file not specified. Usage: $0 restore <backup-file>"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_info "Restoring from backup: $backup_file"
    
    if [ -z "$FORCE" ]; then
        read -p "This will overwrite your current installation. Proceed? [y/N]: " -n 1 -r REPLY
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled."
            exit 0
        fi
    fi
    
    # Stop services first
    stop_services
    
    # Create temporary directory for extraction
    local temp_dir=$(mktemp -d)
    
    # Extract backup
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Find the backup directory (there should be only one)
    local backup_subdir=$(ls -1 "$temp_dir" | head -n 1)
    local backup_path="$temp_dir/$backup_subdir"
    
    cd $INSTALL_PATH
    
    # Restore configuration
    if [ -f "$backup_path/env.bak" ]; then
        cp "$backup_path/env.bak" .env
        log_info "Configuration restored."
    fi
    
    # Restore docker config
    if [ -d "$backup_path/docker-config" ]; then
        cp -r "$backup_path/docker-config/"* "docker/" 2>/dev/null || log_warn "No docker config to restore"
    fi
    
    # Restore database if exists
    if [ -f "$backup_path/db-backup.sql" ]; then
        log_info "Restoring database (this may take a few minutes)..."
        
        # Wait for PostgreSQL to be ready before restoring
        log_info "Waiting for PostgreSQL to be ready..."
        sleep 10
        
        # Restore database
        cat "$backup_path/db-backup.sql" | docker-compose -f docker/docker-compose.yml exec -T postgres psql -U postgres -d citadel_agent 2>/dev/null || {
            log_error "Database restoration failed. Check PostgreSQL is running."
        }
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    # Restart services
    start_services
    
    log_info "Restore completed!"
}

# Cleanup Docker resources
cleanup_resources() {
    log_info "Cleaning up Docker resources..."
    
    if [ -z "$FORCE" ]; then
        read -p "Remove unused Docker images, containers, and networks? [y/N]: " -n 1 -r REPLY
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled."
            exit 0
        fi
    fi
    
    # Remove stopped containers
    docker container prune -f
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused networks
    docker network prune -f
    
    # Remove unused volumes
    docker volume prune -f
    
    log_info "Docker resources cleaned up."
}

# Backup database only
backup_database() {
    log_info "Creating database backup..."
    
    cd $INSTALL_PATH
    
    local db_backup="/tmp/citadel-db-backup-$(date +%Y%m%d-%H%M%S).sql"
    
    if docker-compose -f docker/docker-compose.yml ps postgres >/dev/null 2>&1; then
        # Create backup
        docker-compose -f docker/docker-compose.yml exec postgres pg_dump -U postgres -d citadel_agent > "$db_backup"
        log_info "Database backup created: $db_backup"
        log_info "Size: $(du -h "$db_backup" | cut -f1)"
    else
        log_error "PostgreSQL container not found or not running. Cannot create database backup."
        exit 1
    fi
}

# Monitor system resources
monitor_system() {
    log_info "Monitoring Citadel Agent system resources..."
    
    echo "=== System Information ==="
    echo "Date: $(date)"
    echo "Uptime: $(uptime)"
    echo
    
    echo "=== CPU and Memory ==="
    top -bn1 | head -20
    echo
    
    echo "=== Disk Usage ==="
    df -h | grep -E '^Filesystem|/dev/'
    echo
    
    echo "=== Citadel Agent Containers ==="
    if command -v docker &> /dev/null; then
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo
        docker stats --no-stream
        echo
    fi
    
    echo "=== Citadel Agent Status ==="
    cd $INSTALL_PATH
    ./scripts/status.sh
}

# Show configuration
show_config() {
    log_info "Showing Citadel Agent configuration..."
    
    cd $INSTALL_PATH
    
    echo "=== Installation Path ==="
    echo "$INSTALL_PATH"
    echo
    
    echo "=== Environment Variables (sanitized) ==="
    if [ -f .env ]; then
        cat .env | grep -v "^#" | grep "=" | while read line; do
            KEY=$(echo "$line" | cut -d'=' -f1)
            VALUE=$(echo "$line" | cut -d'=' -f2-)
            
            # Sanitize sensitive values
            if [[ $KEY =~ .*SECRET.*|.*PASSWORD.*|.*KEY.* ]] && [ -n "$VALUE" ]; then
                echo "$KEY=***SANITIZED***"
            else
                echo "$line"
            fi
        done
    else
        echo "No .env file found"
    fi
    echo
    
    echo "=== Docker Compose Services ==="
    docker-compose -f docker/docker-compose.yml config --services
    echo
    
    echo "=== Service Status ==="
    ./scripts/status.sh
}

# Setup SSL certificates
setup_ssl() {
    local domain="$1"
    
    if [ -z "$domain" ]; then
        log_error "Domain not specified. Usage: $0 ssl <domain>"
        exit 1
    fi
    
    log_info "Setting up SSL certificate for $domain using Certbot..."
    
    # Check if certbot is available
    if ! command -v certbot &> /dev/null; then
        log_error "Certbot is not installed. Please install it first: sudo apt install certbot python3-certbot-nginx"
        exit 1
    fi
    
    # Check if nginx is running
    if ! systemctl is-active --quiet nginx; then
        log_error "Nginx is not running. Please start it: sudo systemctl start nginx"
        exit 1
    fi
    
    read -p "Setup SSL for $domain using Certbot? This will modify Nginx configuration. [y/N]: " -n 1 -r REPLY
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Operation cancelled."
        exit 0
    fi
    
    # Run certbot
    sudo certbot --nginx -d "$domain"
    
    if [ $? -eq 0 ]; then
        log_info "SSL certificate setup completed successfully!"
        log_info "Nginx configuration updated for $domain"
    else
        log_error "SSL setup failed. Check Certbot output for details."
        exit 1
    fi
}

# Main function
main() {
    COMMAND="${1:-help}"
    shift || true
    
    # Handle options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--path)
                INSTALL_PATH="$2"
                shift 2
                ;;
            -f|--force)
                FORCE="true"
                shift 1
                ;;
            -*)
                log_error "Unknown option: $1"
                usage
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Validate installation exists for commands that need it
    local commands_needing_validation=("status" "start" "stop" "restart" "logs" "update" "backup" "backup-db" "monitor" "config" "ssl")
    for cmd in "${commands_needing_validation[@]}"; do
        if [ "$COMMAND" == "$cmd" ]; then
            check_installation
            break
        fi
    done
    
    case $COMMAND in
        "status")
            get_status
            ;;
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            view_logs "$@"
            ;;
        "update")
            update_version
            ;;
        "backup")
            create_backup
            ;;
        "restore")
            restore_from_backup "$1"
            ;;
        "cleanup")
            cleanup_resources
            ;;
        "backup-db")
            backup_database
            ;;
        "monitor")
            monitor_system
            ;;
        "config")
            show_config
            ;;
        "ssl")
            setup_ssl "$1"
            ;;
        "help"|*)
            usage
            ;;
    esac
}

# Run main function
main "$@"