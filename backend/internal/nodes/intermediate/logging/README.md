# Intermediate Logging Nodes

This directory contains intermediate logging nodes for the Citadel Agent workflow engine.

## Overview
Intermediate logging nodes provide moderately advanced logging operations including:
- Structured logging with custom formats
- Log aggregation from multiple services
- Log filtering and routing
- Performance and audit logging
- Log rotation and archival

## Available Nodes
- `structured_logger_node.go` - Creates structured log entries with custom schemas
- `aggregation_service_node.go` - Aggregates logs from multiple service instances
- `filter_and_route_node.go` - Filters and routes logs to different destinations
- `audit_logger_node.go` - Creates audit trail logs
- `log_rotation_node.go` - Manages log rotation and archival

## Usage
These nodes are designed for intermediate logging workflows with structured data requirements.

## Requirements
- Log aggregation services
- Structured logging schemas
- Log destination configurations
- Archival storage systems

## Security & Permissions
- Requires log aggregation service permissions
- Audit logging privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform