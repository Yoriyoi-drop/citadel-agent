#!/bin/bash

# Test script for Citadel Agent Terminal Login

echo "================================="
echo "Citadel Agent Terminal Login Test"
echo "================================="
echo

# Build the application
echo "Building the application..."
cd /home/whale-d/fajar/citadel-agent/backend/cmd/terminal-login
go build -o terminal-login main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

echo
echo "Running tests..."
go test -v

if [ $? -eq 0 ]; then
    echo
    echo "✅ All tests passed!"
else
    echo
    echo "❌ Some tests failed"
    exit 1
fi

echo
echo "To run the terminal login interface, execute:"
echo "./terminal-login"