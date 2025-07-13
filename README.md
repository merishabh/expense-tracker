# 💰 Expense Tracker

An intelligent expense tracker that automatically processes bank transaction emails from Gmail and provides AI-powered analytics with beautiful visualizations.

## 🎯 What It Does

- **Automatically reads** bank transaction emails from your Gmail
- **Extracts transaction data** (amount, vendor, date, category) using smart parsing
- **Categorizes expenses** using AI-powered vendor classification
- **Provides analytics** through an interactive web dashboard
- **Offers insights** with spending predictions, budget recommendations, and financial health scores

## ✨ Key Features

- 🤖 **AI-Powered**: Smart vendor categorization and spending insights using Gemini AI
- 📊 **Rich Analytics**: Interactive charts, spending trends, and financial health scoring
- 🔄 **Automated Processing**: Fetches and processes emails automatically
- 💾 **Dual Database**: MongoDB for development, Firestore for production
- 📱 **Modern UI**: Responsive dashboard with beautiful visualizations
- 🐳 **Docker Ready**: Fully containerized setup

## 🚀 Quick Setup

### Prerequisites

- Docker & Docker Compose
- Google Cloud Project with Gmail API enabled
- OAuth 2.0 credentials for Gmail access

### 1. Get OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create/select a project and enable Gmail API
3. Create OAuth 2.0 Client ID (Desktop application)
4. Download `client_secret.json`

### 2. Setup and Run

```bash
# Clone repository
git clone <your-repo-url>
cd expense-tracker

# Add credentials
mkdir credentials
cp /path/to/your/client_secret.json credentials/

# Start services
docker-compose up -d

# View logs and follow OAuth setup
docker-compose logs -f expense-tracker
```

### 3. Complete OAuth Setup

1. Check logs for OAuth URL
2. Open URL in browser and authenticate with Gmail
3. Return to application - authentication will be saved

### 4. Access Dashboard

- **Main Dashboard**: http://localhost:8080
- **MongoDB Admin**: http://localhost:8081 (admin/password)

## 🎨 Dashboard Features

- **Spending Analytics**: Category breakdowns, monthly trends, top vendors
- **AI Chat Assistant**: Ask questions about your spending patterns
- **Smart Insights**: Automatic warnings and budget recommendations
- **Transaction Management**: View and categorize recent transactions
- **Predictions**: AI-powered spending forecasts

## ⚙️ Configuration

### Environment Variables

```bash
# Optional - defaults work for most setups
ENVIRONMENT=development         # or 'production' for Firestore
GEMINI_API_KEY=your_key        # For AI features
MONGODB_URI=mongodb://...      # Custom MongoDB connection
```

### Supported Banks

Currently supports transaction emails from:
- HDFC Bank (Credit Card & Bank Transfers)  
- ICICI Bank (Credit Card)
- Other banks can be added by extending parser functions

## 📖 Usage

1. **First Run**: Complete OAuth authentication
2. **Email Processing**: Application automatically fetches transaction emails
3. **View Analytics**: Open dashboard to see spending insights
4. **AI Assistant**: Ask questions like "How much did I spend on food?" 
5. **Budget Planning**: Review recommendations and insights

## 🔧 Development

### Project Structure

```
expense-tracker/
├── main.go              # Main application entry
├── api.go               # Web API endpoints  
├── parser.go            # Email parsing logic
├── gemini.go            # AI integration
├── database.go          # Database interfaces
├── static/              # Frontend assets
└── docker-compose.yml   # Container configuration
```

### Adding New Banks

Extend `parser.go` with new parsing functions for different email formats.

## 🛠️ Production Deployment

For production, set `ENVIRONMENT=production` and configure Firestore:

```bash
export ENVIRONMENT=production
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

## 📝 Notes

- Transactions are automatically categorized using AI when possible
- Manual category mappings can be customized in `models.go`
- All data is stored securely in your chosen database
- OAuth tokens are refreshed automatically

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Add/extend parser functions for new banks
4. Test with sample emails
5. Submit pull request

---

Built with Go, MongoDB/Firestore, and Gemini AI for intelligent expense tracking. 