#!/bin/bash

# Cron-friendly expense tracker runner
# This script is designed to run in cron jobs without user interaction

# Set script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Set environment variables for cron job detection
export CRON_JOB=true
export DOCKER_CONTAINER=true

# Log file for cron job output
LOG_FILE="$SCRIPT_DIR/cron-expense-tracker.log"

# Function to log with timestamp
log_with_timestamp() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_with_timestamp "Starting expense tracker cron job"

# Check if .env file exists
if [ ! -f .env ]; then
    log_with_timestamp "❌ .env file not found. Please create it with GOOGLE_CLOUD_PROJECT=your-project-id"
    exit 1
fi

# Check if credentials directory exists
if [ ! -d credentials ]; then
log_with_timestamp "❌ credentials directory not found. Please set up credentials first. Go over the README.md file to see how to set up the credentials."
    exit 1
fi

# Check if token exists
if [ ! -f credentials/token.json ]; then
    log_with_timestamp "❌ No OAuth token found. Please run './refresh-token.sh' first to get a valid token."
    exit 1
fi

# Check token age (warn if older than 6 days, OAuth refresh tokens last 7 days by default)
if [ -f credentials/token.json ]; then
    TOKEN_AGE=$(find credentials/token.json -mtime +6 2>/dev/null)
    if [ ! -z "$TOKEN_AGE" ]; then
        log_with_timestamp "⚠️  OAuth token is older than 6 days. Consider refreshing it soon."
    fi
fi

log_with_timestamp "📋 Running expense tracker via Docker Compose"

# Run the expense tracker
if docker-compose up --build 2>&1 | tee -a "$LOG_FILE"; then
    log_with_timestamp "✅ Expense tracker completed successfully"
else
    log_with_timestamp "❌ Expense tracker failed with exit code $?"
    exit 1
fi

log_with_timestamp "🎉 Cron job completed" 