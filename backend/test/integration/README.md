# Integration Tests

This directory contains integration tests for the Citadel Agent workflow engine.

## Overview
Integration tests validate the interaction between different components and systems including:
- Multi-component interaction
- API integration testing
- Database integration
- Third-party service integration
- Security boundary testing

## Test Categories
- `api_integration/` - Tests for API endpoint integration
- `database_integration/` - Tests for database operations
- `node_integration/` - Tests for node-to-node communication
- `auth_integration/` - Tests for authentication flows
- `external_service_integration/` - Tests for third-party service integration

## Test Standards
- Test multi-component interactions
- Validate data flow between components
- Test error handling between components
- Ensure security boundaries are maintained
- Test performance under integrated loads

## Running Tests
```bash
# Run all integration tests
go test ./test/integration/...

# Run with coverage report
go test -cover ./test/integration/...

# Run specific integration test
go test ./test/integration/api_integration/...
```

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform