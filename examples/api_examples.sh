#!/bin/bash

# API Examples Script for Cobra Template
# This script demonstrates how to use the various API endpoints

BASE_URL="http://localhost:8080"

echo "🚀 Cobra Template API Examples"
echo "==============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}$1${NC}"
    echo "----------------------------------------"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Function to check if server is running
check_server() {
    print_info "Checking if server is running..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not running. Please start it with: go run main.go server"
        exit 1
    fi
}

# Health Check Examples
health_checks() {
    print_header "Health Check Examples"
    
    echo "1. Basic Health Check:"
    curl -s "$BASE_URL/health" | jq '.'
    echo ""
    
    echo "2. Readiness Check:"
    curl -s "$BASE_URL/ready" | jq '.'
    echo ""
    
    echo "3. Liveness Check:"
    curl -s "$BASE_URL/live" | jq '.'
    echo ""
}

# User Management Examples
user_examples() {
    print_header "User Management Examples"
    
    echo "1. Create a new user:"
    USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "alice_demo",
            "email": "alice.demo@example.com",
            "first_name": "Alice",
            "last_name": "Demo"
        }')
    echo "$USER_RESPONSE" | jq '.'
    
    # Extract user ID for subsequent operations
    USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id')
    echo ""
    
    echo "2. List all users:"
    curl -s "$BASE_URL/api/v1/users" | jq '.'
    echo ""
    
    if [ "$USER_ID" != "null" ]; then
        echo "3. Get user by ID:"
        curl -s "$BASE_URL/api/v1/users/$USER_ID" | jq '.'
        echo ""
        
        echo "4. Update the user:"
        curl -s -X PUT "$BASE_URL/api/v1/users/$USER_ID" \
            -H "Content-Type: application/json" \
            -d '{
                "first_name": "Alice Updated",
                "active": false
            }' | jq '.'
        echo ""
    fi
    
    echo "5. Get user by email:"
    curl -s "$BASE_URL/api/v1/users/email/john.doe@example.com" | jq '.'
    echo ""
    
    echo "6. Get user by username:"
    curl -s "$BASE_URL/api/v1/users/username/jane_smith" | jq '.'
    echo ""
    
    echo "7. List users with pagination:"
    curl -s "$BASE_URL/api/v1/users?limit=2&offset=0" | jq '.'
    echo ""
    
    if [ "$USER_ID" != "null" ]; then
        echo "8. Delete the created user:"
        curl -s -X DELETE "$BASE_URL/api/v1/users/$USER_ID" | jq '.'
        echo ""
    fi
}

# Validation Examples
validation_examples() {
    print_header "Validation Examples"
    
    echo "1. Create user with invalid email:"
    curl -s -X POST "$BASE_URL/api/v1/users" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser",
            "email": "invalid-email",
            "first_name": "Test",
            "last_name": "User"
        }' | jq '.'
    echo ""
    
    echo "2. Create user with missing required fields:"
    curl -s -X POST "$BASE_URL/api/v1/users" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser"
        }' | jq '.'
    echo ""
    
    echo "3. Create user with username too short:"
    curl -s -X POST "$BASE_URL/api/v1/users" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "ab",
            "email": "test@example.com",
            "first_name": "Test",
            "last_name": "User"
        }' | jq '.'
    echo ""
}

# Error Handling Examples
error_examples() {
    print_header "Error Handling Examples"
    
    echo "1. Get non-existent user:"
    curl -s "$BASE_URL/api/v1/users/00000000-0000-0000-0000-000000000000" | jq '.'
    echo ""
    
    echo "2. Invalid UUID format:"
    curl -s "$BASE_URL/api/v1/users/invalid-uuid" | jq '.'
    echo ""
    
    echo "3. Get user by non-existent email:"
    curl -s "$BASE_URL/api/v1/users/email/nonexistent@example.com" | jq '.'
    echo ""
}

# Performance Examples
performance_examples() {
    print_header "Performance Examples"
    
    echo "1. Measure response time for health check:"
    curl -w "Response time: %{time_total}s\n" -s -o /dev/null "$BASE_URL/health"
    echo ""
    
    echo "2. Measure response time for user list:"
    curl -w "Response time: %{time_total}s\n" -s -o /dev/null "$BASE_URL/api/v1/users"
    echo ""
}

# Documentation Examples
documentation_examples() {
    print_header "Documentation Examples"
    
    echo "1. API Documentation endpoint:"
    curl -s "$BASE_URL/docs" | jq '.'
    echo ""
}

# Main execution
main() {
    echo "Starting API examples..."
    echo ""
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed. Please install jq to format JSON output."
        print_info "On macOS: brew install jq"
        print_info "On Ubuntu: sudo apt-get install jq"
        exit 1
    fi
    
    # Check if server is running
    check_server
    echo ""
    
    # Run examples
    health_checks
    user_examples
    validation_examples
    error_examples
    performance_examples
    documentation_examples
    
    print_success "All examples completed!"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
