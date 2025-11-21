# Citadel Agent - API Documentation

## Base URL
```
https://api.citadel-agent.com/v1
# Or locally:
http://localhost:5001/api/v1
```

## Authentication
All API requests require an Authorization header with a valid JWT token:

```
Authorization: Bearer <your-jwt-token>
```

Some operations also support API key authentication:
```
Authorization: ApiKey <your-api-key-prefix>
```

## Common Response Format

Success responses follow this format:
```json
{
  "data": { ... },
  "pagination": {
    "page": 0,
    "limit": 20,
    "total": 100
  }
}
```

Error responses follow this format:
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2023-01-01T00:00:00Z"
}
```

## API Endpoints

### Workflows

#### Create Workflow
- **POST** `/workflows`
- **Auth required**: Yes
- **Content-Type**: `application/json`
- **Request Body**: 
```json
{
  "name": "My Workflow",
  "description": "Description of my workflow",
  "nodes": [
    {
      "id": "node-1",
      "type": "http_request",
      "name": "HTTP Request Node",
      "config": {
        "url": "https://api.example.com/data",
        "method": "GET"
      },
      "position": {
        "x": 100,
        "y": 100
      }
    }
  ],
  "connections": [
    {
      "source_node_id": "node-1",
      "target_node_id": "node-2"
    }
  ],
  "tags": ["production", "marketing"]
}
```
- **Response**: Workflow object with 201 status

#### Get Workflow
- **GET** `/workflows/{id}`
- **Auth required**: Yes
- **Response**: Workflow object

#### Update Workflow
- **PUT** `/workflows/{id}`
- **Auth required**: Yes
- **Content-Type**: `application/json`
- **Request Body**: Same as Create Workflow
- **Response**: Updated Workflow object

#### Delete Workflow
- **DELETE** `/workflows/{id}`
- **Auth required**: Yes
- **Response**: 204 No Content

#### List Workflows
- **GET** `/workflows`
- **Auth required**: Yes
- **Query Parameters**:
  - `page` (integer, default: 0)
  - `limit` (integer, default: 20, max: 100)
  - `status` (string, optional: filter by status)
  - `tag` (string, optional: filter by tag)
- **Response**: 
```json
{
  "data": [...],
  "pagination": {
    "page": 0,
    "limit": 20,
    "total": 100
  }
}
```

#### Update Workflow Status
- **PUT** `/workflows/{id}/status`
- **Auth required**: Yes
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "status": "active"
}
```
- **Response**: Success message with new status

### Executions

#### Execute Workflow
- **POST** `/workflows/{id}/run`
- **Auth required**: Yes
- **Content-Type**: `application/json` (optional)
- **Request Body** (optional):
```json
{
  "param1": "value1",
  "param2": "value2"
}
```
- **Response**:
```json
{
  "execution_id": "exec-123",
  "message": "Workflow execution started"
}
```

#### Get Execution
- **GET** `/executions/{id}`
- **Auth required**: Yes
- **Response**: Execution object

#### Get Execution Logs
- **GET** `/executions/{id}/logs`
- **Auth required**: Yes
- **Query Parameters**:
  - `page` (integer, default: 0)
  - `limit` (integer, default: 20, max: 100)
- **Response**:
```json
{
  "data": [...],
  "pagination": {
    "page": 0,
    "limit": 20
  }
}
```

#### Retry Execution
- **POST** `/executions/{id}/retry`
- **Auth required**: Yes
- **Response**: Success message with status update

#### Cancel Execution
- **POST** `/executions/{id}/cancel`
- **Auth required**: Yes
- **Response**: Success message with status update

#### Get Workflow Executions
- **GET** `/workflows/{id}/executions`
- **Auth required**: Yes
- **Query Parameters**:
  - `page` (integer, default: 0)
  - `limit` (integer, default: 20, max: 100)
  - `status` (string, optional: filter by status)
- **Response**: List of execution objects

### Users

#### Get Current User
- **GET** `/users/me`
- **Auth required**: Yes
- **Response**: User object

#### Update User Profile
- **PUT** `/users/me`
- **Auth required**: Yes
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "profile": {
    "first_name": "John",
    "last_name": "Doe",
    "timezone": "UTC"
  },
  "preferences": {
    "theme": "dark"
  }
}
```
- **Response**: Updated User object

### API Keys

#### Create API Key
- **POST** `/api-keys`
- **Auth required**: Yes
- **Content-Type**: `application/json`
- **Request Body**:
```json
{
  "name": "My API Key",
  "permissions": ["workflows:read", "workflows:write"],
  "expires_in_days": 30
}
```
- **Response**:
```json
{
  "id": "key-123",
  "name": "My API Key",
  "key": "prefix_actual_key_here", // Only prefix is returned
  "permissions": ["workflows:read", "workflows:write"],
  "created_at": "2023-01-01T00:00:00Z",
  "expires_at": "2023-01-31T00:00:00Z"
}
```

#### List API Keys
- **GET** `/api-keys`
- **Auth required**: Yes
- **Response**: List of API key objects (without full key values)

#### Revoke API Key
- **DELETE** `/api-keys/{id}`
- **Auth required**: Yes
- **Response**: 204 No Content

### Plugins

#### List Plugins
- **GET** `/plugins`
- **Auth required**: Yes
- **Response**: List of plugin objects

#### Get Plugin
- **GET** `/plugins/{id}`
- **Auth required**: Yes
- **Response**: Plugin object

## Error Codes

| Code | Description |
|------|-------------|
| `WORKFLOW_NOT_FOUND` | The specified workflow does not exist |
| `EXECUTION_NOT_FOUND` | The specified execution does not exist |
| `INVALID_WORKFLOW_DEFINITION` | The workflow definition is invalid |
| `CIRCULAR_DEPENDENCY_DETECTED` | The workflow has circular dependencies |
| `UNAUTHORIZED` | The request lacks valid authentication |
| `FORBIDDEN` | The authenticated user lacks permission |
| `VALIDATION_ERROR` | Request data failed validation |
| `INTERNAL_ERROR` | An internal server error occurred |

## Rate Limiting

The API implements rate limiting:
- **Authenticated requests**: 1000 requests per hour per user
- **API key requests**: 5000 requests per hour per key
- **Unauthenticated requests**: 100 requests per hour per IP

Rate limit headers:
- `X-RateLimit-Limit`: The maximum number of requests per time window
- `X-RateLimit-Remaining`: The number of requests remaining in the current time window
- `X-RateLimit-Reset`: The time when the rate limit resets (Unix timestamp)

## Webhook Events

Citadel Agent can send webhook notifications for workflow events. To configure webhooks, use the UI or API.

### Supported Events
- `workflow.started`: When a workflow starts execution
- `workflow.completed`: When a workflow completes successfully
- `workflow.failed`: When a workflow fails
- `execution.status_changed`: When execution status changes

### Webhook Payload
```json
{
  "id": "event-123",
  "event": "workflow.completed",
  "timestamp": "2023-01-01T00:00:00Z",
  "data": {
    "workflow_id": "wf-123",
    "execution_id": "exec-456",
    "result": { ... }
  }
}
```

## Versioning

This documentation covers API version 1.0. The API follows semantic versioning principles, and breaking changes will result in a new major version number.

## SDKs and Libraries

Citadel Agent provides SDKs for multiple languages:
- JavaScript/TypeScript: `@citadel-agent/client`
- Python: `citadel-agent`
- Go: `github.com/citadel-agent/client-go`

## Support

For API support, contact us at api-support@citadel-agent.com or join our developer community at https://community.citadel-agent.com.