#!/bin/bash
# Test script for Citadel Agent

set -e  # Exit on any error

echo "Running Citadel Agent tests..."

# Navigate to backend directory
cd /home/whale-d/fajar/citadel-agent/backend

# Run all tests
echo "Running unit tests..."
go test ./test/unit/... -v

echo "Running tests for internal packages..."
go test ./internal/api/... -v
go test ./internal/auth/... -v
go test ./internal/engine/... -v
go test ./internal/ai/... -v
go test ./internal/runtimes/... -v

echo "All tests completed successfully!"