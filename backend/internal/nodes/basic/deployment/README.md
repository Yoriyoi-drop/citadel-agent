# Basic Deployment Nodes

This directory contains basic deployment nodes for the Citadel Agent workflow engine.

## Overview
Basic deployment nodes provide fundamental deployment operations including:
- Simple file deployment
- Basic service restart
- Environment configuration
- Basic health checks
- Simple rollback operations

## Available Nodes
- `file_deploy_node.go` - Deploys files to target systems
- `service_restart_node.go` - Restarts services
- `env_config_node.go` - Configures environment variables
- `health_check_node.go` - Performs basic health checks

## Usage
These nodes are designed for basic deployment workflows with simple requirements.

## Requirements
- Access to deployment targets
- Appropriate deployment credentials
- Network connectivity to targets

## Security & Permissions
- Requires deployment access permissions
- Basic system operation privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform