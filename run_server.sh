#!/bin/bash

# Script untuk menjalankan Citadel Agent API server
# Simpan sebagai run_server.sh

echo "Menjalankan Citadel Agent API server..."

# Pastikan kita berada di direktori backend
cd /home/whale-d/fajar/citadel-agent/backend

# Build aplikasi
echo "Building aplikasi..."
go build -o bin/api cmd/api/main.go

if [ $? -ne 0 ]; then
    echo "Build gagal"
    exit 1
fi

echo "Build berhasil"

# Menjalankan server
echo "Menjalankan server di port 5001..."
./bin/api

echo "Server telah berhenti"