# Citadel Agent - Developer Guide

## Overview

This document provides comprehensive guidance for developers extending Citadel Agent's functionality, particularly focusing on custom node development, system monitoring, and integration creation.

## Architecture Deep Dive

### Core Engine Components

The Citadel Agent engine consists of several key components:

1. **Workflow Engine** - Orchestrates workflow execution
2. **Node Runtime** - Executes individual nodes
3. **Sandbox System** - Provides secure execution environments
4. **Monitoring System** - Tracks performance and errors
5. **Integration Framework** - Handles external service connections

### Node Architecture Pattern

All nodes follow a consistent pattern:

```go
// NodeConfig defines configuration for the node
type NodeConfig struct {
    // Configuration fields
}

// Node represents the node instance
type Node struct {
    config *NodeConfig
}

// NewNode creates a new node instance
func NewNode(config *NodeConfig) *Node {
    return &Node{config: config}
}

// Execute executes the node with provided inputs
func (n *Node) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Implementation
}
```

## Creating Custom Nodes

### 1. Basic Node Template

Here's a template for creating a new node type:

```go
// backend/internal/nodes/custom/custom_node.go
package custom

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "citadel-agent/backend/internal/workflow/core/engine"
)

// CustomNodeConfig holds configuration for the custom node
type CustomNodeConfig struct {
    Name        string                 `json:"name"`
    Parameters  map[string]interface{} `json:"parameters"`
    Timeout     time.Duration          `json:"timeout"`
}

// CustomNode is the node implementation
type CustomNode struct {
    config *CustomNodeConfig
}

// NewCustomNode creates a new custom node
func NewCustomNode(config *CustomNodeConfig) *CustomNode {
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second
    }
    
    return &CustomNode{
        config: config,
    }
}

// Execute runs the node's logic
func (cn *CustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Override config with inputs if provided
    name := cn.config.Name
    if nameInput, exists := inputs["name"]; exists {
        if nameStr, ok := nameInput.(string); ok {
            name = nameStr
        }
    }
    
    // Perform the main operation
    result := map[string]interface{}{
        "success": true,
        "name":    name,
        "inputs":  inputs,
        "config":  cn.config.Parameters,
        "timestamp": time.Now().Unix(),
    }
    
    return result, nil
}

// RegisterNode registers this node type with the engine
func RegisterNode(registry *engine.NodeRegistry) {
    registry.RegisterNodeType("custom_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
        var name string
        if n, exists := config["name"]; exists {
            if nStr, ok := n.(string); ok {
                name = nStr
            }
        }
        
        var parameters map[string]interface{}
        if params, exists := config["parameters"]; exists {
            if paramsMap, ok := params.(map[string]interface{}); ok {
                parameters = paramsMap
            }
        }
        
        var timeout float64
        if t, exists := config["timeout"]; exists {
            if tFloat, ok := t.(float64); ok {
                timeout = tFloat
            }
        }
        
        nodeConfig := &CustomNodeConfig{
            Name:       name,
            Parameters: parameters,
            Timeout:    time.Duration(timeout) * time.Second,
        }
        
        return NewCustomNode(nodeConfig), nil
    })
}
```

### 2. Advanced Node with External Integration

Here's an example of a node that integrates with an external API:

```go
// backend/internal/nodes/integrations/api_node.go
package integrations

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "citadel-agent/backend/internal/workflow/core/engine"
)

// APINodeConfig holds configuration for API calls
type APINodeConfig struct {
    URL         string            `json:"url"`
    Method      string            `json:"method"`
    Headers     map[string]string `json:"headers"`
    Timeout     time.Duration     `json:"timeout"`
    AuthType    string            `json:"auth_type"` // "none", "bearer", "basic", "api_key"
    AuthConfig  map[string]string `json:"auth_config"`
}

// APINode represents an API integration node
type APINode struct {
    config *APINodeConfig
}

// NewAPINode creates a new API node
func NewAPINode(config *APINodeConfig) *APINode {
    if config.Method == "" {
        config.Method = "GET" // Default method
    }
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second // Default timeout
    }
    
    return &APINode{
        config: config,
    }
}

// Execute makes the API call
func (an *APINode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Override config with inputs
    url := an.config.URL
    if u, exists := inputs["url"]; exists {
        if uStr, ok := u.(string); ok {
            url = uStr
        }
    }
    
    method := an.config.Method
    if m, exists := inputs["method"]; exists {
        if mStr, ok := m.(string); ok {
            method = mStr
        }
    }
    
    // Prepare request body
    var body []byte
    if reqBody, exists := inputs["body"]; exists {
        var err error
        body, err = json.Marshal(reqBody)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
    }
    
    // Create request with timeout context
    reqCtx, cancel := context.WithTimeout(ctx, an.config.Timeout)
    defer cancel()
    
    req, err := http.NewRequestWithContext(reqCtx, method, url, bytes.NewBuffer(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Set headers
    req.Header.Set("Content-Type", "application/json")
    for key, value := range an.config.Headers {
        req.Header.Set(key, value)
    }
    
    // Add authentication
    if err := an.addAuthToRequest(req, inputs); err != nil {
        return nil, fmt.Errorf("failed to add auth to request: %w", err)
    }
    
    // Make the request
    client := &http.Client{Timeout: an.config.Timeout}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make API call: %w", err)
    }
    defer resp.Body.Close()
    
    // Read response
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Parse response
    var responseJSON interface{}
    if len(responseBody) > 0 {
        if err := json.Unmarshal(responseBody, &responseJSON); err != nil {
            // If JSON parsing fails, return as string
            responseJSON = string(responseBody)
        }
    }
    
    // Return result
    result := map[string]interface{}{
        "success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
        "status_code": resp.StatusCode,
        "headers":     resp.Header,
        "response":    responseJSON,
        "request": map[string]interface{}{
            "url":    url,
            "method": method,
            "body":   string(body),
        },
        "timestamp": time.Now().Unix(),
    }
    
    if resp.StatusCode >= 400 {
        result["error"] = fmt.Sprintf("API call failed with status %d", resp.StatusCode)
    }
    
    return result, nil
}

// addAuthToRequest adds authentication headers to the request
func (an *APINode) addAuthToRequest(req *http.Request, inputs map[string]interface{}) error {
    authType := an.config.AuthType
    
    // Override with input if provided
    if authInput, exists := inputs["auth_type"]; exists {
        if authStr, ok := authInput.(string); ok {
            authType = authStr
        }
    }
    
    switch authType {
    case "bearer":
        token := an.config.AuthConfig["token"]
        if tokenInput, exists := inputs["auth_token"]; exists {
            if tokenStr, ok := tokenInput.(string); ok {
                token = tokenStr
            }
        }
        req.Header.Set("Authorization", "Bearer "+token)
        
    case "basic":
        username := an.config.AuthConfig["username"]
        password := an.config.AuthConfig["password"]
        
        if userInput, exists := inputs["auth_username"]; exists {
            if userStr, ok := userInput.(string); ok {
                username = userStr
            }
        }
        
        if passInput, exists := inputs["auth_password"]; exists {
            if passStr, ok := passInput.(string); ok {
                password = passStr
            }
        }
        
        req.SetBasicAuth(username, password)
        
    case "api_key":
        key := an.config.AuthConfig["api_key"]
        header := an.config.AuthConfig["header"] // e.g., "X-API-Key"
        
        if keyInput, exists := inputs["api_key"]; exists {
            if keyStr, ok := keyInput.(string); ok {
                key = keyStr
            }
        }
        
        if headerInput, exists := inputs["api_key_header"]; exists {
            if headerStr, ok := headerInput.(string); ok {
                header = headerStr
            }
        }
        
        if header == "" {
            header = "X-API-Key"
        }
        
        req.Header.Set(header, key)
        
    case "none":
        // No authentication
        fallthrough
    default:
        // No authentication
    }
    
    return nil
}

// RegisterAPINode registers the API node type with the engine
func RegisterAPINode(registry *engine.NodeRegistry) {
    registry.RegisterNodeType("api_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
        var url string
        if u, exists := config["url"]; exists {
            if uStr, ok := u.(string); ok {
                url = uStr
            }
        }
        
        var method string
        if m, exists := config["method"]; exists {
            if mStr, ok := m.(string); ok {
                method = mStr
            }
        }
        
        var headers map[string]string
        if h, exists := config["headers"]; exists {
            if hMap, ok := h.(map[string]interface{}); ok {
                headers = make(map[string]string)
                for k, v := range hMap {
                    if vStr, ok := v.(string); ok {
                        headers[k] = vStr
                    }
                }
            }
        }
        
        var timeout float64
        if t, exists := config["timeout"]; exists {
            if tFloat, ok := t.(float64); ok {
                timeout = tFloat
            }
        }
        
        var authType string
        if at, exists := config["auth_type"]; exists {
            if atStr, ok := at.(string); ok {
                authType = atStr
            }
        }
        
        var authConfig map[string]string
        if ac, exists := config["auth_config"]; exists {
            if acMap, ok := ac.(map[string]interface{}); ok {
                authConfig = make(map[string]string)
                for k, v := range acMap {
                    if vStr, ok := v.(string); ok {
                        authConfig[k] = vStr
                    }
                }
            }
        }
        
        nodeConfig := &APINodeConfig{
            URL:        url,
            Method:     method,
            Headers:    headers,
            Timeout:    time.Duration(timeout) * time.Second,
            AuthType:   authType,
            AuthConfig: authConfig,
        }
        
        return NewAPINode(nodeConfig), nil
    })
}
```

## Integration Examples

### 1. Database Node Implementation

```go
// Database Node with secure connection pooling and query validation
type DBNodeConfig struct {
    ConnectionString string            `json:"connection_string"`
    Query           string            `json:"query"`
    Parameters      map[string]interface{} `json:"parameters"`
    Timeout         time.Duration     `json:"timeout"`
    MaxRetries      int               `json:"max_retries"`
}

func (dn *DBNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Implementation with SQL injection prevention and connection management
    query := dn.config.Query
    if q, exists := inputs["query"]; exists {
        if qStr, ok := q.(string); ok {
            query = qStr
        }
    }
    
    // Validate query against whitelist of allowed operations
    if !dn.isValidQuery(query) {
        return nil, fmt.Errorf("query contains prohibited operations")
    }
    
    // Execute with connection pooling and retry logic
    // ... implementation details
}
```

### 2. AI Node Implementation

```go
// AI Node with model management and token counting
type AINodeConfig struct {
    Model           string            `json:"model"`
    APIKey         string            `json:"api_key"`
    Prompt         string            `json:"prompt"`
    SystemPrompt   string            `json:"system_prompt"`
    Temperature    float64           `json:"temperature"`
    MaxTokens      int               `json:"max_tokens"`
    TopP           float64           `json:"top_p"`
    APIEndpoint    string            `json:"api_endpoint"`
    Timeout        time.Duration     `json:"timeout"`
}

func (an *AINode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Implementation with token counting, rate limiting, and error handling
}
```

## Security Best Practices

### 1. Input Validation

Always validate and sanitize inputs:

```go
func validateInput(input interface{}) error {
    // Check for dangerous patterns
    jsonString, err := json.Marshal(input)
    if err != nil {
        return fmt.Errorf("failed to marshal input for validation: %w", err)
    }
    
    // Check for potential code injection patterns
    dangerousPatterns := []string{
        "<script", "eval(", "exec(", "__import__", 
        "importlib", "subprocess", "os.", "sys.",
    }
    
    lowerStr := strings.ToLower(string(jsonString))
    for _, pattern := range dangerousPatterns {
        if strings.Contains(lowerStr, pattern) {
            return fmt.Errorf("input contains dangerous pattern: %s", pattern)
        }
    }
    
    return nil
}
```

### 2. Resource Limiting

Always implement resource limits:

```go
func (n *Node) safeExecute(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()
    
    // Use channels for communication with timeout
    resultChan := make(chan interface{}, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        result, err := operation()
        if err != nil {
            errorChan <- err
            return
        }
        resultChan <- result
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### 3. Secure Credential Handling

Never store credentials in plain text:

```go
func (n *Node) getSecret(ctx context.Context, secretName string) (string, error) {
    // Use secret manager or environment variable
    secret := os.Getenv(secretName)
    if secret == "" {
        // In production, use a proper secret management system
        // like HashiCorp Vault, AWS Secrets Manager, or Azure Key Vault
        return "", fmt.Errorf("secret %s not found", secretName)
    }
    
    return secret, nil
}
```

## Testing Guidelines

### Unit Tests

Every node should have comprehensive unit tests:

```go
func TestCustomNode_Execute(t *testing.T) {
    config := &CustomNodeConfig{
        Name: "test-node",
        Parameters: map[string]interface{}{
            "param1": "value1",
        },
    }
    
    node := NewCustomNode(config)
    
    inputs := map[string]interface{}{
        "input1": "test-value",
    }
    
    result, err := node.Execute(context.Background(), inputs)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result["success"].(bool))
}
```

### Integration Tests

Test nodes with real external services:

```go
func TestAPINode_Execute(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    config := &APINodeConfig{
        URL:    "https://httpbin.org/post",
        Method: "POST",
        Timeout: 10 * time.Second,
    }
    
    node := NewAPINode(config)
    
    inputs := map[string]interface{}{
        "body": map[string]interface{}{
            "test": "value",
        },
    }
    
    result, err := node.Execute(context.Background(), inputs)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, 200, result["status_code"])
}
```

## Performance Optimization

### 1. Caching

Implement caching for expensive operations:

```go
import "github.com/patrickmn/go-cache"

type CachedNode struct {
    cache *cache.Cache
    node  Node
}

func (cn *CachedNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Create cache key from inputs
    cacheKey := cn.createCacheKey(inputs)
    
    // Check cache first
    if cached, found := cn.cache.Get(cacheKey); found {
        return cached.(map[string]interface{}), nil
    }
    
    // Execute if not in cache
    result, err := cn.node.Execute(ctx, inputs)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    cn.cache.Set(cacheKey, result, cache.DefaultExpiration)
    
    return result, nil
}
```

### 2. Concurrent Execution

For nodes that can benefit from parallelism:

```go
func (n *Node) ExecuteParallel(ctx context.Context, inputs []map[string]interface{}) ([]map[string]interface{}, error) {
    results := make([]map[string]interface{}, len(inputs))
    errors := make([]error, len(inputs))
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Limit concurrent operations
    
    for i, input := range inputs {
        wg.Add(1)
        go func(index int, input map[string]interface{}) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            result, err := n.Execute(ctx, input)
            results[index] = result
            errors[index] = err
        }(i, input)
    }
    
    wg.Wait()
    
    // Check for errors
    for _, err := range errors {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

## Error Handling

### Comprehensive Error Handling

```go
import (
    "errors"
    "fmt"
)

func (n *Node) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Validate inputs
    if err := n.validateInputs(inputs); err != nil {
        return nil, fmt.Errorf("invalid inputs: %w", err)
    }
    
    // Perform operation with recovery
    defer func() {
        if r := recover(); r != nil {
            // Log the panic and return an error
            log.Error("Node execution panicked", "recover", r)
        }
    }()
    
    result, err := n.performOperation(ctx, inputs)
    if err != nil {
        // Wrap error with context
        return nil, fmt.Errorf("node operation failed: %w", err)
    }
    
    return result, nil
}
```

## Monitoring and Observability

### 1. Metrics Collection

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
)

func (n *Node) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Get metrics
    meter := otel.Meter("citadel-agent")
    nodeCounter, _ := meter.Int64Counter("node_executions")
    nodeDuration, _ := meter.Float64Histogram("node_execution_duration")
    
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        
        nodeCounter.Add(ctx, 1, metric.WithAttributes(
            attribute.String("node_type", n.GetType()),
            attribute.String("status", "success"),
        ))
        
        nodeDuration.Record(ctx, duration)
    }()
    
    // Execute operation
    result, err := n.operation(ctx, inputs)
    if err != nil {
        // Record error metric
        nodeCounter.Add(ctx, 1, metric.WithAttributes(
            attribute.String("node_type", n.GetType()),
            attribute.String("status", "error"),
        ))
        
        return nil, err
    }
    
    return result, nil
}
```

### 2. Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (n *Node) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Start a span
    tracer := otel.Tracer("citadel-agent")
    ctx, span := tracer.Start(ctx, "Node.Execute")
    defer span.End()
    
    // Set span attributes
    span.SetAttributes(
        attribute.String("node_type", n.GetType()),
        attribute.Int("input_count", len(inputs)),
    )
    
    // Execute operation
    result, err := n.operation(ctx, inputs)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    return result, nil
}
```

