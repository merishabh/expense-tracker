# 💰 Expense Tracker

An intelligent expense tracker that automatically reads bank transaction emails from Gmail and provides powerful analytics with beautiful visualizations. Features dual database support (MongoDB for local development, Firestore for production) and a modern web dashboard.

## ✨ Features

- 🔄 **Automated Email Processing**: Fetches and processes bank transaction emails from Gmail
- 💾 **Dual Database Support**: MongoDB for local development, Firestore for production
- 📊 **Beautiful Analytics Dashboard**: Interactive charts with enhanced UI/UX
- 🎯 **Smart Predictions**: AI-powered spending predictions and trend analysis
- 🔍 **MongoDB Web UI**: Integrated Mongo Express for easy data browsing
- 📱 **Responsive Design**: Modern, mobile-friendly interface
- 🐳 **Docker Native**: Fully containerized with Docker Compose
- 🔐 **Secure OAuth**: Browser-based Gmail authentication
- 📈 **Advanced Charts**: Monthly trends, top vendors, spending patterns
- 🔔 **Smart Insights**: Spending warnings and budget recommendations
- 🔄 **Auto Token Refresh**: Handles OAuth token expiration automatically

## 🚀 Quick Start

### Prerequisites

1. **Docker & Docker Compose**: Install Docker Desktop
2. **Google Cloud Project**: Create a GCP project
3. **Gmail API Access**: Enable Gmail API in your GCP project
4. **OAuth Credentials**: Create OAuth 2.0 credentials for Gmail access
5. **Firestore (Optional)**: Enable Firestore for production deployment

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd expense-tracker
mkdir credentials
```

### 2. Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the Gmail API
4. Create OAuth 2.0 credentials:
   - Go to APIs & Services > Credentials
   - Create OAuth 2.0 Client ID (Desktop application)
   - Add `http://localhost:8080/oauth2callback` to authorized redirect URIs
   - Download the `client_secret.json` file
5. Copy your OAuth credentials:
   ```bash
   cp /path/to/your/client_secret.json credentials/client_secret.json
   ```

### 3. Start the Application

```bash
# Start MongoDB and the application
docker-compose up -d

# View logs
docker-compose logs -f expense-tracker
```

### 4. First-Time OAuth Setup

When you first run the application, it will guide you through OAuth authentication:

1. Check the application logs for an OAuth URL
2. Open the URL in your browser
3. Complete Gmail authentication
4. The application will automatically save your token

### 5. Access the Dashboard

- **Main Dashboard**: http://localhost:8080
- **MongoDB UI**: http://localhost:8081 (username: `admin`, password: `password`)

## 🗄️ Database Architecture

### Automatic Database Selection

The application intelligently selects the database based on your environment:

```bash
# Local Development (Default)
ENVIRONMENT=development  # → Uses MongoDB

# Production
ENVIRONMENT=production   # → Uses Firestore
```

### MongoDB (Local Development)

- **Container**: Runs in Docker with persistent storage
- **UI**: Mongo Express web interface on port 8081
- **Connection**: Automatic with authentication
- **Data**: Persisted in Docker volumes

### Firestore (Production)

- **Cloud-based**: Google Cloud Firestore
- **Scalable**: Automatic scaling and backup
- **Setup**: Requires service account credentials

## 🎨 Enhanced Dashboard Features

### Interactive Charts

1. **Monthly Spending Trends**
   - Gradient-filled line charts
   - Month-over-month change indicators
   - Smart tooltips with percentage changes

2. **Top Vendors Analysis**
   - Horizontal bar charts with rankings
   - Gold/Silver/Bronze color coding
   - Vendor icons and spending percentages

3. **Spending Predictions**
   - Trend-based forecasting
   - Category-wise predictions
   - Visual trend indicators

### Smart Analytics

- **Financial Health Score**: Overall spending assessment
- **Budget Recommendations**: Personalized suggestions
- **Spending Insights**: Automatic warnings and tips
- **Gemini AI Integration**: Ask questions about your spending

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENVIRONMENT` | `development` or `production` | `development` | No |
| `DB_TYPE` | `mongodb` or `firestore` | Auto-detected | No |
| `MONGODB_URI` | MongoDB connection string | Auto-configured | No |
| `MONGODB_DATABASE` | MongoDB database name | `expense_tracker` | No |
| `GOOGLE_CLOUD_PROJECT` | GCP project ID | - | For Firestore |
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account path | - | For Firestore |
| `GEMINI_API_KEY` | Gemini AI API key | - | For AI features |

### Database Configuration

#### For Local Development (MongoDB)
```bash
# Automatic - no configuration needed
docker-compose up
```

#### For Production (Firestore)
```bash
# Set environment variables
export ENVIRONMENT=production
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Run the application
go run .
```

## 🐳 Docker Setup

### Services

The Docker Compose setup includes:

1. **MongoDB**: Database with authentication
2. **Mongo Express**: Web UI for MongoDB
3. **Expense Tracker**: Main application

### Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Rebuild and start
docker-compose up --build -d

# View specific service logs
docker-compose logs -f expense-tracker
```

