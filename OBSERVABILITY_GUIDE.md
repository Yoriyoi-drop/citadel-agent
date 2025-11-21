# Citadel Agent - Observability Guide

## Overview

This guide covers the implementation and usage of Citadel Agent's observability stack, including metrics, tracing, and logging systems.

## Architecture

### Components Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        OBSERVABILITY STACK                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐│
│  │   METRICS   │  │   TRACING   │  │   LOGGING   │  │  ALERT ││
│  │  PROMETHEUS │  │  OTEL/JAEGER│  │ STRUCTURED  │  │ RULES  ││
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘│
│         │                   │                   │              │
│         ▼                   ▼                   ▼              │
│  ┌─────────────────────────────────────────────────────────────┤
│  │              CITADEL AGENT APPLICATION                        │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          ││
│  │  │ NODE METRICS│ │ WORKFLOW    │ │ SYSTEM      │          ││
│  │  │ COLLECTION  │ │ TRACING     │ │ MONITORING  │          ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘          ││
│  └─────────────────────────────────────────────────────────────┤
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────────┤
│  │              TELEMETRY SERVICE                               ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          ││
│  │  │ METRICS     │ │ TRACE       │ │ EVENT       │          ││
│  │  │ SERVICE     │ │ COLLECTOR   │ │ PROCESSING  │          ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘          ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Metrics Collection

### Built-in Metrics

#### Workflow Execution Metrics
- `citadel_workflow_executions_total` - Total number of workflow executions
- `citadel_workflow_execution_duration_seconds` - Duration of workflow executions
- `citadel_workflow_errors_total` - Total number of workflow errors

#### Node Execution Metrics
- `citadel_node_executions_total` - Total number of node executions
- `citadel_node_execution_duration_seconds` - Duration of node executions
- `citadel_node_errors_total` - Total number of node errors

#### API Request Metrics
- `citadel_api_requests_total` - Total number of API requests
- `citadel_api_request_duration_seconds` - Duration of API requests
- `citadel_api_request_size_bytes` - Size of API requests
- `citadel_api_response_size_bytes` - Size of API responses

#### Resource Usage Metrics
- `citadel_cpu_usage_percent` - CPU usage percentage
- `citadel_memory_usage_bytes` - Memory usage in bytes
- `citadel_goroutines_total` - Total number of goroutines

#### Security Metrics
- `citadel_security_events_total` - Total number of security events
- `citadel_login_attempts_total` - Total number of login attempts
- `citadel_permission_denied_total` - Total number of permission denied events
- `citadel_api_keys_used_total` - Total number of API key usages

### Exposing Metrics

The metrics are exposed via the `/metrics` endpoint:

```bash
curl http://localhost:5001/api/v1/metrics
```

### Custom Metrics

To add custom metrics to your nodes or services:

```go
// In your service
import "citadel-agent/backend/internal/observability"

func (s *MyService) MyOperation(ctx context.Context) error {
    start := time.Now()
    
    // Perform operation
    err := s.doSomething(ctx)
    
    // Record metrics
    if err != nil {
        s.metrics.RecordErrorEvent("my_service", "my_operation", "error", err.Error())
    } else {
        duration := time.Since(start)
        s.metrics.RecordCustomMetric("my_operation_duration", duration.Seconds(), map[string]string{
            "operation": "my_operation",
            "status":    "success",
        })
    }
    
    return err
}
```

## Tracing Implementation

### OpenTelemetry Setup

Citadel Agent uses OpenTelemetry for distributed tracing:

```go
// In your service functions
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflowID string) error {
    // Start a new span
    ctx, span := otel.Tracer("workflow-service").Start(ctx, "ExecuteWorkflow")
    defer span.End()
    
    // Set span attributes
    span.SetAttributes(
        attribute.String("workflow.id", workflowID),
        attribute.String("service.version", "1.0.0"),
    )
    
    // Execute workflow
    err := s.engine.Execute(ctx, workflowID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
    
    return err
}
```

### Tracing Best Practices

1. **Use Semantic Attributes**: Use standard OpenTelemetry attributes for consistency
2. **Set Span Kind**: Specify SERVER or CLIENT spans appropriately
3. **Record Errors**: Always record errors in spans
4. **Add Context**: Include relevant identifiers and metadata

## Logging Configuration

### Structured Logging

All Citadel Agent services use structured logging:

```go
import (
    "github.com/rs/zerolog/log"
)

func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflowID string) error {
    log.Info().
        Str("workflow_id", workflowID).
        Str("user_id", userID).
        Msg("Starting workflow execution")
    
    start := time.Now()
    err := s.engine.Execute(ctx, workflowID)
    duration := time.Since(start)
    
    if err != nil {
        log.Error().
            Err(err).
            Str("workflow_id", workflowID).
            Dur("duration", duration).
            Msg("Workflow execution failed")
    } else {
        log.Info().
            Str("workflow_id", workflowID).
            Dur("duration", duration).
            Msg("Workflow execution completed successfully")
    }
    
    return err
}
```

### Log Levels

