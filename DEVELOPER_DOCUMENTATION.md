# Citadel Agent - Developer Documentation

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Getting Started](#getting-started)
3. [Core Concepts](#core-concepts)
4. [API Reference](#api-reference)
5. [Creating Custom Nodes](#creating-custom-nodes)
6. [Building Integrations](#building-integrations)
7. [Security Best Practices](#security-best-practices)
8. [Monitoring & Observability](#monitoring--observability)
9. [Performance Optimization](#performance-optimization)
10. [Deployment Guide](#deployment-guide)

## Architecture Overview

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                        CITADEL-AGENT ARCHITECTURE               │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐│
│  │   FRONTEND  │  │   ENGINE    │  │     AI      │  │  MONIT. ││
│  │    UI       │  │   CORE      │  │   AGENTS    │  │   SYS   ││
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘│
│         │                   │                   │              │
│         ▼                   ▼                   ▼              │
│  ┌─────────────────────────────────────────────────────────────┤
│  │                  WORKFLOW ENGINE                            ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          ││
│  │  │   RUNNER    │ │  EXECUTOR   │ │ SCHEDULER   │          ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘          ││
│  └─────────────────────────────────────────────────────────────┤
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────────┤
│  │                   NODE RUNTIME                              ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          ││
│  │  │    GO       │ │  JS/TS      │ │ PYTHON      │          ││
│  │  │  RUNTIME    │ │  SANDBOX    │ │  SANDBOX    │          ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘          ││
│  └─────────────────────────────────────────────────────────────┤
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────────┤
│  │                    STORAGE LAYER                            ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          ││
│  │  │   POSTGRES  │ │    REDIS    │ │   MINIO     │          ││
│  │  │             │ │             │ │   FILES     │          ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘          ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. Workflow Engine
The central component that orchestrates workflow execution:
- **Runner**: Manages workflow lifecycle
- **Executor**: Runs nodes concurrently 
- **Scheduler**: Handles scheduled executions
- **Security Manager**: Validates runtime permissions

#### 2. Node Runtime
Isolated execution environments:
- **Go Runtime**: Native execution with resource controls
- **JavaScript Sandbox**: VM isolation with limitations
- **Python Sandbox**: Subprocess isolation
- **AI Runtime**: ML model execution environment

#### 3. Security System
Multi-layer security implementation:
- **Input Sanitization**: Prevents code injection
- **Runtime Validation**: Validates code execution
- **Resource Limiting**: CPU, memory, file, network controls
- **RBAC**: Role-based access control

## Getting Started

### Prerequisites
- Go 1.19+
- Node.js 16+
- PostgreSQL 12+
- Docker & Docker Compose

### Setup Development Environment

#### Backend Setup
```bash
# Clone the repository
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent/backend

# Install Go dependencies
go mod tidy

# Set environment variables
cp .env.example .env
# Edit .env with your database credentials

# Initialize database
go run cmd/migrate/main.go

# Run the API server
go run cmd/api/main.go
```

#### Frontend Setup
```bash
# In a new terminal
cd frontend
npm install
npm run dev
```

#### Docker Setup
```bash
# Run with Docker Compose (development)
docker-compose -f docker-compose.dev.yml up -d

# Run with Docker Compose (production)
docker-compose up -d
```

## Core Concepts

### Workflows
A workflow consists of interconnected nodes that perform operations:

```json
{
  "id": "wf_123",
  "name": "Data Processing Workflow",
  "description": "Process and transform data",
  "nodes": [
    {
      "id": "start_1",
      "type": "start",
      "config": {}
    },
    {
      "id": "http_1", 
      "type": "http_request",
      "config": {
        "method": "GET",
        "url": "https://api.example.com/data"
      }
    },
    {
      "id": "transform_1",
      "type": "data_transform",
      "config": {
        "operation": "json_path",
        "path": "data.items"
      }
    }
  ],
  "connections": [
    {
      "source": "start_1",
      "target": "http_1"
    },
    {
      "source": "http_1",
      "target": "transform_1"
    }
  ]
}
```

### Node Types
Nodes are categorized by functionality:

1. **Basic Nodes**: Core operations
2. **Intermediate Nodes**: Processing and transformation
3. **Advanced Nodes**: Complex integrations
4. **Elite Nodes**: AI and advanced features

### Security Levels
- **Basic**: Input validation and basic runtime checks
- **Advanced**: Resource limiting and network isolation  
- **Enterprise**: Full sandboxing and audit logging

## API Reference

### Authentication
All API endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

### Common Response Format
```json
{
  "success": true,
  "data": {},
  "message": "Operation successful",
  "timestamp": 1634567890
}
```

### Workflow Endpoints

#### Create Workflow
`POST /api/v1/workflows`

Request Body:
```json
{
  "name": "My Workflow",
  "description": "Description of workflow",
  "nodes": [...],
  "connections": [...]
}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "wf_123",
    "name": "My Workflow",
    "status": "draft",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Execute Workflow
`POST /api/v1/workflows/{id}/execute`

Request Body:
```json
{
  "input": {
    "param1": "value1",
    "param2": "value2"
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "execution_id": "exec_456",
    "status": "running",
    "started_at": "2023-01-01T00:00:00Z"
  }
}
```

### Node Endpoints

#### Get Available Node Types
`GET /api/v1/nodes/types`

Response:
```json
{
  "success": true,
  "data": [
    {
      "type": "http_request",
      "category": "integration",
      "name": "HTTP Request",
      "description": "Make HTTP requests",
      "config_schema": {...}
    }
  ]
}
```

#### Test Node Configuration
`POST /api/v1/nodes/test`

Request Body:
```json
{
  "type": "http_request",
  "config": {
    "url": "https://httpbin.org/get",
    "method": "GET"
  }
}
```

## Creating Custom Nodes

### Node Interface
All nodes must implement the `NodeInstance` interface:

```go
type NodeInstance interface {
  Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}
```

### Example Custom Node

#### 1. Define Node Configuration
```go
type CustomNodeConfig struct {
  APIKey    string                 `json:"api_key"`
  Endpoint  string                 `json:"endpoint"`
  Method    string                 `json:"method"`
  Headers   map[string]string      `json:"headers"`
  Payload   map[string]interface{} `json:"payload"`
}
```

#### 2. Create Node Implementation
```go
type CustomNode struct {
  config *CustomNodeConfig
}

func NewCustomNode(config *CustomNodeConfig) *CustomNode {
  return &CustomNode{
    config: config,
  }
}

func (cn *CustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
  // Override config with inputs if provided
  endpoint := cn.config.Endpoint
  if ep, exists := inputs["endpoint"]; exists {
    if epStr, ok := ep.(string); ok {
      endpoint = epStr
    }
  }

  // Perform the node operation
  // This is where you implement your custom logic
  
  result := map[string]interface{}{
    "success": true,
    "data":    "Custom operation completed",
    "inputs":  inputs,
  }

  return result, nil
}
```

#### 3. Register Node Type
```go
func RegisterCustomNode(registry *engine.NodeRegistry) {
  registry.RegisterNodeType("custom_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
    var apikey string
    if key, exists := config["api_key"]; exists {
      if keyStr, ok := key.(string); ok {
        apikey = keyStr
      }
    }
    
    var endpoint string
    if ep, exists := config["endpoint"]; exists {
      if epStr, ok := ep.(string); ok {
        endpoint = epStr
      }
    }
    
    nodeConfig := &CustomNodeConfig{
      APIKey:   apikey,
      Endpoint: endpoint,
    }

    return NewCustomNode(nodeConfig), nil
  })
}
```

### Node Validation
Always implement input validation in your node:

```go
func (cn *CustomNode) validateConfig() error {
  if cn.config.APIKey == "" {
    return errors.New("API key is required")
  }
  
  if cn.config.Endpoint == "" {
    return errors.New("endpoint is required")
  }
  
  // Validate URL format
  _, err := url.ParseRequestURI(cn.config.Endpoint)
  if err != nil {
    return fmt.Errorf("invalid endpoint URL: %w", err)
  }
  
  return nil
}
```

## Building Integrations

### GitHub Integration Node
```go
// backend/internal/nodes/integrations/github_node.go
type GitHubNodeConfig struct {
  GithubToken string `json:"github_token"`
  Repository  string `json:"repository"` // owner/repo format
  Endpoint    string `json:"endpoint"`   // API endpoint
  Method      string `json:"method"`     // GET, POST, etc.
}

// Implementation handles GitHub API calls with authentication
```

### Slack Integration Node
```go
// backend/internal/nodes/integrations/slack_node.go
type SlackNodeConfig struct {
  SlackToken string   `json:"slack_token"`
  WebhookURL string   `json:"webhook_url"`
  Channel    string   `json:"channel"`
  Username   string   `json:"username"`
  // ... additional configuration
}
```

### Email Integration Node
```go
// backend/internal/nodes/integrations/email_node.go
type EmailNodeConfig struct {
  SMTPServer   string   `json:"smtp_server"`
  SMTPPort     int      `json:"smtp_port"`
  SMTPUsername string   `json:"smtp_username"`
  SMTPPassword string   `json:"smtp_password"`
  FromAddress  string   `json:"from_address"`
  Recipients   []string `json:"recipients"`
  Subject      string   `json:"subject"`
  Body         string   `json:"body"`
}
```

## Security Best Practices

### Input Validation
Always validate user inputs:

```go
func validateInput(input string) error {
  // Check for dangerous patterns
  if strings.Contains(input, "<script>") {
    return errors.New("potentially dangerous input detected")
  }
  
  // Validate length
  if len(input) > 1000 {
    return errors.New("input too long")
  }
  
  return nil
}
```

### Sandbox Safety
Implement proper isolation:

```go
func executeSafely(code string) (string, error) {
  // Use timeouts
  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()
  
  // Restrict resources
  // Limit memory, CPU, network access
  
  // Execute in isolated environment
  result, err := sandbox.Execute(ctx, code)
  if err != nil {
    return "", fmt.Errorf("execution failed: %w", err)
  }
  
  return result, nil
}
```

### Authentication & Authorization
Follow these principles:

1. **JWT Tokens**: Use signed tokens with short expiration
2. **RBAC**: Implement role-based access control
3. **Rate Limiting**: Prevent abuse
4. **Input Sanitization**: Prevent injection attacks

### Secrets Management
Never hardcode secrets:

```go
// Use environment variables or secure vault
apiKey := os.Getenv("GITHUB_API_KEY")

// Or use a secrets manager
secret, err := secretManager.Get(ctx, "github-api-key")
if err != nil {
  return fmt.Errorf("failed to get secret: %w", err)
}
```

## Monitoring & Observability

### Metrics Collection
Citadel Agent provides comprehensive metrics:

- **Workflow Execution Metrics**: Duration, success rate, error rate
- **Node Execution Metrics**: Per-node type performance
- **API Request Metrics**: Response times, error rates, throughput
- **Resource Usage Metrics**: CPU, memory, goroutine count

### Tracing
OpenTelemetry tracing is implemented:

```go
// In your service functions
func (s *WorkflowService) Execute(ctx context.Context, workflowID string) error {
  ctx, span := telemetry.StartSpan(ctx, "WorkflowService.Execute")
  defer span.End()
  
  // Add attributes for tracing
  telemetry.SetAttribute(ctx, "workflow.id", workflowID)
  
  // Perform operation
  err := s.engine.ExecuteWorkflow(ctx, workflowID)
  if err != nil {
    telemetry.RecordError(ctx, err)
    return err
  }
  
  return nil
}
```

### Health Checks
Health endpoints are available:

```
GET /api/v1/health
GET /api/v1/metrics
```

### Logging
Structured logging is implemented throughout the system:

```go
// Use structured logging
log.Info().Str("workflow_id", workflowID).Msg("Workflow execution started")
log.Error().Err(err).Str("workflow_id", workflowID).Msg("Workflow execution failed")
```

## Performance Optimization

### Database Queries
- Use connection pooling
- Implement indexing strategies
- Use database connection limits
- Implement query timeouts

### Caching Strategies
- Redis for session storage
- In-memory cache for frequently accessed data
- Cache invalidation strategies

### Resource Management
- Implement request timeouts
- Use context for cancellation
- Limit concurrent operations
- Implement circuit breakers

### Memory Management
- Use sync.Pool for object reuse
- Implement proper cleanup of resources
- Monitor garbage collection

## Deployment Guide

### Production Requirements

#### Infrastructure
- Load balancer (Nginx/AWS ALB)
- SSL/TLS termination
- Database cluster
- Redis cluster
- Monitoring solution (Prometheus/Grafana)

#### Environment Variables
```bash
# Database
DB_HOST=primary-db.example.com
DB_PORT=5432
DB_USER=citadel_user
DB_PASSWORD=secure_password
DB_NAME=citadel_agent

# Redis
REDIS_HOST=redis-cluster.example.com
REDIS_PORT=6379

# Security
JWT_SECRET=very_long_and_secure_secret_key_at_least_32_characters
API_RATE_LIMIT=1000

# Monitoring
OTEL_EXPORTER_OTLP_ENDPOINT=collector.example.com:4317
```

### Docker Deployment

#### Production Docker Compose
```yaml
version: '3.8'
services:
  api:
    image: citadel-agent/api:${VERSION}
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    ports:
      - "5001:5001"
    restart: unless-stopped
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 2G
        reservations:
          cpus: '0.25'
          memory: 512M

  postgres:
    image: postgres:14-alpine
    environment:
      - POSTGRES_DB=citadel_agent
      - POSTGRES_USER=citadel_user
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment
For Kubernetes, use the provided manifests in the `deploy/k8s/` directory.

### Backup & Recovery
- Regular PostgreSQL dumps
- Redis snapshotting
- File storage backups
- Versioned configurations

## Troubleshooting

### Common Issues

#### Workflow Not Executing
1. Check node configurations
2. Verify dependencies between nodes
3. Check execution logs
4. Validate resource limits

#### Performance Problems
1. Monitor resource usage
2. Check for memory leaks
3. Optimize database queries
4. Review concurrent execution limits

#### Security Issues
1. Enable detailed logging
2. Review sandbox configurations
3. Check permission settings
4. Validate input sanitization

### Debug Mode
Run in debug mode to get detailed logs:

```bash
DEBUG=true LOG_LEVEL=debug go run cmd/api/main.go
```

## Support Resources

- [Documentation](https://citadel-agent.com/docs)
- [GitHub Issues](https://github.com/citadel-agent/citadel-agent/issues)
- [Community Forum](https://community.citadel-agent.com)
- [API Reference](https://api.citadel-agent.com/reference)
- [Video Tutorials](https://youtube.com/citadel-agent)

For enterprise support, contact [support@citadel-agent.com].