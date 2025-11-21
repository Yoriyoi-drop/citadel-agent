# Citadel Agent - Advanced Developer Guide

## Table of Contents

1. [System Architecture Deep Dive](#system-architecture-deep-dive)
2. [Node Development Best Practices](#node-development-best-practices)
3. [Security Implementation Patterns](#security-implementation-patterns)
4. [Monitoring & Observability Deep Dive](#monitoring--observability-deep-dive)
5. [Performance Optimization](#performance-optimization)
6. [Integration Patterns](#integration-patterns)
7. [Testing Strategies](#testing-strategies)
8. [Deployment & Operations](#deployment--operations)

## System Architecture Deep Dive

### Core Engine Architecture

The Citadel Agent core engine consists of 5 main layers:

```
┌─────────────────────────────────────────────────────────────┐
│                        API LAYER                           │
│  ┌─────────────────┐ ┌─────────────────┐ ┌───────────────┐ │
│  │   HTTP API      │ │   GraphQL API   │ │ WebSocket API │ │
│  │   (RESTful)     │ │   (Real-time)   │ │   (Live)      │ │
│  └─────────────────┘ └─────────────────┘ └───────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     WORKFLOW ENGINE                        │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐  │
│  │   RUNNER    │ │  EXECUTOR   │ │   SCHEDULER        │  │
│  │ (Lifecycle)  │ │(Execution) │ │ (Cron/Event-Based) │  │
│  └─────────────┘ └─────────────┘ └──────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    NODE RUNTIME                           │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐  │
│  │   GO        │ │   JS/TS     │ │ PYTHON/JAVA        │  │
│  │ (Native)    │ │ (Sandbox)   │ │  (Container)       │  │
│  └─────────────┘ └─────────────┘ └──────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    SECURITY LAYER                          │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐  │
│  │   SANDBOX   │ │  VALIDATOR  │ │ RBAC/PERMISSIONS   │  │
│  │ (Isolation) │ │ (Input)     │ │ (Access Control)   │  │
│  └─────────────┘ └─────────────┘ └──────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    DATA LAYER                             │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐  │
│  │  POSTGRES   │ │    REDIS    │ │   MINIO/FILES      │  │
│  │ (Storage)   │ │ (Cache/Queue)│ │ (Media/Blobs)     │  │
│  └─────────────┘ └─────────────┘ └──────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

#### 1. Workflow Engine Core
```go
// The engine orchestrates the execution of workflows
type Engine struct {
    Runners     map[string]*WorkflowRunner
    Executors   map[string]*NodeExecutor  
    Scheduler   *CronScheduler
    SecurityMgr *SecurityManager
    Metrics     *MetricsCollector
}

// The runner manages individual workflow lifecycles
type WorkflowRunner struct {
    ID          string
    Definition  *WorkflowDefinition
    State       *WorkflowState
    Context     context.Context
    Cancel      context.CancelFunc
    Events      chan *WorkflowEvent
}
```

#### 2. Node Execution Runtime
```go
// Each node type has its own execution environment
type NodeExecutor interface {
    Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
    ValidateConfig(config map[string]interface{}) error
    GetMetadata() *NodeMetadata
}

// Secure execution environments
type SecureRuntime struct {
    Sandbox    SandboxInterface  // Isolated execution
    Validator  *InputValidator   // Input sanitization
    ResourceLimiter *ResourceLimiter // CPU/Memory limits
    Logger     *StructuredLogger // Execution logging
}
```

### Node Development Best Practices

#### 1. Node Interface Contract

```go
// Every node must implement this interface
type Node interface {
    Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
    Validate(inputs map[string]interface{}) error
    GetMetadata() *NodeMetadata
    GetSchema() *NodeSchema
}
```

#### 2. Configuration Pattern

```go
// Use strongly typed configuration
type MyNodeConfig struct {
    RequiredField string                 `json:"required_field" validate:"required"`
    OptionalField *string               `json:"optional_field,omitempty"`
    Parameters    map[string]interface{} `json:"parameters"`
    Timeout       time.Duration          `json:"timeout" default:"30s"`
    Retries       int                   `json:"retries" default:"3"`
    Resources     ResourceLimits        `json:"resources"`
}

// Provide good defaults
func NewMyNode(config *MyNodeConfig) *MyNode {
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second
    }
    if config.Retries == 0 {
        config.Retries = 3
    }
    
    return &MyNode{config: config}
}
```

#### 3. Error Handling Pattern

```go
func (n *MyNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // 1. Validate inputs
    if err := n.validateInputs(inputs); err != nil {
        return nil, fmt.Errorf("input validation failed: %w", err)
    }
    
    // 2. Apply timeout context
    ctx, cancel := context.WithTimeout(ctx, n.config.Timeout)
    defer cancel()
    
    // 3. Create span for tracing
    ctx, span := telemetry.StartSpan(ctx, "MyNode.Execute")
    defer span.End()
    
    // 4. Record metrics
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RecordDuration(ctx, "node_execution_duration", duration, map[string]string{
            "node_type": n.GetType(),
            "status":    "success", // Will change if error occurs
        })
    }()
    
    // 5. Execute main logic
    result, err := n.internalExecute(ctx, inputs)
    if err != nil {
        // 6. Handle error properly
        span.RecordError(err)
        metrics.RecordCounter(ctx, "node_execution_errors_total", 1, map[string]string{
            "node_type": n.GetType(),
            "error_type": "execution_error",
        })
        
        return nil, fmt.Errorf("execution failed: %w", err)
    }
    
    return result, nil
}
```

#### 4. Security Pattern

```go
func (n *Node) ExecuteSecure(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // 1. Sanitize all inputs
    sanitizedInputs, err := n.sanitize(inputs)
    if err != nil {
        return nil, fmt.Errorf("input sanitization failed: %w", err)
    }
    
    // 2. Check permissions
    if err := n.checkPermissions(ctx); err != nil {
        return nil, fmt.Errorf("permission denied: %w", err)
    }
    
    // 3. Validate against policies
    if err := n.enforcePolicies(sanitizedInputs); err != nil {
        return nil, fmt.Errorf("policy violation: %w", err)
    }
    
    // 4. Execute in sandbox
    return n.executeInSandbox(ctx, sanitizedInputs)
}
```

### Security Implementation Patterns

#### 1. Input Sanitization Pipeline

```go
type InputSanitizer struct {
    validators []Validator
    cleaners   []Cleaner
}

func NewInputSanitizer() *InputSanitizer {
    return &InputSanitizer{
        validators: []Validator{
            NewSQLInjectionValidator(),
            NewXSSValidator(),
            NewCommandInjectionValidator(),
            NewPathTraversalValidator(),
        },
        cleaners: []Cleaner{
            NewHTMLEntityEncoder(),
            NewSpecialCharacterSanitizer(),
            NewLengthLimiter(),
        },
    }
}

func (is *InputSanitizer) Sanitize(input interface{}) (interface{}, error) {
    // Validate first
    for _, validator := range is.validators {
        if err := validator.Validate(input); err != nil {
            return nil, fmt.Errorf("validation failed: %w", err)
        }
    }
    
    // Clean second
    cleanInput := input
    for _, cleaner := range is.cleaners {
        cleanInput = cleaner.Clean(cleanInput)
    }
    
    return cleanInput, nil
}
```

#### 2. Resource Limiting Pattern

```go
type ResourceLimiter struct {
    maxCPU       float64
    maxMemory    int64
    maxTime      time.Duration
    maxNetwork   int64
    maxFiles     int64
}

func (rl *ResourceLimiter) ExecuteWithLimits(fn func() (interface{}, error)) (interface{}, error) {
    // Set up resource monitoring
    monitor := NewResourceMonitor()
    defer monitor.Stop()
    
    // Execute in goroutine with resource tracking
    resultChan := make(chan interface{}, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        // Monitor resources periodically
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()
        
        go func() {
            for {
                select {
                case <-ticker.C:
                    if err := monitor.CheckLimits(rl); err != nil {
                        // Trigger resource violation
                        return
                    }
                case <-done:
                    return
                }
            }
        }()
        
        result, err := fn()
        if err != nil {
            errorChan <- err
        } else {
            resultChan <- result
        }
        
        close(done)
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-time.After(rl.MaxTime):
        return nil, fmt.Errorf("execution timed out after %v", rl.MaxTime)
    }
}
```

#### 3. RBAC Pattern

```go
type PermissionChecker struct {
    evaluator *CasbinEnforcer
}

type CasbinPermissionChecker struct {
    enforcer *casbin.Enforcer
}

func (cpc *CasbinPermissionChecker) CheckPermission(userID, resource, action string) (bool, error) {
    permitted, err := cpc.enforcer.Enforce(userID, resource, action)
    if err != nil {
        return false, fmt.Errorf("permission check failed: %w", err)
    }
    
    return permitted, nil
}

// In node execution
func (n *Node) ExecuteWithRBAC(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    userID := ctx.Value("user_id").(string)
    
    allowed, err := n.permissionChecker.CheckPermission(userID, n.GetResourceName(), "execute")
    if err != nil {
        return nil, fmt.Errorf("permission check error: %w", err)
    }
    
    if !allowed {
        return nil, fmt.Errorf("permission denied: user %s cannot execute %s", userID, n.GetType())
    }
    
    return n.Execute(ctx, inputs)
}
```

## Monitoring & Observability Deep Dive

### 1. Metrics Strategy

```go
// Custom metrics for different systems
type MetricsCollector struct {
    // Core metrics
    workflowExecutions *prometheus.CounterVec
    workflowDuration   *prometheus.HistogramVec
    workflowErrors     *prometheus.CounterVec
    
    // Node-specific metrics
    nodeExecutions     *prometheus.CounterVec
    nodeDuration       *prometheus.HistogramVec
    nodeErrors         *prometheus.CounterVec
    
    // Resource metrics
    cpuUsage          *prometheus.GaugeVec
    memoryUsage       *prometheus.GaugeVec
    goroutines        prometheus.Gauge
    
    // Security metrics
    authAttempts      *prometheus.CounterVec
    securityEvents    *prometheus.CounterVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        workflowExecutions: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "citadel_workflow_executions_total",
                Help: "Total number of workflow executions",
            },
            []string{"workflow_type", "status", "tenant_id"},
        ),
        workflowDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "citadel_workflow_duration_seconds",
                Help: "Workflow execution duration",
                Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
            },
            []string{"workflow_type", "status", "tenant_id"},
        ),
        // ... other metrics
    }
}

// Record metrics with context
func (mc *MetricsCollector) RecordWorkflowExecution(workflowType, status, tenantID string, duration time.Duration) {
    mc.workflowExecutions.WithLabelValues(workflowType, status, tenantID).Inc()
    mc.workflowDuration.WithLabelValues(workflowType, status, tenantID).Observe(duration.Seconds())
}
```

### 2. Tracing Pattern

```go
// Distributed tracing with OpenTelemetry
func (n *Node) ExecuteWithTracing(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Start trace
    tracer := otel.Tracer("citadel-agent")
    ctx, span := tracer.Start(ctx, "Node.Execute")
    defer span.End()
    
    // Set span attributes
    span.SetAttributes(
        attribute.String("node.type", n.GetType()),
        attribute.String("node.id", n.GetID()),
        attribute.Int("input.size", len(inputs)),
    )
    
    // Add trace context to logs
    ctx = log.WithContext(ctx, "trace_id", span.SpanContext().TraceID().String())
    
    // Execute and record errors
    result, err := n.Execute(ctx, inputs)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    // Add result info to span
    span.SetAttributes(attribute.Bool("success", true))
    
    return result, nil
}
```

### 3. Logging Pattern

```go
// Structured logging with correlation IDs
type StructuredLogger struct {
    logger zerolog.Logger
}

func (sl *StructuredLogger) LogExecution(ctx context.Context, nodeType, nodeID string, inputs, outputs map[string]interface{}, err error, duration time.Duration) {
    event := sl.logger.Info()
    
    if err != nil {
        event = sl.logger.Error().Err(err)
    }
    
    event.
        Str("node_type", nodeType).
        Str("node_id", nodeID).
        Str("correlation_id", ctx.Value("correlation_id").(string)).
        Dur("duration_ms", duration).
        Int("input_size", len(inputs)).
        Int("output_size", len(outputs)).
        Msg("Node execution completed")
}
```

## Performance Optimization

### 1. Caching Strategies

```go
// Multi-layer caching with proper invalidation
type CacheManager struct {
    l1Cache *ristretto.Cache  // In-memory cache
    l2Cache *redis.Client     // Distributed cache
    ttl     time.Duration
}

func (cm *CacheManager) GetOrCompute(ctx context.Context, key string, computeFn func() (interface{}, error)) (interface{}, error) {
    // Try L1 first
    if value, found := cm.l1Cache.Get(key); found {
        return value, nil
    }
    
    // Try L2
    if value, err := cm.l2Cache.Get(ctx, key).Result(); err == nil {
        // Set in L1
        cm.l1Cache.Set(key, value, cm.ttl)
        return value, nil
    }
    
    // Compute and cache
    value, err := computeFn()
    if err != nil {
        return nil, err
    }
    
    // Set in both caches
    cm.l1Cache.Set(key, value, cm.ttl)
    cm.l2Cache.SetEX(ctx, key, value, cm.ttl)
    
    return value, nil
}
```

### 2. Connection Pooling

```go
// Database connection pools with adaptive sizing
type ConnectionPoolManager struct {
    pools map[string]*sqlx.DB
    config *PoolConfig
}

type PoolConfig struct {
    MinConns    int
    MaxConns    int
    MaxIdleTime time.Duration
    MaxLifeTime time.Duration
    Adaptive    bool  // Adjust pool size based on load
}

func (cpm *ConnectionPoolManager) GetOptimalPool(ctx context.Context, workloadHint string) *sqlx.DB {
    if !cpm.config.Adaptive {
        return cpm.pools["default"]
    }
    
    // Adjust pool size based on recent load patterns
    recentLoad := cpm.measureRecentLoad(ctx)
    optimalSize := cpm.calculateOptimalPoolSize(recentLoad, workloadHint)
    
    pool := cpm.pools[workloadHint]
    if pool != nil {
        pool.SetMaxOpenConns(optimalSize)
    }
    
    return pool
}
```

### 3. Batch Processing

```go
// Batch processing for high-throughput scenarios
type BatchProcessor struct {
    batchSize int
    timeout   time.Duration
    processor BatchFunction
    buffer    chan interface{}
    batches   chan []interface{}
}

func (bp *BatchProcessor) ProcessSingle(item interface{}) {
    select {
    case bp.buffer <- item:
    default:
        // Buffer is full, process immediately
        go bp.processBatch([]interface{}{item})
    }
}

func (bp *BatchProcessor) batchAggregator() {
    batch := make([]interface{}, 0, bp.batchSize)
    timer := time.NewTimer(bp.timeout)
    
    for {
        select {
        case item := <-bp.buffer:
            batch = append(batch, item)
            
            if len(batch) >= bp.batchSize {
                bp.batches <- batch
                batch = make([]interface{}, 0, bp.batchSize)
                timer.Reset(bp.timeout)
            }
        case <-timer.C:
            if len(batch) > 0 {
                bp.batches <- batch
                batch = make([]interface{}, 0, bp.batchSize)
            }
            timer.Reset(bp.timeout)
        }
    }
}
```

## Integration Patterns

### 1. API Integration Pattern

```go
// Generic API integration with circuit breaker
type APIIntegration struct {
    client        *http.Client
    circuitBreaker *gobreaker.CircuitBreaker
    retryPolicy   *RetryPolicy
    rateLimiter   *RateLimiter
}

func (ai *APIIntegration) CallAPI(ctx context.Context, endpoint string, payload interface{}) (interface{}, error) {
    // Check circuit breaker
    if !ai.circuitBreaker.Ready() {
        return nil, fmt.Errorf("circuit breaker is open")
    }
    
    // Apply rate limiting
    if err := ai.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }
    
    // Execute with retry policy
    var result interface{}
    err := retry.Do(
        func() error {
            req, err := ai.createRequest(ctx, endpoint, payload)
            if err != nil {
                return retry.Unrecoverable(err)
            }
            
            resp, err := ai.client.Do(req)
            if err != nil {
                return err // Retryable error
            }
            defer resp.Body.Close()
            
            // Check response status
            if resp.StatusCode >= 400 {
                return fmt.Errorf("API returned error status: %d", resp.StatusCode)
            }
            
            // Parse response
            result, err = ai.parseResponse(resp)
            return err
        },
        retry.Attempts(uint(ai.retryPolicy.MaxRetries)),
        retry.Delay(ai.retryPolicy.BackoffTime),
        retry.MaxJitter(ai.retryPolicy.MaxJitter),
    )
    
    if err != nil {
        ai.circuitBreaker.Fail()
        return nil, fmt.Errorf("API call failed after retries: %w", err)
    }
    
    ai.circuitBreaker.Success()
    return result, nil
}
```

### 2. Streaming Integration Pattern

```go
// Event streaming with backpressure
type StreamProcessor struct {
    consumers []Consumer
    producers []Producer
    buffer    chan *StreamMessage
    backpressure int
    maxBufferSize int
}

type StreamMessage struct {
    Data      interface{}
    Metadata  map[string]string
    Timestamp time.Time
    Topic     string
}

func (sp *StreamProcessor) ProcessStream(ctx context.Context, topic string) error {
    consumer := sp.getConsumerForTopic(topic)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-consumer.Messages():
            // Check buffer size to avoid backpressure
            if len(sp.buffer) >= sp.maxBufferSize {
                // Apply backpressure strategy: drop, buffer, or wait
                if sp.backpressure == DropMessages {
                    continue
                } else if sp.backpressure == Block {
                    sp.buffer <- msg
                } else if sp.backpressure == WaitAndRetry {
                    if !sp.tryAddWithTimeout(msg, 5*time.Second) {
                        // Still couldn't add, apply fallback strategy
                        continue
                    }
                }
            } else {
                sp.buffer <- msg
            }
        }
    }
}
```

## Testing Strategies

### 1. Unit Testing Pattern

```go
func TestMyNode_Execute(t *testing.T) {
    type testCase struct {
        name           string
        config         *MyNodeConfig
        inputs         map[string]interface{}
        expectedOutput map[string]interface{}
        expectedError  string
        mockSetup      func() * mocks.MockDependency
    }
    
    testCases := []testCase{
        {
            name: "successful execution",
            config: &MyNodeConfig{Timeout: 5 * time.Second},
            inputs: map[string]interface{}{"param": "value"},
            expectedOutput: map[string]interface{}{"result": "success"},
        },
        {
            name: "validation error",
            config: &MyNodeConfig{Timeout: 5 * time.Second},
            inputs: map[string]interface{}{"bad_param": "value"},
            expectedError: "validation failed",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Create node with config
            node := NewMyNode(tc.config)
            
            // Execute
            output, err := node.Execute(context.Background(), tc.inputs)
            
            // Assertions
            if tc.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tc.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expectedOutput, output)
            }
        })
    }
}
```

### 2. Integration Testing Pattern

```go
func TestWorkflowIntegration(t *testing.T) {
    // Setup test infrastructure
    db := setupTestDatabase(t)
    redis := setupTestRedis(t)
    es := setupTestEventStore(t)
    
    defer cleanupTestInfrastructure(db, redis, es)
    
    // Create workflow engine with test dependencies
    engine, cleanup := setupTestEngine(t, db, redis, es)
    defer cleanup()
    
    // Create test workflow
    workflow := createTestWorkflow()
    
    // Register test nodes
    engine.RegisterNode("test_http", &TestHTTPNode{})
    engine.RegisterNode("test_db", &TestDBNode{})
    
    // Execute workflow
    result, err := engine.ExecuteWorkflow(context.Background(), workflow, map[string]interface{}{
        "test_data": "value",
    })
    
    // Assertions
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "completed", result.Status)
}
```

### 3. Performance Testing Pattern

```go
func BenchmarkWorkflowExecution(b *testing.B) {
    // Initialize the engine once
    engine := initBenchmarkEngine()
    workflow := loadBenchmarkWorkflow()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := engine.ExecuteWorkflow(context.Background(), workflow, map[string]interface{}{
            "iteration": i,
        })
        if err != nil {
            b.Fatal(err)
        }
    }
    
    b.StopTimer()
    // Calculate throughput metrics
    b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
}
```

## Deployment & Operations

### 1. Configuration Management

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

database:
  host: "${DB_HOST:localhost}"
  port: "${DB_PORT:5432}"
  name: "${DB_NAME:citadel}"
  user: "${DB_USER:citadel}"
  password: "${DB_PASSWORD:}"
  pool:
    max_open: 20
    max_idle: 5
    max_lifetime: "30m"

redis:
  address: "${REDIS_ADDR:localhost:6379}"
  password: "${REDIS_PASS:}"
  db: 0
  pool:
    size: 10

security:
  jwt_secret: "${JWT_SECRET:default_secret_change_in_prod}"
  enable_rate_limiting: true
  rate_limit:
    requests: 1000
    window: "1m"

logging:
  level: "${LOG_LEVEL:info}"
  format: "json"
  output: "stdout"

tracing:
  enabled: true
  endpoint: "${OTEL_ENDPOINT:http://localhost:4317}"

metrics:
  enabled: true
  port: 9090

monitoring:
  enabled: true
  health_check_interval: "10s"
  alert_channels:
    - "email"
    - "slack"
```

### 2. Health Check Pattern

```go
type HealthChecker struct {
    checks map[string]HealthCheck
}

type HealthCheck func(ctx context.Context) (HealthStatus, error)

type HealthStatus struct {
    Status  string                 `json:"status"`  // "pass", "fail", "warn"
    Output  string                 `json:"output,omitempty"`
    Time    time.Time              `json:"time"`
    Service string                 `json:"service"`
    Checks  map[string]interface{} `json:"checks,omitempty"`
}

func (hc *HealthChecker) AddCheck(name string, check HealthCheck) {
    hc.checks[name] = check
}

func (hc *HealthChecker) CheckAll(ctx context.Context) *HealthStatus {
    results := make(map[string]interface{})
    overallStatus := "pass"
    
    for name, check := range hc.checks {
        status, err := check(ctx)
        if err != nil {
            status.Status = "fail"
            status.Output = err.Error()
        }
        
        results[name] = status
        
        if status.Status == "fail" {
            overallStatus = "fail"
        } else if status.Status == "warn" && overallStatus == "pass" {
            overallStatus = "warn"
        }
    }
    
    return &HealthStatus{
        Status:  overallStatus,
        Time:    time.Now(),
        Service: "citadel-agent",
        Checks:  results,
    }
}

func (hc *HealthChecker) DatabaseCheck(ctx context.Context) (HealthStatus, error) {
    // Test database connection
    err := hc.db.PingContext(ctx)
    if err != nil {
        return HealthStatus{
            Status: "fail",
            Output: fmt.Sprintf("Database connection failed: %v", err),
            Time:   time.Now(),
            Service: "database",
        }, nil
    }
    
    return HealthStatus{
        Status: "pass",
        Time:   time.Now(),
        Service: "database",
    }, nil
}
```

### 3. Deployment Manifest (Kubernetes)

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: citadel-agent
  labels:
    app: citadel-agent
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
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        fsGroup: 65532
      containers:
      - name: api
        image: citadel-agent:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: jwt-secret
        envFrom:
        - configMapRef:
            name: citadel-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
---
apiVersion: v1
kind: Service
metadata:
  name: citadel-agent-service
spec:
  selector:
    app: citadel-agent
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Best Practices Summary

### Code Quality
1. Always handle context cancellation
2. Use structured logging with correlation IDs
3. Implement proper error wrapping with `%w`
4. Follow input validation patterns consistently
5. Apply security measures at every layer

### Performance
1. Use connection pooling for external services
2. Implement caching with proper TTLs
3. Apply resource limiting to prevent abuse
4. Use batch processing for bulk operations
5. Monitor and optimize hot paths

### Security
1. Validate and sanitize all inputs
2. Use parameterized queries to prevent injection
3. Implement proper authentication and authorization
4. Apply principle of least privilege
5. Encrypt sensitive data at rest and in transit

### Maintainability
1. Document complex business logic
2. Write comprehensive tests
3. Use consistent naming patterns
4. Follow SOLID principles
5. Keep functions focused and small

This guide provides the foundation for developing with Citadel Agent at an enterprise scale, following industry best practices for security, performance, and maintainability.