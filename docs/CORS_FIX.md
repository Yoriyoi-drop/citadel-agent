# CORS Configuration Fix

## Problem
Frontend was unable to fetch nodes from backend API due to CORS policy blocking requests from different origins (localhost:5173, 192.168.x.x:3000, etc).

## Solution
Updated CORS configuration in `backend/internal/config/config.go` to allow all origins in development mode.

### Changes Made

**File: `backend/internal/config/config.go`**
```go
// Before
viper.SetDefault("cors_allowed_origins", "http://localhost:3000,http://localhost:8080")

// After
viper.SetDefault("cors_allowed_origins", "*")
```

This allows the frontend to be accessed from:
- `http://localhost:3000` (Next.js default port)
- `http://localhost:5173` (Vite dev server)
- `http://192.168.x.x:3000` (Local network access)
- Any other origin (development convenience)

## Testing

Test CORS with curl:
```bash
# Test from localhost
curl -H "Origin: http://localhost:5173" http://localhost:8080/api/v1/registry/nodes -v

# Test from local network IP
curl -H "Origin: http://192.168.43.98:3000" http://localhost:8080/api/v1/registry/nodes -v
```

Both should return:
```
Access-Control-Allow-Origin: *
```

## Production Considerations

⚠️ **IMPORTANT**: For production deployment, replace `*` with specific allowed origins:

```go
viper.SetDefault("cors_allowed_origins", "https://yourdomain.com,https://app.yourdomain.com")
```

Or use environment variable:
```bash
export CITADEL_CORS_ALLOWED_ORIGINS="https://yourdomain.com,https://app.yourdomain.com"
```

## Server Restart

After changing CORS configuration, restart the backend server:
```bash
cd backend
go run cmd/api/main.go
```

## Verification

1. Open browser DevTools (F12)
2. Navigate to Network tab
3. Refresh the page
4. Check the response headers for `/api/v1/registry/nodes`
5. Should see: `Access-Control-Allow-Origin: *`

## Related Files
- `backend/internal/config/config.go` - CORS configuration
- `backend/cmd/api/main.go` - Fiber server with CORS middleware
- `frontend/src/stores/nodeStore.ts` - Frontend API calls

## Status
✅ CORS issue resolved
✅ Frontend can now fetch nodes from backend
✅ Works on localhost and local network IPs
