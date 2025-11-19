# Use the official Golang image to create a build artifact
FROM golang:1.22-alpine AS builder

# Set destination for the binary
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o citadel-scheduler cmd/scheduler/main.go

# Start a new stage from scratch
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/citadel-scheduler .

# Create a non-root user
RUN addgroup -g 65532 nonroot &&\
    adduser -D -u 65532 -G nonroot nonroot

# Change ownership of the binary to non-root user
RUN chown nonroot:nonroot ./citadel-scheduler

# Switch to non-root user
USER nonroot

# Command to run the executable
CMD ["./citadel-scheduler"]