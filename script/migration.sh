#!/usr/bin/env bash
if [ ! -f ".env" ]
then
  touch .env
fi
set -o allexport
. .env set
echo "Running migrations"
POSTGRES_URL="host=${DB_HOST} port=${DB_PORT} user=${DB_USERNAME} password=${DB_PASSWORD} dbname=${DB_DATABASE_NAME} sslmode=disable"
goose --dir './internal/database/migrations' postgres "${POSTGRES_URL}" up
