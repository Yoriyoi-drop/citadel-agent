# Intermediate Messaging Nodes

This directory contains intermediate messaging nodes for the Citadel Agent workflow engine.

## Overview
Intermediate messaging nodes provide moderately advanced messaging operations including:
- Message queue management
- Multi-protocol message handling
- Message routing with complex rules
- Message transformation and enrichment
- Guaranteed delivery mechanisms

## Available Nodes
- `mq_manager_node.go` - Manages message queues (RabbitMQ, Kafka, etc.)
- `multi_protocol_handler_node.go` - Handles multiple messaging protocols
- `complex_router_node.go` - Routes messages based on complex rules
- `message_enricher_node.go` - Enriches messages with additional data
- `guaranteed_delivery_node.go` - Ensures reliable message delivery

## Usage
These nodes are designed for intermediate messaging workflows with multiple protocol support.

## Requirements
- Message broker connectivity
- Multi-protocol messaging libraries
- Message routing rules
- Reliable delivery mechanisms

## Security & Permissions
- Requires message broker access permissions
- Multi-protocol messaging privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform