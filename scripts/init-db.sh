#!/bin/bash
set -e

# Wait for PostgreSQL to be ready
until pg_isready -h postgres -p 5432 -U otellab; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done

echo "PostgreSQL is ready. Running migrations..."

# Run migrations
psql -h postgres -p 5432 -U otellab -d taskdb -f /migrations/001_create_tasks_table.sql

echo "Migrations completed successfully!"