# @citadel-agent/cli

Citadel Agent - Enterprise Workflow Automation Platform CLI

Citadel Agent is a powerful workflow automation platform designed to handle complex systems with 200+ built-in nodes, enterprise security, and cloud-native scalability. It's built as a more modern, faster, and lighter alternative to n8n.

## Installation

To install Citadel Agent globally:

```bash
npm install -g @citadel-agent/cli
```

## Quick Start

1. Install Citadel Agent:
```bash
citadel install
```

2. Start the services:
```bash
citadel start
```

3. Access the platform:
- API: http://localhost:5001
- UI: http://localhost:3000 (when available)

4. Stop the services:
```bash
citadel stop
```

## Commands

- `citadel install` - Install Citadel Agent locally
- `citadel start` - Start Citadel Agent services
- `citadel stop` - Stop Citadel Agent services
- `citadel status` - Check status of services
- `citadel reset` - Reset all data (⚠️ irreversible)
- `citadel version` - Show version information

## Requirements

- Node.js v16 or higher
- Docker
- Docker Compose

## Features

- 🏗️ **Foundation Engine**: Robust workflow execution with dependency resolution
- 🔐 **Enterprise Security**: Node sandboxing, SSRF protection, RBAC
- ⚡ **High Performance**: Optimized for speed and scalability
- 🧩 **Extensible Nodes**: 200+ built-in nodes with plugin system
- 🌐 **Real-time Updates**: WebSocket support for live workflow monitoring
- 📊 **Monitoring**: Built-in metrics and observability

## Architecture

Citadel Agent follows a microservices architecture:
- **API Service**: Handles REST API requests and workflow management
- **Worker Service**: Executes workflow nodes in isolated environments
- **Scheduler Service**: Manages scheduled workflows and triggers

## Documentation

For full documentation, visit [Citadel Agent Documentation](https://citadel-agent.com/docs)

## Contributing

We welcome contributions! Please see our [contributing guide](CONTRIBUTING.md) for more details.

## License

Apache 2.0 - see the [LICENSE](LICENSE) file for details.
