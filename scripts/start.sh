#!/bin/bash

# Startup script for Citadel Agent

set -e  # Exit immediately if a command exits with a non-zero status

echo "🚀 Starting Citadel Agent Stack..."

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    if ! command -v docker compose &> /dev/null; then
        echo "❌ Neither 'docker-compose' nor 'docker compose' is available."
        exit 1
    fi
fi

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "⚠️  Port $port is already in use. Please free it and try again."
        return 1
    fi
    return 0
}

# Check if required ports are free
echo "🔍 Checking required ports..."
check_port 5001 || exit 1  # API server
check_port 3000 || exit 1  # Frontend (if exists)

echo "🔧 Bringing up services..."

# Pull latest images before starting
echo "📦 Pulling latest images..."
docker-compose -f docker/docker-compose.yml pull --ignore-pull-failures

# Start services with dependencies in mind
echo "🔄 Starting services..."
docker-compose -f docker/docker-compose.yml up -d --remove-orphans

# Wait for services to be healthy
echo "⏳ Waiting for services to become healthy..."

# Check PostgreSQL health
max_attempts=30
attempts=0
while [ $attempts -lt $max_attempts ]; do
    status=$(docker-compose -f docker/docker-compose.yml ps postgres --format "table {{.Status}}" | grep -v Status || echo "missing")
    if [[ "$status" == *"healthy"* ]]; then
        echo "✅ PostgreSQL is healthy"
        break
    elif [[ "$status" == *"missing"* ]]; then
        echo "❌ PostgreSQL container not found!"
        exit 1
    fi
    
    sleep 2
    ((attempts++))
done

if [ $attempts -eq $max_attempts ]; then
    echo "❌ PostgreSQL did not become healthy in time"
    docker-compose -f docker/docker-compose.yml logs postgres
    exit 1
fi

# Check Redis health
attempts=0
while [ $attempts -lt $max_attempts ]; do
    status=$(docker-compose -f docker/docker-compose.yml ps redis --format "table {{.Status}}" | grep -v Status || echo "missing")
    if [[ "$status" == *"healthy"* ]]; then
        echo "✅ Redis is healthy"
        break
    elif [[ "$status" == *"missing"* ]]; then
        echo "❌ Redis container not found!"
        exit 1
    fi
    
    sleep 2
    ((attempts++))
done

if [ $attempts -eq $max_attempts ]; then
    echo "❌ Redis did not become healthy in time"
    docker-compose -f docker/docker-compose.yml logs redis
    exit 1
fi

# Give some time for the databases to fully initialize
sleep 5

# Check API service health
attempts=0
while [ $attempts -lt $max_attempts ]; do
    status=$(docker-compose -f docker/docker-compose.yml ps api --format "table {{.Status}}" | grep -v Status || echo "missing")
    if [[ "$status" == *"Up"* ]]; then
        echo "✅ API service is running"
        
        # Optionally try to reach the health endpoint
        sleep 2
        if curl -sf http://localhost:5001/health >/dev/null 2>&1; then
            echo "✅ API service health check passed"
            break
        elif [ $status != *"healthy"* ]; then
            # Service is running but not healthy yet
            sleep 2
            ((attempts++))
            continue
        fi
    elif [[ "$status" == *"missing"* ]]; then
        echo "❌ API container not found!"
        docker-compose -f docker/docker-compose.yml ps
        exit 1
    fi
    
    sleep 2
    ((attempts++))
done

if [ $attempts -eq $max_attempts ]; then
    echo "❌ API service did not start properly"
    echo "📋 Last 20 lines of API logs:"
    docker-compose -f docker/docker-compose.yml logs --tail=20 api
    exit 1
fi

# Check worker service
attempts=0
while [ $attempts -lt $max_attempts ]; do
    status=$(docker-compose -f docker/docker-compose.yml ps worker --format "table {{.Status}}" | grep -v Status || echo "missing")
    if [[ "$status" == *"Up"* ]] || [[ "$status" == *"healthy"* ]]; then
        echo "✅ Worker service is running"
        break
    elif [[ "$status" == *"missing"* ]]; then
        echo "⚠️  Worker container not found (this may be OK depending on your setup)"
        break
    fi
    
    sleep 2
    ((attempts++))
done

# Check scheduler service
attempts=0
while [ $attempts -lt $max_attempts ]; do
    status=$(docker-compose -f docker/docker-compose.yml ps scheduler --format "table {{.Status}}" | grep -v Status || echo "missing")
    if [[ "$status" == *"Up"* ]] || [[ "$status" == *"healthy"* ]]; then
        echo "✅ Scheduler service is running"
        break
    elif [[ "$status" == *"missing"* ]]; then
        echo "⚠️  Scheduler container not found (this may be OK depending on your setup)"
        break
    fi
    
    sleep 2
    ((attempts++))
done

echo
echo "🎉 Citadel Agent stack is up and running!"
echo
echo "🔗 Services:"
echo "   API: http://localhost:5001 (health: http://localhost:5001/health)"
echo
echo "📊 Docker Compose Status:"
docker-compose -f docker/docker-compose.yml ps
echo
echo "💡 Tips:"
echo "   • To view logs: docker-compose -f docker/docker-compose.yml logs -f"
echo "   • To stop: docker-compose -f docker/docker-compose.yml down"
echo "   • To restart: ./scripts/start.sh"