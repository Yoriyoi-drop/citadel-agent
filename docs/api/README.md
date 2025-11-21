# API Documentation

## Base URL
```
http://localhost:5001/api/v1
```

## Authentication
Most endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Authentication

#### POST /auth/login
Login with email and password to get JWT token.

**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "token": "string",
  "user": {
    "id": "string",
    "email": "string",
    "username": "string"
  }
}
```

#### POST /auth/register
Register a new user account.

**Request Body:**
```json
{
  "email": "string",
  "password": "string",
  "username": "string",
  "first_name": "string",
  "last_name": "string"
}
```

**Response:**
```json
{
  "token": "string",
  "user": {
    "id": "string",
    "email": "string",
    "username": "string"
  }
}
```

#### GET /auth/github
Initiate GitHub OAuth flow.

**Response:**
Redirects to GitHub authorization page.

#### GET /auth/github/callback
Handle GitHub OAuth callback.

**Query Parameters:**
- code: Authorization code from GitHub

**Response:**
```json
{
  "message": "GitHub authentication successful",
  "user": {
    "id": "string",
    "name": "string",
    "email": "string",
    "avatar_url": "string"
  },
  "token": "string"
}
```

#### GET /auth/google
Initiate Google OAuth flow.

**Response:**
Redirects to Google authorization page.

#### GET /auth/google/callback
Handle Google OAuth callback.

**Query Parameters:**
- code: Authorization code from Google

**Response:**
```json
{
  "message": "Google authentication successful",
  "user": {
    "id": "string",
    "name": "string",
    "email": "string",
    "avatar_url": "string"
  },
  "token": "string"
}
```

### Workflows

#### GET /workflows
Get list of workflows for the authenticated user.

**Response:**
```json
[
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "nodes": [
      {
        "id": "string",
        "type": "string",
        "config": {},
        "input": {},
        "output": {},
        "status": "string"
      }
    ],
    "status": "string",
    "created_at": "int64",
    "updated_at": "int64"
  }
]
```

#### POST /workflows
Create a new workflow.

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "config": {},
      "input": {},
      "output": {},
      "status": "string"
    }
  ]
}
```

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "config": {},
      "input": {},
      "output": {},
      "status": "string"
    }
  ],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### GET /workflows/{id}
Get a specific workflow by ID.

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "config": {},
      "input": {},
      "output": {},
      "status": "string"
    }
  ],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### PUT /workflows/{id}
Update an existing workflow.

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "config": {},
      "input": {},
      "output": {},
      "status": "string"
    }
  ]
}
```

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "config": {},
      "input": {},
      "output": {},
      "status": "string"
    }
  ],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### DELETE /workflows/{id}
Delete a workflow by ID.

**Response:**
204 No Content

#### POST /workflows/{id}/run
Execute a workflow.

**Response:**
```json
{
  "message": "Workflow executed successfully",
  "workflow_id": "string",
  "execution_id": "string",
  "status": "string"
}
```

### Nodes

#### GET /nodes
Get list of available nodes.

**Response:**
```json
{
  "nodes": [
    {
      "id": "string",
      "type": "string",
      "name": "string",
      "description": "string",
      "inputs": {},
      "outputs": {},
      "config": {},
      "icon": "string",
      "category": "string"
    }
  ]
}
```

#### GET /nodes/types
Get list of available node types.

**Response:**
```json
{
  "node_types": [
    "http_request",
    "delay",
    "function",
    "trigger",
    "data_process",
    "go_code",
    "javascript_code",
    "python_code",
    "java_code",
    "ruby_code",
    "php_code",
    "rust_code",
    "csharp_code",
    "shell_script",
    "ai_agent",
    "multi_runtime"
  ]
}
```

### AI Agents

#### POST /ai-agents
Create a new AI agent.

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "config": {},
  "tools": []
}
```

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "config": {},
  "tools": [],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### GET /ai-agents
Get list of AI agents.

**Response:**
```json
[
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "config": {},
    "tools": [],
    "status": "string",
    "created_at": "int64",
    "updated_at": "int64"
  }
]
```

#### GET /ai-agents/{id}
Get a specific AI agent by ID.

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "config": {},
  "tools": [],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### PUT /ai-agents/{id}
Update an existing AI agent.

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "config": {},
  "tools": []
}
```

**Response:**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "config": {},
  "tools": [],
  "status": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### DELETE /ai-agents/{id}
Delete an AI agent by ID.

**Response:**
204 No Content

#### POST /ai-agents/{id}/execute
Execute an AI agent.

**Response:**
```json
{
  "message": "AI agent executed successfully",
  "agent_id": "string",
  "execution_id": "string",
  "result": {}
}
```

#### POST /ai-agents/{id}/tools
Add a tool to an AI agent.

**Request Body:**
```json
{
  "tool_id": "string",
  "config": {}
}
```

**Response:**
```json
{
  "message": "Tool added to AI agent successfully",
  "agent_id": "string",
  "tool_id": "string"
}
```

### Runtime Execution

#### POST /runtime/execute
Execute code in a specific runtime environment.

**Request Body:**
```json
{
  "runtime_type": "string",
  "code": "string",
  "input": {},
  "timeout": 30
}
```

**Response:**
```json
{
  "result": {
    "output": "string",
    "error": "string",
    "success": true,
    "execution_time": 100,
    "resources": {}
  }
}
```

### User Management

#### GET /users/me
Get the authenticated user's profile.

**Response:**
```json
{
  "id": "string",
  "email": "string",
  "username": "string",
  "first_name": "string",
  "last_name": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

#### PUT /users/me
Update the authenticated user's profile.

**Request Body:**
```json
{
  "username": "string",
  "first_name": "string",
  "last_name": "string",
  "email": "string"
}
```

**Response:**
```json
{
  "id": "string",
  "email": "string",
  "username": "string",
  "first_name": "string",
  "last_name": "string",
  "created_at": "int64",
  "updated_at": "int64"
}
```

### Health Check

#### GET /health
Check the health status of the API.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "string",
  "version": "string"
}
```