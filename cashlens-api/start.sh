#!/bin/bash

# cashlens API startup script
# This script starts the Go API server with proper environment configuration

set -e  # Exit on error

echo "🚀 Starting cashlens API..."
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found"
    echo "   Please copy .env.example to .env and configure it"
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
    echo "⚠️  Warning: PostgreSQL is not running on localhost:5432"
    echo "   The API will fail to start without a database connection"
    echo ""
fi

# Run the API server
go run cmd/api/main.go
