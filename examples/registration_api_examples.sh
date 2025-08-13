#!/bin/bash

# Course Registration System API Examples
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1"

echo "=== Course Registration System API Examples ==="
echo

# Function to make HTTP requests with proper formatting
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    echo "🔗 $method $endpoint"
    if [ -n "$data" ]; then
        echo "📤 Request Body: $data"
    fi
    
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X "$method" "$BASE_URL$endpoint")
    fi
    
    echo "📥 Response:"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo
    echo "---"
    echo
}

# Health Check
echo "1. Health Check"
make_request "GET" "/../../health"

# Register for courses
echo "2. Course Registration"
make_request "POST" "/register" '{
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "section_ids": [
    "223e4567-e89b-12d3-a456-426614174001",
    "323e4567-e89b-12d3-a456-426614174002"
  ]
}'

# Get student registrations
echo "3. Get Student Registrations"
make_request "GET" "/students/123e4567-e89b-12d3-a456-426614174000/registrations"

# Get waitlist status
echo "4. Get Waitlist Status"
make_request "GET" "/students/123e4567-e89b-12d3-a456-426614174000/waitlist"

# Get available sections
echo "5. Get Available Sections"
make_request "GET" "/sections/available?semester_id=423e4567-e89b-12d3-a456-426614174003"

# Drop a course
echo "6. Drop Course"
make_request "POST" "/register/drop" '{
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "section_id": "223e4567-e89b-12d3-a456-426614174001"
}'

echo "=== Performance Test ==="
echo "7. Concurrent Registration Test (5 students, same section)"

# Generate UUIDs for test students
student_ids=(
    "523e4567-e89b-12d3-a456-426614174001"
    "523e4567-e89b-12d3-a456-426614174002"
    "523e4567-e89b-12d3-a456-426614174003"
    "523e4567-e89b-12d3-a456-426614174004"
    "523e4567-e89b-12d3-a456-426614174005"
)

section_id="323e4567-e89b-12d3-a456-426614174002"

# Run concurrent registrations
for student_id in "${student_ids[@]}"; do
    {
        echo "🏃 Registering student $student_id"
        curl -s -X POST "$BASE_URL/register" \
            -H "Content-Type: application/json" \
            -d "{\"student_id\": \"$student_id\", \"section_ids\": [\"$section_id\"]}" \
            | jq '.'
    } &
done

# Wait for all background jobs to complete
wait

echo
echo "=== Load Test with curl ==="
echo "8. Load Test (10 concurrent requests)"

# Simple load test
for i in {1..10}; do
    {
        student_id="load-test-$(uuidgen)"
        curl -s -X POST "$BASE_URL/register" \
            -H "Content-Type: application/json" \
            -d "{\"student_id\": \"$student_id\", \"section_ids\": [\"$section_id\"]}" \
            > /dev/null
        echo "Request $i completed"
    } &
done

wait
echo "Load test completed!"

echo
echo "=== Error Handling Examples ==="

echo "9. Invalid Student ID Format"
make_request "POST" "/register" '{
  "student_id": "invalid-uuid",
  "section_ids": ["223e4567-e89b-12d3-a456-426614174001"]
}'

echo "10. Missing Required Fields"
make_request "POST" "/register" '{
  "student_id": "123e4567-e89b-12d3-a456-426614174000"
}'

echo "11. Empty Section IDs"
make_request "POST" "/register" '{
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "section_ids": []
}'

echo "=== API Testing Complete ==="
