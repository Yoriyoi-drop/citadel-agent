# Citadel Agent - Visual Workflow Builder & Plugin System

## Features Completed

### 1. Visual Workflow Builder
A fully-featured visual workflow builder that allows users to create automated workflows through a drag-and-drop interface.

#### Components:
- **WorkflowCanvas**: Interactive canvas for building workflows
- **NodePalette**: Drag-and-drop node selection panel
- **PropertiesPanel**: Configuration panel for selected nodes
- **NodeComponents**: Individual node types (HTTP, Database, AI, Decision, etc.)

#### Node Types Supported:
- Start/End nodes
- HTTP Request nodes
- Database Query nodes
- Decision/Conditional nodes
- Delay/Timer nodes
- AI Agent nodes
- Notification nodes

#### Features:
- Drag-and-drop workflow creation
- Real-time node configuration
- Connection validation
- Workflow validation
- Node property editing
- Visual debugging support

### 2. Integration Plugins

#### GitHub Integration Node:
- Makes authenticated requests to GitHub API
- Supports all GitHub API endpoints
- Configurable HTTP methods (GET, POST, PUT, PATCH, DELETE)
- Query parameter and request body support
- Automatic error handling

#### Slack Integration Node:
- Sends messages to Slack channels via webhook or API
- Supports rich text formatting and attachments
- Threaded message support
- Multiple channel targeting
- Message customization

#### Email Integration Node:
- Sends emails via SMTP
- Supports HTML and plain text content
- CC, BCC, and attachment support
- TLS encryption support
- Multi-recipient functionality

## Architecture

### Frontend Components
```
src/
├── components/
│   ├── workflow-builder/
│   │   ├── WorkflowBuilder.jsx
│   │   ├── PropertiesPanel.jsx
│   │   ├── nodes/
│   │   │   ├── BaseNode.jsx
│   │   │   ├── StartNode.jsx
│   │   │   ├── EndNode.jsx
│   │   │   ├── HTTPNode.jsx
│   │   │   ├── DatabaseNode.jsx
│   │   │   ├── DecisionNode.jsx
│   │   │   ├── DelayNode.jsx
│   │   │   ├── AINode.jsx
│   │   │   └── NotificationNode.jsx
│   ├── Dashboard.jsx
├── hooks/
│   └── useWorkflow.js
├── services/
│   └── workflowService.js
```

### Backend Nodes
```
backend/
├── internal/
│   ├── nodes/
│   │   ├── integrations/
│   │   │   ├── github_node.go
│   │   │   ├── slack_node.go
│   │   │   ├── email_node.go
│   │   │   └── registry.go
```

## API Endpoints

The frontend communicates with the backend via these API endpoints:

- `GET /api/v1/workflows` - Get all workflows
- `GET /api/v1/workflows/:id` - Get specific workflow
- `POST /api/v1/workflows` - Create new workflow
- `PUT /api/v1/workflows/:id` - Update workflow
- `DELETE /api/v1/workflows/:id` - Delete workflow
- `POST /api/v1/workflows/:id/execute` - Execute workflow
- `GET /api/v1/workflows/:id/executions` - Get workflow executions
- `GET /api/v1/executions/:id` - Get execution details
- `POST /api/v1/nodes/test` - Test node configuration
- `POST /api/v1/workflows/validate` - Validate workflow

## Getting Started

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### Backend Development
```bash
cd backend
go run cmd/api/main.go
```

## Configuration

### Environment Variables
```bash
REACT_APP_API_URL=http://localhost:5001/api/v1
```

### Node Registration
All integration nodes are registered automatically through the `integrations.RegisterAllIntegrations()` function that gets called when the backend service starts.

## Security Considerations

- All API communications are secured with JWT authentication
- Node configurations are validated before execution
- Sandboxed execution for all code-based nodes
- Rate limiting for API integrations
- Input sanitization for all user-provided data

## Future Enhancements

Planned additions to the workflow builder and plugin system:

- Database integration nodes
- Loop and iteration nodes
- Error handling nodes
- Advanced scheduling nodes
- More third-party integrations (Discord, Telegram, etc.)
- Workflow versioning and history
- Real-time collaboration features
- Advanced monitoring and alerting
- Plugin marketplace

## Contributing

1. Create a new node in the `backend/internal/nodes/` directory
2. Implement the Node interface
3. Register your node in the appropriate registry file
4. Create a corresponding React component for the UI
5. Add the node to the palette in the workflow builder component
6. Update documentation in this README

## Testing

The visual workflow builder includes comprehensive testing capabilities:
- Node configuration validation
- Connection validation
- Workflow validation
- Error handling tests
- Integration tests for all plugin nodes

## Performance

The workflow builder is optimized for:
- Smooth drag-and-drop interactions
- Efficient rendering of complex workflows
- Real-time validation
- Low-latency API communication
- Optimized data structures for large workflow graphs