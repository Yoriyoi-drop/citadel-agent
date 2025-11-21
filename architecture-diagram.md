# Citadel Agent - Arsitektur Lengkap

## Overview
Citadel Agent adalah platform otomasi workflow modern dengan kemampuan AI agent, multi-language runtime, dan sandboxing keamanan lanjutan. Arsitektur dirancang untuk skalabilitas, keamanan, dan modularitas.

## High-Level Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        A[Web UI / Dashboard]
        B[CLI Tools]
        C[API Clients]
        D[Mobile App]
    end
    
    subgraph "Load Balancer / Gateway"
        E[API Gateway]
        F[Rate Limiter]
        G[CORS Handler]
    end
    
    subgraph "Authentication Service"
        H[OAuth 2.0]
        I[JWT Token Service]
        J[User Management]
    end
    
    subgraph "Core Services"
        K[API Service]
        L[Workflow Engine]
        M[Scheduler Service]
        N[Worker Service]
    end
    
    subgraph "Runtime Services"
        O[Go Runtime]
        P[Python Runtime]
        Q[JavaScript Runtime]
        R[Java Runtime]
        S[Container Runtime]
        T[AI Agent Runtime]
    end
    
    subgraph "Data Layer"
        U[(PostgreSQL)]
        V[(Redis Cache)]
        W[(Object Storage)]
    end
    
    subgraph "Monitoring & Logging"
        X[Prometheus]
        Y[Grafana]
        Z[ELK Stack]
    end
    
    subgraph "Security Layer"
        AA[Sandbox Manager]
        BB[Policy Engine]
        CC[Audit Logger]
    end
    
    A --> E
    B --> E
    C --> E
    D --> E
    
    E --> H
    H --> K
    I --> K
    J --> K
    
    K --> L
    K --> M
    K --> N
    
    L --> O
    L --> P
    L --> Q
    L --> R
    L --> S
    L --> T
    
    M --> L
    N --> L
    
    O --> AA
    P --> AA
    Q --> AA
    R --> AA
    S --> AA
    T --> AA
    
    AA --> BB
    AA --> CC
    
    K --> U
    L --> U
    M --> U
    N --> U
    
    K --> V
    L --> V
    M --> V
    N --> V
    
    O --> W
    P --> W
    Q --> W
    R --> W
    S --> W
    T --> W
    
    CC --> Z
    X --> Y
```

## Microservices Architecture

```mermaid
graph LR
    subgraph "citadel-agent-cluster"
        direction TB
        
        subgraph "Gateway Layer"
            APIGW[API Gateway]
            LB[Load Balancer]
        end
        
        subgraph "Core Services"
            API[API Service]
            WORKER[Worker Service]
            SCHED[Scheduler Service]
        end
        
        subgraph "Authentication"
            AUTH[Auth Service]
            OAUTH[OAuth Service]
        end
        
        subgraph "Runtime Pool"
            RT1[Runtime Instance 1]
            RT2[Runtime Instance 2]
            RT3[Runtime Instance 3]
            RTN[Runtime Instance N]
        end
        
        subgraph "Data Layer"
            POSTGRES[(PostgreSQL Cluster)]
            REDIS[(Redis Cluster)]
            MINIO[(MinIO Object Store)]
        end
        
        subgraph "Monitoring"
            PROM[Prometheus]
            GRAFANA[Grafana]
            JAEGER[Jaeger Tracing]
        end
    end
    
    LB --> APIGW
    APIGW --> API
    APIGW --> AUTH
    APIGW --> OAUTH
    
    API --> WORKER
    API --> SCHED
    AUTH --> API
    
    SCHED --> WORKER
    WORKER --> RT1
    WORKER --> RT2
    WORKER --> RT3
    WORKER --> RTN
    
    API --> POSTGRES
    API --> REDIS
    WORKER --> POSTGRES
    WORKER --> REDIS
    SCHED --> POSTGRES
    SCHED --> REDIS
    
    RT1 --> MINIO
    RT2 --> MINIO
    RT3 --> MINIO
    RTN --> MINIO
    
    API -.-> PROM
    WORKER -.-> PROM
    SCHED -.-> PROM
    RT1 -.-> PROM
    RT2 -.-> PROM
    RT3 -.-> PROM
    RTN -.-> PROM
    
    PROM --> GRAFANA
    API -.-> JAEGER
    WORKER -.-> JAEGER
    SCHED -.-> JAEGER
    RT1 -.-> JAEGER
    RT2 -.-> JAEGER
    RT3 -.-> JAEGER
    RTN -.-> JAEGER
