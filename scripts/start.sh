#!/bin/bash
# Start script for Citadel Agent

set -e

echo "Starting Citadel Agent..."

# Start backend services in background
echo "Starting backend services..."
cd backend

echo "Starting API service in background..."
./api &
API_PID=$!

echo "Starting Worker service in background..."
./worker &
WORKER_PID=$!

echo "Starting Scheduler service in background..."
./scheduler &
SCHEDULER_PID=$!

cd ..

# Start frontend if in development
echo "Starting frontend..."
if [ -d "frontend" ]; then
  cd frontend
  npm start &
  FRONTEND_PID=$!
  cd ..
fi

echo "Services started!"
echo "API: http://localhost:5001"
echo "Frontend: http://localhost:3000"

# Function to stop services
cleanup() {
    echo "Stopping services..."
    kill $API_PID $WORKER_PID $SCHEDULER_PID $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

# Wait for all processes
wait $API_PID $WORKER_PID $SCHEDULER_PID $FRONTEND_PID