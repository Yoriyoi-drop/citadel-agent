#!/bin/bash

# Run script for Citadel Agent Terminal Login

echo "====================================="
echo "Running Citadel Agent Terminal Login"
echo "====================================="
echo

cd /home/whale-d/fajar/citadel-agent/backend/cmd/terminal-login

if [ -f "terminal-login" ]; then
    echo "Found existing binary, removing..."
    rm terminal-login
fi

echo "Building application..."
go build -o terminal-login main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"
echo "Running terminal login interface..."
echo

./terminal-login