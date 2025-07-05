# Expense Tracker

An automated expense tracker that reads bank transaction emails from Gmail and stores them in Google Firestore. Fully containerized with Docker support and seamless OAuth token management.

## Prerequisites

1. **Google Cloud Project**: Create a GCP project
2. **Gmail API Access**: Enable Gmail API in your GCP project
3. **Firestore Database**: Enable Firestore in your GCP project
4. **Service Account**: Create a service account with Firestore access
5. **OAuth Credentials**: Create OAuth 2.0 credentials for Gmail access
6. **Docker**: Install Docker and Docker Compose

## Setup Instructions

### 1. Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the following APIs:
   - Gmail API
   - Firestore API
4. Create a service account:
   - Go to IAM & Admin > Service Accounts
   - Create a new service account with Firestore Admin role
   - Download the JSON key file
5. Create OAuth 2.0 credentials:
   - Go to APIs & Services > Credentials
   - Create OAuth 2.0 Client ID (Desktop application)
   - Add `http://localhost:8080/oauth2callback` to authorized redirect URIs
   - Download the `client_secret.json` file

### 2. Local Setup

1. Clone this repository
2. Create a `credentials` directory:
   ```bash
   mkdir credentials
   ```
3. Copy your files to the credentials directory:
   ```bash
   cp /path/to/your/service-account-key.json credentials/service-account.json
   cp /path/to/your/client_secret.json credentials/client_secret.json
   ```
4. Create a `.env` file:
   ```bash
   echo "GOOGLE_CLOUD_PROJECT=your-project-id" > .env
   ```

### 3. Running with Docker (Recommended)

#### First Time Setup - OAuth Token Refresh

For the first run or when your OAuth token expires, use the refresh script:

```bash
# Make the script executable (if not already)
chmod +x refresh-token.sh

# Run the OAuth token refresh
./refresh-token.sh
```

**What happens:**
1. Script removes any expired tokens
2. Starts Docker container with OAuth flow
3. Container displays an OAuth URL
4. You open the URL in your browser and complete authentication
5. Container saves the new token and processes emails
6. Fresh token is ready for future runs

#### Regular Usage

Once you have a valid token, run normally:

```bash
# Regular run (uses existing token)
docker-compose up

# Run in background
docker-compose up -d

# Rebuild and run
docker-compose up --build
```

#### Manual Docker Commands (Alternative)

```bash
# Build the image
docker build -f Dockerfile.flexible -t expense-tracker .

# Run with environment variables
docker run --rm \
  -e GOOGLE_CLOUD_PROJECT=your-project-id \
  -e GOOGLE_APPLICATION_CREDENTIALS=/root/credentials/service-account.json \
  -v $(pwd)/credentials:/root/credentials \
  -p 8080:8080 \
  expense-tracker
```

### 4. Running Locally (Go Development)

```bash
# Set environment variables
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/credentials/service-account.json

# Build and run
go run .
```

### 5. Cron Job Setup (Automated Scheduling)

For automated daily runs, you can set up a cron job:

#### Step 1: Prepare Cron Script

```bash
# Make the cron script executable
chmod +x run-cron.sh

# Test the cron script
./run-cron.sh
```

#### Step 2: Set Up Cron Job

```bash
# Edit crontab
crontab -e

# Add a daily run at 9 AM (adjust path to your project)
0 9 * * * /path/to/expense-tracker/run-cron.sh

# Or run every 6 hours
0 */6 * * * /path/to/expense-tracker/run-cron.sh
```

#### Step 3: Monitor Cron Jobs

```bash
# View cron logs
tail -f cron-expense-tracker.log

# Check cron job status
crontab -l
```

#### Cron Job Features

- **✅ Automatic Token Refresh**: Handles OAuth token expiration automatically
- **✅ Non-Interactive**: Runs without user interaction
- **✅ Comprehensive Logging**: Logs all activities with timestamps
- **✅ Error Handling**: Graceful failure with proper exit codes
- **✅ Token Age Warning**: Warns when tokens are getting old
- **✅ Pre-flight Checks**: Validates environment before running

#### Token Management for Cron Jobs

The system handles OAuth tokens intelligently:

1. **Fresh Token**: Uses existing valid token
2. **Expired Token**: Automatically refreshes using refresh token
3. **Invalid Refresh Token**: Logs error and exits (requires manual refresh)
4. **No Token**: Logs error and exits (requires initial setup)

**Important**: OAuth refresh tokens typically last 7 days. For long-term automation, consider:
- Running the refresh script weekly
- Setting up monitoring for token expiration
- Using Google Cloud's service account authentication for production

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GOOGLE_CLOUD_PROJECT` | Your GCP project ID | Yes |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account JSON | Yes |
| `CRON_JOB` | Set to "true" for cron job detection | Auto-detected |
| `DOCKER_CONTAINER` | Set to "true" for Docker detection | Auto-detected |

### File Structure

```
expense-tracker/
├── credentials/
│   ├── service-account.json    # Your service account key
│   ├── client_secret.json      # OAuth client secret
│   └── token.json             # OAuth token (auto-generated)
├── .env                       # Environment variables
├── docker-compose.yml         # Docker Compose configuration
├── refresh-token.sh          # OAuth token refresh script
├── run-cron.sh              # Cron job runner script
├── cron-expense-tracker.log  # Cron job log file (auto-generated)
└── Dockerfile.flexible       # Multi-stage Docker build
```

## OAuth Token Management

### When to Refresh Tokens

- **First time setup**: No token exists
- **Token expired**: Get "invalid_grant" or "Token has been expired" errors
- **Revoked access**: Token was manually revoked in Google Account settings
- **Cron job failures**: Check logs for token-related errors

### Token Refresh Methods

#### Method 1: Using Refresh Script (Recommended)
```bash
./refresh-token.sh
```

#### Method 2: Manual Process
```bash
# Remove expired token
rm credentials/token.json

