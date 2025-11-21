# End-to-End Tests

This directory contains end-to-end tests for the Citadel Agent workflow engine.

## Overview
End-to-end tests validate complete user journeys and system workflows including:
- Complete user scenarios
- Full workflow execution
- API to UI round-trips
- Data flow through entire system
- Performance under real-world conditions

## Test Scenarios
- `user_workflow/` - Tests user login to workflow execution
- `api_full_cycle/` - Tests complete API interaction cycles
- `complex_workflow/` - Tests multi-step complex workflows
- `error_handling/` - Tests system-wide error handling
- `performance/` - Tests performance under load
- `upgrade/` - Tests upgrade scenarios

## Test Standards
- Simulate real user workflows
- Test complete data lifecycle
- Include error scenario testing
- Measure performance metrics
- Validate data integrity across components

## Running Tests
```bash
# Run all end-to-end tests
npm run test:e2e

# Run specific e2e test
npm run test:e2e -- --spec=user_workflow

# Run with headless mode
npm run test:e2e -- --headless
```

---

**Part of Citadel Agent v1** - Advanced Workflow Automation Platform