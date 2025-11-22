# Complete Citadel Agent Architecture

## Overview

Citadel Agent is a comprehensive workflow automation and orchestration system designed with a modular, scalable, and robust architecture. The system combines three core technologies as recommended:

1. **Temporal.io** - For reliable workflow orchestration and state management
2. **go-plugin** - For modular, isolated node execution 
3. **Fiber** - For high-performance API server and RESTful endpoints

## Architecture Components

### 1. Core Engine Layer

#### a) Temporal Integration
- **Location**: `/backend/internal/temporal/`
- **Purpose**: Provides distributed workflow orchestration with built-in retry, timeout, and failure recovery mechanisms
- **Components**:
  - `client.go` - Temporal server communication layer
  - `workflows.go` - Main workflow definitions and orchestration logic
  - `activities.go` - Individual node execution activities
  - `service.go` - Service layer bridging Citadel Agent and Temporal
  - `plugin_integration.go` - Integration between plugins and Temporal workflows
  - `config.go` - Temporal configuration and validation
  - `example.go` - Usage examples

#### b) Plugin System
- **Location**: `/backend/internal/plugins/`
- **Purpose**: Enables modular, isolated node execution with support for external code
- **Components**:
  - `node_plugin.go` - Plugin interface and RPC server/client
  - `node_manager.go` - Plugin lifecycle management
  - `node_adapter.go` - Adapter for local nodes to plugin interface
  - `registry.go` - Plugin-aware node registry
  - `engine_adapter.go` - Engine supporting both local and plugin nodes
  - `example_usage.go` - Usage examples

### 2. API Layer

#### Fiber API Server
- **Location**: `/backend/internal/api/`
- **Purpose**: High-performance RESTful API server with comprehensive endpoints
- **Components**:
  - `server.go` - Server configuration and middleware setup
  - `workflow_handlers.go` - Workflow management endpoints
  - `node_handlers.go` - Node type management
  - `plugin_handlers.go` - Plugin management endpoints
  - `example.go` - Usage examples
  - `cmd/api/main.go` - Application entry point

### 3. Data Layer

#### Models and Interfaces
- **Location**: `/backend/internal/models/`, `/backend/internal/interfaces/`
- **Purpose**: Data structures and interface contracts
- **Components**:
  - Updated Execution model with all required fields
  - NodeInstance interface for consistent node implementation
  - Configuration schemas

## System Integration

### Workflow Flow
1. **API Request**: Client makes request to Fiber API server
2. **Validation**: Request validated by API server
3. **Orchestration**: Temporal workflow created/scheduled 
4. **Node Execution**: Individual nodes executed as Temporal activities
5. **Plugin Integration**: Plugin nodes executed in isolated processes
6. **Result Aggregation**: Results collected and returned via API

### Plugin Flow
1. **Registration**: Plugins registered with PluginManager
2. **Workflow Integration**: Available for use in Temporal workflows
3. **Isolated Execution**: Each plugin runs in separate process
4. **Communication**: RPC-based communication with main process
5. **Result**: Execution results returned to workflow engine

## Key Features

### 1. Scalability
- Horizontal scaling through Temporal cluster
- Plugin system allows adding functionality without engine restart
- Fiber's performance handles high throughput

### 2. Reliability 
- Temporal provides fault tolerance and retry mechanisms
- Circuit breakers and failure isolation
- Graceful degradation on partial failures

### 3. Modularity
- Plugin system allows custom node types
- Loose coupling between components
- Easy extension and customization

### 4. Performance
- Fiber provides high-performance API layer
- Fasthttp-based for maximum throughput
- Optimized memory usage

### 5. Observability
- Comprehensive logging
- Health check endpoints
- Workflow execution tracking
- Plugin status monitoring

## API Endpoints

### Workflows
```
GET    /api/v1/workflows           # List all workflows
POST   /api/v1/workflows           # Create new workflow
GET    /api/v1/workflows/:id       # Get workflow details
POST   /api/v1/workflows/:id/execute  # Execute workflow
GET    /api/v1/workflows/:id/status  # Get execution status
POST   /api/v1/workflows/:id/cancel  # Cancel execution
POST   /api/v1/workflows/:id/terminate # Terminate execution
```

### Nodes
```
GET    /api/v1/nodes               # List all node types
GET    /api/v1/nodes/plugins       # List plugin nodes
```

### Plugins
```
GET    /api/v1/plugins             # List all plugins
POST   /api/v1/plugins/register    # Register new plugin
GET    /api/v1/plugins/:id         # Get plugin details
POST   /api/v1/plugins/:id/execute # Execute plugin
DELETE /api/v1/plugins/:id         # Unregister plugin
```

### Engine
```
GET    /api/v1/engine/status       # Engine status
GET    /api/v1/engine/stats        # Engine statistics
```

### System
```
GET    /health                     # Health check
```

## Deployment Architecture

### Production Setup
```
[Load Balancer]
    ↓
[Multiple API Instances (Fiber)]
    ↓
[Temporal Cluster]
    ↓
[Multiple Worker Nodes (for Temporal)]
    ↓
[Plugin Processes (managed by PluginManager)]
```

### Configuration
- **Temporal**: Distributed workflow orchestration
- **API Server**: Load-balanced Fiber instances
- **Plugins**: Dynamically loaded processes
- **Database**: For persistent data storage (not implemented here but extensible)

## Security Considerations

### API Security
- Authentication/Authorization should be implemented per deployment
- Input validation for all endpoints
- Rate limiting can be added via Fiber middleware
- Plugin execution isolation

### Plugin Security
- Plugin execution in separate processes
- Limited system access by default
- Configuration validation required

## Migration Path

### From Legacy to New Architecture
1. **Phase 1**: Deploy new system alongside legacy
2. **Phase 2**: Migrate workflows gradually to use Temporal
3. **Phase 3**: Replace API layer with Fiber
4. **Phase 4**: Convert nodes to plugins for better isolation

## Extensibility Points

### Adding New Node Types
1. Create plugin implementation
2. Register with PluginManager
3. Available in API automatically

### Custom Workflows
1. Define workflow structure
2. Register with Temporal service
3. Execute via API endpoints

### Middleware Extensions
1. Add Fiber middleware
2. Extend API server configuration
3. Apply to global or specific routes

## Performance Benchmarks

### Expected Performance
- Fiber API server: Thousands of requests/second
- Temporal workflows: Distributed execution scaling horizontally
- Plugin system: Isolated execution without affecting main process
- Memory usage: Optimized through Fasthttp

## Development Guidelines

### Adding New Features
1. Follow existing pattern in `/backend/internal/` structure
2. Maintain interface contracts
3. Add corresponding API endpoints
4. Include proper error handling
5. Add documentation

### Plugin Development
1. Implement `NodePlugin` interface
2. Include proper metadata and schema
3. Test with plugin manager
4. Register with system
5. Use via API endpoints

This architecture provides a solid foundation for a scalable, maintainable, and robust workflow automation system that can grow with changing requirements while maintaining reliability and performance.