#!/bin/bash

# Stop script for Citadel Agent

echo "🛑 Stopping Citadel Agent Stack..."

# Check if docker-compose.yml exists
if [ ! -f "docker/docker-compose.yml" ]; then
    echo "❌ docker-compose.yml not found in docker/ directory"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    if ! command -v docker compose &> /dev/null; then
        echo "❌ Neither 'docker-compose' nor 'docker compose' is available."
        exit 1
    fi
fi

# Stop all services
docker-compose -f docker/docker-compose.yml down

echo "✅ Citadel Agent stack stopped."