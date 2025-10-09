#!/bin/bash
set -e

DB_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable"}

echo "🔄 Running database migrations..."
echo "Database: $DB_URL"

migrate -path migrations -database "$DB_URL" up

echo "✅ Migrations completed successfully!"
