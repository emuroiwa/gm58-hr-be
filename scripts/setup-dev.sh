#!/bin/bash

echo "Setting up GM58-HR Backend for development..."

# Create .env file from example if it doesn't exist
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Created .env file from template. Please update with your configuration."
fi

# Download dependencies
echo "Downloading Go dependencies..."
go mod download

# Run migrations
echo "Running database migrations..."
go run scripts/migrate.go -cmd=up

echo "Development setup complete!"
echo "Run 'go run cmd/server/main.go' to start the development server"
