# Citadel Agent - Advanced Developer Documentation

## Table of Contents

- [Introduction](#introduction)
- [Architecture Deep Dive](#architecture-deep-dive)
- [Node Development Guide](#node-development-guide)
- [AI Agent System](#ai-agent-system)
- [Workflow Engine Internals](#workflow-engine-internals)
- [Security Implementation](#security-implementation)
- [Monitoring & Observability](#monitoring--observability)
- [Performance Optimization](#performance-optimization)
- [Best Practices](#best-practices)

## Introduction

This document provides deep technical insights into Citadel Agent's architecture, development patterns, and implementation details. It serves as a comprehensive guide for developers who wish to extend the system, contribute to the core, or integrate with Citadel Agent in advanced ways.

## Architecture Deep Dive

### Core Architecture Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CITADEL-AGENT ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐        │
│  │   FRONTEND      │    │    ENGINE       │    │  AI AGENTS      │        │
│  │   (React)       │◄──►│  (Go/Fiber)     │◄──►│  (LangChain)    │        │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘        │
│         │                       │                       │                  │
│         ▼                       ▼                       ▼                  │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                         WORKFLOW ENGINE                           │  │
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐      │  │
│  │  │   RUNNER        │ │   EXECUTOR      │ │  SCHEDULER      │      │  │
│  │  │ (Workflow      │ │ (Node          │ │ (Cron/Event-   │      │  │
│  │  │  Lifecycle)    │ │  Execution)    │ │  based)       │      │  │
│  │  └─────────────────┘ └─────────────────┘ └─────────────────┘      │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                             │                                           │
│                             ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                        NODE RUNTIME                               │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  │  │
│  │  │  GO         │ │  JS/TS      │ │ PYTHON      │ │  AI/ML      │  │  │
│  │  │  NATIVE     │ │  VM         │ │  SUBPROCESS │ │  RUNTIME    │  │  │
│  │  │  RUNTIME    │ │  SANDBOX    │ │  ISOLATION  │ │  (TENSORRT) │  │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                             │                                           │
│                             ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                      SECURITY LAYER                               │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  │  │
│  │  │  RUNTIME    │ │  NETWORK    │ │   FILE      │ │  PERMISSION │  │  │
│  │  │  VALIDATOR  │ │  ISOLATION  │ │  ACCESS     │ │  CHECKER    │  │  │
│  │  │             │ │             │ │  CONTROL    │ │             │  │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                             │                                           │
│                             ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                        DATA LAYER                                 │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  │  │
│  │  │   POSTGRES  │ │    REDIS    │ │  MINIO      │ │   PROMETHEUS│  │  │
│  │  │   DATABASE  │ │   CACHE/QUEUES│ │  STORAGE    │ │   METRICS   │  │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Architecture Principles

1. **Modular Design**: Components are loosely coupled and highly cohesive
2. **Security First**: Security is implemented at every layer
3. **Performance Optimized**: Efficient algorithms and resource management
4. **Observability Built-in**: Metrics, logs, and traces from day one
5. **Extensibility**: Plugin system for custom functionality

## Node Development Guide

### Node Interface Contract

Every node in Citadel Agent must implement the Node interface:

```go
type Node interface {
    Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
    ValidateConfig(config map[string]interface{}) error
    GetMetadata() *NodeMetadata
    GetSchema() *NodeSchema
    GetDocumentation() *NodeDocumentation
}
```

### Node Execution Lifecycle

The node execution follows this sequence:

1. **Validation Phase**: Input validation against schema
2. **Preprocessing Phase**: Parameter transformation and enrichment
3. **Execution Phase**: Core business logic execution
4. **Post-processing Phase**: Result formatting and validation
5. **Output Phase**: Returning structured results

### Security Patterns in Nodes

When creating custom nodes, always implement these security patterns:

```go
func (n *MyCustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // 1. Input sanitization
    sanitizedInputs, err := n.sanitizeInputs(inputs)
    if err != nil {
        return nil, fmt.Errorf("input sanitization failed: %w", err)
    }

    // 2. Context timeout application
    execCtx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()

    // 3. Permission checking
    if err := n.checkPermissions(execCtx); err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }

    // 4. Resource limiting
    if err := n.checkResourceLimits(execCtx); err != nil {
        return nil, fmt.Errorf("resource limits exceeded: %w", err)
    }

    // 5. Execute main logic
    result, err := n.executeMainLogic(execCtx, sanitizedInputs)
    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }

    // 6. Output validation
    if err := n.validateOutput(result); err != nil {
        return nil, fmt.Errorf("output validation failed: %w", err)
    }

    return result, nil
}
```

### Creating Custom Nodes

#### Basic Node Template

```go
// my_custom_node.go
package nodes

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "citadel-agent/backend/internal/workflow/core/engine"
)

// MyCustomNodeConfig represents node-specific configuration
type MyCustomNodeConfig struct {
    APIEndpoint string `json:"api_endpoint"`
    APIToken    string `json:"api_token"`
    Timeout     time.Duration `json:"timeout"`
    RetryCount  int    `json:"retry_count"`
    // Add your specific configuration fields
}

// MyCustomNode implements the Node interface
type MyCustomNode struct {
    config *MyCustomNodeConfig
}

// NewMyCustomNode creates a new instance of MyCustomNode
func NewMyCustomNode(config *MyCustomNodeConfig) *MyCustomNode {
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second
    }
    if config.RetryCount == 0 {
        config.RetryCount = 3
    }

    return &MyCustomNode{
        config: config,
    }
}

// Execute executes the node with given inputs
func (n *MyCustomNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Override configuration with input values if specified
    endpoint := n.config.APIEndpoint
    if ep, exists := inputs["api_endpoint"]; exists {
        if epStr, ok := ep.(string); ok {
            endpoint = epStr
        }
    }

    // Validate inputs
    if err := n.validateInputs(inputs); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }

    // Create execution context
    execCtx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()

    // Perform the actual operation
    result, err := n.performOperation(execCtx, endpoint, inputs)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "success":      true,
        "result":       result,
        "input_values": inputs,
        "timestamp":    time.Now().Unix(),
        "node_type":    "my_custom_node",
    }, nil
}

// ValidateConfig validates the node configuration
func (n *MyCustomNode) ValidateConfig(config map[string]interface{}) error {
    if config["api_endpoint"] == nil {
        return fmt.Errorf("api_endpoint is required")
    }

    return nil
}

// GetMetadata provides node metadata
func (n *MyCustomNode) GetMetadata() *engine.NodeMetadata {
    return &engine.NodeMetadata{
        Type:        "my_custom_node",
        Name:        "My Custom Node",
        Category:    "integration", // "data", "logic", "integration", "ai", etc.
        Description: "A custom node that performs specialized operations",
        Version:     "1.0.0",
        Author:      "Your Name",
        License:     "MIT",
        Complexity:  "intermediate", // "basic", "intermediate", "advanced", "elite"
    }
}

// GetSchema defines input/output schemas
func (n *MyCustomNode) GetSchema() *engine.NodeSchema {
    return &engine.NodeSchema{
        Inputs: map[string]engine.SchemaField{
            "api_endpoint": {
                Type:        "string",
                Required:    false, // false because it's configurable
                Description: "API endpoint override",
                DefaultValue: n.config.APIEndpoint,
            },
            "input_data": {
                Type:        "object",
                Required:    true,
                Description: "Data to process",
            },
        },
        Outputs: map[string]engine.SchemaField{
            "success": {
                Type:        "boolean",
                Description: "Whether the operation was successful",
            },
            "result": {
                Type:        "any",
                Description: "Result of the operation",
            },
            "timestamp": {
                Type:        "integer",
                Description: "Unix timestamp of execution",
            },
        },
    }
}

// GetDocumentation provides usage documentation
func (n *MyCustomNode) GetDocumentation() *engine.NodeDocumentation {
    return &engine.NodeDocumentation{
        Usage: `This node performs specialized operations using an external API.`,
        Examples: []string{
            `{ "input_data": { "param1": "value1", "param2": "value2" } }`,
        },
        Notes: []string{
            "Requires valid API token in configuration",
            "Timeout configuration affects all API calls",
        },
    }
}

// validateInputs validates specific inputs
func (n *MyCustomNode) validateInputs(inputs map[string]interface{}) error {
    if inputs["input_data"] == nil {
        return fmt.Errorf("'input_data' is required")
    }
    
    return nil
}

// performOperation executes the main operation
func (n *MyCustomNode) performOperation(ctx context.Context, endpoint string, inputs map[string]interface{}) (interface{}, error) {
    // Implement your specific business logic here
    // For example: make HTTP requests, database queries, etc.
    
    // Return the result of the operation
    result := map[string]interface{}{
        "processed_data": inputs["input_data"],
        "operation": "custom_operation",
        "endpoint": endpoint,
    }
    
    return result, nil
}
```

#### Node Registration

After creating your custom node, register it with the engine:

```go
// register_my_custom_node.go
func RegisterMyCustomNode(registry *engine.NodeRegistry) {
    registry.RegisterNodeType("my_custom_node", func(config map[string]interface{}) (engine.NodeInstance, error) {
        // Extract configuration values
        var apiEndpoint string
        if ep, exists := config["api_endpoint"]; exists {
            if epStr, ok := ep.(string); ok {
                apiEndpoint = epStr
            }
        }

        var apiToken string
        if token, exists := config["api_token"]; exists {
            if tokenStr, ok := token.(string); ok {
                apiToken = tokenStr
            }
        }

        var timeout float64
        if t, exists := config["timeout_seconds"]; exists {
            if tFloat, ok := t.(float64); ok {
                timeout = tFloat
            }
        }

        var retryCount float64
        if retries, exists := config["retry_count"]; exists {
            if retriesFloat, ok := retries.(float64); ok {
                retryCount = retriesFloat
            }
        }

        // Create node configuration
        nodeConfig := &MyCustomNodeConfig{
            APIEndpoint: apiEndpoint,
            APIToken:    apiToken,
            Timeout:     time.Duration(timeout) * time.Second,
            RetryCount:  int(retryCount),
        }

        // Create and return node instance
        return NewMyCustomNode(nodeConfig), nil
    })
}
```

### Advanced Node Patterns

#### Stateful Node Pattern

For nodes that need to maintain state between executions:

```go
type StatefulNode struct {
    config *NodeConfig
    state  StateManager  // Interface for state persistence
}

type StateManager interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}) error
    Delete(ctx context.Context, key string) error
}

func (n *StatefulNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Get current state
    state, err := n.state.Get(ctx, fmt.Sprintf("node_%s_state", n.config.Name))
    if err != nil && !errors.Is(err, ErrStateNotFound) {
        return nil, fmt.Errorf("failed to get state: %w", err)
    }

    // Process inputs with state
    result, newState := n.processWithState(inputs, state)

    // Update state
    if newState != nil {
        err = n.state.Set(ctx, fmt.Sprintf("node_%s_state", n.config.Name), newState)
        if err != nil {
            return nil, fmt.Errorf("failed to update state: %w", err)
        }
    }

    return result, nil
}
```

#### Streaming Node Pattern

For nodes that handle streaming data:

```go
type StreamingNode struct {
    config *StreamingNodeConfig
}

func (n *StreamingNode) ExecuteStream(ctx context.Context, inputStream <-chan interface{}) (<-chan map[string]interface{}, <-chan error) {
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
                    return  // Channel closed
                }
                
                result, err := n.processItem(ctx, input)
                if err != nil {
                    select {
                    case errorStream <- err:
                    case <-ctx.Done():
                        return
                    }
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

## AI Agent System

### AI Agent Architecture

The AI Agent system is built on these core principles:

1. **Memory System**: Both short-term and long-term memory
2. **Tool Integration**: Pluggable tools for external service access
3. **Multi-Agent Coordination**: Agents collaborating on tasks
4. **Human-in-the-Loop**: Integration with human oversight
5. **Security Isolation**: Safe AI execution environments

### Memory System Design

Citadel Agent implements a hierarchical memory system:

#### Short-Term Memory
- Stores recent conversation history
- Implements sliding window with configurable size
- Optional automatic summarization for long conversations

#### Long-Term Memory
- Persistent storage using vector databases
- Semantic search capabilities
- Memory consolidation and cleanup routines
- Cross-workflow memory sharing

```go
// Memory interface for AI agents
type Memory interface {
    Add(ctx context.Context, entry *MemoryEntry) error
    Search(ctx context.Context, query string, limit int) ([]*MemoryEntry, error)
    Summarize(ctx context.Context, entries []*MemoryEntry) (*MemoryEntry, error)
    GetContext(ctx context.Context, query string) ([]*MemoryEntry, error)
    Cleanup(ctx context.Context) error
}

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
    ID          string    `json:"id"`
    Content     string    `json:"content"`
    Embedding   []float32 `json:"embedding"`
    Timestamp   time.Time `json:"timestamp"`
    Importance  float64   `json:"importance"`  // 0.0-1.0
    Tags        []string  `json:"tags"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### Multi-Agent Coordination

The multi-agent coordination system enables agents to work together on complex tasks:

#### Coordination Protocols
- **Task Assignment**: Fair task distribution among agents
- **Resource Sharing**: Coordinated access to shared resources
- **Communication**: Structured message passing between agents
- **Synchronization**: Coordinated execution ordering

#### Agent Roles
- **Manager**: Orchestrates task distribution
- **Worker**: Executes assigned tasks
- **Critic**: Reviews and validates results
- **Planner**: Creates task plans
- **Coordinator**: Manages inter-agents communication

### Creating Custom AI Agent Nodes

```go
// Custom AI agent node implementation
type CustomAINode struct {
    config *CustomAIConfig
    llm    llms.Model  // LangChain Go LLM interface
    memory Memory
    tools  []tools.Tool
}

// Execute runs the AI agent operation
func (n *CustomAINode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Prepare the prompt
    prompt := n.preparePrompt(inputs)
    
    // Get relevant context from memory
    context, err := n.memory.GetContext(ctx, prompt)
    if err != nil {
        // Log error but continue execution
        fmt.Printf("Warning: failed to get memory context: %v\n", err)
    }
    
    // Include context in the prompt
    fullPrompt := n.includeContext(prompt, context)
    
    // Execute with LLM
    result, err := n.llm.Call(ctx, fullPrompt)
    if err != nil {
        return nil, fmt.Errorf("LLM call failed: %w", err)
    }
    
    // Store interaction in memory
    memoryEntry := &MemoryEntry{
        Content:   fmt.Sprintf("Input: %v\nOutput: %s", inputs, result),
        Timestamp: time.Now(),
        Importance: 0.7, // Default importance
        Tags:      []string{"ai_interaction", n.config.ModelName},
    }
    
    if err := n.memory.Add(ctx, memoryEntry); err != nil {
        // Log error but don't fail the operation
        fmt.Printf("Warning: failed to store to memory: %v\n", err)
    }
    
    return map[string]interface{}{
        "success":     true,
        "response":    result,
        "input":       inputs,
        "model":       n.config.ModelName,
        "timestamp":   time.Now().Unix(),
        "tokens_used": estimateTokenCount(result),
    }, nil
}
```

## Workflow Engine Internals

### Execution Model

The workflow engine implements a sophisticated execution model:

#### Execution States
1. **Pending**: Workflow created but not started
2. **Running**: Current execution in progress
3. **Waiting**: Waiting for external events or conditions
4. **Paused**: Execution suspended by user
5. **Completed**: Execution finished successfully
6. **Failed**: Execution failed with errors
7. **Cancelled**: Execution cancelled by user

#### Parallel Execution

The engine supports parallel execution of nodes that have no dependency relationships:

```go
// Parallel execution implementation
func (e *Engine) executeNodesInParallel(ctx context.Context, execution *Execution, workflow *Workflow) error {
    // Build dependency graph
    graph, err := e.buildDependencyGraph(workflow)
    if err != nil {
        return fmt.Errorf("failed to build dependency graph: %w", err)
    }

    // Initialize execution state
    state := &ExecutionState{
        CompletedNodes: make(map[string]bool),
        Results:        make(map[string]*NodeResult),
        Mutex:          &sync.RWMutex{},
    }

    // Track which nodes are ready to execute
    readyNodes := make(map[string]bool)
    remainingNodes := len(workflow.Nodes)

    // Initially, add all nodes with no dependencies to ready list
    for _, node := range workflow.Nodes {
        if len(node.Dependencies) == 0 {
            readyNodes[node.ID] = true
        }
    }

    // Semaphore to limit concurrent execution
    semaphore := make(chan struct{}, e.config.MaxConcurrentNodes)
    
    for remainingNodes > 0 {
        // Execute all ready nodes concurrently
        var wg sync.WaitGroup
        executedThisRound := 0
        
        for nodeID, isReady := range readyNodes {
            if !isReady {
                continue
            }
            
            // Check if we've reached the concurrency limit
            select {
            case semaphore <- struct{}{}: // Acquire
                // Continue to execute
            default:
                // Reached concurrency limit, stop this round
                break
            }
            
            wg.Add(1)
            executedThisRound++
            
            // Execute node in goroutine
            go func(nodeID string) {
                defer wg.Done()
                defer func() { <-semaphore }() // Release
                
                node := e.getNodeByID(workflow, nodeID)
                result, err := e.executeSingleNode(ctx, execution, node, state)
                
                // Update execution state
                state.Mutex.Lock()
                delete(readyNodes, nodeID)
                state.CompletedNodes[nodeID] = true
                state.Results[nodeID] = result
                remainingNodes--
                
                // Update dependent nodes that may now be ready
                for _, n := range workflow.Nodes {
                    if _, completed := state.CompletedNodes[n.ID]; completed {
                        continue
                    }
                    
                    if e.allDependenciesMet(n.Dependencies, state.CompletedNodes) {
                        readyNodes[n.ID] = true
                    }
                }
                state.Mutex.Unlock()
                
                // Handle errors
                if err != nil {
                    // Log error, update execution status
                    e.updateExecutionWithError(execution, nodeID, err)
                }
            }(nodeID)
        }
        
        wg.Wait()

        // If no nodes were executed, we have a circular dependency
        if executedThisRound == 0 {
            return fmt.Errorf("circular dependency detected in workflow")
        }
    }

    return nil
}
```

### Scheduling System

The advanced scheduling system supports multiple scheduling patterns:

#### Cron-based Scheduling
- Standard cron expressions
- Named schedules (e.g., "@every 1m", "@daily", "@weekly")
- Timezone support

#### Event-based Scheduling
- Webhook triggers
- Database change notifications
- File system events
- Message queue events

#### Complex Scheduling
- Time windows (only execute during specific hours)
- Day-of-week constraints
- Calendar-based scheduling
- Seasonal or holiday-aware scheduling

## Security Implementation

### Runtime Security

Citadel Agent implements multiple layers of security:

#### Code Execution Security
1. **Static Analysis**: AST parsing to detect dangerous patterns
2. **Sandboxing**: Process isolation and resource limiting
3. **Runtime Validation**: Input/output validation during execution
4. **Network Isolation**: Egress filtering and whitelist controls
5. **File System Isolation**: Restricted file access with permissions

#### Example of Security Validation
```go
func (n *SecureNode) validateCode(code string) error {
    // Static analysis for dangerous patterns
    dangerousPatterns := []string{
        "eval(", "exec(", "function.constructor", "__import__", 
        "importlib.", "subprocess.", "os.", "sys.", "commands.",
        "code.", "compile(", "execfile(", "file(", "open(",
    }
    
    codeLower := strings.ToLower(code)
    for _, pattern := range dangerousPatterns {
        if strings.Contains(codeLower, pattern) {
            return fmt.Errorf("code contains dangerous pattern: %s", pattern)
        }
    }
    
    // Validate syntax for supported languages
    switch n.language {
    case "javascript":
        return n.validateJavaScriptSyntax(code)
    case "python":
        return n.validatePythonSyntax(code)
    default:
        return nil // Other languages may have different validators
    }
}

// validateJavaScriptSyntax validates JavaScript syntax
func (n *SecureNode) validateJavaScriptSyntax(code string) error {
    // In a real implementation, we would parse the JavaScript AST
    // and perform security checks using libraries like goja
    
    // For now, we'll just do a simple check
    if strings.Contains(code, "import") || strings.Contains(code, "export") {
        // In a real system, we'd validate module imports for security
    }
    
    return nil
}
```

### RBAC Implementation

Role-Based Access Control is implemented at multiple levels:

1. **API Level**: JWT token validation and role checking
2. **Workflow Level**: Tenant isolation and permission checks
3. **Node Level**: Node-specific access controls
4. **Resource Level**: Fine-grained resource access

### Data Protection

#### Encryption
- **At-Rest**: All sensitive data encrypted with AES-256
- **In-Transit**: All communication uses TLS 1.3
- **Key Management**: Hardware Security Module (HSM) integration

#### Masking
- **PII Data**: Automatic detection and masking of personally identifiable information
- **API Keys**: Never logged or exposed unnecessarily
- **Credentials**: Proper sanitization in all logs

## Monitoring & Observability

### Metrics Collection

Citadel Agent provides comprehensive metrics:

#### System Metrics
- CPU, Memory, and Disk usage
- Goroutine and thread counts
- GC pause times and heap statistics
- Network I/O statistics

#### Application Metrics
- Workflow execution success/failure rates
- Node execution times and error rates
- Request rates per endpoint
- Queue depths and processing times

#### Custom Metrics
- Business-specific metrics defined by workflows
- User-defined KPIs and measurements
- Custom alert thresholds

#### Example Metrics Collection
```go
func (e *Engine) executeWithMetrics(ctx context.Context, workflow *Workflow) (map[string]interface{}, error) {
    start := time.Now()
    
    // Record execution attempt
    metrics.Counter("workflow_executions_total").With(
        "workflow_id", workflow.ID,
        "tenant_id", getTenantID(ctx),
        "user_id", getUserID(ctx),
    ).Add(1)
    
    result, err := e.execute(ctx, workflow)
    duration := time.Since(start)
    
    // Record execution duration
    metrics.Histogram("workflow_execution_duration_seconds").With(
        "workflow_id", workflow.ID,
        "tenant_id", getTenantID(ctx),
        "status", getStatusString(err),
    ).Observe(duration.Seconds())
    
    // Record success/failure
    if err != nil {
        metrics.Counter("workflow_execution_errors_total").With(
            "workflow_id", workflow.ID,
            "error_type", getErrorType(err),
        ).Add(1)
    }
    
    return result, err
}
```

### Distributed Tracing

All requests support distributed tracing:

```go
func (n *Node) ExecuteWithTracing(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Create span
    ctx, span := tracer.Start(ctx, "Node.Execute", trace.WithAttributes(
        attribute.String("node.type", n.GetType()),
        attribute.String("node.id", n.GetID()),
        attribute.Int("input.count", len(inputs)),
    ))
    defer span.End()
    
    // Execute operation
    result, err := n.Execute(ctx, inputs)
    
    // Record error if occurred
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        span.SetStatus(codes.Ok, "")
    }
    
    return result, err
}
```

### Log Collection

Structured logging with correlation IDs and proper context:

```go
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflowID string) error {
    logger := log.WithContext(ctx).With(
        slog.String("workflow_id", workflowID),
        slog.String("execution_id", getExecutionID(ctx)),
        slog.String("user_id", getUserID(ctx)),
        slog.String("tenant_id", getTenantID(ctx)),
    )
    
    logger.Info("Starting workflow execution")
    start := time.Now()
    
    err := e.execute(ctx, workflowID)
    
    duration := time.Since(start)
    
    if err != nil {
        logger.Error("Workflow execution failed", 
            slog.Any("error", err),
            slog.Duration("duration", duration))
    } else {
        logger.Info("Workflow execution completed", 
            slog.Duration("duration", duration))
    }
    
    return err
}
```

## Performance Optimization

### Caching Strategies

#### L1 Cache (In-Memory)
- For frequently accessed, small pieces of data
- Using concurrent maps or memory-optimized structures

#### L2 Cache (Redis)
- For larger datasets and cross-instance caching
- With proper TTL and eviction policies

#### Cache-Aside Pattern
```go
func (n *Node) GetData(ctx context.Context, key string) (interface{}, error) {
    // Try L1 cache first
    if result := n.l1cache.Get(key); result != nil {
        return result, nil
    }
    
    // Try L2 cache (Redis)
    if result := n.l2cache.Get(key); result != nil {
        // Populate L1 cache
        n.l1cache.Set(key, result, time.Minute)
        return result, nil
    }
    
    // Load from source
    data, err := n.loadDataFromSource(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    n.l2cache.Set(key, data, 10*time.Minute)
    n.l1cache.Set(key, data, time.Minute)
    
    return data, nil
}
```

### Database Optimization

#### Connection Pooling
- Properly sized connection pools
- Connection lifetime management
- Idle connection handling

#### Query Optimization
- Prepared statements for repeated queries
- Proper indexing strategies
- Efficient pagination for large datasets

#### Example Query Optimization
```go
type OptimizedQueryExecutor struct {
    db        *sqlx.DB
    stmtCache *StmtCache // Cache for prepared statements
}

func (oqe *OptimizedQueryExecutor) QueryWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
    // Use prepared statement from cache
    stmt := oqe.stmtCache.Get("SELECT * FROM workflows WHERE id = ?")
    
    var workflow Workflow
    err := stmt.GetContext(ctx, &workflow, workflowID)
    if err != nil {
        return nil, err
    }
    
    return &workflow, nil
}
```

## Best Practices

### Code Organization
- Follow Go project layout recommendations
- Separate concerns with distinct packages
- Use interfaces for loose coupling
- Implement proper error wrapping with `%w`

### Testing Strategy
- Unit tests for pure functions
- Integration tests for component interactions
- End-to-end tests for complete workflows
- Performance tests for critical paths
- Security tests for all input vectors

### Security Best Practices
- Never trust user input
- Implement defense in depth
- Use principle of least privilege
- Regular security audits and pen testing
- Secure defaults for all configurations

### Performance Best Practices
- Measure performance regularly
- Profile bottlenecks before optimizing
- Implement proper buffering and batching
- Use efficient data structures
- Consider caching strategies

### Monitoring Best Practices
- Implement structured logging
- Use distributed tracing
- Set up proper alerting
- Monitor business metrics, not just system metrics
- Implement SLI/SLO monitoring

This documentation provides the foundational knowledge needed to extend and work with Citadel Agent at an advanced level. The system is designed to be secure, scalable, and extensible while maintaining enterprise-grade reliability.