# Intermediate Deployment Nodes

This directory contains intermediate deployment nodes for the Citadel Agent workflow engine.

## Overview
Intermediate deployment nodes provide moderately advanced deployment operations including:
- Blue-green deployment strategies
- Canary release implementations
- Rolling deployment management
- Health-aware deployments
- Configuration management

## Available Nodes
- `blue_green_deployment_node.go` - Implements blue-green deployment patterns
- `canary_release_node.go` - Manages canary release deployments
- `rolling_deployment_node.go` - Handles rolling deployment strategies
- `health_aware_deploy_node.go` - Deploys based on system health indicators
- `config_management_node.go` - Manages configuration deployments

## Usage
These nodes are designed for intermediate deployment workflows with advanced orchestration.

## Requirements
- Deployment orchestration platform access
- Health monitoring systems
- Configuration management tools
- Deployment target connectivity

## Security & Permissions
- Requires deployment orchestration permissions
- Configuration management privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform