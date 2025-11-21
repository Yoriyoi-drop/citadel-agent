# Citadel Agent - Technical Architecture Documentation

## Overview
Citadel Agent is an advanced workflow automation platform with AI agent capabilities, multi-language runtime, and advanced sandboxing security. This document outlines the technical architecture, components, and security implementations.

## System Architecture

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CITADEL-AGENT ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   FRONTEND  │  │   API       │  │ WORKFLOW    │  │     AI      │    │
│  │    UI       │  │   GATEWAY   │  │   ENGINE    │  │   AGENTS    │    │
│  │             │  │             │  │             │  │             │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│         │                   │                   │                │     │
│         ▼                   ▼                   ▼                ▼     │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                    SECURITY LAYER                                   ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  ││
│  │  │ AUTH        │ │ VALIDATION  │ │ RATE LIMIT  │ │ POLICY      │  ││
│  │  │ MIDDLEWARE  │ │ MIDDLEWARE  │ │ MIDDLEWARE  │ │ ENFORCEMENT │  ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                              │                                          │
│                              ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                   DATA LAYER                                        ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  ││
│  │  │ POSTGRES    │ │    REDIS    │ │  SECURITY   │ │   LOGGING   │  ││
│  │  │   DB        │ │   CACHE     │ │   POLICY    │ │             │  ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Workflow Engine
The workflow engine is the core of the platform that manages and executes workflows.

#### Key Features:
- Parallel execution of workflow nodes
- Dependency resolution and management
- Advanced error handling and retry mechanisms
- Comprehensive monitoring and observability

#### Architecture:
- **Runner**: Manages workflow execution lifecycle
- **Executor**: Executes nodes concurrently
- **Scheduler**: Schedules workflow execution
- **Security Manager**: Manages runtime validation and permissions
- **Monitoring System**: Collects metrics and tracks execution

### 2. Security Implementation

#### Runtime Sandboxing
The system implements multi-layer security for code execution:

```go
type RuntimeSandbox struct {
    config *SandboxConfig
}

type SandboxConfig struct {
    MaxExecutionTime    time.Duration
    MaxMemory           int64
    MaxOutputLength     int
    AllowedHosts        []string
    BlockedPaths        []string
    AllowedCapabilities []string
    EnableNetwork       bool
    EnableFileAccess    bool
}
```

#### Security Policy
A comprehensive security policy engine validates all operations:

```go
type SecurityPolicy struct {
    AllowedHosts           []string
    BlockedIPs             []*net.IPNet
    AllowedIPs             []*net.IPNet
    RateLimits             map[string]*RateLimit
    ContentSecurityPolicy  *CSPConfig
    NetworkFilter          *NetworkFilter
    AuditLoggingEnabled    bool
    MaxRequestSize         int64
    MaxUploadSize          int64
    SessionTimeout         time.Duration
    PasswordPolicy         *PasswordPolicy
    APIKeyPolicy           *APIKeyPolicy
}
```

#### API Security
- JWT-based authentication
- Role-based access control (RBAC)
- API key management with permission scoping
- Rate limiting and request validation
- Audit logging for all security-relevant events

### 3. Authentication & Authorization System

#### Architecture:
- **AuthService**: Handles authentication and authorization operations
- **UserRepository**: Database operations for user management
- **APIKeyRepository**: API key management
- **TeamRepository**: Team and member management

#### Features:
- User registration and login
- OAuth integration (GitHub, Google)
- Session management and refresh tokens
- API key creation with scoped permissions
- Role-based access control
- Password policy enforcement

### 4. Frontend Architecture

#### Technology Stack:
- React 18 with functional components
- React Router for navigation
- React Flow for visual workflow builder
- Tailwind CSS for styling
- Zustand for state management
- Axios for API communication

#### Key Components:
- **AuthContext**: Authentication state management
- **WorkflowContext**: Workflow state management
- **MainLayout**: Main application layout with sidebar
- **WorkflowBuilder**: Visual workflow editor
- **Dashboard**: Analytics and workflow management

### 5. API Layer

#### Security Middleware:
- Request validation and sanitization
- Rate limiting with configurable policies
- Content Security Policy headers
- CORS configuration
- XSS protection headers
- Authentication and authorization
- Audit logging

#### Endpoints:
- `/api/v1/auth/*` - Authentication endpoints
- `/api/v1/workflows/*` - Workflow management
- `/api/v1/executions/*` - Execution management
- `/api/v1/users/*` - User management
- `/api/v1/teams/*` - Team management

## Security Implementation Details

### 1. Code Execution Sandboxing
The system provides secure execution environments for multiple languages:

#### JavaScript Execution:
- Syntax and pattern validation
- Blocked dangerous functions (eval, Function, etc.)
- Limited to safe APIs
- Timeout enforcement

#### Python Execution:
- Subprocess isolation
- Blocked dangerous modules (os, subprocess, etc.)
- Limited built-in functions
- Resource limiting

#### Shell Execution:
- Command validation
- Blocked dangerous commands
- Limited to allowed operations
- Output length limiting

### 2. Network Security
- Egress proxy with domain whitelisting
- IP blocking and allowing
- SSRF protection
- Maximum redirect limits
- Network call validation

### 3. Data Security
- Input sanitization and validation
- Output encoding
- SQL injection prevention
- NoSQL injection prevention
- Directory traversal protection

### 4. API Security
- JWT token validation
- API key authentication
- Rate limiting by endpoint
- Request size limits
- Content type validation

### 5. Authentication Security
- Bcrypt password hashing
- Secure token generation
- Session timeout enforcement
- Account lockout after failed attempts
- Password complexity requirements

## Deployment Architecture

### Backend Services:
1. **API Service**: Handles HTTP requests and business logic
2. **Worker Service**: Executes workflow nodes
3. **Scheduler Service**: Manages workflow scheduling
4. **Database Service**: PostgreSQL for data storage
5. **Cache Service**: Redis for caching and queuing

### Security in Deployment:
- Network isolation between services
- Resource limits per container
- Health checks and monitoring
- Automated backups
- SSL/TLS termination

## Monitoring and Observability

### Metrics Collection:
- Request rates and latencies
- Error rates and types
- Resource utilization
- Security events
- Business metrics

### Tracing:
- End-to-end request tracing
- Performance bottleneck identification
- Node execution tracking
- Error correlation

### Alerting:
- Threshold-based alerts
- Anomaly detection
- Multi-channel notifications
- Escalation policies

## Development Best Practices

### Security Practices:
- Input validation at all boundaries
- Principle of least privilege
- Regular security audits
- Automated security scanning
- Secure coding guidelines

### Code Quality:
- Comprehensive testing (unit, integration, e2e)
- Code review processes
- Static analysis tools
- Dependency management
- Documentation standards

### Performance:
- Load testing
- Performance monitoring
- Caching strategies
- Database optimization
- Resource efficiency

## Next Steps

### Phase 1: Core Features Completion (Completed)
- ✅ Authentication & Authorization System
- ✅ Frontend Dashboard Foundation  
- ✅ Security Sandboxing Implementation

### Phase 2: Advanced Features
- Multi-agent coordination
- Advanced workflow patterns
- Plugin marketplace
- Enhanced analytics

### Phase 3: Production Readiness
- Performance optimization
- Security audit
- Scalability testing
- Production deployment guides

## Support and Community

For support, documentation, and community resources, visit the official Citadel Agent platform.