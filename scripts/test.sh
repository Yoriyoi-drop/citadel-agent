#!/bin/bash
# Test script for Citadel Agent

set -e

echo "Running tests for Citadel Agent..."

# Run backend tests
echo "Running backend tests..."
cd backend
go test -v ./internal/...
go test -v ./test/...

cd ..

# Run frontend tests if available
if [ -d "frontend" ]; then
  echo "Running frontend tests..."
  cd frontend
  npm test -- --coverage
  cd ..
fi

# Run integration tests
if [ -d "tests" ]; then
  echo "Running integration tests..."
  cd tests
  # Placeholder for integration tests
  echo "Integration tests would run here"
  cd ..
fi

echo "All tests completed!"