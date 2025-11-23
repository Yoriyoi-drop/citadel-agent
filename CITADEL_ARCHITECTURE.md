# Citadel Agent Architecture Documentation

## 🎯 Vision
> Enterprise-grade, AI-powered workflow automation platform - Better than n8n with local AI capabilities

## 🏗️ System Architecture

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Web App    │  │  Mobile App  │  │   CLI Tool   │         │
│  │   (React)    │  │ (React Native)│ │    (Go)      │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼──────────────────┼──────────────────┼────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                      API GATEWAY (Fiber)                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  • Authentication & Authorization (JWT, OAuth2)          │  │
│  │  • Rate Limiting & Throttling                            │  │
│  │  • Request Validation & Sanitization                     │  │
│  │  • API Versioning (/api/v1, /api/v2)                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
┌─────────▼─────────┐ ┌─────▼────────┐ ┌──────▼──────────┐
│  WORKFLOW ENGINE  │ │  AI MANAGER  │ │  NODE REGISTRY  │
│   (Temporal.io)   │ │   (Hybrid)   │ │   (Dynamic)     │
│                   │ │              │ │                 │
│  • Orchestration  │ │  • Local AI  │ │  • 40+ Nodes    │
│  • Scheduling     │ │  • API AI    │ │  • Validation   │
│  • State Mgmt     │ │  • Routing   │ │  • Lifecycle    │
│  • Error Recovery │ │  • Cost Mgmt │ │  • Versioning   │
└─────────┬─────────┘ └──────┬───────┘ └─────────┬───────┘
          │                  │                   │
          └──────────────────┼───────────────────┘
                             │
          ┌──────────────────┼──────────────────────────┐
          │                  │                          │
┌─────────▼─────────┐ ┌─────▼────────┐ ┌──────────────▼─────────┐
│   DATA LAYER      │ │  CACHE LAYER │ │  MESSAGE QUEUE         │
│                   │ │              │ │                        │
│  • PostgreSQL     │ │  • Redis     │ │  • RabbitMQ           │
│  • DuckDB         │ │  • Memory    │ │  • Kafka (Optional)   │
│  • S3/Minio       │ │  • TTL       │ │  • Dead Letter Queue  │
└─────────┬─────────┘ └──────┬───────┘ └────────┬──────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                   OBSERVABILITY LAYER                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Prometheus  │  │ OpenTelemetry│  │   Grafana    │         │
│  │  (Metrics)   │  │   (Tracing)  │  │ (Dashboards) │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## 📁 Folder Structure

### Backend Structure
```
citadel-agent/backend/
├── cmd/                    # Main applications
│   ├── api/               # API server
│   ├── worker/            # Background worker
│   └── migrate/           # Database migrations
├── internal/              # Internal packages
│   ├── api/              # API handlers
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── validators/   # Request validators
│   ├── workflow/         # Workflow engine
│   │   ├── core/        # Core engine logic
│   │   │   ├── types/   # Type definitions
│   │   │   ├── engine/  # Execution engine
│   │   │   └── middleware/ # Engine middleware
│   │   ├── temporal/    # Temporal integration
│   │   └── observability/ # Monitoring
│   ├── nodes/            # Node implementations
│   │   ├── http/        # HTTP nodes
│   │   ├── database/    # Database nodes
│   │   ├── ai/          # AI nodes
│   │   ├── utility/     # Utility nodes
│   │   ├── security/    # Security nodes
│   │   ├── integration/ # Integration nodes
│   │   ├── scheduler/   # Scheduling nodes
│   │   └── analytics/   # Analytics nodes
│   ├── database/         # Database layer
│   ├── cache/            # Caching layer
│   ├── queue/            # Message queue
│   ├── auth/             # Authentication
│   ├── config/           # Configuration
│   └── interfaces/       # Interfaces for breaking cycles
├── pkg/                  # Public packages
│   ├── utils/           # Utility functions
│   ├── errors/          # Custom errors
│   └── logger/          # Logging utilities
└── bin/                  # Compiled binaries
```

