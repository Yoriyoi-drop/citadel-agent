# Citadel Agent

> Enterprise-grade, AI-powered workflow automation platform - Better than n8n with local AI capabilities

## 🚀 Features

- **Visual Workflow Builder**: Drag-and-drop interface like n8n
- **AI-Powered Nodes**: Local and API-based AI integration
- **40+ Production-ready Nodes**: HTTP, Database, AI, Security, Utility, etc.
- **Scalable Architecture**: Handle 10,000+ concurrent workflows
- **Privacy-First**: Data stays on your infrastructure
- **Cost-Effective**: 90% cheaper than competitors
- **Self-Contained**: All dependencies included

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Fiber v2
- **Workflow Engine**: Temporal.io
- **Database**: PostgreSQL 15+, DuckDB
- **Cache**: Redis 7+
- **AI/ML**: Local models with llama.cpp, whisper.cpp

### Frontend
- **Framework**: React 18 + TypeScript
- **Styling**: Tailwind CSS
- **Workflow Builder**: ReactFlow
- **State Management**: Zustand

## 📁 Project Structure

```
citadel-agent/
├── backend/                    # Go backend
│   ├── cmd/                   # Main applications
│   ├── internal/              # Internal packages
│   │   ├── api/              # API handlers
│   │   ├── workflow/         # Workflow engine
│   │   ├── nodes/            # Node implementations
│   │   └── ...
│   └── pkg/                  # Public libraries
├── frontend/                  # React frontend
├── ai-models/                 # Local AI models (15GB)
├── database/                  # Database setup
├── docker/                    # Docker configuration
├── docs/                      # Documentation
├── tests/                     # Test suites
└── scripts/                   # Automation scripts
```

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/your-org/citadel-agent.git

# Navigate to project directory
cd citadel-agent

# Setup environment
cp .env.example .env
# Edit .env with your configuration

# Download AI models (optional, for local AI)
./scripts/download-models.sh

# Start the application
docker-compose up --build

# Or run locally
# Backend
cd backend && go run cmd/api/main.go
# Frontend
cd frontend && npm install && npm run dev
```

## 📖 Documentation

- [Getting Started](./docs/getting-started/quick-start.md)
- [Workflow Guide](./docs/guides/workflow-design.md)
- [Node Development](./docs/guides/node-development.md)
- [AI Integration](./docs/guides/ai-integration.md)
- [API Reference](./docs/api-reference/rest-api.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](./LICENSE) file for details.