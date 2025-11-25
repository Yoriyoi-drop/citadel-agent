# Citadel Agent Workflow Examples

This directory contains example workflows that demonstrate various capabilities of Citadel Agent. You can import these workflows directly into the application.

## Available Examples

### 1. HTTP Data Processing (`http-processing-workflow.json`)
Demonstrates how to fetch data from an external API, transform the JSON response, and save the results to a database.
- **Nodes used**: HTTP Request, JSON Transform, Postgres Query
- **Use case**: ETL pipelines, data synchronization

### 2. Scheduled Task (`scheduled-task.json`)
Shows how to run a recurring task using a Cron trigger.
- **Nodes used**: Cron Trigger, HTTP Request, Logger
- **Use case**: Health checks, periodic reports, cleanup jobs

### 3. API Integration (`api-integration.json`)
A complex workflow that chains multiple API requests together, passing data from one to the next.
- **Nodes used**: Webhook Trigger, HTTP Request
- **Use case**: Order processing, cross-service automation

## How to Import

1. Open the Citadel Agent UI
2. Go to the "Workflows" page
3. Click "Import Workflow"
4. Select one of the `.json` files from this directory
5. Configure any necessary credentials (e.g., database connection, API keys)
6. Click "Activate" to start the workflow