# Start container for OAuth flow
docker-compose up --build

# Follow the OAuth URL in the container logs
```

### How OAuth Works in Docker

1. **Port 8080** is exposed for OAuth callback
2. **Credentials volume** is mounted as read-write
3. **Browser authentication** completes the OAuth flow
4. **Token persistence** saves to host filesystem
5. **Future runs** use the saved token automatically

### Automatic Token Refresh

The enhanced OAuth system automatically:
- ✅ Detects expired tokens
- ✅ Attempts refresh using refresh token
- ✅ Falls back to interactive auth if refresh fails
- ✅ Detects non-interactive environments (cron jobs)
- ✅ Provides clear error messages for troubleshooting

## Security Notes

- ✅ All credential files are in `credentials/` directory
- ✅ `.env` file is ignored by Git
- ✅ OAuth tokens are securely stored and auto-refreshed
- ✅ Service account has minimal required permissions
- ⚠️ Never commit credential files to Git
- ⚠️ Use environment variables for sensitive configuration
- 🔒 In production, use Google Cloud's built-in authentication

## Troubleshooting

### Common Issues

1. **OAuth Token Expired**:
   ```
   Error: oauth2: "invalid_grant" "Token has been expired or revoked"
   ```
   **Solution**: Run `./refresh-token.sh`

2. **Cron Job Token Issues**:
   ```
   Error: No valid token found and running in non-interactive environment
   ```
   **Solution**: Run `./refresh-token.sh` to get a fresh token, then retry cron job

3. **Authentication Error**:
   ```
   Error: rpc error: code = PermissionDenied
   ```
   **Solution**: Ensure service account has Firestore Admin role

4. **Gmail Access Error**:
   ```
   Error: Gmail API access denied
   ```
   **Solution**: Verify Gmail API is enabled and OAuth consent screen is configured

5. **Port 8080 In Use**:
   ```
   Error: bind: address already in use
   ```
   **Solution**: Stop other services using port 8080 or change port in docker-compose.yml

### Viewing Logs

```bash
# View container logs
docker-compose logs -f expense-tracker

# View cron job logs
tail -f cron-expense-tracker.log

# View logs for specific container
docker logs expense-tracker-expense-tracker-1
```

### Debug Mode

For detailed debugging, modify the container to run interactively:

```bash
# Run container with shell access
docker run -it --rm \
  -e GOOGLE_CLOUD_PROJECT=your-project-id \
  -e GOOGLE_APPLICATION_CREDENTIALS=/root/credentials/service-account.json \
  -v $(pwd)/credentials:/root/credentials \
  -p 8080:8080 \
  expense-tracker sh
```

## Production Deployment

### Google Cloud Run

```bash
# Build and push to Google Container Registry
docker build -f Dockerfile.flexible -t gcr.io/your-project-id/expense-tracker .
docker push gcr.io/your-project-id/expense-tracker

# Deploy to Cloud Run
gcloud run deploy expense-tracker \
  --image gcr.io/your-project-id/expense-tracker \
  --platform managed \
  --region us-central1 \
  --set-env-vars GOOGLE_CLOUD_PROJECT=your-project-id
```

### Kubernetes

Use Google Cloud's Workload Identity for secure authentication instead of mounting credential files.

### Cloud Scheduler + Cloud Run

For production automation, use Google Cloud Scheduler to trigger Cloud Run jobs:

```bash
# Create a Cloud Scheduler job
gcloud scheduler jobs create http expense-tracker-daily \
  --schedule="0 9 * * *" \
  --uri="https://your-cloud-run-url" \
  --http-method=POST \
  --oidc-service-account-email=your-service-account@your-project.iam.gserviceaccount.com
```

## Features

- 🔄 **Automated Email Processing**: Fetches HDFC Bank transaction emails
- 📊 **Transaction Parsing**: Extracts amount, vendor, date, and card details
- 🔥 **Firestore Integration**: Stores transactions with unique IDs
- 🐳 **Docker Native**: Fully containerized with Docker Compose
- 🔐 **Secure OAuth**: Browser-based authentication with token persistence
- 📝 **Comprehensive Logging**: Detailed transaction processing logs
- 🛠️ **Easy Setup**: One-script OAuth token refresh
- ⏰ **Cron Job Ready**: Automated scheduling with intelligent token management
- 🔄 **Auto Token Refresh**: Handles token expiration automatically
- 📋 **Production Ready**: Cloud Run and Kubernetes deployment options

## Supported Transaction Types

- **HDFC Bank Credit Card**: ZOMATO, SWIGGY, and other merchants
- **Extensible Parser**: Easy to add new bank formats and transaction types 