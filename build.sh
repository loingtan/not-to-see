#!/bin/bash

# Build script for Course Registration API
set -e

echo "🏗️  Building Course Registration API Docker Image..."

# Build the Docker image
docker build -t course-registration-api:latest .

echo "✅ Docker image built successfully: course-registration-api:latest"

# Optional: Tag with version if provided
if [ ! -z "$1" ]; then
    docker tag course-registration-api:latest course-registration-api:$1
    echo "✅ Tagged image with version: course-registration-api:$1"
fi

echo "🎯 Image ready for deployment!"
echo ""
echo "📋 Next steps:"
echo "  1. Start database services: make db-up"
echo "  2. Start application stack: make up"
echo "  3. Run migrations: make migrate-up"
echo "  4. Check status: make logs"
