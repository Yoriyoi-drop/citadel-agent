#!/bin/bash

# Citadel Agent Setup Script
# This script sets up the development environment for Citadel Agent

set -e  # Exit on any error

echo "Setting up Citadel Agent development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -o 'go[0-9]\.[0-9]*')
if [[ $(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1) != "1.21" ]]; then
    echo "Go version is too old. Please install Go 1.21 or later."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Node.js is not installed. Please install Node.js 18 or later."
    exit 1
fi

# Check Node version
NODE_VERSION=$(node --version | grep -o '[0-9]*\.[0-9]*\.[0-9]*' | cut -d. -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "Node.js version is too old. Please install Node.js 18 or later."
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

# Initialize Go modules if not already done
if [ ! -f backend/go.mod ]; then
    echo "Initializing Go module..."
    cd backend
    go mod init github.com/citadel-agent/backend
    go mod tidy
    cd ..
fi

# Install backend tools
echo "Installing backend development tools..."
go install github.com/cosmtrek/air@latest  # Live reload for Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linter

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd frontend
npm install
cd ..

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
# Citadel Agent Environment Variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=citadel
DB_PASSWORD=citadel
DB_NAME=citadel
REDIS_URL=localhost:6379
TEMPORAL_ADDRESS=localhost:7233
JWT_SECRET=change-this-to-a-secure-random-string
LOG_LEVEL=info
PORT=8080
EOF
    echo "Created .env file with default values. Please update with your specific settings."
fi

# Create required directories if they don't exist
mkdir -p logs
mkdir -p ai-models

echo "Setup completed successfully!"
echo ""
echo "To start the application:"
echo "1. Start services: docker-compose up -d"
echo "2. Run backend: cd backend && go run main.go (in another terminal)"
echo "3. Run frontend: cd frontend && npm run dev (in another terminal)"
echo ""
echo "The application will be available at:"
echo "- Backend API: http://localhost:8080"
echo "- Frontend: http://localhost:5173"
echo "- Temporal UI: http://localhost:8081"
echo "- Grafana: http://localhost:3000 (admin/admin)"
echo "- RabbitMQ: http://localhost:15672 (guest/guest)"