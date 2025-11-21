# Development Dockerfile for API service
FROM golang:1.21-alpine AS builder

# Install git and other dependencies needed for development
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ .

# Install fresh for hot reloading
RUN go install github.com/pilu/fresh@latest

# Final stage
FROM golang:1.21-alpine

# Install git for fresh
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy the entire backend source
COPY backend/ .

# Copy the go modules from builder
COPY --from=builder /go/pkg/mod /go/pkg/mod

# Copy the fresh binary
COPY --from=builder /go/bin/fresh /go/bin/fresh

# Expose port
EXPOSE 5001

# Install fresh in the PATH
ENV PATH="/go/bin:$PATH"

# Run the application with fresh for hot reload
CMD ["sh", "-c", "cd /app && fresh"]