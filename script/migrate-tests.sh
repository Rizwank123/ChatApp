#!/usr/bin/env bash
if [ -f "test.env" ]; then
  set -o allexport
  . test.env set
elif [ -f ".env" ]; then
  set -o allexport
  . .env set
fi

echo "Running migrations for tests..."
if [ -z "${DB_HOST}" ]; then
  echo "No database configured in test.env or .env. Skipping database migrations for unit tests."
  exit 0
fi

POSTGRES_URL="host=${DB_HOST} port=${DB_PORT:-5432} user=${DB_USERNAME:-postgres} password=${DB_PASSWORD:-postgres} dbname=${DB_DATABASE_NAME:-chatapp} sslmode=disable"

if command -v goose >/dev/null 2>&1; then
  goose --dir './internal/database/migrations' postgres "${POSTGRES_URL}" up || {
    echo "Warning: Database migration failed. Continuing with unit tests..."
  }
else
  echo "goose not installed, skipping test migrations."
fi
