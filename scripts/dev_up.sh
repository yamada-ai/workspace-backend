#!/bin/bash
set -e

echo "🚀 Starting development environment..."

# Start database
echo "1️⃣  Starting PostgreSQL..."
docker compose -f infrastructure/compose.dev.yml up -d db

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 5

# Run migrations
echo "2️⃣  Running migrations..."
bash scripts/migrate.sh

echo "✅ Development environment is ready!"
echo ""
echo "To start the server:"
echo "  DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable' go run cmd/work-tracker/main.go"
echo ""
echo "Or run the full stack:"
echo "  docker compose -f infrastructure/compose.dev.yml up"
