# OAuth Setup Guide

This guide will help you set up Google and GitHub OAuth authentication for your Forum application.

## Step 1: Create OAuth Applications

### Google OAuth Setup

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or use an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Set the application type to "Web application" 
6. Add authorized redirect URIs:
   - Development: `http://localhost:8080/api/public/auth/google/callback`
   - Production: `https://your-domain.com/api/public/auth/google/callback`
7. Note the Client ID and Client Secret

### GitHub OAuth Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the application details:
   - Application name: "Forum App" (or your preferred name)
   - Homepage URL: `http://localhost:8080` (development) or your production URL
   - Application description: Optional
   - Authorization callback URL: 
     - Development: `http://localhost:8080/api/public/auth/github/callback`
     - Production: `https://your-domain.com/api/public/auth/github/callback`
4. Register the application
5. Note the Client ID and Client Secret
6. Optionally, generate a new client secret if needed

## Step 2: Configure Environment Variables

1. Open the `env.sh` file and update it with your OAuth credentials:

```bash
# Google OAuth credentials
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"

# GitHub OAuth credentials
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"

# Base URL for the application
export BASE_URL="http://localhost:8080"  # For development
# export BASE_URL="https://your-domain.com"  # For production
```

## Step 3: Run the Application

1. Make the scripts executable:
   ```bash
   chmod +x env.sh start.sh
   ```

2. Start the application:
   ```bash
   ./start.sh
   ```

3. Open your browser and navigate to `http://localhost:8080`

4. Click the "Login with Google" or "Login with GitHub" buttons to test the OAuth flow

## Troubleshooting

### Common Issues

1. **Redirect URI Mismatch**: Ensure the redirect URIs in your OAuth provider settings match exactly with what's configured in the application.

2. **Environment Variables Not Set**: Check that your environment variables are correctly set by running:
   ```bash
   echo $GOOGLE_CLIENT_ID
   echo $GITHUB_CLIENT_ID
   ```

3. **CORS Issues**: If you're experiencing CORS issues, ensure the BASE_URL is correctly set and matches your application's URL.

4. **OAuth Scopes**: If you're not getting all user information, ensure the OAuth scopes are correctly set in the config.go file.

### Debugging

Enable detailed logging by setting:
```bash
export DEBUG=true
```

This will output more detailed logs about the OAuth process, which can help diagnose issues. 