# Security Guide & Best Practices

This document outlines the security features of Citadel Agent and provides best practices for securing your deployment.

## Security Overview

Citadel Agent is designed with security-first principles:
- Plugin isolation and sandboxing
- Authentication and authorization
- Data encryption
- Network security
- Audit logging
- Resource limiting

## Authentication & Authorization

### 1. API Authentication

#### JWT-based Authentication
```go
// Example JWT middleware
func AuthMiddleware() fiber.Handler {
    return jwtware.New(jwtware.Config{
        SigningKey: []byte(os.Getenv("JWT_SECRET")),
        TokenLookup: "header:Authorization:Bearer ",
    })
}
```

#### API Key Authentication
```go
// Example API key middleware
func APIKeyMiddleware() fiber.Handler {
    return keyauth.New(keyauth.Config{
        KeyLookup: "header:X-API-Key",
        Validator: func(c *fiber.Ctx, key string) (bool, error) {
            // Validate API key against database
            return validateAPIKey(key), nil
        },
    })
}
```

### 2. Role-Based Access Control (RBAC)

Roles and permissions:
- `admin`: Full system access
- `user`: Workflow creation and execution
- `viewer`: Read-only access
- `plugin_developer`: Plugin management

### 3. OAuth2 Integration

Citadel Agent supports OAuth2 for:
- Single Sign-On (SSO)
- Third-party authentication
- Fine-grained permission control

## Data Encryption

### 1. Encryption at Rest

All sensitive data is encrypted:
- Workflow configurations
- User credentials
- API keys
- Plugin data
- Audit logs

### 2. Encryption in Transit

All communications use TLS:
- API requests/responses
- Database connections
- Temporal communication
- Plugin RPC calls

### 3. Secrets Management

Best practices for managing secrets:
```yaml
# Use environment variables for secrets
JWT_SECRET: ${JWT_SECRET}
DB_PASSWORD: ${DB_PASSWORD}
TEMPORAL_AUTH: ${TEMPORAL_AUTH}
PLUGIN_SECRETS: ${PLUGIN_SECRETS}
```

## Plugin Security

### 1. Isolation

Each plugin runs in:
- Separate process space
- Limited resource environment
- Restricted network access
- Sandboxed execution

### 2. Resource Limits

Plugin execution is limited by:
- Memory usage (default: 256MB)
- CPU percentage (default: 80%)
- Execution time (default: 30s)
- File system access
- Network access

### 3. Permission System

Plugins can be configured with specific permissions:
```go
type PluginPermissions struct {
    NetworkAccess   bool `json:"network_access"`
    FilesystemAccess bool `json:"filesystem_access"`
    DatabaseAccess  bool `json:"database_access"`
    SystemCalls     bool `json:"system_calls"`
    EnvironmentAccess bool `json:"environment_access"`
}
```

## Network Security

### 1. API Protection

#### Rate Limiting
```go
// Rate limiting middleware
app.Use(fiberthrottle.New(fiberthrottle.Config{
    Max: 100,           // requests
    Duration: 60 * time.Second,  // per 60 seconds
    Message: "Too many requests",
}))
```

#### CORS Configuration
```yaml
security:
  cors:
    allow_origins:
      - "https://trusted-domain.com"
    allow_methods:
      - GET
      - POST
      - OPTIONS
    allow_headers:
      - Authorization
      - Content-Type
    max_age: 86400
```

### 2. Internal Communication

Secure communication between services:
- Mutual TLS for service-to-service
- Certificate pinning
- Service mesh (optional)
- Network policies (Kubernetes)

## Input Validation

### 1. Request Validation

All API requests are validated:
- JSON schema validation
- Type checking
- Size limitations
- Content filtering

### 2. Plugin Input Sanitization

Plugin inputs are sanitized:
- SQL injection prevention
- XSS protection
- Command injection prevention
- Path traversal prevention

## Security Headers

API responses include security headers:
```go
// Security headers middleware
app.Use(helmet.New(helmet.Config{
    XSSProtection: "1; mode=block",
    ContentTypeNosniff: "nosniff",
    XFrameOptions: "SAMEORIGIN",
    HSTSMaxAge: 31536000,
    HSTSExcludeSubdomains: false,
    HSTSPreloadEnabled: true,
}))
```

## Audit & Logging

### 1. Audit Trail

All significant events are logged:
- User authentication
- API access
- Workflow execution
- Plugin registration
- Configuration changes
- Security events

### 2. Log Protection

Audit logs are protected:
- Immutable storage
- Tamper-evident logging
- Secure transmission
- Access controls

## Vulnerability Management

### 1. Regular Scanning

- Container vulnerability scanning
- Dependency security scanning
- Network vulnerability scanning
- Code security scanning

### 2. Patch Management

- Automated security updates
- Regular dependency updates
- Quick response to CVEs
- Security bulletin monitoring

## Security Configuration

### 1. Environment Variables

Secure environment configuration:
```bash
# .env.secure
JWT_SECRET=very_long_random_string
DB_PASSWORD=complex_password_with_special_chars
TEMPORAL_AUTH=secure_temporal_auth_token
PLUGIN_SANDBOX=true
RATE_LIMIT_ENABLED=true
CORS_ORIGINS=trusted-domains.com
```

### 2. Configuration File

Secure configuration example:
```yaml
security:
  jwt:
    secret: ${JWT_SECRET}
    expiry: 24h
    refresh_expiry: 168h
  
  https:
    enabled: true
    redirect_http: true
    hsts: true
    hsts_max_age: 31536000
  
  rate_limit:
    enabled: true
    requests_per_second: 100
    burst_size: 200
  
  cors:
    allow_origins: ["https://yourdomain.com"]
    allow_methods: ["GET", "POST", "OPTIONS"]
    allow_headers: ["Authorization", "Content-Type"]
  
  plugins:
    sandbox_enabled: true
    allow_network_access: false
    max_memory_mb: 256
    max_cpu_percentage: 80
    timeout: 30s
```

## Security Best Practices

### 1. Deployment Security

- Run containers as non-root user
- Use read-only file systems when possible
- Implement resource limits
- Use trusted base images
- Scan images for vulnerabilities

### 2. Network Security

- Use private networks for internal services
- Implement network segmentation
- Use SSL/TLS for all connections
- Implement firewall rules
- Monitor network traffic

### 3. Data Security

- Encrypt sensitive data
- Use secure key management
- Implement data masking
- Regular security audits
- Data retention policies

### 4. Access Control

- Principle of least privilege
- Regular access reviews
- Multi-factor authentication
- Session management
- Password policies

### 5. Incident Response

#### Security Event Classification
- **Critical**: System compromise, data breach
- **High**: Potential security issue
- **Medium**: Suspicious activity
- **Low**: Informational

#### Response Procedures
1. Containment
2. Evidence preservation
3. Analysis
4. Remediation
5. Documentation
6. Communication

## Security Testing

### 1. Penetration Testing

Regular penetration testing:
- Network penetration testing
- Application security testing
- API security testing
- Social engineering testing

### 2. Vulnerability Assessment

Automated security testing:
- Static code analysis
- Dynamic application testing
- Infrastructure scanning
- Dependency scanning

### 3. Security Audits

Regular security audits:
- Configuration reviews
- Access control reviews
- Policy compliance
- Security training

## Compliance

### 1. Security Standards

Citadel Agent follows security standards:
- OWASP Top 10
- NIST Cybersecurity Framework
- ISO 27001
- SOC 2 Type II

### 2. Industry Compliance

Support for compliance requirements:
- GDPR (privacy)
- HIPAA (healthcare)
- PCI DSS (payments)
- SOX (financial)

## Security Monitoring

### 1. Security Events

Monitor for security events:
- Failed authentication attempts
- Unauthorized access attempts
- Suspicious API usage
- Anomalous plugin behavior
- Configuration changes
- Network anomalies

### 2. Security Dashboards

Key security metrics:
- Authentication success/failure rates
- API request patterns
- Plugin execution patterns
- Network traffic anomalies
- System resource usage

## Emergency Procedures

### 1. Security Incident Response

In case of security incident:
1. Activate incident response team
2. Isolate affected systems
3. Preserve evidence
4. Assess impact
5. Implement containment
6. Communicate appropriately
7. Document everything
8. Recover and restore
9. Post-incident review

### 2. Breach Notification

Breach notification procedures:
- Internal notification within 1 hour
- Customer notification within 24 hours
- Regulatory notification per requirements
- Public disclosure if required

This security framework ensures Citadel Agent is deployed and operated with security as a primary concern.