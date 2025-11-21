# Troubleshooting OAuth Login Issues

## Firefox GitHub Login "Not Found" Error

If you're experiencing a "not found" error when trying to log in with GitHub using Firefox, here are some steps to resolve the issue:

### 1. Verify Server is Running
Make sure the Citadel Agent API server is running on port 5001:
```bash
./run_server.sh
```

### 2. Check Callback URLs
Ensure your GitHub OAuth application is configured with the correct callback URL:
- **Development**: `http://localhost:5001/api/v1/auth/github/callback`
- **Production**: `https://yourdomain.com/api/v1/auth/github/callback`

### 3. Environment Configuration
Verify your `.env` file contains the correct configuration:
```env
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_CALLBACK_URL=http://localhost:5001/api/v1/auth/github/callback
SERVER_PORT=5001
```

### 4. Cross-Origin Resource Sharing (CORS)
The issue might be related to CORS policies. Ensure your server allows requests from your origin.

### 5. Firefox-Specific Issues
- Check if Firefox has strict tracking protection enabled which might block the OAuth flow
- Try temporarily disabling browser extensions that might interfere with OAuth
- Clear browser cache and cookies for the site

### 6. Testing OAuth Endpoints
You can test if the OAuth endpoints are working properly:
```bash
# Test the initation endpoint
curl -v http://localhost:5001/api/v1/auth/github

# Test the callback endpoint (though this will fail without a proper code parameter)
curl -v "http://localhost:5001/api/v1/auth/github/callback?code=sample_code"
```

### 7. Check Server Logs
Monitor your server logs for any error messages when attempting OAuth login to identify specific issues.

### 8. Alternative Browsers
If the issue persists in Firefox, try using Chrome or another browser to determine if it's a browser-specific issue.