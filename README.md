# Expense Tracker

A personal expense tracker that parses bank transaction emails from Gmail, supports Google Pay activity imports, and provides an AI chat assistant powered by Claude.

## What it does

- Syncs transaction emails from Gmail (HDFC, ICICI credit card)
- Imports Google Pay activity HTML exports from Google Takeout (incremental — only new transactions are added)
- Stores transactions in MongoDB (dev) or Firestore (prod)
- Web dashboard with spending summaries, category breakdowns, and trends
- Claude-powered chat to query your spending in natural language
- Persistent memory so the assistant remembers context across conversations

## Setup

### Prerequisites

- Go 1.24+
- MongoDB (for local dev) or a Firestore project (for prod)
- Google Cloud project with Gmail API enabled
- OAuth 2.0 client credentials (`client_secret.json`)
- Anthropic API key (for AI chat)

### Credentials

```bash
mkdir credentials
cp /path/to/client_secret.json credentials/
# For production Firestore:
cp /path/to/service-account.json credentials/
```

### Environment variables

```bash
ENVIRONMENT=production          # omit or set to anything else for MongoDB
ANTHROPIC_API_KEY=sk-...        # required for AI chat
MONGODB_URI=mongodb://...       # optional, defaults to localhost
GOOGLE_CLOUD_PROJECT=your-id   # required for Firestore (production)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

### Run

```bash
# Start API server
go run . api

# Or sync Gmail emails once (CLI mode)
go run .
```

With Docker:

```bash
docker-compose up -d
```

Dashboard: http://localhost:8080

## Google Pay import

Export your Google Pay activity from [Google Takeout](https://takeout.google.com), then upload the HTML file via the dashboard. Only transactions newer than your latest stored transaction are imported — re-uploading a full export is fast.

## Project structure

```
ai/          Claude client and tool definitions
handlers/    HTTP handlers and routing
models/      Database interfaces (MongoDB + Firestore)
services/    Email parsing, Google Pay import, reporting, memory
frontend/    Web dashboard
```

## Supported banks

- HDFC Bank (email alerts)
- ICICI Bank credit card (email alerts)