## 🔍 MongoDB Web UI

Access the MongoDB interface at http://localhost:8081:

### Features
- **Browse Collections**: View transactions and unparsed emails
- **Filter Data**: Use MongoDB query syntax
- **Export Data**: Download as JSON or CSV
- **Edit Records**: Modify documents directly
- **Search**: Find specific transactions

### Common Queries
```javascript
// Find all transactions
{}

// Transactions above $100
{"amount": {"$gt": 100}}

// Transactions from specific vendor
{"vendor": "Amazon"}

// Transactions by date range
{"date": {"$gte": "2024-01-01", "$lte": "2024-12-31"}}
```

## 📁 File Structure

```
expense-tracker/
├── credentials/
│   ├── client_secret.json      # OAuth credentials
│   └── token.json             # OAuth token (auto-generated)
├── static/                    # Frontend assets
│   ├── index.html            # Main dashboard
│   ├── script.js             # Dashboard JavaScript
│   └── style.css             # Enhanced styling
├── docker-compose.yml         # Docker services
├── database.go               # Database abstraction layer
├── mongodb.go                # MongoDB implementation
├── firestore.go              # Firestore implementation
├── api.go                    # API endpoints
├── auth.go                   # OAuth authentication
├── analytics.go              # Analytics and insights
├── gemini.go                 # AI integration
├── main.go                   # Application entry point
└── README.md                 # This file
```

## 🚀 Development

### Running Locally

```bash
# Set environment variables for MongoDB
export MONGODB_URI="mongodb://admin:password@localhost:27017/expense_tracker?authSource=admin"
export MONGODB_DATABASE="expense_tracker"
export DB_TYPE="mongodb"

# Run the application
go run .
```

### API Endpoints

- `GET /` - Dashboard
- `GET /api/transactions` - Get all transactions
- `GET /analytics` - Spending analytics
- `GET /insights` - Smart insights
- `GET /predictions` - Spending predictions
- `GET /score` - Financial health score
- `POST /ask-gemini` - AI-powered questions

## 🔧 Troubleshooting

### Common Issues

1. **MongoDB Connection Error**:
   ```
   Error: server selection error: server selection timeout
   ```
   **Solution**: Ensure MongoDB container is running and environment variables are set correctly

2. **Port 8080 Already in Use**:
   ```
   Error: bind: address already in use
   ```
   **Solution**: Stop other services using port 8080 or change port in docker-compose.yml

3. **OAuth Token Expired**:
   ```
   Error: oauth2: "invalid_grant" "Token has been expired or revoked"
   ```
   **Solution**: Delete `credentials/token.json` and restart the application

4. **Mongo Express Login Issues**:
   - **URL**: http://localhost:8081
   - **Username**: `admin`
   - **Password**: `password`

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f expense-tracker
docker-compose logs -f mongodb
docker-compose logs -f mongo-express

# Application logs only
docker-compose logs -f expense-tracker
```

## 📊 Supported Transaction Types

- **HDFC Bank Credit Card**: ZOMATO, SWIGGY, and other merchants
- **Extensible Parser**: Easy to add new bank formats and transaction types
- **Smart Categorization**: Automatic transaction categorization
- **Vendor Recognition**: Intelligent vendor name extraction

## 🔒 Security

- ✅ OAuth 2.0 authentication for Gmail access
- ✅ Secure credential storage
- ✅ Docker container isolation
- ✅ Environment-based configuration
- ✅ No hardcoded secrets
- ⚠️ Keep credentials directory secure
- ⚠️ Never commit credentials to Git

## 🚢 Production Deployment

### Google Cloud Run

```bash
# Build and deploy to Cloud Run
gcloud run deploy expense-tracker \
  --source . \
  --platform managed \
  --region us-central1 \
  --set-env-vars ENVIRONMENT=production,GOOGLE_CLOUD_PROJECT=your-project-id
```

### Environment Variables for Production

```bash
export ENVIRONMENT=production
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export GEMINI_API_KEY=your-gemini-key
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with MongoDB locally
5. Submit a pull request

## 📝 License

MIT License - See LICENSE file for details

---

**Happy Expense Tracking!** 💰📊✨ 