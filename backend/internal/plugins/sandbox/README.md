# Plugin Sandbox

This directory contains the sandbox implementation for Citadel Agent plugins.

## Overview
The plugin sandbox provides a secure execution environment for third-party plugins including:
- Isolated execution environments
- Security boundary enforcement
- Resource limitation and monitoring
- API access control
- File system isolation

## Components
- `javascript_sandbox.go` - Secure JavaScript execution environment
- `python_sandbox.go` - Secure Python execution environment
- `go_plugin_sandbox.go` - Secure Go plugin execution environment
- `security_enforcer.go` - Enforces security policies within sandboxes
- `resource_limiter.go` - Limits and monitors resource usage
- `api_controller.go` - Controls API access from sandboxed plugins

## Security Features
- Process isolation
- Network access control
- File system restrictions
- Resource quotas (CPU, memory, disk)
- Time and operation limits

## Usage
The sandbox ensures that third-party plugins execute safely without compromising the system.

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform