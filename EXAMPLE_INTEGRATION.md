# Complete Integration Example

This document shows a complete example of how to use all components of Citadel Agent together.

## Complete Example Application

Here's a complete example showing how to set up and use the entire system:

### main.go
```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/citadel-agent/backend/internal/api"
    "github.com/citadel-agent/backend/internal/temporal"
    "github.com/citadel-agent/backend/internal/plugins"
    "github.com/citadel-agent/backend/internal/workflow/core/engine"
)

func main() {
    // Initialize base engine
    baseEngine := engine.NewEngine(&engine.Config{
        Parallelism: 10,
    })

    // Initialize plugin manager
    pluginManager := plugins.NewNodeManager()

    // Register example plugins (optional)
    // err := pluginManager.RegisterPluginAtPath("example_plugin", "./plugins/example_plugin")
    // if err != nil {
    //     log.Printf("Could not register example plugin: %v", err)
    // }

    // Initialize Temporal client
    temporalClient, err := temporal.NewTemporalClient(&temporal.Config{
        Address:   os.Getenv("TEMPORAL_ADDRESS", "localhost:7233"),
        Namespace: os.Getenv("TEMPORAL_NAMESPACE", "default"),
    })
    if err != nil {
        log.Fatal("Failed to create Temporal client:", err)
    }

    // Initialize Temporal workflow service
    workflowService := temporal.NewTemporalWorkflowService(temporalClient, baseEngine)
    
    // Register node types for compatibility
    workflowService.RegisterNodeTypes()

    // Create API server
    server := api.NewServer(workflowService, pluginManager, baseEngine, nil)

    // Set up graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // Start server in a goroutine
    go func() {
        if err := server.Start(); err != nil {
            log.Fatal("Server failed to start:", err)
        }
    }()

    log.Println("Citadel Agent API server started. Press Ctrl+C to stop.")
    
    // Wait for interrupt signal
    <-stop
    log.Println("Shutting down server...")

    // Gracefully shutdown the server
    if err := server.Shutdown(); err != nil {
        log.Fatal("Error during server shutdown:", err)
    }

    log.Println("Server stopped")
}
```

### Creating and Executing a Workflow

#### 1. Define a workflow via API:

```bash
curl -X POST http://localhost:3000/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "id": "data-processing-workflow",
    "name": "Data Processing Workflow",
    "description": "Process data with multiple steps",
    "nodes": [
      {
        "id": "http-node",
        "type": "http_request",
        "name": "Fetch Data",
        "config": {
          "url": "https://api.example.com/data",
          "method": "GET"
        },
        "options": {
          "retry_attempts": 3,
          "timeout": 30000000000,
          "retry_on_failure": true
        }
      },
      {
        "id": "transform-node", 
        "type": "data_transformer",
        "name": "Transform Data",
        "config": {
          "operation": "json_path",
          "path": "$.results[*]"
        },
        "options": {
          "retry_attempts": 2,
          "timeout": 20000000000
        }
      },
      {
        "id": "ai-node",
        "type": "ai_agent", 
        "name": "AI Analysis",
        "config": {
          "provider": "openai",
          "model": "gpt-4",
          "task": "analyze"
        },
        "options": {
          "retry_attempts": 2,
          "timeout": 60000000000
        }
      }
    ],
    "connections": [
      {
        "source_node_id": "http-node",
        "target_node_id": "transform-node"
      },
      {
        "source_node_id": "transform-node", 
        "target_node_id": "ai-node"
      }
    ],
    "options": {
      "parallelism": 3,
      "timeout": 1800000000000,
      "error_handling": "continue",
      "retry_policy": {
        "initial_interval": 1000000000,
        "backoff_coefficient": 2.0,
        "maximum_interval": 60000000000,
        "maximum_attempts": 5
      }
    }
  }'
```

#### 2. Execute the workflow:

```bash
curl -X POST http://localhost:3000/api/v1/workflows/data-processing-workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "parameters": {
      "user_id": "12345",
      "request_type": "batch_process",
      "priority": "high"
    }
  }'
```

#### 3. Check workflow status:

```bash
curl http://localhost:3000/api/v1/workflows/data-processing-workflow/status
```

