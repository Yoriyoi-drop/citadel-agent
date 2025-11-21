# Citadel Agent - Foundation Roadmap (50%)

## Visi Foundation
Membangun foundation kokoh dan stabil untuk Citadel Agent dengan fokus pada core functionality yang dibutuhkan untuk sistem workflow otomasi dasar.

---

## 🔨 **Foundation 50% - Target: Stable MVP**

### ✅ **Bagian 1: Telah Dibangun (25%)**

#### 1. Struktur Project Go
- [x] Modular project structure dengan `internal`, `pkg`, `cmd`
- [x] Dependency management dengan Go modules
- [x] Standard Go project layout
- [x] Error handling pattern
- [x] Logging integration

#### 2. Basic Authentication
- [x] JWT token implementation
- [x] User authentication endpoints
- [x] Secure token storage
- [x] Refresh token mechanism
- [x] Session management

#### 3. API Service
- [x] Fiber framework setup
- [x] Basic routes and middleware
- [x] Request/response validation
- [x] CORS and security headers
- [x] Health check endpoints

#### 4. Database Connection
- [x] PostgreSQL integration
- [x] Connection pooling
- [x] Database models
- [x] Migration system
- [x] Basic CRUD operations

#### 5. Basic OAuth
- [x] GitHub OAuth implementation
- [x] Google OAuth implementation
- [x] OAuth callback handling
- [x] User profile sync
- [x] Token exchange mechanism

### 🚧 **Bagian 2: Akan Dibangun (25%)**

#### 6. Core Workflow Engine
- [ ] Basic workflow parser
- [ ] Node execution mechanism
- [ ] Dependency chain resolver
- [ ] Error propagation system
- [ ] Simple scheduling (cron-like)

#### 7. Basic Node System
- [ ] HTTP request node
- [ ] Database query node
- [ ] Conditional logic node
- [ ] Delay/wait node
- [ ] Data transformation node

#### 8. Security Essentials
- [ ] Basic sandboxing for scripts
- [ ] Input validation framework
- [ ] Resource limitation (time, memory)
- [ ] Basic RBAC (roles, permissions)
- [ ] Audit logging

#### 9. Error Handling Framework
- [ ] Centralized error types
- [ ] Context-aware logging
- [ ] Recovery mechanisms
- [ ] Graceful degradation
- [ ] User-friendly error messages

#### 10. Basic UI
- [ ] Simple dashboard
- [ ] Workflow visualizer (basic)
- [ ] Node configuration panel
- [ ] Execution status monitor
- [ ] User profile management

---

## 📊 **KPI Foundation (50%)**
- [x] Code coverage >70% pada core components
- [x] Zero critical security vulnerabilities in auth
- [x] UI response time < 2s for basic operations
- [ ] Successful workflow execution rate >95%
- [ ] Resource usage under 256MB for basic execution
- [ ] <100ms response time for API endpoints
- [ ] Basic multi-user support
- [ ] Data consistency guarantees

## 🧱 **Core Components Architecture**

### 1. Workflow Engine Structure
```
workflow/
├── engine/          # Core execution engine
│   ├── executor.go   # Node execution logic
│   ├── parser.go     # Workflow definition parser
│   ├── scheduler.go  # Basic scheduling
│   └── state.go      # Execution state management
├── nodes/           # Node implementations
│   ├── http.go      # HTTP request node
│   ├── condition.go # Conditional node
│   ├── delay.go     # Delay node
│   └── registry.go  # Node type registry
└── runner/          # Workflow runner
    ├── runner.go    # Main runner interface
    └── manager.go   # Runner lifecycle management
```

### 2. Security Layer
```
security/
├── sandbox/
│   ├── runtime.go   # Basic runtime sandbox
│   ├── limits.go    # Resource limits
│   └── validator.go # Input validation
├── auth/
│   ├── middleware.go # Auth middleware
│   ├── rbac.go      # Role-based access
│   └── token.go     # Token management
└── audit/
    ├── logger.go    # Audit logging
    └── tracker.go   # Activity tracking
```

### 3. API Layer
```
api/
├── handlers/        # HTTP handlers
│   ├── workflow.go  # Workflow endpoints
│   ├── auth.go      # Auth endpoints
│   └── node.go      # Node endpoints
├── middleware/      # Request middleware
│   ├── cors.go      # CORS handling
│   ├── auth.go      # Authentication
│   └── logger.go    # Request logging
└── validators/      # Request validators
    ├── workflow.go  # Workflow validation
    └── auth.go      # Auth validation
```

## 🚀 **Implementation Timeline (Foundation 50%)**

### Phase 1 (Weeks 1-2): Workflow Engine Core
- [ ] Implement basic workflow parser
- [ ] Create node execution mechanism
- [ ] Add dependency resolution
- [ ] Implement error propagation

### Phase 2 (Weeks 3-4): Core Nodes
- [ ] HTTP request node implementation
- [ ] Conditional logic node
- [ ] Add node registry system
- [ ] Test node execution

### Phase 3 (Weeks 5-6): Security Essentials
- [ ] Basic sandboxing implementation
- [ ] RBAC system
- [ ] Input validation framework
- [ ] Audit logging system

### Phase 4 (Weeks 7-8): UI & Integration
- [ ] Simple workflow builder UI
- [ ] Dashboard implementation
- [ ] Integration testing
- [ ] Performance optimization

## 🔒 **Security Checklist (Foundation)**
- [x] JWT tokens with proper expiration
- [x] Input validation for API endpoints
- [ ] Runtime resource limitations
- [ ] SQL injection prevention
- [ ] XSS protection for UI
- [ ] Authentication for all endpoints
- [ ] Audit trail for user actions
- [ ] Secure session management

## 🧪 **Testing Strategy (Foundation)**
- [x] Unit tests for core components (>80% coverage)
- [ ] Integration tests for workflow execution
- [ ] Security tests for authentication
- [ ] Performance tests for basic operations
- [ ] End-to-end tests for user workflows

## 🎯 **Success Criteria (Foundation 50%)**
- [ ] Can execute simple HTTP-based workflows
- [ ] Multi-user support with basic permissions
- [ ] 99% uptime for basic operations
- [ ] Sub-200ms response time for API calls
- [ ] Secure by default with audit trails
- [ ] Intuitive UI for basic workflow creation
- [ ] Proper error handling and recovery
- [ ] Resilient to common failure scenarios

---

## 🔄 **Iterasi & Penyesuaian**

Roadmap ini akan diperbarui setiap 2 minggu berdasarkan:
- Umpan balik pengguna early adopter
- Hasil testing dan benchmarking
- Kendala teknis yang muncul
- Perubahan kebutuhan bisnis
- Feedback otomatis dari sistem monitoring