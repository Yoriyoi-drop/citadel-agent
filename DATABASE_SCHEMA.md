# DATABASE SCHEMA

## 🗄️ ENTIRE DATABASE SCHEMA FOR AUTOMATION PLATFORM

### 1. USERS & AUTHENTICATION TABLES

```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(32),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(512) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User role assignments
CREATE TABLE IF NOT EXISTS user_role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES user_roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id)
);

-- API keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    scopes JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- OAuth providers
CREATE TABLE IF NOT EXISTS oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'github', 'gitlab', etc.
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    scopes JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);
```

### 2. WORKFLOW CORE TABLES

```sql
-- Workspaces/organizations
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id),
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Workspace members
CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL, -- 'owner', 'admin', 'editor', 'viewer'
    invited_by UUID REFERENCES users(id),
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, user_id)
);

-- Workflow definitions
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    nodes JSONB NOT NULL, -- Stores the workflow nodes as JSON
    edges JSONB NOT NULL, -- Stores the workflow edges as JSON
    settings JSONB, -- Stores workflow settings like triggers, schedule, etc.
    status VARCHAR(50) DEFAULT 'draft', -- draft, active, inactive, archived
    version INTEGER DEFAULT 1,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Workflow versions for versioning system
CREATE TABLE IF NOT EXISTS workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    nodes JSONB NOT NULL,
    edges JSONB NOT NULL,
    settings JSONB,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workflow_id, version_number)
);

-- Workflow tags
CREATE TABLE IF NOT EXISTS workflow_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7), -- hex color code
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, name)
);

-- Workflow tag assignments
CREATE TABLE IF NOT EXISTS workflow_tag_assignments (
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES workflow_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (workflow_id, tag_id)
);

-- Workflow triggers
CREATE TABLE IF NOT EXISTS workflow_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    trigger_type VARCHAR(100) NOT NULL, -- 'webhook', 'schedule', 'manual', 'event'
    settings JSONB NOT NULL, -- Trigger-specific settings
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Workflow schedules
CREATE TABLE IF NOT EXISTS workflow_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    is_active BOOLEAN DEFAULT TRUE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Webhook endpoints
CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    endpoint_path VARCHAR(255) UNIQUE NOT NULL,
    secret VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 3. EXECUTION & RUNTIME TABLES

```sql
-- Workflow executions
CREATE TABLE IF NOT EXISTS workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    workflow_version INTEGER, -- Version of workflow when executed
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- running, completed, failed, cancelled, paused
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT, -- Execution duration in milliseconds
    variables JSONB, -- Execution variables
    result JSONB, -- Final result of execution
    error TEXT, -- Error message if execution failed
    trigger_data JSONB, -- Data that triggered the execution
    created_by UUID REFERENCES users(id), -- User who started the execution
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Node executions (individual node runs within a workflow execution)
CREATE TABLE IF NOT EXISTS node_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL, -- ID of the node in the workflow
    node_type VARCHAR(255) NOT NULL, -- Type of the node
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- running, completed, failed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_ms BIGINT, -- Execution duration in milliseconds
    input_data JSONB, -- Input data to the node
    output_data JSONB, -- Output data from the node
    error TEXT, -- Error message if node execution failed
    retry_count INTEGER DEFAULT 0,
    worker_id VARCHAR(255), -- ID of the worker that executed this node
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Execution logs
CREATE TABLE IF NOT EXISTS execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    node_execution_id UUID REFERENCES node_executions(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL, -- debug, info, warn, error
    message TEXT NOT NULL,
    metadata JSONB, -- Additional metadata about the log
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Execution queue (for job scheduling)
CREATE TABLE IF NOT EXISTS execution_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES workflow_executions(id) ON DELETE CASCADE, -- May be null if not yet created
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, failed, cancelled
    priority INTEGER DEFAULT 0, -- Higher number = higher priority
    payload JSONB NOT NULL, -- Queue payload
    max_retries INTEGER DEFAULT 3,
    retry_count INTEGER DEFAULT 0,
    error TEXT,
    worker_id VARCHAR(255), -- ID of worker processing this job
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 4. NODE & PLUGIN SYSTEM TABLES

```sql
-- Built-in node definitions
CREATE TABLE IF NOT EXISTS node_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL, -- trigger, action, condition, loop, etc.
    category VARCHAR(100), -- http, database, ai, etc.
    icon VARCHAR(100),
    settings_schema JSONB, -- JSON schema for node settings
    input_schema JSONB, -- JSON schema for node input
    output_schema JSONB, -- JSON schema for node output
    is_builtin BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    version VARCHAR(20) DEFAULT '1.0.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Custom nodes (user-created)
CREATE TABLE IF NOT EXISTS custom_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    icon VARCHAR(100),
    settings_schema JSONB,
    input_schema JSONB,
    output_schema JSONB,
    code TEXT, -- Node implementation code
    language VARCHAR(50), -- 'javascript', 'python', 'go', etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    version VARCHAR(20) DEFAULT '1.0.0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Node templates
CREATE TABLE IF NOT EXISTS node_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    node_type VARCHAR(255) NOT NULL,
    template_data JSONB NOT NULL, -- Pre-configured node settings
    is_public BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Node marketplace/plugins
CREATE TABLE IF NOT EXISTS node_marketplace (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    publisher VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL,
    node_type VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    settings_schema JSONB,
    download_url TEXT NOT NULL,
    manifest JSONB NOT NULL, -- Node manifest file
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    download_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00, -- Average rating
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Installed marketplace nodes
CREATE TABLE IF NOT EXISTS installed_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    marketplace_node_id UUID REFERENCES node_marketplace(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL,
    config JSONB, -- Installation-specific configuration
    is_active BOOLEAN DEFAULT TRUE,
    installed_by UUID REFERENCES users(id),
    installed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 5. INTEGRATION & CONNECTIONS TABLES

```sql
-- Connection credentials
CREATE TABLE IF NOT EXISTS connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- 'http', 'database', 'aws', 'slack', etc.
    config JSONB NOT NULL, -- Connection configuration (masked)
    encrypted_config TEXT NOT NULL, -- Encrypted configuration values
    verification_status VARCHAR(20) DEFAULT 'unverified', -- unverified, verified, failed
    last_verified_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- OAuth connections
CREATE TABLE IF NOT EXISTS oauth_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    scopes JSONB DEFAULT '[]',
    external_user_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Environment variables
CREATE TABLE IF NOT EXISTS environment_variables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    is_secret BOOLEAN DEFAULT FALSE, -- If true, value is encrypted
    scope VARCHAR(50) DEFAULT 'workflow', -- 'workflow', 'global', 'user'
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 6. MONITORING & ANALYTICS TABLES

```sql
-- Workflow metrics
CREATE TABLE IF NOT EXISTS workflow_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    execution_date DATE NOT NULL,
    successful_executions INTEGER DEFAULT 0,
    failed_executions INTEGER DEFAULT 0,
    total_executions INTEGER DEFAULT 0,
    avg_duration_ms DECIMAL(10,2),
    max_duration_ms BIGINT,
    min_duration_ms BIGINT,
    error_rate DECIMAL(5,2), -- Percentage
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workflow_id, execution_date)
);

-- Node metrics
CREATE TABLE IF NOT EXISTS node_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(255) NOT NULL,
    execution_date DATE NOT NULL,
    execution_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    avg_duration_ms DECIMAL(10,2),
    max_duration_ms BIGINT,
    min_duration_ms BIGINT,
    error_rate DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workflow_id, node_id, execution_date)
);

-- System metrics
CREATE TABLE IF NOT EXISTS system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(100) NOT NULL, -- 'api_requests', 'queue_size', 'worker_load', etc.
    metric_value DECIMAL(15,2) NOT NULL,
    unit VARCHAR(50), -- 'count', 'ms', 'bytes', 'percentage', etc.
    tags JSONB, -- Additional metric tags
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    workspace_id UUID REFERENCES workspaces(id),
    action VARCHAR(100) NOT NULL, -- 'workflow_created', 'node_executed', 'user_signed_in', etc.
    resource_type VARCHAR(50), -- 'workflow', 'node', 'user', 'connection', etc.
    resource_id VARCHAR(255), -- ID of the resource
    old_values JSONB, -- Previous state of the resource
    new_values JSONB, -- New state of the resource
    ip_address INET,
    user_agent TEXT,
    metadata JSONB, -- Additional metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 7. BILLING & USAGE TABLES

```sql
-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL, -- active, inactive, cancelled, past_due
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Usage records
CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    metric VARCHAR(100) NOT NULL, -- 'workflow_executions', 'api_calls', 'storage_gb', etc.
    amount BIGINT NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Billing plans
CREATE TABLE IF NOT EXISTS billing_plans (
    id VARCHAR(100) PRIMARY KEY, -- e.g., 'free', 'pro', 'team', 'enterprise'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR(10) DEFAULT 'month', -- month, year
    features JSONB DEFAULT '[]', -- Features included in the plan
    limits JSONB DEFAULT '{}', -- Usage limits for the plan
    is_active BOOLEAN DEFAULT TRUE,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 8. INDEXES & PERFORMANCE

```sql
-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Sessions indexes
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(token);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);

-- Workspaces indexes
CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_created_at ON workspaces(created_at);

-- Workspace members indexes
CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);

-- Workflows indexes
CREATE INDEX idx_workflows_workspace ON workflows(workspace_id);
CREATE INDEX idx_workflows_created_by ON workflows(created_by);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_created_at ON workflows(created_at);
CREATE INDEX idx_workflows_updated_at ON workflows(updated_at);
CREATE INDEX idx_workflows_published_at ON workflows(published_at);

-- Workflow executions indexes
CREATE INDEX idx_workflow_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX idx_workflow_executions_created_by ON workflow_executions(created_by);
CREATE INDEX idx_workflow_executions_started_at ON workflow_executions(started_at);
CREATE INDEX idx_workflow_executions_created_at ON workflow_executions(created_at);

-- Node executions indexes
CREATE INDEX idx_node_executions_execution ON node_executions(execution_id);
CREATE INDEX idx_node_executions_node_id ON node_executions(node_id);
CREATE INDEX idx_node_executions_status ON node_executions(status);
CREATE INDEX idx_node_executions_started_at ON node_executions(started_at);

-- Execution queue indexes
CREATE INDEX idx_execution_queue_status ON execution_queue(status);
CREATE INDEX idx_execution_queue_priority ON execution_queue(priority DESC, created_at);
CREATE INDEX idx_execution_queue_worker ON execution_queue(worker_id);

-- Connections indexes
CREATE INDEX idx_connections_workspace ON connections(workspace_id);
CREATE INDEX idx_connections_type ON connections(type);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- System metrics indexes
CREATE INDEX idx_system_metrics_type ON system_metrics(metric_type);
CREATE INDEX idx_system_metrics_recorded_at ON system_metrics(recorded_at);

-- Webhook endpoints indexes
CREATE INDEX idx_webhook_endpoints_path ON webhook_endpoints(endpoint_path);
CREATE INDEX idx_webhook_endpoints_workflow ON webhook_endpoints(workflow_id);

-- Triggers indexes
CREATE INDEX idx_workflow_triggers_workflow ON workflow_triggers(workflow_id);
CREATE INDEX idx_workflow_triggers_type ON workflow_triggers(trigger_type);

-- Schedules indexes
CREATE INDEX idx_workflow_schedules_workflow ON workflow_schedules(workflow_id);
CREATE INDEX idx_workflow_schedules_next_run ON workflow_schedules(next_run_at);
CREATE INDEX idx_workflow_schedules_active ON workflow_schedules(is_active, next_run_at);

-- API keys indexes
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
```

### 9. TRIGGERS & FUNCTIONS

```sql
-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workspaces_updated_at 
    BEFORE UPDATE ON workspaces 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflows_updated_at 
    BEFORE UPDATE ON workflows 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflow_executions_updated_at 
    BEFORE UPDATE ON workflow_executions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_node_executions_updated_at 
    BEFORE UPDATE ON node_executions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_connections_updated_at 
    BEFORE UPDATE ON connections 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_execution_queue_updated_at 
    BEFORE UPDATE ON execution_queue 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate execution duration
CREATE OR REPLACE FUNCTION calculate_execution_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.ended_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_ms := EXTRACT(EPOCH FROM (NEW.ended_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_workflow_execution_duration
    BEFORE UPDATE ON workflow_executions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_execution_duration();

CREATE TRIGGER calculate_node_execution_duration
    BEFORE UPDATE ON node_executions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_execution_duration();
```

### 10. VIEWS FOR ANALYTICS

```sql
-- View for workflow execution summary
CREATE VIEW workflow_execution_summary AS
SELECT 
    w.id as workflow_id,
    w.name as workflow_name,
    w.workspace_id,
    COUNT(we.id) as total_executions,
    COUNT(CASE WHEN we.status = 'completed' THEN 1 END) as successful_executions,
    COUNT(CASE WHEN we.status = 'failed' THEN 1 END) as failed_executions,
    COUNT(CASE WHEN we.status = 'cancelled' THEN 1 END) as cancelled_executions,
    AVG(we.duration_ms) as avg_duration_ms,
    MAX(we.duration_ms) as max_duration_ms,
    MIN(we.duration_ms) as min_duration_ms,
    AVG(CASE WHEN we.status = 'failed' THEN 1.0 ELSE 0.0 END) * 100 as failure_rate
FROM workflows w
LEFT JOIN workflow_executions we ON w.id = we.workflow_id
WHERE we.created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY w.id, w.name, w.workspace_id;

-- View for user activity
CREATE VIEW user_activity_summary AS
SELECT 
    u.id as user_id,
    u.name as user_name,
    u.email,
    w.id as workspace_id,
    w.name as workspace_name,
    COUNT(we.id) as executions_started,
    COUNT(DISTINCT wf.id) as workflows_created,
    MAX(we.created_at) as last_activity_at
FROM users u
LEFT JOIN workspaces w ON u.id = w.owner_id
LEFT JOIN workflows wf ON w.id = wf.workspace_id AND wf.created_by = u.id
LEFT JOIN workflow_executions we ON w.id = we.workflow_id AND we.created_by = u.id
WHERE we.created_at >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY u.id, u.name, u.email, w.id, w.name;
```

---

## 🗃️ DATABASE DESIGN NOTES

### Security Considerations:
1. **Encryption**: Sensitive data like API keys and connection credentials are encrypted
2. **Row-level security**: Workspace-based isolation
3. **Audit trails**: Complete audit logging for compliance
4. **Input validation**: Schema validation at database level

### Performance Considerations:
1. **Indexing strategy**: Comprehensive indexes for common queries
2. **Partitioning**: Date-based partitioning for logs and metrics
3. **Connection pooling**: Optimized for high-concurrency scenarios
4. **Caching layer**: Designed to work with Redis/Memcached

### Scalability:
1. **Horizontal scaling**: Designed for multi-region deployment
2. **Time-series data**: Optimized for metrics and log storage
3. **Queue system**: Built-in job queue for background processing
4. **Event-driven**: Support for real-time notifications

### Compliance:
1. **GDPR ready**: User data deletion capabilities
2. **Audit logging**: Complete trail of all actions
3. **Data retention**: Configurable retention policies
4. **Access controls**: Granular permissions system