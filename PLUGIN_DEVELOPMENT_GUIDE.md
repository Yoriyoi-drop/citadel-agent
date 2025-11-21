# Citadel Agent - Plugin Development Guide

## Overview

This guide details how to develop plugins for Citadel Agent. Citadel Agent provides a robust plugin architecture that allows developers to extend the system's functionality with custom nodes, services, and integrations.

## Plugin Architecture

### Core Concepts

Citadel Agent's plugin system is built around the following concepts:

1. **Node Instances**: Individual executable units within workflows
2. **Node Registry**: Centralized registration system for node types
3. **Node Interfaces**: Standardized interfaces for plugin development
4. **Security Sandboxing**: Isolated execution environments

### Node Types Classification

Nodes are classified into four tiers based on functionality and complexity:

#### Tier A: Elite Nodes
- **Purpose**: Advanced, highly complex operations
- **Examples**: AI Agents, Multi-model integrations, Advanced orchestration
- **Security**: Highest isolation level
- **Performance**: Optimized for complex operations

#### Tier B: Advanced Nodes
- **Purpose**: Feature-rich, multiple integration points
- **Examples**: Database operations, Advanced API integrations, Complex data transformations
- **Security**: High isolation with resource limits
- **Performance**: Good balance of features and performance

#### Tier C: Intermediate Nodes
- **Purpose**: Moderate complexity operations
- **Examples**: File operations, Basic API calls, Conditional logic
- **Security**: Moderate isolation
- **Performance**: Good for routine operations

#### Tier D: Basic Nodes
- **Purpose**: Simple, foundational operations
- **Examples**: Logging, Basic conditionals, Simple data manipulation
- **Security**: Standard validation
- **Performance**: Optimized for speed

## Developing Custom Nodes

### 1. Node Interface

All custom nodes must implement the `NodeInstance` interface:

```go
type NodeInstance interface {
    Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
    ValidateConfig(config map[string]interface{}) error
    GetMetadata() *NodeMetadata
    GetSchema() *NodeSchema
}
```

### 2. Node Configuration Structure

```go
type NodeConfig struct {
    Type    string                 `json:"type"`
    Name    string                 `json:"name"`
    Description string             `json:"description"`
    Params  map[string]interface{} `json:"params"`
    Timeout time.Duration          `json:"timeout"`
    RetryPolicy RetryConfig        `json:"retry_policy"`
}

type RetryConfig struct {
    MaxRetries int           `json:"max_retries"`
    Backoff    time.Duration `json:"backoff"`
    Jitter     bool          `json:"jitter"`
}
```

### 3. Basic Node Template

Here's a template for creating a new node:

```go
// my_custom_node.go
package nodes

import (
    "context"
    "fmt"
    "time"
)

// MyCustomNodeConfig represents the configuration for the node
type MyCustomNodeConfig struct {
    APIKey     string            `json:"api_key"`
    Endpoint   string            `json:"endpoint"`
    Timeout    time.Duration     `json:"timeout"`
    Parameters map[string]interface{} `json:"parameters"`
}

// MyCustomNode represents the node implementation
type MyCustomNode struct {
    config *MyCustomNodeConfig
}

// NewMyCustomNode creates a new instance of the node
func NewMyCustomNode(config *MyCustomNodeConfig) *MyCustomNode {
    return &MyCustomNode{
        config: config,
    }
}

// Execute executes the node with the given inputs
func (n *MyCustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Override config values with inputs if provided
    endpoint := n.config.Endpoint
    if ep, exists := inputs["endpoint"]; exists {
        if epStr, ok := ep.(string); ok {
            endpoint = epStr
        }
    }

    // Validate inputs
    if endpoint == "" {
        return nil, fmt.Errorf("endpoint is required")
    }

    // Prepare execution context with timeout
    execCtx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()

    // Your node's business logic goes here
    result := map[string]interface{}{
        "success":      true,
        "data":         "example result",
        "endpoint":     endpoint,
        "input_values": inputs,
        "timestamp":    time.Now().Unix(),
    }

    return result, nil
}

// ValidateConfig validates the node configuration
func (n *MyCustomNode) ValidateConfig(config map[string]interface{}) error {
    if config["endpoint"] == nil {
        return fmt.Errorf("endpoint is required in configuration")
    }
    
    if endpoint, ok := config["endpoint"].(string); ok && endpoint == "" {
        return fmt.Errorf("endpoint cannot be empty")
    }
    
    return nil
}

// GetMetadata returns node metadata
func (n *MyCustomNode) GetMetadata() *NodeMetadata {
    return &NodeMetadata{
        Type:        "my_custom_operation",
        Name:        "My Custom Node",
        Category:    "integration",
        Description: "A custom node that performs specific operations",
        Version:     "1.0.0",
        Author:      "Your Name",
        Repository:  "https://github.com/your-org/my-custom-node",
        License:     "MIT",
    }
}

// GetSchema returns the node's input/output schema
func (n *MyCustomNode) GetSchema() *NodeSchema {
    return &NodeSchema{
        Inputs: map[string]SchemaField{
            "endpoint": {
                Type:        "string",
                Required:    true,
                Description: "API endpoint URL",
            },
            "api_key": {
                Type:        "string",
                Required:    false,
                Description: "API key for authentication",
            },
            "parameters": {
                Type:        "object",
                Required:    false,
                Description: "Additional parameters",
            },
        },
        Outputs: map[string]SchemaField{
            "success": {
                Type:        "boolean",
                Description: "Whether the operation was successful",
            },
            "data": {
                Type:        "any",
                Description: "Response data from the operation",
            },
            "timestamp": {
                Type:        "integer",
                Description: "Unix timestamp of execution",
            },
        },
    }
}
```

### 4. Node Registration

After creating your node, register it with the engine:

```go
// register_my_custom_node.go

// RegisterMyCustomNode registers the custom node type with the engine
func RegisterMyCustomNode(registry *engine.NodeRegistry) {
    registry.RegisterNodeType("my_custom_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
        // Extract configuration values
        var endpoint string
        if ep, exists := config["endpoint"]; exists {
            if epStr, ok := ep.(string); ok {
                endpoint = epStr
            }
        }

        var apiKey string
        if key, exists := config["api_key"]; exists {
            if keyStr, ok := key.(string); ok {
                apiKey = keyStr
            }
        }

        var timeout float64
        if t, exists := config["timeout_seconds"]; exists {
            if tFloat, ok := t.(float64); ok {
                timeout = tFloat
            }
        }

        var parameters map[string]interface{}
        if params, exists := config["parameters"]; exists {
            if paramsMap, ok := params.(map[string]interface{}); ok {
                parameters = paramsMap
            }
        }

        // Create node configuration
        nodeConfig := &MyCustomNodeConfig{
            Endpoint:   endpoint,
            APIKey:     apiKey,
            Timeout:    time.Duration(timeout) * time.Second,
            Parameters: parameters,
        }

        // Create and return node instance
        return NewMyCustomNode(nodeConfig), nil
    })
}
```

## Security Best Practices

### 1. Input Validation
Always validate and sanitize inputs:

```go
func (n *MyCustomNode) validateInput(input interface{}) error {
    // Check for dangerous patterns
    jsonString, err := json.Marshal(input)
    if err != nil {
        return fmt.Errorf("failed to marshal input for validation: %w", err)
    }
    
    dangerousPatterns := []string{
        "<script", "eval(", "exec(", "__import__", 
        "subprocess", "os.", "sys.", "importlib.",
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
Implement resource limits:

```go
func (n *MyCustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Apply timeout from context
    ctx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()
    
    // Use channels for communication with timeout
    resultChan := make(chan map[string]interface{}, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        // Execute in goroutine
        result, err := n.performOperation(ctx, inputs)
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
func (n *MyCustomNode) getSecureCredential(ctx context.Context, credentialName string) (string, error) {
    // Use secret management system
    // For example, AWS Secrets Manager, HashiCorp Vault, etc.
    
    // In a real implementation:
    // 1. Fetch from secret manager
    // 2. Use environment variables
    // 3. Or encrypt credentials at rest
    
    credential := os.Getenv(credentialName)
    if credential == "" {
        return "", fmt.Errorf("credential %s not found", credentialName)
    }
    
    return credential, nil
}
```

## Advanced Node Capabilities

### 1. State Management
For nodes that need to maintain state:

```go
type StatefulNode struct {
    config *NodeConfig
    state  StateManager  // Interface for state persistence
}

type StateManager interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}) error
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) (map[string]interface{}, error)
}

