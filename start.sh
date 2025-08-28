#!/bin/bash

# Course Registration System Startup Script
set -e

echo "🎓 Course Registration System - Complete Startup"
echo "================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check if a service is healthy
wait_for_service() {
    local service=$1
    local timeout=${2:-60}
    local count=0
    
    print_step "Waiting for $service to be healthy..."
    
    while [ $count -lt $timeout ]; do
        if docker-compose -f docker-compose.db.yml ps -q $service | xargs docker inspect --format='{{.State.Health.Status}}' 2>/dev/null | grep -q "healthy"; then
            print_success "$service is healthy!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        count=$((count + 2))
    done
    
    print_error "$service failed to become healthy within $timeout seconds"
    return 1
}

#
print_step "Creating Docker network..."
./scripts/network-manager.sh create


print_step "Starting database services..."
docker-compose -f docker-compose.db.yml up -d


print_step "Waiting for database services to initialize..."
wait_for_service "pg-primary" 120
wait_for_service "pg-replica-1" 60
wait_for_service "pg-replica-2" 60
wait_for_service "pgpool" 60
wait_for_service "pgbouncer" 60
print_step "Building API Docker image..."
./build.sh

# Step 5: Start application services
print_step "Starting application services..."
docker-compose up -d

# Step 6: Wait for API services to be ready
print_step "Waiting for API services to start..."
sleep 10

# Step 7: Run database migrations
print_step "Running database migrations..."
make migrate-up || print_warning "Migration may have failed or already applied"

# Step 8: Display service status
print_step "Checking service status..."
echo ""
echo "🔍 Database Services:"
docker-compose -f docker-compose.db.yml ps

echo ""
echo "🔍 Application Services:"
docker-compose ps

echo ""
echo "🌐 Service URLs:"
echo "  📊 Load Balancer (nginx): http://localhost"
echo "  🗄️  Database Admin (adminer): http://localhost:8080"
echo "  🔴 Redis Master: localhost:6379"
echo "  🔴 Redis Sentinels: localhost:26379, localhost:26380, localhost:26381"
echo "  🐘 PostgreSQL Primary: localhost:15432"
echo "  🐘 PGPool: localhost:9999"
echo "  🐘 PGBouncer: localhost:6432"

echo ""
echo "🚀 API Endpoints (via load balancer):"
echo "  POST http://localhost/api/v1/register"
echo "  GET  http://localhost/api/v1/students/{id}/registrations"
echo "  GET  http://localhost/health"

echo ""
echo "🔧 Individual API instances (for development):"
echo "  API 1: http://localhost:8081"
echo "  API 2: http://localhost:8082"
echo "  API 3: http://localhost:8083"
echo "  API 4: http://localhost:8084"

echo ""
print_success "Course Registration System is ready!"
echo ""
echo "📚 Next steps:"
echo "  • Test the health endpoint: curl http://localhost/health"
echo "  • View logs: make logs"
echo "  • Run load tests: cd performance_tests/k6 && make load-test"
echo "  • Stop services: make full-down"
