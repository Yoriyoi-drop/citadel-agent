# Monitoring and Observability Guide

This document provides comprehensive guidance for monitoring and observing Citadel Agent in production.

## Overview

Citadel Agent includes built-in monitoring and observability features to help you understand system health, performance, and behavior. This guide covers:

- Health checks and status endpoints
- Logging configuration and formats
- Metrics collection
- Tracing (when enabled)
- Alerting configuration
- Performance monitoring

## Health Checks

### 1. Basic Health Check

Endpoint: `GET /health`

```json
{
  "status": "healthy",
  "message": "Citadel Agent API server is running",
  "timestamp": 1234567890,
  "uptime": "1h23m45s",
  "request_processing_time": "2ms"
}
```

### 2. Engine Status

Endpoint: `GET /api/v1/engine/status`

```json
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

### 3. Engine Statistics

Endpoint: `GET /api/v1/engine/stats`

```json
{
  "total_workflows_executed": 150,
  "total_nodes_executed": 2450,
  "active_workflows": 5,
  "average_execution_time": "2.3s",
  "success_rate": "98.5%",
  "error_rate": "1.5%",
  "plugins_registered": 3,
  "node_types_available": 10,
  "memory_usage": "456MB",
  "timestamp": 1234567890
}
```

## Logging

### 1. Log Format

Citadel Agent uses structured JSON logging:

```json
{
  "level": "info",
  "time": "2023-12-01T10:00:00Z",
  "message": "Request completed",
  "request_id": "req-abc123",
  "method": "POST",
  "url": "/api/v1/workflows/execute",
  "status": 200,
  "latency": "150ms",
  "user_agent": "curl/7.68.0",
  "ip": "192.168.1.100"
}
```

### 2. Log Levels

- `debug`: Detailed information for troubleshooting
- `info`: General operational information
- `warn`: Potentially problematic situations
- `error`: Error events that don't prevent operation
- `fatal`: Critical errors that cause shutdown

### 3. Access Logging

All HTTP requests are logged with:
- Request ID for tracing
- HTTP method and URL
- Response status code
- Response time
- Client IP address
- User agent
- Request size
- Response size

## Metrics

### 1. Prometheus Metrics

When enabled, metrics are available at `/metrics`:

```
# Citadel Agent Metrics
citadel_workflows_total{status="success"} 150
citadel_workflows_total{status="failed"} 2
citadel_workflow_duration_seconds{quantile="0.5"} 1.2
citadel_workflow_duration_seconds{quantile="0.9"} 3.4
citadel_workflow_duration_seconds{quantile="0.99"} 7.8
citadel_nodes_executed_total 2450
citadel_plugins_loaded 3
citadel_api_requests_total{method="GET",status="200",endpoint="/health"} 1500
citadel_api_requests_duration_seconds_sum{method="POST",endpoint="/api/v1/workflows/execute"} 345.6
citadel_api_requests_duration_seconds_count{method="POST",endpoint="/api/v1/workflows/execute"} 120
```

### 2. Key Metrics

#### Workflow Metrics
- `citadel_workflows_total` - Total workflow executions
- `citadel_workflow_duration_seconds` - Workflow execution time
- `citadel_workflow_success_rate` - Success rate of workflows
- `citadel_workflows_active` - Currently active workflows

#### API Metrics
- `citadel_api_requests_total` - Total API requests
- `citadel_api_requests_duration_seconds` - Request duration
- `citadel_api_requests_by_endpoint` - Requests by endpoint
- `citadel_api_error_rate` - Error rate

#### Plugin Metrics
- `citadel_plugins_loaded` - Number of loaded plugins
- `citadel_plugin_executions_total` - Total plugin executions
- `citadel_plugin_duration_seconds` - Plugin execution time

#### Resource Metrics
- `citadel_memory_usage_bytes` - Memory usage
- `citadel_cpu_usage_percent` - CPU usage percentage
- `citadel_goroutines_count` - Number of goroutines

## Tracing (Optional)

When distributed tracing is enabled:

- Requests are traced across service boundaries
- Workflow execution is traced end-to-end
- Plugin execution is traced
- Database calls are traced

Trace IDs are included in logs and HTTP headers for correlation.

## Performance Monitoring

### 1. Key Performance Indicators

#### Throughput
- Requests per second (RPS)
- Workflows executed per minute
- Nodes processed per second

#### Latency
- API response times
- Workflow execution times
- Plugin execution times

#### Availability
- Service uptime
- Error rates
- Failed request percentages

### 2. Performance Benchmarks

Baseline performance expectations:
- API response time: <100ms for simple operations
- Workflow execution: <1s for simple workflows
- Throughput: 1000+ RPS on average hardware

## Alerting

### 1. Critical Alerts

#### Service Availability
- API server down
- Temporal connection lost
- Database connection failed

#### Performance
- High error rates (>5%)
- Slow response times (>5s)
- High memory usage (>80%)
- High CPU usage (>80%)

#### Workflows
- Workflow execution failures
- Workflow timeouts
- Long-running workflows

### 2. Warning Alerts

#### Capacity
- Approaching resource limits
- High queue depths
- Plugin execution queue buildup

#### Security
- Unauthorized access attempts
- Rate limiting triggered
- Suspicious patterns

## Monitoring Setup Examples

### 1. Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'citadel-agent'
    static_configs:
      - targets: ['citadel-agent:3000']
    scrape_interval: 15s
    metrics_path: /metrics
```

