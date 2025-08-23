# Manual API Testing Guide for Course Registration System

This guide provides step-by-step instructions for manually testing the course registration API endpoints using curl commands and your Cobra CLI application.

## Prerequisites

### 1. Start the System
```bash

cd /Users/la60716/cobra-template
make up
python3 generate_test_data_db.py
make run
make start
```

### 2. Verify System is Running
```bash
# Check health endpoint
curl -X GET http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"2025-08-23T12:30:00Z","database":"connected","cache":"connected"}
```

## Get Sample Data for Testing

Before testing registration endpoints, you need to get sample student and section IDs from the database:

### Get Sample Student ID
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT student_id, student_number, first_name, last_name FROM students WHERE enrollment_status = 'active' LIMIT 5;"
```

### Get Sample Section IDs
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT s.section_id, c.course_code, c.course_name, s.section_number, s.available_seats FROM sections s JOIN courses c ON s.course_id = c.course_id LIMIT 10;"
```

### Get Sample Semester ID
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT semester_id, semester_code, semester_name FROM semesters WHERE is_active = true;"
```

**Copy the IDs from the output above for use in the following tests.**

## API Endpoint Testing

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health \
  -H "Content-Type: application/json" \
  -w "\nStatus Code: %{http_code}\n"
```
**Expected**: Status 200, JSON response with system health

---

### 2. Get Available Sections
```bash
# Replace SEMESTER_ID with actual semester ID from database
curl -X GET "http://localhost:8080/api/v1/sections/available?semester_id=SEMESTER_ID" \
  -H "Content-Type: application/json" \
  -w "\nStatus Code: %{http_code}\n"
```

**Expected**: Status 200, Array of available sections

**Example with actual ID**:
```bash
curl -X GET "http://localhost:8080/api/v1/sections/available?semester_id=123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json"
```{"success":true,"message":"Student registrations retrieved successfully","data":{"registrations":[{"registration_id":"16c03a1e-7858-4dd4-b6c7-7a564567e99e","student_id":"afbb77da-008e-4dc3-ad88-28b58ff22eed","section_id":"87cdd665-0b12-412b-93ca-668b865f4b7e","status":"enrolled","registration_date":"2025-08-23T12:59:01.571793+07:00","created_at":"2025-08-23T12:59:01.571793+07:00","updated_at":"2025-08-23T12:59:01.571793+07:00","version":1,"student":{"student_id":"afbb77da-008e-4dc3-ad88-28b58ff22eed","student_number":"S2024000001","first_name":"Michelle","last_name":"Gonzalez","enrollment_status":"active","created_at":"2025-08-23T19:31:42.47748+07:00","updated_at":"2025-08-23T19:31:42.477484+07:00","version":1},"section":{"section_id":"87cdd665-0b12-412b-93ca-668b865f4b7e","course_id":"5183538a-851e-4c8b-9780-9b0e19bb26f8","semester_id":"2c40494b-e001-4a8d-8298-6482f1f4169c","section_number":"001","total_seats":30,"available_seats":30,"is_active":true,"created_at":"2025-08-23T19:31:47.212967+07:00","updated_at":"2025-08-23T19:31:47.212975+07:00","version":1,"course":{"course_id":"00000000-0000-0000-0000-000000000000","course_code":"","course_name":"","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","version":0},"semester":{"semester_id":"00000000-0000-0000-0000-000000000000","semester_code":"","semester_name":"","start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","registration_start":"0001-01-01T00:00:00Z","registration_end":"0001-01-01T00:00:00Z","is_active":false,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}}]}}

---

### 3. Register for Courses
```bash
# Replace STUDENT_ID and SECTION_ID with actual IDs
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "STUDENT_ID",
    "section_ids": ["SECTION_ID1", "SECTION_ID2"]
  }' \
  -w "\nStatus Code: %{http_code}\n"
```

**Expected**: Status 200, Registration results for each section

**Example with actual IDs**:
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "123e4567-e89b-12d3-a456-426614174001",
    "section_ids": ["123e4567-e89b-12d3-a456-426614174002"]
  }'
```

---

### 4. Get Student Registrations
```bash
# Replace STUDENT_ID with actual student ID
curl -X GET "http://localhost:8080/api/v1/students/STUDENT_ID/registrations" \
  -H "Content-Type: application/json" \
  -w "\nStatus Code: %{http_code}\n"
```

**Expected**: Status 200, Array of student's registrations

---

### 5. Get Student Waitlist Status
```bash
# Replace STUDENT_ID with actual student ID
curl -X GET "http://localhost:8080/api/v1/students/STUDENT_ID/waitlist" \
  -H "Content-Type: application/json" \
  -w "\nStatus Code: %{http_code}\n"
```

**Expected**: Status 200, Array of waitlist entries for the student

---

### 6. Drop a Course
```bash
# Replace STUDENT_ID and SECTION_ID with actual IDs
curl -X POST http://localhost:8080/api/v1/register/drop \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "STUDENT_ID",
    "section_id": "SECTION_ID"
  }' \
  -w "\nStatus Code: %{http_code}\n"
```

**Expected**: Status 200, Success message

---

## Test Scenarios

### Scenario 1: Successful Registration Flow
1. Get available sections for a semester
2. Register a student for 2-3 courses
3. Check student's registrations
4. Drop one course
5. Verify the course was dropped

### Scenario 2: Waitlist Testing
1. Register multiple students for the same popular section until it's full
2. Register additional students to test waitlist functionality
3. Check waitlist status for students
4. Drop a course from an enrolled student
5. Verify waitlist processing

### Scenario 3: Error Handling
1. Try to register with invalid student ID
2. Try to register for non-existent section
3. Try to register the same student for the same course twice
4. Try to drop a course the student isn't enrolled in

## Detailed Test Commands

### Test with Real Data (Example)

First, get real IDs:
```bash
# Get a real student ID
STUDENT_ID=$(docker exec course_registration_postgres psql -U postgres -d course_registration -t -c "SELECT student_id FROM students WHERE enrollment_status = 'active' LIMIT 1;" | tr -d ' ')
echo "Student ID: $STUDENT_ID"

# Get a real section ID
SECTION_ID=$(docker exec course_registration_postgres psql -U postgres -d course_registration -t -c "SELECT section_id FROM sections LIMIT 1;" | tr -d ' ')
echo "Section ID: $SECTION_ID"
```

Then test registration:
```bash
# Register the student
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d "{
    \"student_id\": \"$STUDENT_ID\",
    \"section_ids\": [\"$SECTION_ID\"]
  }"

# Check registrations
curl -X GET "http://localhost:8080/api/v1/students/$STUDENT_ID/registrations" \
  -H "Content-Type: application/json"
```

## Response Examples

### Successful Registration Response
```json
{
  "results": [
    {
      "section_id": "123e4567-e89b-12d3-a456-426614174002",
      "status": "enrolled",
      "message": "Registration completed successfully"
    }
  ]
}
```

### Waitlist Response
```json
{
  "results": [
    {
      "section_id": "123e4567-e89b-12d3-a456-426614174002",
      "status": "waitlisted",
      "message": "Added to waitlist",
      "position": 3
    }
  ]
}
```

### Error Response
```json
{
  "error": "student not found",
  "details": "invalid student ID provided"
}
```

## Database Verification

After each test, you can verify the results in the database:

### Check Registrations Table
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT r.*, s.first_name, s.last_name FROM registrations r JOIN students s ON r.student_id = s.student_id ORDER BY r.created_at DESC LIMIT 10;"
```

### Check Section Availability
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT s.section_id, c.course_code, s.section_number, s.total_seats, s.available_seats FROM sections s JOIN courses c ON s.course_id = c.course_id WHERE s.available_seats < s.total_seats;"
```

### Check Waitlist
```bash
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT w.*, s.first_name, s.last_name FROM waitlist w JOIN students s ON w.student_id = s.student_id ORDER BY w.section_id, w.position;"
```

## Load Testing

For basic load testing, you can run multiple curl commands in parallel:

```bash
# Run 10 concurrent registrations
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/register \
    -H "Content-Type: application/json" \
    -d "{\"student_id\": \"$STUDENT_ID\", \"section_ids\": [\"$SECTION_ID\"]}" &
done
wait
```

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure the server is running on the correct port
2. **Database Connection Error**: Check PostgreSQL is running (`docker ps`)
3. **Invalid UUID Error**: Ensure you're using actual UUIDs from the database
4. **404 Not Found**: Verify the endpoint URL is correct

### Debug Commands
```bash
# Check server logs
tail -f /path/to/logfile

# Check if server is listening
lsof -i :8080

# Check database connection
docker exec course_registration_postgres psql -U postgres -d course_registration -c "SELECT 1;"
```

## Expected Behavior

- **Registration**: Should succeed for available seats, add to waitlist when full
- **Waitlist**: Should automatically process when seats become available
- **Drop**: Should free up seats and process waitlist
- **Concurrency**: Multiple simultaneous requests should be handled correctly
- **Error Handling**: Invalid requests should return appropriate error messages

This manual testing approach will help you verify all functionality of your course registration system.
