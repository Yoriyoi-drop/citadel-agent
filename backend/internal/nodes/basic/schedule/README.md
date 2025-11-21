# Basic Schedule Nodes

This directory contains basic schedule nodes for the Citadel Agent workflow engine.

## Overview
Basic schedule nodes provide fundamental scheduling operations including:
- Simple one-time scheduling
- Basic recurring tasks
- Time-based triggers
- Simple cron-like functionality
- Basic schedule management

## Available Nodes
- `one_time_scheduler_node.go` - Schedules one-time tasks
- `recurring_task_node.go` - Manages recurring tasks
- `time_trigger_node.go` - Triggers based on time conditions
- `simple_cron_node.go` - Implements basic cron-like scheduling

## Usage
These nodes are designed for basic scheduling workflows with simple requirements.

## Requirements
- System time access
- Scheduling service access
- Valid time specifications

## Security & Permissions
- Requires system scheduling permissions
- Basic time operation privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform