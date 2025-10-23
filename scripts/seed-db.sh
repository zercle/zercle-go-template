#!/bin/bash
set -e

# Seed database with initial data
# Usage: ./scripts/seed-db.sh [environment]

ENVIRONMENT=${1:-local}
CONFIG_FILE="configs/${ENVIRONMENT}.yaml"

echo "Seeding database for ${ENVIRONMENT} environment..."

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql is not installed"
    exit 1
fi

# Run migrations first
echo "Running migrations..."
# Migration command would go here
# Example: psql -U postgres -d postgres -f migrations/latest.sql

echo "Database seeding completed!"
