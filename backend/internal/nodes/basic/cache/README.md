# Basic Cache Nodes

This directory contains basic cache nodes for the Citadel Agent workflow engine.

## Overview
Basic cache nodes provide fundamental caching operations including:
- Simple key-value storage
- Basic cache read/write operations
- Time-based expiration
- Simple cache validation
- Basic cache statistics

## Available Nodes
- `simple_cache_node.go` - Basic key-value cache operations
- `ttl_cache_node.go` - Cache with time-to-live expiration
- `cache_validator_node.go` - Validates cache entries
- `cache_stats_node.go` - Provides basic cache statistics

## Usage
These nodes are designed for basic caching workflows with simple requirements.

## Requirements
- Cache storage backend (Redis, Memcached, etc.)
- Network connectivity to cache service
- Basic cache credentials

## Security & Permissions
- Requires cache service access permissions
- Basic network access privileges

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform