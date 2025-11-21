# Arsitektur Citadel Agent

## Overview
Citadel Agent adalah platform otomasi workflow modern dengan kemampuan AI agent, multi-language runtime, dan sandboxing keamanan lanjutan. Arsitektur dirancang untuk skalabilitas, keamanan, dan modularitas.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CITADEL AGENT SYSTEM                               │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────────┐  │
│  │   FRONTEND UI   │    │  CLI TOOLS      │    │   EXTERNAL CLIENTS      │  │
│  │                 │    │                 │    │                         │  │
│  │  • Dashboard    │    │  • CLI Tool     │    │  • REST API            │  │
│  │  • Workflow     │    │  • Advanced CLI │    │  • Webhooks            │  │
│  │    Editor       │    │                 │    │  • Third-party Apps    │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────────────┘  │
│                           │         │                          │            │
│                           │         │                          │            │
│                           ▼         ▼                          ▼            │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                     LOAD BALANCER / REVERSE PROXY                       │  │
│  │                        (NGINX / TRAEFIK)                                │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                           │         │                                        │
│                           │         │                                        │
│                           ▼         ▼                                        │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────────┐  │
│  │    API GATEWAY  │    │   AUTH SERVICE  │    │    MONITORING & LOGS    │  │
│  │                 │    │                 │    │                         │  │
│  │  • Request      │    │  • OAuth       │    │  • Prometheus          │  │
│  │    Routing      │    │  • JWT         │    │  • Grafana             │  │
│  │  • Rate Limit   │    │  • User Mgmt   │    │  • ELK Stack           │  │
│  │  • API Version  │    │  • SSO         │    │  • Audit Logging       │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────────────┘  │
│         │                       │                           │                │
│         │                       │                           │                │
│         ▼                       ▼                           ▼                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                      MICROSERVICES LAYER                                │  │
│  ├─────────────────┬─────────────────┬─────────────────────────────────────┤  │
│  │    API SERVICE  │   WORKER        │   SCHEDULER                         │  │
│  │                 │   SERVICE       │   SERVICE                           │  │
│  │ • User mgmt     │ • Workflow      │ • Cron scheduling                   │  │
│  │ • Workflow CRUD │   execution     │ • Event scheduling                  │  │
│  │ • Node mgmt     │ • Node runtime  │ • Workflow triggers                 │  │
│  │ • Auth API      │ • Task queue    │ • Time-based events                 │  │
│  │ • Agent API     │ • Sandboxing    │                                     │  │
│  └─────────────────┴─────────────────┴─────────────────────────────────────┘  │
│              │              │                    │                           │
│              │              │                    │                           │
│              ▼              ▼                    ▼                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                      DATA LAYER                                         │  │
│  ├─────────────────┬─────────────────┬─────────────────────────────────────┤  │
│  │   POSTGRES      │     REDIS       │        FILE STORAGE                 │  │
│  │                 │                 │                                     │  │
│  │ • User data     │ • Session       │ • Workflow assets                   │  │
│  │ • Workflow def  │ • Cache         │ • Logs                              │  │
│  │ • Node configs  │ • Task queue    │ • Temp files                        │  │
│  │ • Audit logs    │ • Rate limit    │                                     │  │
│  └─────────────────┴─────────────────┴─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Service Architecture

### 1. API Service
**Responsibility**: Menyediakan REST API, mengelola otentikasi, manajemen pengguna dan workflow

```
┌─────────────────────────────────────┐
│            API SERVICE              │
├─────────────────────────────────────┤
│ • Fiber HTTP Framework              │
│ • JWT Authentication                │
│ • Rate Limiting                     │
│ • Request Validation                │
│ • API Versioning                    │
│ • CORS Management                   │
│ • Request Logging                   │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│         CONTROLLERS                 │
├─────────────────────────────────────┤
│ • Auth Controller                   │
│ • Workflow Controller               │
│ • Node Controller                   │
│ • AI Agent Controller               │
│ • User Controller                   │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│          SERVICES                   │
├─────────────────────────────────────┤
│ • Auth Service                      │
│ • Workflow Service                  │
│ • Node Service                      │
│ • AI Agent Service                  │
│ • User Service                      │
└─────────────────────────────────────┘
```