- **DEBUG**: Detailed diagnostic information
- **INFO**: General operational messages
- **WARN**: Warning conditions
- **ERROR**: Error conditions
- **FATAL**: Critical errors requiring shutdown

### Log Format

Logs are output in JSON format for easy parsing:

```json
{
  "level": "info",
  "time": "2023-10-01T12:00:00Z",
  "message": "Workflow execution completed successfully",
  "workflow_id": "wf_123",
  "duration": 1.234,
  "user_id": "user_456"
}
```

## Dashboard and Visualization

### Prometheus Configuration

Sample Prometheus configuration to scrape Citadel Agent metrics:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'citadel-agent'
    static_configs:
      - targets: ['citadel-agent:5001']
    metrics_path: '/api/v1/metrics'
    scrape_interval: 5s
```

### Grafana Dashboards

Citadel Agent ships with pre-configured Grafana dashboards:

#### System Performance Dashboard
- CPU and memory usage
- Goroutine count
- GC pause times
- Request rates and latencies

#### Workflow Execution Dashboard
- Workflow execution rates
- Success/error rates
- Execution durations
- Top slow workflows

#### API Performance Dashboard
- Request latency percentiles
- Error rates by endpoint
- Request throughput
- Response sizes

### Alerting Rules

Sample PromQL queries for alerting:

```promql
# High error rate alert
increase(citadel_api_requests_total{status_code=~"5.."}[5m]) > 0

# Slow requests alert
histogram_quantile(0.95, rate(citadel_api_request_duration_seconds_bucket[5m])) > 1

# Workflow failure alert
increase(citadel_workflow_errors_total[10m]) > 5

# High memory usage
citadel_memory_usage_bytes / 1024 / 1024 > 500  # > 500 MB
```

## Monitoring Integration

### Exporter Configuration

Citadel Agent supports multiple export destinations:

#### OpenTelemetry Collector
```yaml
exporters:
  otlp:
    endpoint: collector.example.com:4317
    insecure: false
```

#### Prometheus Remote Write
```yaml
exporters:
  prometheusremotewrite:
    endpoint: https://prometheus.example.com/api/v1/write
    headers:
      authorization: Bearer ${PROMETHEUS_TOKEN}
```

#### Jaeger for Tracing
```yaml
exporters:
  jaeger:
    endpoint: jaeger-collector:14250
    tls:
      insecure: true
```

## Health Checks

### Health Endpoint

The health endpoint provides status information:

```bash
curl -X GET http://localhost:5001/api/v1/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": 1634567890,
  "uptime": 3600,
  "goroutines": 50,
  "database_connected": true,
  "redis_connected": true,
  "disk_space_available": true,
  "memory_available": true
}
```

### Liveness and Readiness Probes

For Kubernetes deployments:

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 5001
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /api/v1/health
    port: 5001
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

## Performance Monitoring

### Key Performance Indicators

Monitor these KPIs for optimal Citadel Agent performance:

#### Request Performance
- 95th percentile response time < 500ms
- 99th percentile response time < 1s
- Error rate < 0.1%

#### Workflow Execution
- Average workflow execution time
- Failed executions rate
- Resource utilization per workflow

#### System Resources
- CPU utilization < 80%
- Memory utilization < 80%
- Disk space usage < 80%

#### Concurrency
- Active goroutines monitoring
- Database connection pool usage
- Redis connection pool usage

## Troubleshooting Monitoring

### Common Monitoring Issues and Solutions

#### Missing Metrics
1. Verify the metrics endpoint is accessible
2. Check Prometheus configuration targets
3. Ensure proper service discovery
4. Validate metric names and labels

#### High Memory Usage
1. Monitor garbage collection metrics
2. Check for memory leaks in long-running operations
3. Review connection pool sizes
4. Validate proper resource cleanup

#### Slow Requests
1. Enable detailed tracing for slow requests
2. Monitor database query performance
3. Check for resource contention
4. Validate caching effectiveness

#### Spikes in Error Rates
1. Check application logs for error details
2. Verify upstream service availability
3. Review recent deployments
4. Monitor system resource usage

## Advanced Monitoring

### Custom Dashboards

Create custom dashboards for specific use cases:

#### Tenant-Specific Monitoring
- Workflow execution per tenant
- API usage per tenant
- Storage usage per tenant

#### Node-Type Monitoring
- Performance by node type
- Error rates by node category
- Resource usage by node type

### Alerting Configuration

Configure custom alerting rules for your specific needs:

#### Workflow Success Rate
```
alert: LowWorkflowSuccessRate
expr: (sum(rate(citadel_workflow_executions_total{status="success"}[5m])) / 
      sum(rate(citadel_workflow_executions_total[5m]))) < 0.95
for: 5m
```

#### Node Performance
```
alert: SlowNodeExecution
expr: histogram_quantile(0.95, 
      rate(citadel_node_execution_duration_seconds_bucket[5m])) > 5
for: 2m
```

This observability stack ensures Citadel Agent is highly monitorable, debuggable, and maintainable in production environments.