```

## Workflow Execution Flow

```mermaid
sequenceDiagram
    participant Client as Client
    participant API as API Service
    participant DB as Database
    participant Scheduler as Scheduler
    participant Worker as Worker
    participant Runtime as Runtime
    participant Sandbox as Sandbox
    participant ExtSvc as External Service

    Client->>API: Create Workflow
    API->>DB: Store workflow definition
    API-->>Client: Workflow Created

    Client->>API: Execute Workflow
    API->>DB: Validate workflow exists
    API->>Worker: Queue workflow execution
    Worker->>DB: Fetch workflow definition
    Worker->>Sandbox: Initialize secure runtime
    loop For each node in workflow
        Sandbox->>Runtime: Execute node
        alt Runtime is external service
            Runtime->>ExtSvc: Call external service
            ExtSvc-->>Runtime: Response
        else Runtime is script
            Runtime-->>Sandbox: Execution result
        end
        Sandbox-->>Worker: Node execution result
    end
    Worker->>DB: Update execution status
    Worker-->>API: Execution completed
    API-->>Client: Execution result
```

## Component Interaction Diagram

```mermaid
graph LR
    subgraph "Frontend"
        A[React Dashboard]
        B[Flow Builder UI]
        C[Real-time Logs]
    end
    
    subgraph "Backend Services"
        D[REST API]
        E[GraphQL API]
        F[WebSocket Service]
    end
    
    subgraph "Workflow Engine"
        G[Workflow Parser]
        H[Node Executor]
        I[Condition Evaluator]
        J[Loop Processor]
    end
    
    subgraph "Runtime Managers"
        K[Go Runtime Manager]
        L[Python Runtime Manager]
        M[JS Runtime Manager]
        N[Container Runtime Manager]
        O[AI Agent Manager]
    end
    
    subgraph "Security"
        P[Sandbox Manager]
        Q[Policy Evaluator]
        R[Secrets Provider]
    end
    
    subgraph "Storage"
        S[PostgreSQL]
        T[Redis]
        U[Object Storage]
    end

    A --> D
    B --> D
    C --> F
    
    D --> G
    E --> G
    F --> C
    
    G --> H
    G --> I
    G --> J
    
    H --> K
    H --> L
    H --> M
    H --> N
    H --> O
    
    K --> P
    L --> P
    M --> P
    N --> P
    O --> P
    
    P --> Q
    R --> P
    R --> K
    R --> L
    R --> M
    R --> N
    R --> O
    
    D --> S
    G --> S
    H --> S
    K --> U
    L --> U
    M --> U
    N --> U
    O --> U
    
    D --> T
    G --> T
    H --> T
    P --> T
```

## Security Architecture

```mermaid
graph TD
    subgraph "Security Controls"
        A[Network Segmentation]
        B[Firewall Rules]
        C[Identity & Access Mgmt]
        D[Secrets Management]
        E[Runtime Isolation]
        F[Audit Logging]
        G[Vulnerability Scanning]
        H[Intrusion Detection]
    end
    
    subgraph "Application Layer"
        I[OAuth 2.0/JWT]
        J[RBAC System]
        K[Sandboxing]
        L[Code Signing]
        M[Image Scanning]
    end
    
    subgraph "Infrastructure"
        N[Container Security]
        O[Kubernetes Security]
        P[API Gateway Security]
        Q[Transport Encryption]
    end

    A --> C
    B --> C
    C --> I
    D --> I
    E --> K
    F --> I
    G --> N
    H --> A
    
    I --> J
    J --> K
    K --> L
    L --> M
    M --> N
    N --> O
    P --> Q
```

## Scalability Architecture

```mermaid
graph TB
    subgraph "Horizontal Scaling"
        A[API Service Cluster]
        B[Worker Service Pool]
        C[Scheduler Service Cluster]
        D[Runtime Instances]
    end
    
    subgraph "Load Distribution"
        E[Load Balancer]
        F[Service Discovery]
        G[Auto-scaling Groups]
    end
    
    subgraph "Data Partitioning"
        H[Database Sharding]
        I[Cache Clustering]
        J[Message Queues]
    end
    
    subgraph "Monitoring"
        K[Metrics Collection]
        L[Auto-scaling Triggers]
        M[Health Checks]
    end
    
    E --> A
    E --> B
    E --> C
    F --> A
    F --> B
    F --> C
    G --> A
    G --> B
    G --> C
    G --> D
    
    H --> A
    I --> A
    J --> B
    
    K --> L
    L --> G
    M --> G
    M --> A
    M --> B
    M --> C
    M --> D
```

## Data Flow Architecture

```mermaid
graph LR
    A[User Input] --> B[API Validation]
    B --> C[Authentication]
    C --> D{Workflow Type?}
    
    D -->|Simple| E[Direct Execution]
    D -->|Complex| F[Queue Management]
    D -->|Scheduled| G[Scheduler Service]
    
    E --> H[Runtime Selection]
    F --> H
    G --> H
    
    H --> I{Runtime Type?}
    I -->|Go| J[Go Runtime]
    I -->|Python| K[Python Sandbox]
    I -->|JS| L[Node.js Runtime]
    I -->|Container| M[Docker Runtime]
    I -->|AI Agent| N[AI Agent Runtime]
    
    J --> O[Sandbox Enforcement]
    K --> O
    L --> O
    M --> O
    N --> O
    
    O --> P[Execution Result]
    P --> Q[Result Storage]
    P --> R[Event Broadcasting]
    Q --> S[Audit Log]
    R --> T[Notification Service]
```

## DevOps Architecture

```mermaid
graph LR
    subgraph "Development"
        A[Code Repository]
        B[CI Pipeline]
        C[Security Scanning]
        D[Unit Tests]
    end
    
    subgraph "Testing"
        E[Integration Tests]
        F[E2E Tests]
        G[Security Tests]
        H[Load Tests]
    end
    
    subgraph "Staging"
        I[Staging Cluster]
        J[Canary Deployments]
        K[Test Workflows]
        L[Performance Baselines]
    end
    
    subgraph "Production"
        M[Prod Cluster]
        N[Blue-Green Deploy]
        O[Live Workflows]
        P[Production Monitoring]
    end
    
    A --> B
    B --> C
    B --> D
    C --> E
    D --> E
    E --> F
    E --> G
    E --> H
    F --> I
    G --> I
    H --> I
    I --> J
    I --> K
    I --> L
    J --> M
    K --> M
    L --> M
    M --> N
    N --> O
    O --> P
```

## Deployment Architecture

```mermaid
graph LR
    subgraph "On-Premises Deployment"
        A[Kubernetes Cluster]
        B[Load Balancer]
        C[External DB]
        D[Object Store]
    end
    
    subgraph "Cloud Deployment"
        E[EKS/AKS/GKE]
        F[Cloud Load Balancer]
        G[RDS/DynamoDB]
        H[S3/GCS]
    end
    
    subgraph "Hybrid Deployment"
        F1[Edge Nodes]
        F2[Regional Clusters]
        F3[Central Control Plane]
    end
    
    A --> B
    A --> C
    A --> D
    
    E --> F
    E --> G
    E --> H
    
    F1 --> F2
    F2 --> F3
```