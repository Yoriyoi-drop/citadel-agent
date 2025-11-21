#!/bin/bash
# Build script for Citadel Agent

set -e

echo "Building Citadel Agent..."

# Build backend services
echo "Building backend services..."
cd backend

echo "Building API service..."
go build -o ../bin/api cmd/api/main.go

echo "Building Worker service..."
go build -o ../bin/worker cmd/worker/main.go

echo "Building Scheduler service..."
go build -o ../bin/scheduler cmd/scheduler/main.go

cd ..

# Build frontend
echo "Building frontend..."
if [ -d "frontend" ]; then
  cd frontend
  npm run build
  cd ..
fi

echo "Build completed successfully!"
echo "Binaries available in bin/ directory"