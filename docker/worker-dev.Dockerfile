# Development Dockerfile for Worker
FROM golang:1.22-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY backend/ .

# Install nodemon for auto-reload
RUN go install github.com/pilu/fresh@latest

# Command to run in development mode
CMD ["fresh", "-c", "../fresh.conf", "-b", "worker"]