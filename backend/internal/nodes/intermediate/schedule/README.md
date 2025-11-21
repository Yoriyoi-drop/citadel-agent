# Intermediate Schedule Nodes

This directory contains intermediate schedule nodes for the Citadel Agent workflow engine.

## Overview
Intermediate schedule nodes provide moderately advanced scheduling operations including:
- Calendar-based scheduling
- Resource-constrained scheduling
- Priority-based task ordering
- Schedule conflict resolution
- Time zone aware scheduling

## Available Nodes
- `calendar_scheduler_node.go` - Schedules tasks based on calendar rules
- `resource_aware_scheduler_node.go` - Considers resource availability in scheduling
- `priority_scheduler_node.go` - Orders tasks based on priority levels
- `conflict_resolver_node.go` - Resolves conflicting schedule requests
- `timezone_aware_scheduler_node.go` - Handles scheduling across time zones

## Usage
These nodes are designed for intermediate scheduling workflows with enhanced functionality.

## Requirements
- Calendar systems access
- Resource availability information
- Priority management systems
- Time zone conversion utilities

## Security & Permissions
- Requires scheduling system permissions
- Resource management privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform