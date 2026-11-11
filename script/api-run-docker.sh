#!/bin/bash

## Install Goose
echo "Installing goose"
go install github.com/pressly/goose/v3/cmd/goose@latest

## Run migrations
echo "Running migrations"
POSTGRES_URL="host=${DB_HOST} port=${DB_PORT} user=${DB_USERNAME} password=${DB_PASSWORD} dbname=${DB_DATABASE_NAME} sslmode=disable"
goose --dir './internal/database/migrations' postgres "${POSTGRES_URL}" up

## Start the server
go run github.com/agrowex/bhoomi-platform/cmd