### Working with Plugins

#### 1. Register a custom plugin:

```bash
curl -X POST http://localhost:3000/api/v1/plugins/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "custom_data_processor",
    "name": "Custom Data Processor",
    "description": "Custom data processing plugin",
    "path": "/path/to/custom_processor_plugin",
    "type": "custom"
  }'
```

#### 2. Execute the plugin directly:

```bash
curl -X POST http://localhost:3000/api/v1/plugins/custom_data_processor/execute \
  -H "Content-Type: application/json" \
  -d '{
    "data": "input_data",
    "options": {
      "format": "json",
      "output_path": "/tmp/result"
    }
  }'
```

### Using Plugin Nodes in Workflows

Once registered, plugin nodes can be used in workflow definitions:

```json
{
  "id": "workflow-with-plugin",
  "name": "Workflow with Plugin Node",
  "nodes": [
    {
      "id": "plugin-node",
      "type": "custom_data_processor",
      "name": "Custom Processing",
      "config": {
        "processor_param": "value"
      }
    }
  ]
}
```

## Complete System Architecture Flow

```
┌─────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   API Client    │───▶│  Fiber API       │───▶│  Temporal        │
│                 │    │  Server          │    │  Server          │
│  (HTTP/REST)    │    │  (Go Fiber)      │    │  (Cluster)       │
└─────────────────┘    └──────────────────┘    └──────────────────┘
                           │   │                    │
                           │   │                    │
                    ┌──────▼───▼────────────────────▼─────────────┐
                    │             Core System                    │
                    │                                              │
                    │  ┌─────────────────┐  ┌──────────────────┐  │
                    │  │  Workflow       │  │  Plugin          │  │
                    │  │  Service        │  │  Manager         │  │
                    │  │  (Temporal)     │  │  (go-plugin)     │  │
                    │  └─────────────────┘  └──────────────────┘  │
                    │                                              │
                    │  ┌─────────────────┐  ┌──────────────────┐  │
                    │  │  Base Engine    │  │  Node Registry   │  │
                    │  │  (Local)        │  │  (Plugin-aware)  │  │
                    │  └─────────────────┘  └──────────────────┘  │
                    └─────────────────────────────────────────────┘
                                          │
                                          │
                    ┌─────────────────────▼─────────────────────┐
                    │            Plugin Processes              │
                    │  (Isolated execution, secure sandbox)    │
                    └───────────────────────────────────────────┘
```

## Error Handling and Recovery

The system implements comprehensive error handling at multiple levels:

1. **API Layer (Fiber)**:
   - Panic recovery middleware
   - Request validation
   - Graceful error responses

2. **Workflow Layer (Temporal)**:
   - Automatic retry with configurable policies
   - Circuit breaker patterns
   - Timeout handling
   - Failure recovery

3. **Plugin Layer (go-plugin)**:
   - Process isolation
   - Graceful failure handling
   - Resource limits

## Monitoring and Observability

### Health Check
```
GET /health
Response:
{
  "status": "healthy",
  "message": "Citadel Agent API server is running",
  "timestamp": 1234567890,
  "uptime": "1h23m45s",
  "request_processing_time": "2ms"
}
```

### Engine Status
```
GET /api/v1/engine/status
Response:
{
  "engine": "temporal",
  "status": "running",
  "temporal_connected": true,
  "plugins_loaded": 3,
  "workflows_running": 2,
  "uptime": "1h23m45s",
  "timestamp": 1234567890
}
```

## Configuration

### Environment Variables
```
PORT=3000                           # API server port
HOST=0.0.0.0                        # API server host
TEMPORAL_ADDRESS=localhost:7233     # Temporal server address
TEMPORAL_NAMESPACE=default          # Temporal namespace  
CORS_ORIGINS=*                      # CORS allowed origins
DEBUG=false                         # Debug mode
```

## Development Workflow

1. **Add new node types**: Create plugin or extend base engine
2. **Define workflows**: Use API or configuration files
3. **Register plugins**: Register with plugin manager
4. **Test integration**: Execute workflows via API
5. **Monitor execution**: Check status and results

This integration example demonstrates how all components work together to provide a powerful, scalable, and maintainable workflow automation system.