func (n *StatefulNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Retrieve previous state
    currentState, err := n.state.Get(ctx, fmt.Sprintf("node-%s-state", n.config.Name))
    if err != nil && !errors.Is(err, ErrStateNotFound) {
        return nil, fmt.Errorf("failed to get state: %w", err)
    }
    
    // Process inputs with state
    result, newState := n.processWithState(inputs, currentState)
    
    // Save new state
    err = n.state.Set(ctx, fmt.Sprintf("node-%s-state", n.config.Name), newState)
    if err != nil {
        return nil, fmt.Errorf("failed to save state: %w", err)
    }
    
    return result, nil
}
```

### 2. Streaming Operations
For nodes that work with streams:

```go
type StreamNode struct {
    config *StreamNodeConfig
}

func (n *StreamNode) ExecuteStream(ctx context.Context, inputStream <-chan interface{}) (<-chan map[string]interface{}, <-chan error) {
    outputStream := make(chan map[string]interface{}, 100)
    errorStream := make(chan error, 10)
    
    go func() {
        defer close(outputStream)
        defer close(errorStream)
        
        for {
            select {
            case <-ctx.Done():
                return
            case input, ok := <-inputStream:
                if !ok {
                    return
                }
                
                result, err := n.processItem(ctx, input)
                if err != nil {
                    errorStream <- err
                    continue
                }
                
                select {
                case outputStream <- result:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    
    return outputStream, errorStream
}
```

### 3. Event-Based Operations
For nodes that respond to events:

```go
type EventDrivenNode struct {
    config *EventNodeConfig
    events chan Event
}

type Event struct {
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`
    Timestamp time.Time              `json:"timestamp"`
    Payload   map[string]interface{} `json:"payload"`
}

func (n *EventDrivenNode) SubscribeToEvents(ctx context.Context, eventTypes []string) error {
    // Subscribe to event bus
    return n.eventBus.Subscribe(ctx, eventTypes, n.handleEvent)
}

func (n *EventDrivenNode) handleEvent(ctx context.Context, event Event) error {
    // Process the event
    result, err := n.processEvent(ctx, event)
    if err != nil {
        return fmt.Errorf("failed to process event: %w", err)
    }
    
    // Store result if needed
    n.results[event.ID] = result
    
    return nil
}
```

## Plugin Distribution

### 1. Package Structure
A complete plugin should have the following structure:

```
my-plugin/
├── go.mod
├── go.sum
├── main.go
├── node/
│   ├── my_plugin_node.go
│   └── register.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── utils/
│       └── helpers.go
├── test/
│   └── integration_test.go
├── docs/
│   ├── README.md
│   └── configuration.md
└── examples/
    └── workflow_example.json
```

### 2. Plugin Manifest
Create a plugin manifest file:

```json
{
  "manifest_version": "1.0",
  "id": "my-awesome-plugin",
  "name": "My Awesome Plugin",
  "version": "1.0.0",
  "description": "A plugin that does awesome things",
  "author": "Your Name",
  "license": "MIT",
  "compatibility": {
    "citadel_agent_version": ">=0.1.0"
  },
  "dependencies": {
    "database": ["postgresql", "mysql"],
    "external_services": [
      "https://api.example.com"
    ]
  },
  "nodes": [
    {
      "type": "my_awesome_operation",
      "class": "Tier B: Advanced",
      "description": "Performs awesome operations"
    }
  ],
  "config_schema": {
    "properties": {
      "api_key": {
        "type": "string",
        "description": "API key for service authentication"
      }
    },
    "required": ["api_key"]
  }
}
```

### 3. Testing Your Plugin
Create comprehensive tests:

```go
func TestMyCustomNode_Execute(t *testing.T) {
    config := &MyCustomNodeConfig{
        Endpoint: "https://httpbin.org/get",
        Timeout:  5 * time.Second,
    }
    
    node := NewMyCustomNode(config)
    
    inputs := map[string]interface{}{
        "param": "test_value",
    }
    
    result, err := node.Execute(context.Background(), inputs)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result["success"].(bool))
}

func TestMyCustomNode_Execute_ValidationError(t *testing.T) {
    config := &MyCustomNodeConfig{
        Endpoint: "", // Invalid: empty endpoint
        Timeout:  5 * time.Second,
    }
    
    node := NewMyCustomNode(config)
    
    inputs := map[string]interface{}{}
    
    _, err := node.Execute(context.Background(), inputs)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "endpoint is required")
}

func TestMyCustomNode_ValidateConfig(t *testing.T) {
    node := &MyCustomNode{}
    
    validConfig := map[string]interface{}{
        "endpoint": "https://example.com",
    }
    
    err := node.ValidateConfig(validConfig)
    assert.NoError(t, err)
    
    invalidConfig := map[string]interface{}{}
    err = node.ValidateConfig(invalidConfig)
    assert.Error(t, err)
}
```

## Performance Optimization

### 1. Connection Pooling
For nodes that use external services:

```go
type ConnectionPoolNode struct {
    config *NodeConfig
    pool   *ConnectionPool // Custom connection pool
}

func (n *ConnectionPoolNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    conn, err := n.pool.Acquire(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire connection: %w", err)
    }
    defer n.pool.Release(conn)
    
    // Use the connection
    result, err := n.makeRequest(ctx, conn, inputs)
    if err != nil {
        // Handle error, possibly marking connection as unhealthy
        n.pool.MarkUnhealthy(conn)
        return nil, err
    }
    
    return result, nil
}
```

### 2. Caching Strategies
Implement caching for expensive operations:

```go
type CachedNode struct {
    config    *NodeConfig
    cache     CacheProvider
    upstream  UpstreamProvider
}

func (n *CachedNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    cacheKey := n.generateCacheKey(inputs)
    
    // Try to get from cache first
    if cached, exists := n.cache.Get(cacheKey); exists {
        return cached.(map[string]interface{}), nil
    }
    
    // If not in cache, execute upstream and cache result
    result, err := n.upstream.Execute(ctx, inputs)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    n.cache.Set(cacheKey, result, n.config.CacheTTL)
    
    return result, nil
}
```

### 3. Batch Processing
For high-volume operations:

```go
type BatchNode struct {
    config      *NodeConfig
    batchBuffer chan BatchItem
    batchSize   int
    batchTimer  *time.Timer
}

type BatchItem struct {
    ID      string
    Input   map[string]interface{}
    Result  chan<- map[string]interface{}
    Error   chan<- error
}

func (n *BatchNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    resultChan := make(chan map[string]interface{}, 1)
    errorChan := make(chan error, 1)
    
    item := BatchItem{
        ID:    uuid.New().String(),
        Input: inputs,
        Result: resultChan,
        Error:  errorChan,
    }
    
    select {
    case n.batchBuffer <- item:
        // Item queued successfully
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    // Wait for result or error
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (n *BatchNode) processBatch() {
    batch := make([]BatchItem, 0, n.batchSize)
    
    // Collect items for batch
    for len(batch) < n.batchSize {
        select {
        case item := <-n.batchBuffer:
            batch = append(batch, item)
        case <-time.After(100 * time.Millisecond):
            // Timeout reached, process available items
            break
        }
    }
    
    if len(batch) == 0 {
        return
    }
    
    // Process batch
    results, errors := n.executeBatch(context.Background(), batch)
    
    // Send results back to individual requesters
    for i, item := range batch {
        if i < len(errors) && errors[i] != nil {
            item.Error <- errors[i]
        } else if i < len(results) {
            item.Result <- results[i]
        }
    }
}
```

## Integration Examples

### 1. API Integration Node
Example of integrating with external APIs:

```go
// api_integration_node.go
package nodes

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/cenkalti/backoff/v4"
)

type APIIntegrationNode struct {
    config *APIIntegrationConfig
    client *http.Client
}

type APIIntegrationConfig struct {
    Endpoint    string            `json:"endpoint"`
    Method      string            `json:"method"`
    Headers     map[string]string `json:"headers"`
    Timeout     time.Duration     `json:"timeout"`
    RetryPolicy RetryConfig       `json:"retry_policy"`
    Auth        AuthConfig        `json:"auth"`
}

type AuthConfig struct {
    Type   string `json:"type"`  // "bearer", "basic", "api_key"
    Token  string `json:"token"`
    User   string `json:"user"`
    Pass   string `json:"pass"`
    Header string `json:"header"`  // For API key header
}

func (n *APIIntegrationNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Override config with inputs if provided
    endpoint := n.config.Endpoint
    if ep, exists := inputs["endpoint"]; exists {
        if epStr, ok := ep.(string); ok {
            endpoint = epStr
        }
    }
    
    method := n.config.Method
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
    
    // Create request with retry logic
    var result map[string]interface{}
    operation := func() error {
        req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(body))
        if err != nil {
            return backoff.Permanent(fmt.Errorf("failed to create request: %w", err))
        }
        
        // Set headers
        req.Header.Set("Content-Type", "application/json")
        for k, v := range n.config.Headers {
            req.Header.Set(k, v)
        }
        
        // Add authentication
        n.addAuthToRequest(req)
        
        // Execute request
        resp, err := n.client.Do(req)
        if err != nil {
            return fmt.Errorf("failed to execute request: %w", err)
        }
        defer resp.Body.Close()
        
        // Read response
        respBody, err := io.ReadAll(resp.Body)
        if err != nil {
            return backoff.Permanent(fmt.Errorf("failed to read response: %w", err))
        }
        
        // Parse response
        if err := json.Unmarshal(respBody, &result); err != nil {
            // If JSON parsing fails, return raw response
            result = map[string]interface{}{
                "raw_response": string(respBody),
                "status_code": resp.StatusCode,
            }
        } else {
            result["status_code"] = resp.StatusCode
        }
        
        // Return error for retryable status codes
        if resp.StatusCode >= 500 || resp.StatusCode == 429 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }
        
        // Permanent errors for 4xx codes
        if resp.StatusCode >= 400 && resp.StatusCode < 500 {
            return backoff.Permanent(fmt.Errorf("client error: %d", resp.StatusCode))
        }
        
        return nil
    }
    
    // Execute with exponential backoff
    backoffCfg := backoff.NewExponentialBackOff()
    backoffCfg.MaxElapsedTime = n.config.RetryPolicy.MaxElapsedTime
    backoffCfg.InitialInterval = n.config.RetryPolicy.Backoff
    
    if err := backoff.Retry(operation, backoffCfg); err != nil {
        return nil, fmt.Errorf("failed to execute API request after retries: %w", err)
    }
    
    result["success"] = true
    result["timestamp"] = time.Now().Unix()
    
    return result, nil
}

func (n *APIIntegrationNode) addAuthToRequest(req *http.Request) {
    switch n.config.Auth.Type {
    case "bearer":
        req.Header.Set("Authorization", "Bearer "+n.config.Auth.Token)
    case "basic":
        auth := n.config.Auth.User + ":" + n.config.Auth.Pass
        encoded := base64.StdEncoding.EncodeToString([]byte(auth))
        req.Header.Set("Authorization", "Basic "+encoded)
    case "api_key":
        header := n.config.Auth.Header
        if header == "" {
            header = "X-API-Key"
        }
        req.Header.Set(header, n.config.Auth.Token)
    }
}
```

### 2. Database Integration Node
Example of database operations:

```go
// db_integration_node.go
package nodes

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/mattn/go-sqlite3"
)

type DBIntegrationNode struct {
    config *DBIntegrationConfig
    db     *sql.DB
}

type DBIntegrationConfig struct {
    ConnectionString string    `json:"connection_string"`
    Driver           string    `json:"driver"`  // "postgres", "mysql", "sqlite3"
    MaxConns         int       `json:"max_conns"`
    MaxIdleConns     int       `json:"max_idle_conns"`
    ConnMaxLifetime  time.Duration `json:"conn_max_lifetime"`
    QueryTimeout     time.Duration `json:"query_timeout"`
    Type             string    `json:"type"`  // "select", "insert", "update", "delete"
    Query            string    `json:"query"`
    Params           []interface{} `json:"params"`
}

func (n *DBIntegrationNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    query := n.config.Query
    if q, exists := inputs["query"]; exists {
        if qStr, ok := q.(string); ok {
            query = qStr
        }
    }
    
    params := n.config.Params
    if p, exists := inputs["params"]; exists {
        if pSlice, ok := p.([]interface{}); ok {
            params = pSlice
        }
    }
    
    // Apply query timeout
    ctx, cancel := context.WithTimeout(ctx, n.config.QueryTimeout)
    defer cancel()
    
    var result interface{}
    var err error
    
    switch n.config.Type {
    case "select", "SELECT":
        result, err = n.executeSelect(ctx, query, params)
    case "insert", "INSERT":
        result, err = n.executeInsert(ctx, query, params)
    case "update", "UPDATE":
        result, err = n.executeUpdate(ctx, query, params)
    case "delete", "DELETE":
        result, err = n.executeDelete(ctx, query, params)
    default:
        return nil, fmt.Errorf("unsupported query type: %s", n.config.Type)
    }
    
    if err != nil {
        return nil, fmt.Errorf("database operation failed: %w", err)
    }
    
    return map[string]interface{}{
        "success": true,
        "result":  result,
        "query":   query,
        "type":    n.config.Type,
        "params":  params,
        "timestamp": time.Now().Unix(),
    }, nil
}

func (n *DBIntegrationNode) executeSelect(ctx context.Context, query string, params []interface{}) (interface{}, error) {
    rows, err := n.db.QueryContext(ctx, query, params...)
    if err != nil {
        return nil, fmt.Errorf("failed to execute select query: %w", err)
    }
    defer rows.Close()
    
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %w", err)
    }
    
    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range columns {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            val := values[i]
            b, ok := val.([]byte)
            if ok {
                val = string(b)
            }
            row[col] = val
        }
        
        results = append(results, row)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %w", err)
    }
    
    return results, nil
}

func (n *DBIntegrationNode) executeInsert(ctx context.Context, query string, params []interface{}) (interface{}, error) {
    result, err := n.db.ExecContext(ctx, query, params...)
    if err != nil {
        return nil, fmt.Errorf("failed to execute insert query: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return nil, fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    lastInsertID, err := result.LastInsertId()
    if err != nil {
        // Not all databases support LastInsertId
        lastInsertID = -1
    }
    
    return map[string]interface{}{
        "rows_affected":    rowsAffected,
        "last_insert_id":   lastInsertID,
    }, nil
}
```

## Debugging and Testing

### 1. Debug Configuration
Enable debug mode for your node:

```go
type DebugConfig struct {
    EnableDebug   bool     `json:"enable_debug"`
    LogLevel      string   `json:"log_level"`  // "debug", "info", "warn", "error"
    LogSensitive  bool     `json:"log_sensitive"`  // Whether to log sensitive data
    TraceSpans    bool     `json:"trace_spans"`
    Profiling     bool     `json:"profiling"`
}

func (n *MyNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    if n.config.Debug.EnableDebug {
        n.logger.Info("Executing node", "inputs", n.maybeRedact(inputs))
        n.logger.Info("Node config", "config", n.maybeRedact(n.config))
    }
    
    // ... node execution logic ...
    
    if n.config.Debug.EnableDebug {
        n.logger.Info("Node execution completed", "result", n.maybeRedact(result))
    }
    
    return result, nil
}

func (n *MyNode) maybeRedact(data interface{}) interface{} {
    if n.config.Debug.LogSensitive {
        return data
    }
    
    // Implement redaction logic for sensitive fields
    
    switch v := data.(type) {
    case map[string]interface{}:
        redacted := make(map[string]interface{})
        for k, val := range v {
            if n.isSensitiveField(k) {
                redacted[k] = "[REDACTED]"
            } else {
                redacted[k] = n.maybeRedact(val)
            }
        }
        return redacted
    case []interface{}:
        redacted := make([]interface{}, len(v))
        for i, item := range v {
            redacted[i] = n.maybeRedact(item)
        }
        return redacted
    default:
        return data
    }
}

func (n *MyNode) isSensitiveField(field string) bool {
    sensitiveFields := []string{
        "password", "secret", "token", "key", "auth", "credential",
        "api_key", "access_token", "refresh_token", "private",
    }
    
    fieldLower := strings.ToLower(field)
    for _, sensitive := range sensitiveFields {
        if strings.Contains(fieldLower, sensitive) {
            return true
        }
    }
    
    return false
}
```

### 2. Mocking External Dependencies
For testing without external services:

```go
// For testing
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

// In tests
func TestMyAPIIntegrationNode(t *testing.T) {
    mockClient := &MockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            // Return mock response
            response := &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
            }
            return response, nil
        },
    }
    
    // Use mock client in node
    node := &APIIntegrationNode{client: mockClient}
    
    result, err := node.Execute(context.Background(), map[string]interface{}{})
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "success", result["message"])
}
```

## Deployment and Distribution

### 1. Docker Packaging
Create a Docker image for your plugin:

```Dockerfile
# Dockerfile.plugin
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plugin main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -g 65532 nonroot && adduser -D -u 65532 -G nonroot nonroot

WORKDIR /root/
COPY --from=builder /app/plugin .
USER nonroot

EXPOSE 8080
CMD ["./plugin"]
```

### 2. Kubernetes Deployment
Deploy your plugin as a Kubernetes service:

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-plugin
  labels:
    app: my-plugin
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-plugin
  template:
    metadata:
      labels:
        app: my-plugin
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        fsGroup: 65532
      containers:
      - name: plugin
        image: my-registry/my-plugin:latest
        ports:
        - containerPort: 8080
        env:
        - name: PLUGIN_CONFIG_PATH
          value: "/etc/plugin/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /etc/plugin
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: my-plugin-config
---
apiVersion: v1
kind: Service
metadata:
  name: my-plugin-service
spec:
  selector:
    app: my-plugin
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## Best Practices Summary

### 1. Security First
- Always validate and sanitize inputs
- Implement proper authentication and authorization
- Use isolated execution environments
- Limit resource usage
- Encrypt sensitive data in transit and at rest

### 2. Performance Optimization
- Use connection pooling where appropriate
- Implement efficient caching strategies
- Optimize database queries
- Use batch processing for high-volume operations
- Implement proper timeout and retry mechanisms

### 3. Error Handling
- Use proper error wrapping with `%w`
- Implement circuit breakers for external dependencies
- Use exponential backoff for retries
- Provide meaningful error messages to users
- Log errors appropriately without exposing sensitive information

### 4. Maintainability
- Write comprehensive unit and integration tests
- Follow consistent naming and documentation conventions
- Use configuration files for parameters
- Implement proper logging with contextual information
- Follow SOLID principles in design

This guide provides a comprehensive foundation for developing plugins for Citadel Agent. Following these patterns will ensure compatibility with the system's architecture and maintain security and performance standards.