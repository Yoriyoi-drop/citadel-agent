# Installation Guide

## System Requirements

Before installing Citadel Agent, make sure your system meets the following requirements:

- **Node.js**: v16.0.0 or higher
- **NPM**: v8.0.0 or higher (usually comes with Node.js)
- **Docker**: v20.10.0 or higher
- **Docker Compose**: v2.0.0 or higher
- **Operating System**: Linux, macOS, or Windows with WSL2
- **Memory**: 4GB RAM recommended
- **Storage**: 2GB available space

## Quick Installation

Install Citadel Agent globally using npm:

```bash
npm install -g @citadel-agent/cli
```

## Alternative Installation Methods

### Using Yarn
```bash
yarn global add @citadel-agent/cli
```

### Using npx (No installation required)
```bash
npx @citadel-agent/cli install
npx @citadel-agent/cli start
```

## Post-Installation Setup

1. **Install Citadel Agent:**
```bash
citadel install
```

2. **Start the services:**
```bash
citadel start
```

3. **Verify installation:**
```bash
citadel status
```

## Configuration

Citadel Agent will create configuration files in `~/.citadel-agent/` after the first install. The default configuration includes:

- API server on port 5001
- PostgreSQL database with default credentials
- Redis for caching and session management

You can customize the configuration by editing the `.env` file in `~/.citadel-agent/`.

## First Run

After installation and startup, Citadel Agent will be available at:

- API: http://localhost:5001
- Health check: http://localhost:5001/health

## Troubleshooting

### Docker Permission Issues
If you encounter Docker permission issues:

```bash
sudo usermod -aG docker $USER
```

Then log out and log back in.

### Port Already in Use
If ports are already in use, modify the `.env` file in `~/.citadel-agent/` to use different ports.

### Reset Installation
To reset the entire installation:

```bash
citadel reset
citadel install
```

## Updating

To update Citadel Agent to the latest version:

```bash
npm update -g @citadel-agent/cli
```

## Uninstall

To uninstall Citadel Agent:

```bash
npm uninstall -g @citadel-agent/cli
```

Additionally, to remove all data and configuration files:

```bash
rm -rf ~/.citadel-agent/
```

## Support

For support and questions, please visit our [GitHub repository](https://github.com/citadel-agent/citadel-agent).