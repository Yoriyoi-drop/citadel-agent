# Citadel Agent - Automation Workflow Platform

Citadel Agent adalah platform automation workflow enterprise-grade yang dirancang untuk menangani sistem kompleks seperti n8n. Platform ini dirancang dengan backend Go, frontend React, dan dukungan plugin untuk ekstensibilitas maksimal.

## 🏗️ Arsitektur

### Backend (Go)
- **API Layer**: Fiber/Gin web framework
- **Worker Executor**: Eksekusi workflow dan task
- **Scheduler**: Penjadwalan cron, interval, dan trigger
- **Core Engine**: Workflow execution engine
- **Plugin Runtime**: Node.js plugin system

### Frontend (React + TypeScript)
- **React Flow**: Canvas drag-and-drop untuk workflow
- **Zustand**: State management
- **TypeScript**: Type safety

### Database & Caching
- **PostgreSQL**: Penyimpanan data utama
- **Redis**: Session dan caching

### Deployment
- **Docker**: Containerisasi
- **Docker Compose**: Multi-service orchestration

## 📁 Struktur Project

```
/automation-platform
│
├── backend/
│   ├── cmd/
│   │   ├── api/                   # main API server
│   │   ├── worker/                # workflow executor worker
│   │   └── scheduler/             # scheduler (cron, interval, trigger)
│   │
│   ├── internal/
│   │   ├── config/                # environment, config loader
│   │   ├── database/              # PostgreSQL & Redis connections
│   │   ├── models/                # struct models (Workflow, Node, User, etc.)
│   │   ├── repositories/          # database CRUD logic
│   │   ├── services/              # business logic
│   │   ├── engine/                # CORE WORKFLOW ENGINE
│   │   ├── plugins/               # Plugin loader and sandbox
│   │   ├── api/                   # API controllers and routes
│   │   ├── utils/                 # utility functions
│   │   └── auth/                  # authentication & authorization
│   │
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── WorkflowCanvas/    # React Flow components
│   │   │   ├── Sidebar/
│   │   │   ├── Inspector/
│   │   │   └── Dashboard/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── store/
│   │   ├── services/
│   │   ├── utils/
│   │   └── types/
│   │
│   └── package.json
│
├── plugins/                       # marketplace plugin user
│   ├── js/
│   └── python/
│
├── docker/
│   ├── api.Dockerfile
│   ├── worker.Dockerfile
│   ├── frontend.Dockerfile
│   └── docker-compose.yml
│
└── scripts/
    ├── start.sh
    ├── migrate.sh
    └── seed.sh
```

## 🚀 Cara Menjalankan

### Development
```bash
# Jalankan semua service dengan docker-compose
docker-compose -f docker/docker-compose.yml up --build
```

### Development Manual
```bash
# Setup database
./scripts/migrate.sh

# Jalankan API server
cd backend && go run cmd/api/main.go

# Jalankan worker
cd backend && go run cmd/worker/main.go

# Jalankan frontend
cd frontend && npm start
```

## 🛠️ Teknologi yang Digunakan

### Backend
- **Go**: Bahasa utama untuk performa tinggi
- **Fiber**: Web framework cepat
- **GORM**: ORM untuk database
- **Redis**: Caching dan session
- **PostgreSQL**: Database relasional

### Frontend
- **React**: UI library
- **TypeScript**: Type safety
- **React Flow**: Workflow canvas
- **Zustand**: State management
- **Axios**: HTTP client

### Deployment
- **Docker**: Containerisasi
- **Docker Compose**: Orkestrasi multi-container

## 🧩 Plugin System

Platform ini mendukung plugin untuk ekstensibilitas:

- **JavaScript Plugin**: Di sandbox untuk keamanan
- **Python Plugin**: Untuk AI/ML tasks

## 🔐 Keamanan

- **JWT Authentication**: Untuk session management
- **RBAC**: Role-based access control
- **Sandboxed Plugins**: Untuk keamanan plugin

## 📊 Fitur Utama

- Workflow designer drag-and-drop
- Node scheduling (cron, interval)
- Real-time execution monitoring
- Plugin marketplace
- Multi-tenant support
- Audit logging
- REST API dan WebSocket

## 🤝 Kontribusi

Lihat `CONTRIBUTING.md` untuk panduan berkontribusi.