# Dual Database Setup Guide

This expense tracker supports both **MongoDB** (for local development) and **Firestore** (for production) databases.

## Database Configuration

The application automatically chooses the database based on environment variables:

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_TYPE` | Database type: `mongodb` or `firestore` | Determined by `ENVIRONMENT` |
| `ENVIRONMENT` | Environment: `development` or `production` | - |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | MongoDB database name | `expense_tracker` |
| `GOOGLE_CLOUD_PROJECT` | Google Cloud project ID (for Firestore) | Required for Firestore |
| `GEMINI_API_KEY` | Google Gemini API key | Required |

### Database Selection Logic

1. If `DB_TYPE` is set → Use specified database
2. If `ENVIRONMENT=production` → Use Firestore
3. Default → Use MongoDB (local development)

## Local Development Setup (MongoDB)

### Option 1: Using Docker Compose (Recommended)

```bash
# 1. Set environment variables
export DB_TYPE=mongodb
export GEMINI_API_KEY=your_gemini_api_key_here

# 2. Start MongoDB and the application
docker-compose up
```

### Option 2: Local MongoDB Installation

```bash
# 1. Install MongoDB locally
brew install mongodb/brew/mongodb-community

# 2. Start MongoDB
brew services start mongodb-community

# 3. Set environment variables
export DB_TYPE=mongodb
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=expense_tracker
export GEMINI_API_KEY=your_gemini_api_key_here

# 4. Run the application
go run . api
```

## Production Setup (Firestore)

### Prerequisites

1. **Google Cloud Project**: Create a project in [Google Cloud Console](https://console.cloud.google.com/)
2. **Enable Firestore**: Enable Firestore in Native mode
3. **Service Account**: Create a service account with Firestore permissions
4. **API Keys**: Enable and get Gemini API key

### Environment Variables for Production

```bash
export DB_TYPE=firestore
export ENVIRONMENT=production
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export GEMINI_API_KEY=your_gemini_api_key_here
```

### Cloud Run Deployment

```bash
# Build and deploy
gcloud builds submit --tag gcr.io/your-project-id/expense-tracker
gcloud run deploy expense-tracker \
  --image gcr.io/your-project-id/expense-tracker \
  --platform managed \
  --set-env-vars DB_TYPE=firestore,ENVIRONMENT=production,GOOGLE_CLOUD_PROJECT=your-project-id
```

## Environment Configuration Examples

### Local Development (.env file)

```env
# Database Configuration
DB_TYPE=mongodb
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=expense_tracker
ENVIRONMENT=development

# API Keys
GEMINI_API_KEY=your_gemini_api_key_here
```

### Production (.env file)

```env
# Database Configuration
DB_TYPE=firestore
ENVIRONMENT=production
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# API Keys
GEMINI_API_KEY=your_gemini_api_key_here
```

## Running the Application

### Start API Server

```bash
go run . api
```

### Process Emails (with database auto-selection)

```bash
go run .
```

## Docker Compose Configuration

The `docker-compose.yml` includes:

- **MongoDB service**: Persistent storage with authentication
- **App service**: Configured for MongoDB by default
- **Networks**: Isolated network for services
- **Volumes**: Persistent data storage

### Services

- **MongoDB**: `mongodb://admin:password@mongodb:27017/expense_tracker`
- **App**: http://localhost:8080

## Database Schemas

### MongoDB Collections

- `transactions`: Store parsed transactions
- `unparsed_emails`: Store unparsed email data

### Firestore Collections

- `transactions`: Store parsed transactions  
- `unparsed_emails`: Store unparsed email data

## Testing Database Connectivity

```bash
# Test MongoDB connection
mongo mongodb://localhost:27017/expense_tracker

# Test Firestore connection (requires credentials)
export GOOGLE_CLOUD_PROJECT=your-project-id
go run . api
```

## Troubleshooting

### MongoDB Issues

1. **Connection refused**: Ensure MongoDB is running
2. **Authentication failed**: Check credentials in docker-compose.yml
3. **Database not found**: MongoDB creates database automatically

### Firestore Issues

1. **Permission denied**: Check service account permissions
2. **Project not found**: Verify `GOOGLE_CLOUD_PROJECT` variable
3. **API not enabled**: Enable Firestore API in Google Cloud Console

## Migration Between Databases

Currently, manual migration is required. Future versions will include migration tools.

### Export from MongoDB

```bash
mongoexport --db expense_tracker --collection transactions --out transactions.json
```

### Import to Firestore

Use the Firebase CLI or custom scripts to import JSON data to Firestore.

## Performance Considerations

- **MongoDB**: Better for high-frequency writes and complex queries
- **Firestore**: Better for real-time sync and global scale
- **Local Development**: MongoDB is faster for development/testing
- **Production**: Firestore offers better managed service capabilities 