#!/bin/bash

# Build script for Citadel Agent
set -e

echo "Building Citadel Agent binaries..."

# Create dist directory if not exists
mkdir -p dist

# Build binaries for different platforms
echo "Building API server..."
GOOS=linux GOARCH=amd64 go build -o dist/citadel-api-linux-amd64 ./backend/cmd/api
GOOS=darwin GOARCH=amd64 go build -o dist/citadel-api-darwin-amd64 ./backend/cmd/api
GOOS=windows GOARCH=amd64 go build -o dist/citadel-api-windows-amd64.exe ./backend/cmd/api

echo "Building Worker service..."
GOOS=linux GOARCH=amd64 go build -o dist/citadel-worker-linux-amd64 ./backend/cmd/worker
GOOS=darwin GOARCH=amd64 go build -o dist/citadel-worker-darwin-amd64 ./backend/cmd/worker
GOOS=windows GOARCH=amd64 go build -o dist/citadel-worker-windows-amd64.exe ./backend/cmd/worker

echo "Building Scheduler service..."
GOOS=linux GOARCH=amd64 go build -o dist/citadel-scheduler-linux-amd64 ./backend/cmd/scheduler
GOOS=darwin GOARCH=amd64 go build -o dist/citadel-scheduler-darwin-amd64 ./backend/cmd/scheduler
GOOS=windows GOARCH=amd64 go build -o dist/citadel-scheduler-windows-amd64.exe ./backend/cmd/scheduler

echo "Build completed! Binaries are in the 'dist' directory."