### 2. Grafana Dashboard

Example dashboard panels:
- API Request Rate
- API Response Time
- Workflow Execution Rate
- Workflow Duration
- Error Rate
- Memory Usage
- CPU Usage
- Active Workflows
- Plugin Status

### 3. ELK Stack Configuration

```json
{
  "index_patterns": ["citadel-agent-*"],
  "template": {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 1
    },
    "mappings": {
      "properties": {
        "level": {"type": "keyword"},
        "time": {"type": "date"},
        "message": {"type": "text"},
        "request_id": {"type": "keyword"},
        "status": {"type": "integer"},
        "latency": {"type": "float"},
        "method": {"type": "keyword"},
        "url": {"type": "keyword"}
      }
    }
  }
}
```

## Security Monitoring

### 1. Access Logs

Monitor for:
- Suspicious request patterns
- Repeated failed authentication attempts
- Unauthorized endpoint access
- Large payload sizes

### 2. Audit Logs

Track:
- Workflow creation/execution
- Plugin registration
- User actions
- Configuration changes

## Troubleshooting

### 1. Common Monitoring Issues

#### No Metrics Available
- Check if metrics endpoint is enabled
- Verify Prometheus configuration
- Check network connectivity

#### High Memory Usage
- Monitor memory metrics over time
- Check for memory leaks
- Review workflow complexity

#### Slow Performance
- Check workflow execution times
- Monitor database performance
- Review plugin performance

### 2. Debugging Commands

```bash
# Check health
curl http://citadel-agent:3000/health

# Get detailed status
curl http://citadel-agent:3000/api/v1/engine/status

# Check metrics
curl http://citadel-agent:3000/metrics

# Monitor logs in real-time
docker logs -f citadel-agent
```

## Best Practices

1. **Set Up Appropriate Alerts**: Not too many, not too few
2. **Monitor Both System and Business Metrics**: Technical and user-facing
3. **Use Histograms for Duration**: Better than just averages
4. **Tag Metrics Properly**: For effective filtering and analysis
5. **Set Up Dashboards**: For quick status assessment
6. **Regular Review**: Update metrics and alerts based on usage patterns
7. **Document Runbooks**: For alert responses
8. **Test Alerting**: Verify alerts work as expected

## Integration with External Systems

### 1. Service Mesh (Istio, Linkerd)
- Automatic metrics collection
- Distributed tracing
- Traffic management

### 2. APM Tools (DataDog, New Relic)
- Detailed application performance
- End-user experience monitoring
- Infrastructure monitoring

### 3. Cloud Monitoring (CloudWatch, Stackdriver)
- Cloud resource metrics
- Log aggregation
- Alerting integration

This monitoring and observability framework ensures you have full visibility into your Citadel Agent deployment and can respond quickly to issues.