# Temporal Integration for Citadel Agent

## Overview
Citadel Agent now integrates with Temporal.io for robust workflow orchestration, providing features like:
- Reliable workflow execution with persistence
- Built-in retry mechanisms and circuit breakers
- Distributed execution across multiple workers
- Advanced scheduling and monitoring capabilities
- Automatic fault tolerance and recovery

## Architecture
- `TemporalClient`: Wrapper around Temporal SDK client
- `TemporalWorkflowService`: Service layer that bridges Citadel Agent and Temporal
- `Workflows`: Main workflow definitions that orchestrate nodes
- `Activities`: Individual node execution activities
- `PluginIntegration`: Integration with Citadel Agent's plugin system

## Setting Up Temporal Server

### Option 1: Local Development (Docker)
```bash
docker run --rm -p 7233:7233 -p 7243:7243 temporalio/auto-setup:1.19.0
```

### Option 2: Production
Deploy Temporal server cluster following the official Temporal documentation.

## Configuration

### Environment Variables
```bash
CITADEL_TEMPORAL_ADDRESS=localhost:7233
CITADEL_TEMPORAL_NAMESPACE=default
CITADEL_TEMPORAL_TASK_QUEUE=citadel-agent-workflows
CITADEL_TEMPORAL_WORKFLOW_TIMEOUT=60m
CITADEL_TEMPORAL_ACTIVITY_TIMEOUT=10m
```

### Code Configuration
```go
config := temporal.GetDefaultConfig()
config.Address = "your-temporal-server:7233"
config.Namespace = "your-namespace"

if err := config.Validate(); err != nil {
    log.Fatal("Invalid Temporal config:", err)
}
```

## Usage

### 1. Initialize Temporal Service
```go
temporalClient, err := temporal.NewTemporalClient(&temporal.Config{
    Address:   os.Getenv("CITADEL_TEMPORAL_ADDRESS"),
    Namespace: os.Getenv("CITADEL_TEMPORAL_NAMESPACE"),
})
if err != nil {
    log.Fatal("Failed to create Temporal client:", err)
}

workflowService := temporal.NewTemporalWorkflowService(temporalClient, baseEngine)
workflowService.RegisterNodeTypes()
```

### 2. Define a Workflow
```go
workflowDef := &temporal.WorkflowDefinition{
    ID:          "my-workflow",
    Name:        "My Workflow",
    Description: "A sample workflow",
    Options: temporal.WorkflowOptions{
        Parallelism:   5,
        Timeout:       time.Minute * 30,
        RetryAttempts: 3,
        ErrorHandling: "continue",
    },
    Nodes: []temporal.NodeDefinition{
        {
            ID:   "node-1",
            Type: "http_request",
            Name: "HTTP Request",
            Config: map[string]interface{}{
                "url": "https://api.example.com",
                "method": "GET",
            },
            Options: temporal.NodeExecutionOptions{
                RetryAttempts:  3,
                Timeout:        time.Second * 30,
                RetryOnFailure: true,
            },
        },
        // Add more nodes...
    },
    Connections: []temporal.ConnectionDefinition{
        {
            SourceNodeID: "node-1",
            TargetNodeID: "node-2",
        },
        // Add more connections...
    },
}

workflowService.RegisterWorkflowDefinition(workflowDef)
```

### 3. Execute a Workflow
```go
params := map[string]interface{}{
    "input1": "value1",
    "input2": "value2",
}

workflowID, err := workflowService.ExecuteWorkflow(context.Background(), "my-workflow", params)
if err != nil {
    log.Fatal("Failed to execute workflow:", err)
}

fmt.Printf("Started workflow with ID: %s\n", workflowID)
```

## Node Integration

### Built-in Nodes
The system supports all existing Citadel Agent node types:
- `http_request`
- `condition`
- `delay`
- `database_query`
- `script_execution`
- `ai_agent`
- `data_transformer`
- `notification`
- `loop`
- `error_handler`

### Plugin Nodes
Plugin nodes registered with the PluginManager are also available in Temporal workflows:
```go
pluginManager.RegisterPluginAtPath("custom_ai_processor", "./plugins/ai_processor")
workflowService.RegisterNodeTypes() // This registers plugin nodes too
```

## Advanced Features

### Retry Policies
Custom retry policies can be configured for workflows and individual nodes:
```go
retryPolicy := temporal.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2.0,
    MaximumInterval:    time.Minute,
    MaximumAttempts:    5,
    NonRetryableErrors: []string{"SystemError"},
}
```

### Circuit Breakers
Circuit breaker functionality is built into node execution to prevent cascading failures.

### Parallel Execution
Workflows support parallel execution of nodes when dependencies allow:
```go
workflowOptions := temporal.WorkflowOptions{
    Parallelism:   10,  // Max concurrent nodes
    MaxConcurrent: 5,   // Max concurrent activities
}
```

## Migration Path

The Temporal integration is designed to be backward compatible:

1. **Phase 1**: Run Temporal alongside existing engine
2. **Phase 2**: Migrate workflows gradually to use Temporal
3. **Phase 3**: Deprecate old engine (optional)

Both systems can coexist, allowing for gradual migration of workflows.

## Error Handling

### Workflow-Level Errors
- `continue`: Continue execution when a node fails
- `stop`: Stop the entire workflow when a node fails  
- `retry`: Retry the failed node with backoff

### Node-Level Errors
Each node can have specific retry policies and error handling configurations.