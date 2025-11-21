# OAuth Configuration for Citadel Agent

## Callback URLs

The Authorization callback URLs for GitHub and Google OAuth in Citadel Agent are typically configured as follows:

### GitHub OAuth
- **Authorization callback URL**: `http://localhost:5001/api/v1/auth/github/callback` (for development)
- **Authorization callback URL**: `https://yourdomain.com/api/v1/auth/github/callback` (for production)

### Google OAuth
- **Authorization callback URL**: `http://localhost:5001/api/v1/auth/google/callback` (for development)
- **Authorization callback URL**: `https://yourdomain.com/api/v1/auth/google/callback` (for production)

## Configuration in .env file

These callbacks would typically be used with the following environment variables:

```
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_CALLBACK_URL=http://localhost:5001/api/v1/auth/github/callback

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:5001/api/v1/auth/google/callback
```

## Implementation Details

The OAuth flow would typically handle the following endpoints:

### GitHub OAuth Endpoints:
- **GET** `/api/v1/auth/github` - Initiates GitHub OAuth flow
- **GET** `/api/v1/auth/github/callback` - Handles GitHub OAuth callback

### Google OAuth Endpoints:
- **GET** `/api/v1/auth/google` - Initiates Google OAuth flow
- **GET** `/api/v1/auth/google/callback` - Handles Google OAuth callback

## Setup Instructions

### For GitHub:
1. Go to GitHub Developer Settings
2. Create a new OAuth application
3. Set Homepage URL to: `http://localhost:5001` (or your domain)
4. Set Authorization callback URL to: `http://localhost:5001/api/v1/auth/github/callback`

### For Google:
1. Go to Google Cloud Console
2. Create a new OAuth 2.0 client ID
3. Set Authorized JavaScript origins to: `http://localhost:5001`
4. Set Authorized redirect URIs to: `http://localhost:5001/api/v1/auth/google/callback`