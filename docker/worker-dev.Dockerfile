# Development Dockerfile for Worker service
FROM golang:1.21-alpine

# Set working directory
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ .

# Expose any necessary port for debugging
EXPOSE 8080

# Run the worker service
CMD ["sh", "-c", "cd /app && go run cmd/worker/main.go"]