#!/bin/bash

echo "Initializing New-API environment..."

# Create necessary directories
echo "Creating directories..."
mkdir -p logs
mkdir -p public/static
mkdir -p public/uploads
mkdir -p redis_data
mkdir -p mysql_data

# Set permissions
echo "Setting permissions..."
chmod -R 777 logs
chmod -R 777 public/static
chmod -R 777 public/uploads
chmod -R 777 redis_data
chmod -R 777 mysql_data

echo "Directory initialization completed!"
echo "Created:"
echo "  - logs/"
echo "  - public/static/"
echo "  - public/uploads/"
echo "  - redis_data/"
echo "  - mysql_data/"
echo ""
echo "All directories have been set with 777 permissions."

# Check if .env file exists
if [ ! -f .env ]; then
    echo ""
    echo "Warning: .env file not found. Please create it based on .env.example"
    echo "Required environment variables for your docker-compose.yml:"
    echo "  - SQL_DSN"
    echo "  - CONSOLE_JWT_SECRET"
    echo "  - CONSOLE_API_DOMAIN"
    echo "  - GENERATE_DEFAULT_TOKEN"
    echo "  - REDIS_CONN_STRING"
    echo "  - NODE_TYPE"
    echo "  - API_REQUEST_LOG_ENABLED=true"
else
    echo ""
    echo "✓ .env file found"
fi

echo ""
echo "Initialization complete. You can now run:"
echo "  docker-compose up -d"