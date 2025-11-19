#!/bin/bash

# Status script for Citadel Agent

echo "🔍 Checking Citadel Agent Stack Status..."

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

# Show status of all services
docker-compose -f docker/docker-compose.yml ps

# Check if API is responding
echo
echo "🌐 Testing API connectivity..."
if curl -sf http://localhost:5001/health >/dev/null 2>&1; then
    echo "✅ API is responding: http://localhost:5001/health"
else
    echo "⚠️  API is not responding on http://localhost:5001/health"
    echo "   This might be normal if the service just started - try again in a moment"
fi

echo
echo "💡 Tips:"
echo "   • To start: ./scripts/start.sh"
echo "   • To view logs: docker-compose -f docker/docker-compose.yml logs -f"