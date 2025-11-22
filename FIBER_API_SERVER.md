# Fiber API Server for Citadel Agent

## Overview
Citadel Agent now includes a high-performance API server built with Fiber, providing RESTful endpoints for workflow management, node operations, and plugin management. Fiber was chosen for its speed, low memory footprint, and Express.js-like syntax.

## Features
- **High Performance**: Built on Fasthttp for maximum speed
- **Low Memory Footprint**: Optimized for resource efficiency  
- **RESTful API**: Clean, intuitive endpoints
- **Middleware Support**: CORS, logging, recovery, request ID
- **Real-time Workflow**: Integration with Temporal for workflow orchestration
- **Plugin System**: Dynamic plugin loading and management
- **Scalability**: Designed for horizontal scaling

## Architecture
- `server.go` - Main server implementation with Fiber configuration
- `workflow_handlers.go` - Handlers for workflow operations
- `node_handlers.go` - Handlers for node management
- `plugin_handlers.go` - Handlers for plugin operations

## Endpoints

### Health Check
```
GET /health
```

### Workflows
```
GET    /api/v1/workflows           # List all workflows
POST   /api/v1/workflows           # Create a new workflow
GET    /api/v1/workflows/:id       # Get workflow details
POST   /api/v1/workflows/:id/execute  # Execute a workflow
GET    /api/v1/workflows/:id/status  # Get workflow execution status
POST   /api/v1/workflows/:id/cancel  # Cancel workflow execution
POST   /api/v1/workflows/:id/terminate  # Terminate workflow execution
```

### Nodes
```
GET    /api/v1/nodes               # List all node types
GET    /api/v1/nodes/plugins       # List plugin nodes only
```

### Plugins
```
GET    /api/v1/plugins             # List all plugins
POST   /api/v1/plugins/register    # Register a new plugin
GET    /api/v1/plugins/:id         # Get plugin details
POST   /api/v1/plugins/:id/execute # Execute a plugin
DELETE /api/v1/plugins/:id         # Unregister a plugin
```

### Engine
```
GET    /api/v1/engine/status       # Get engine status
GET    /api/v1/engine/stats        # Get engine statistics
```

## Configuration

### Environment Variables
```bash
PORT=3000                           # Server port
HOST=0.0.0.0                       # Server host
CORS_ORIGINS=*                     # CORS allowed origins
DEBUG=false                        # Enable debug mode
TEMPORAL_ADDRESS=localhost:7233    # Temporal server address
```

### Default Configuration
```go
config := &api.Config{
    Port:             "3000",           # Default port
    Host:             "0.0.0.0",        # Bind to all interfaces
    ReadTimeout:      30s,              # Read timeout
    WriteTimeout:     30s,              # Write timeout
    IdleTimeout:      60s,              # Idle timeout
    ShutdownTimeout:  10s,              # Graceful shutdown timeout
    EnableCORS:       true,             # Enable CORS middleware
    EnableLogger:     true,             # Enable request logging
    EnableRecover:    true,             # Enable panic recovery
    BasePath:         "/api/v1",        # API base path
    DebugMode:        false,            # Debug mode
    RequestIDHeader:  "X-Request-ID",   # Request ID header
}
```

## Usage Examples

### Starting the Server
```go
package main

import (
    "log"
    
    "github.com/citadel-agent/backend/internal/api"
    "github.com/citadel-agent/backend/internal/temporal"
    "github.com/citadel-agent/backend/internal/plugins"
    "github.com/citadel-agent/backend/internal/workflow/core/engine"
)

func main() {
    // Initialize components
    baseEngine := engine.NewEngine(&engine.Config{
        Parallelism: 5,
    })
    
    pluginManager := plugins.NewNodeManager()
    
    // Initialize Temporal client and service
    temporalClient, err := temporal.NewTemporalClient(&temporal.Config{
        Address:   "localhost:7233",
        Namespace: "default",
    })
    if err != nil {
        log.Fatal("Failed to create Temporal client:", err)
    }
    
    temporalService := temporal.NewTemporalWorkflowService(temporalClient, baseEngine)
    
    // Create API server
    server := api.NewServer(temporalService, pluginManager, baseEngine, nil)
    
    // Start the server
    if err := server.Start(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}
```

### Creating a Workflow via API
```bash
curl -X POST http://localhost:3000/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-workflow",
    "name": "My Workflow",
    "description": "A sample workflow",
    "nodes": [
      {
        "id": "node-1",
        "type": "http_request",
        "name": "HTTP Request",
        "config": {
          "url": "https://api.example.com",
          "method": "GET"
        }
      }
    ],
    "connections": [],
    "options": {
      "parallelism": 5,
      "timeout": 1800000000000,
      "error_handling": "continue"
    }
  }'
```

### Executing a Workflow
```bash
curl -X POST http://localhost:3000/api/v1/workflows/my-workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": {
      "input1": "value1",
      "input2": "value2"
    }
  }'
```

## Middleware
- **CORS**: Handles cross-origin resource sharing
- **Logger**: Logs all requests with configurable format
- **Recover**: Recovers from panics to prevent server crashes
- **Request ID**: Assigns unique request IDs for tracing

## Security Considerations
- All endpoints should be protected with authentication in production
- Plugin registration should be restricted to admin users
- Input validation should be implemented for all request bodies
- Rate limiting should be considered for production deployments

## Performance
- Fiber is designed to handle thousands of concurrent requests
- Optimized memory allocation reduces garbage collection
- Built on Fasthttp for maximum throughput
- Supports graceful shutdown for zero-downtime deployments