### 2. Worker Service
**Responsibility**: Menjalankan eksekusi workflow dan node, menangani task dari queue

```
┌─────────────────────────────────────┐
│           WORKER SERVICE            │
├─────────────────────────────────────┤
│ • Task Queue Consumer               │
│ • Workflow Execution Engine         │
│ • Node Runtime Manager              │
│ • Sandboxing System                 │
│ • Resource Monitoring               │
│ • Error Handling & Retry            │
└─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────┐
│        EXECUTION RUNTIMES           │
├─────────────────────────────────────┤
│ • Go Runtime                        │
│ • JavaScript/Node Runtime           │
│ • Python Runtime                    │
│ • Java Runtime                      │
│ • Ruby Runtime                      │
│ • PHP Runtime                       │
│ • Rust Runtime                      │
│ • Container Runtime                 │
└─────────────────────────────────────┘
```

### 3. Scheduler Service
**Responsibility**: Menjadwalkan eksekusi workflow berbasis waktu dan event

```
┌─────────────────────────────────────┐
│         SCHEDULER SERVICE           │
├─────────────────────────────────────┤
│ • Cron Expression Parser            │
│ • Time-based Scheduling             │
│ • Event-driven Scheduling           │
│ • Workflow Trigger Management       │
│ • Schedule Persistence              │
│ • Dead Letter Queue                 │
└─────────────────────────────────────┘
```

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SECURITY LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │ AUTHENTICATION  │  │ AUTHORIZATION   │  │    RUNTIME SECURITY         │  │
│  │                 │  │                 │  │                             │  │
│  │ • OAuth 2.0     │  │ • Role-based    │  │ • Container Sandboxing    │  │
│  │ • JWT Tokens    │  │ • Permission    │  │ • VM Isolation            │  │
│  │ • Multi-factor  │  │ • Policy-based  │  │ • Resource Limits         │  │
│  │ • SSO Support   │  │ • RBAC System   │  │ • Network Isolation       │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│              │                   │                      │                   │
│              ▼                   ▼                      ▼                   │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                       DATA SECURITY                                     │  │
│  │ • Encryption at Rest                                                    │  │
│  │ • Encryption in Transit                                                 │  │
│  │ • Audit Logging                                                         │  │
│  │ • Data Loss Prevention                                                  │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Workflow Execution Flow
```
Client Request → API Service → Validate → Store Workflow → Queue Task → Worker Service → Execute Nodes → Update Status → Response
```

### Node Execution Flow
```
Workflow Node → Runtime Selector → Sandboxing → Code Execution → Result Capture → Status Update
```

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DEPLOYMENT ARCHITECTURE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │   KUBERNETES    │  │  DOCKER       │  │    MONOLITHIC                 │  │
│  │   CLUSTER       │  │  COMPOSE      │  │    DEPLOYMENT               │  │
│  │                 │  │                 │  │                             │  │
│  │ • Auto-scaling  │  │ • Multi-service │  │ • Single binary deployment  │  │
│  │ • Load balancing│  │ • Network       │  │ • All-in-one executable     │  │
│  │ • Health checks │  │ • Volumes       │  │ • Simple setup              │  │
│  │ • Rollout mgmt  │  │ • Secrets       │  │                             │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Web Framework**: Fiber
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: Redis/RabbitMQ
- **Runtime**: Multiple (Go, JS, Python, Java, etc.)

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes (optional)
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack

### Security
- **Authentication**: JWT, OAuth 2.0
- **Authorization**: RBAC
- **Runtime Isolation**: Containers, VMs, Sandboxing
- **Encryption**: TLS, AES

## Scalability Considerations

1. **Vertical Scaling**: Each service can be scaled independently
2. **Horizontal Scaling**: Services can be replicated
3. **Database Scaling**: Read replicas, connection pooling
4. **Caching**: Multi-layer caching with Redis
5. **Load Distribution**: Multiple worker instances
6. **Event-driven**: Asynchronous processing for high throughput