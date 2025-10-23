#!/bin/bash
set -e

# Run development server
# Usage: ./scripts/run-dev.sh

echo "Starting development server..."

# Set development environment
export ENV=local
export DEBUG=true

# Build and run
go run ./cmd/server