### Frontend Structure
```
citadel-agent/frontend/
├── public/               # Static assets
├── src/
│   ├── components/      # UI components
│   │   ├── workflow/    # Workflow builder
│   │   ├── nodes/       # Node components
│   │   ├── execution/   # Execution views
│   │   └── ui/          # Reusable UI components
│   ├── pages/           # Route components
│   ├── hooks/           # Custom React hooks
│   ├── stores/          # State management (Zustand)
│   ├── api/             # API client
│   ├── types/           # TypeScript definitions
│   ├── styles/          # CSS styles
│   └── utils/           # Utility functions
├── package.json
└── tsconfig.json
```

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Fiber v2
- **Workflow Engine**: Temporal.io
- **Database**: PostgreSQL 15+, DuckDB
- **Cache**: Redis 7+
- **ORM**: GORM
- **Authentication**: JWT, OAuth2
- **Monitoring**: Prometheus, Grafana, OpenTelemetry
- **Testing**: testify, gomock

### Frontend
- **Framework**: React 18 + TypeScript
- **Styling**: Tailwind CSS
- **Workflow Builder**: ReactFlow
- **State Management**: Zustand
- **Routing**: React Router v6
- **Forms**: React Hook Form + Zod
- **Testing**: Vitest + React Testing Library
- **Charts**: Recharts

### AI/ML
- **Local Models**: llama.cpp, whisper.cpp, ONNX Runtime
- **Providers**: OpenAI, Anthropic, Groq, Hugging Face
- **Types**: LLM, Vision, Speech, NLP models

## 🔧 Core Components

### 1. Node System
- **40+ Production-ready Nodes**: HTTP, Database, AI, Utility, Security, etc.
- **Dynamic Registration**: Register nodes at runtime
- **Validation**: Comprehensive input/output validation
- **Versioning**: Node version management
- **Testing**: Built-in testing framework

### 2. Workflow Engine
- **Orchestration**: Temporal-based workflow execution
- **Scheduling**: Cron, interval, event-based triggers
- **State Management**: Persistent state across executions
- **Error Handling**: Retry logic, circuit breaker, fallbacks
- **Monitoring**: Real-time execution tracking

### 3. AI Integration
- **Hybrid Approach**: Local + API fallback
- **Model Management**: Dynamic model loading
- **Cost Optimization**: Smart routing to cheapest provider
- **Privacy**: Local processing by default
- **Scalability**: Horizontal scaling of AI services

### 4. Security Model
- **Authentication**: JWT, OAuth2, API Keys
- **Authorization**: RBAC with granular permissions
- **Encryption**: Data at rest and in transit
- **Audit Logging**: Comprehensive activity tracking
- **Sandboxing**: Isolated node execution

## 🚀 Key Features

### Visual Workflow Builder
- **Drag-and-drop Interface**: Like n8n but more intuitive
- **Node Library**: 40+ pre-built nodes
- **Real-time Execution**: See results as they happen
- **Collaboration**: Multi-user editing
- **Version Control**: Git-like workflow versioning

### AI Capabilities
- **Local Processing**: First choice for privacy
- **Multiple Providers**: Auto-fallback between providers
- **Cost Management**: Real-time cost tracking
- **Model Selection**: Smart model routing
- **Response Caching**: Reduce costs and latency

### Enterprise Features
- **Multi-tenancy**: Isolated environments
- **Role-based Access**: Fine-grained permissions
- **Audit Trail**: Complete execution history
- **SLA Monitoring**: Uptime and performance tracking
- **Disaster Recovery**: Automated backups

## 📊 Performance Targets

- **Latency**: <100ms for simple workflows
- **Throughput**: 10,000+ concurrent workflows
- **Uptime**: 99.9% availability
- **Scalability**: Linear scaling to 1M+ executions/day
- **Cost**: 90% cheaper than competitors

## 🏁 Roadmap

### Phase 1: Foundation (Weeks 1-4)
- Complete architecture implementation
- Core workflow engine
- Basic node library (10 nodes)
- User authentication
- Basic UI

### Phase 2: AI Integration (Weeks 5-8)
- Local AI model integration
- Hybrid AI engine
- Advanced AI nodes (15 nodes)
- Cost optimization features

### Phase 3: Enterprise Features (Weeks 9-12)
- Multi-tenancy
- Advanced security
- Monitoring & observability
- Performance optimization
- Remaining node library (15+ nodes)

### Phase 4: Polish & Scale (Weeks 13-16)
- Performance tuning
- Documentation
- Testing
- Production deployment
- Community features

## 📦 Size Breakdown (24GB Total)

- **Backend Code**: 1.5GB
- **Frontend Code**: 1GB
- **AI Models**: 15GB
- **Database**: 2GB
- **Docker Images**: 3GB
- **Documentation**: 0.5GB
- **Tests**: 1GB
- **Tools & Scripts**: 1GB

> **Note**: This architecture is designed to be modular, scalable, and production-ready from day one.