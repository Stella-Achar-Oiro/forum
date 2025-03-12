#!/bin/bash
# Environment variables for forum application

# Base URL for the application
export BASE_URL="http://localhost:8080"

# Google OAuth credentials
export GOOGLE_CLIENT_ID="YOUR_GOOGLE_CLIENT_ID"
export GOOGLE_CLIENT_SECRET="YOUR_GOOGLE_CLIENT_SECRET"

# GitHub OAuth credentials
export GITHUB_CLIENT_ID="YOUR_GITHUB_CLIENT_ID"
export GITHUB_CLIENT_SECRET="YOUR_GITHUB_CLIENT_SECRET"

# Secret key for session management
export SESSION_SECRET="a-very-long-and-secure-random-string-for-session-encryption"

# Database connection string
export DATABASE_URL="your-database-connection-string"

echo "Environment variables set successfully!" 