## Deployment Considerations

### 1. Containerization

Create an optimized Dockerfile:

```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o citadel-agent ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -g 65532 nonroot && adduser -D -u 65532 -G nonroot nonroot

WORKDIR /root/
COPY --from=builder /app/citadel-agent .
USER nonroot

EXPOSE 5001
CMD ["./citadel-agent"]
```

### 2. Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: citadel-agent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: citadel-agent
  template:
    metadata:
      labels:
        app: citadel-agent
    spec:
      containers:
      - name: api
        image: citadel-agent:latest
        ports:
        - containerPort: 5001
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        securityContext:
          runAsNonRoot: true
          runAsUser: 65532
          readOnlyRootFilesystem: true
          allowPrivilegeEscalation: false
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 5001
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 5001
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Advanced Features

### 1. Custom Node Registry

```go
// Create a custom node registry
type CustomNodeRegistry struct {
    nodes map[string]func(map[string]interface{}) (engine.NodeInstance, error)
}

func NewCustomNodeRegistry() *CustomNodeRegistry {
    registry := &CustomNodeRegistry{
        nodes: make(map[string]func(map[string]interface{}) (engine.NodeInstance, error)),
    }
    
    // Register default nodes
    registry.RegisterNodeType("custom_api", NewAPINodeFromConfig)
    registry.RegisterNodeType("custom_db", NewDBNodeFromConfig)
    
    return registry
}

func (r *CustomNodeRegistry) RegisterNodeType(name string, factory func(map[string]interface{}) (engine.NodeInstance, error)) {
    r.nodes[name] = factory
}

func (r *CustomNodeRegistry) CreateNode(nodeType string, config map[string]interface{}) (engine.NodeInstance, error) {
    factory, exists := r.nodes[nodeType]
    if !exists {
        return nil, fmt.Errorf("unknown node type: %s", nodeType)
    }
    
    return factory(config)
}
```

### 2. Dynamic Node Loading

```go
// Load nodes dynamically from plugins
func (r *CustomNodeRegistry) LoadNodePlugins(pluginDir string) error {
    files, err := ioutil.ReadDir(pluginDir)
    if err != nil {
        return fmt.Errorf("failed to read plugin directory: %w", err)
    }
    
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".so") { // Shared object files
            plugin, err := plugin.Open(filepath.Join(pluginDir, file.Name()))
            if err != nil {
                log.Printf("Failed to load plugin %s: %v", file.Name(), err)
                continue
            }
            
            registerFunc, err := plugin.Lookup("RegisterNodes")
            if err != nil {
                log.Printf("Plugin %s does not have RegisterNodes function: %v", file.Name(), err)
                continue
            }
            
            if register, ok := registerFunc.(*func(*CustomNodeRegistry)); ok {
                (*register)(r)
            }
        }
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues and Solutions

1. **Memory Leaks**: Always close resources and use context timeouts
2. **Deadlocks**: Use channels and mutexes carefully, prefer non-blocking operations
3. **Race Conditions**: Use sync primitives appropriately
4. **Security Vulnerabilities**: Validate all inputs and limit resource usage
5. **Performance Issues**: Profile and optimize hot paths

### Debugging Tips

1. Use structured logging: `log.Info().Str("field", "value").Msg("message")`
2. Implement circuit breakers for external service calls
3. Use distributed tracing to identify bottlenecks
4. Monitor resource usage with metrics
5. Implement health checks for all services

## Conclusion

This guide provides a comprehensive overview of extending Citadel Agent's functionality. Remember to always prioritize security, performance, and maintainability when developing new features. The system is designed to be modular and extensible, so take advantage of the existing patterns and abstractions to ensure consistency across the codebase.