# Unit Tests

This directory contains unit tests for the Citadel Agent workflow engine.

## Overview
Unit tests provide the most granular level of testing for individual components including:
- Individual function tests
- Component isolation testing
- Edge case validation
- Performance benchmarking
- Regression testing

## Test Organization
- `engine/` - Tests for workflow engine components
- `nodes/` - Tests for individual node implementations
- `auth/` - Tests for authentication and authorization
- `services/` - Tests for service layer components
- `utils/` - Tests for utility functions
- `models/` - Tests for data models

## Test Standards
- Each function should have associated unit tests
- Test coverage should exceed 80%
- Tests should run in isolation
- Performance benchmarks should be included
- Negative test cases should be covered

## Running Tests
```bash
# Run all unit tests
go test ./test/unit/...

# Run with coverage report
go test -cover ./test/unit/...

# Run with verbose output
go test -v ./test/unit/...
```

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform