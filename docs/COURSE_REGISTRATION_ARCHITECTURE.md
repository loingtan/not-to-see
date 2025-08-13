# Course Registration System - Architecture Documentation

## Part 1: System Design and Data Modeling

### High-Level Architecture

The course registration system follows a microservices-inspired architecture with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   API Gateway   │    │   Client Apps   │
│   (Production)  │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Registration  │    │   Course        │    │   Student       │
│   Service       │    │   Service       │    │   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Redis Cache   │    │   Message Queue │    │   PostgreSQL    │
│                 │    │   (RabbitMQ)    │    │   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

1. **API Gateway/Load Balancer**: Routes requests and handles load distribution
2. **Registration Service**: Core business logic for course registration
3. **Course Service**: Manages course and section information
4. **Student Service**: Handles student data and authentication
5. **Redis Cache**: Caches seat availability and course details
6. **Message Queue**: Handles asynchronous registration processing
7. **PostgreSQL Database**: Primary data store with ACID compliance

### Database Schema Design

The system uses PostgreSQL with the following core entities:

#### Students Table
```sql
CREATE TABLE students (
    student_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_number VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    enrollment_status VARCHAR(20) DEFAULT 'active',
    academic_level VARCHAR(20) DEFAULT 'undergraduate',
    major VARCHAR(100),
    gpa DECIMAL(3,2) CHECK (gpa >= 0.0 AND gpa <= 4.0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);
```

#### Courses Table
```sql
CREATE TABLE courses (
    course_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_code VARCHAR(20) UNIQUE NOT NULL,
    course_name VARCHAR(200) NOT NULL,
    description TEXT,
    credits INTEGER NOT NULL CHECK (credits > 0),
    department VARCHAR(100) NOT NULL,
    prerequisites TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);
```

#### Sections Table (Critical for Concurrency)
```sql
CREATE TABLE sections (
    section_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id UUID NOT NULL REFERENCES courses(course_id),
    semester_id UUID NOT NULL REFERENCES semesters(semester_id),
    section_number VARCHAR(10) NOT NULL,
    instructor_name VARCHAR(200),
    meeting_times JSONB,
    location VARCHAR(200),
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    available_seats INTEGER NOT NULL CHECK (available_seats >= 0),
    reserved_seats INTEGER DEFAULT 0 CHECK (reserved_seats >= 0),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1, -- Critical for optimistic locking
    UNIQUE(course_id, semester_id, section_number),
    CONSTRAINT check_seat_consistency CHECK (available_seats + reserved_seats <= total_seats)
);
```

#### Registrations Table
```sql
CREATE TABLE registrations (
    registration_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(student_id),
    section_id UUID NOT NULL REFERENCES sections(section_id),
    status registration_status NOT NULL DEFAULT 'enrolled',
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    grade VARCHAR(5),
    credits_earned INTEGER,
    is_audit BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    UNIQUE(student_id, section_id)
);
```

#### Waitlist Table
```sql
CREATE TABLE waitlist (
    waitlist_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(student_id),
    section_id UUID NOT NULL REFERENCES sections(section_id),
    position INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notification_sent BOOLEAN DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(student_id, section_id)
);
```

## Part 2: Core Registration Logic and Concurrency Handling

### Registration Endpoint

**POST /api/v1/register**

```json
{
  "student_id": "uuid",
  "section_ids": ["uuid1", "uuid2", "uuid3"]
}
```

### Pessimistic Locking Strategy

**Use Case**: When you need guaranteed consistency and can tolerate reduced throughput.

```sql
-- Transaction flow:
BEGIN;
  SELECT available_seats FROM sections 
  WHERE section_id = $1 FOR UPDATE; -- Row-level lock
  
  IF available_seats > 0 THEN
    UPDATE sections 
    SET available_seats = available_seats - 1 
    WHERE section_id = $1;
    
    INSERT INTO registrations (...) VALUES (...);
  END IF;
COMMIT;
```

**Performance Analysis**:
- **Pros**: Guaranteed data consistency, no lost updates
- **Cons**: Increased lock contention, potential deadlocks, reduced throughput
- **Bottlenecks**: Under high contention (>100 concurrent requests), lock wait times increase exponentially

### Optimistic Locking Strategy (Recommended)

**Use Case**: High-throughput scenarios where conflicts are relatively rare.

```go
func (s *RegistrationService) attemptRegistration(ctx context.Context, studentID, sectionID uuid.UUID) (bool, error) {
    // Get current section state
    section, err := s.sectionRepo.GetByID(ctx, sectionID)
    if err != nil {
        return false, fmt.Errorf("failed to get section: %w", err)
    }

    if section.AvailableSeats <= 0 {
        return false, nil // No seats available
    }

    // Attempt to decrement available seats with optimistic locking
    section.AvailableSeats--
    section.Version++

    err = s.sectionRepo.UpdateWithOptimisticLock(ctx, section)
    if err != nil {
        // Optimistic lock failure - retry
        return false, err
    }

    // Create registration record
    registration := &domain.Registration{
        RegistrationID:   uuid.New(),
        StudentID:        studentID,
        SectionID:        sectionID,
        Status:           domain.StatusEnrolled,
        RegistrationDate: time.Now(),
    }

    if err := s.registrationRepo.Create(ctx, registration); err != nil {
        // Registration failed, need to rollback seat decrement
        section.AvailableSeats++
        s.sectionRepo.UpdateWithOptimisticLock(ctx, section)
        return false, fmt.Errorf("failed to create registration: %w", err)
    }

    return true, nil
}
```

**SQL Implementation**:
```sql
UPDATE sections 
SET available_seats = $1, version = version + 1, updated_at = NOW()
WHERE section_id = $2 AND version = $3;
-- If RowsAffected = 0, optimistic lock failed
```

## Part 3: Caching and Asynchronous Processing

### Caching Strategy (Cache-Aside Pattern)

```go
func (s *RegistrationService) registerForSection(ctx context.Context, studentID, sectionID uuid.UUID) RegistrationResult {
    // 1. Check cache for seat availability
    available, err := s.cacheService.GetAvailableSeats(ctx, sectionID)
    if err == nil && available <= 0 {
        // No seats available, add to waitlist immediately
        position, err := s.addToWaitlist(ctx, studentID, sectionID)
        return RegistrationResult{Status: "waitlisted", Position: &position}
    }

    // 2. Enqueue registration request for async processing
    registrationJob := RegistrationJob{
        StudentID: studentID,
        SectionID: sectionID,
        Timestamp: time.Now(),
    }

    s.queueService.EnqueueRegistration(ctx, registrationJob)
    
    return RegistrationResult{Status: "processing"}
}
```

### Redis Cache Implementation

**Key Patterns**:
- `section:seats:{section_id}` - Available seat count
- `section:details:{section_id}` - Section details
- `course:details:{course_id}` - Course information

**Atomic Seat Decrement (Lua Script)**:
```lua
local key = KEYS[1]
local current = redis.call("GET", key)
if current == false then
    return redis.error_reply("Key does not exist")
end
local value = tonumber(current)
if value <= 0 then
    return redis.error_reply("No seats available")
end
return redis.call("DECR", key)
```

### Asynchronous Queue Processing

```go
func (s *RegistrationService) ProcessRegistrationJob(ctx context.Context, job RegistrationJob) error {
    maxRetries := 3
    for attempt := 1; attempt <= maxRetries; attempt++ {
        success, err := s.attemptRegistration(ctx, job.StudentID, job.SectionID)
        if err != nil {
            if attempt == maxRetries {
                // Final attempt failed, add to waitlist
                s.addToWaitlist(ctx, job.StudentID, job.SectionID)
                return fmt.Errorf("registration failed after %d attempts: %w", maxRetries, err)
            }
            // Brief delay before retry
            time.Sleep(time.Duration(attempt*100) * time.Millisecond)
            continue
        }

        if success {
            // Update cache
            s.cacheService.DecrementAvailableSeats(ctx, job.SectionID)
            return nil
        }

        // No seats available, add to waitlist
        s.addToWaitlist(ctx, job.StudentID, job.SectionID)
        return nil
    }
    return errors.New("unexpected end of registration processing")
}
```

### Waitlist Handling

When a student drops a course:

```go
func (s *RegistrationService) DropCourse(ctx context.Context, studentID, sectionID uuid.UUID) error {
    // 1. Update registration status to 'dropped'
    // 2. Increment available seats
    // 3. Process waitlist
    
    nextEntry, err := s.waitlistRepo.GetNextInLine(ctx, sectionID)
    if err != nil || nextEntry == nil {
        return nil // No one in waitlist
    }

    // Remove from waitlist and enqueue registration job
    s.waitlistRepo.Delete(ctx, nextEntry.WaitlistID)
    
    registrationJob := RegistrationJob{
        StudentID: nextEntry.StudentID,
        SectionID: sectionID,
        Timestamp: time.Now(),
    }

    return s.queueService.EnqueueRegistration(ctx, registrationJob)
}
```

## Part 4: Database Internals and Optimization

### Isolation Levels

**Chosen Level: READ COMMITTED**

**Rationale**:
- Prevents dirty reads
- Allows non-repeatable reads (acceptable for seat counts)
- Good balance between consistency and performance
- Avoids phantom read issues in our use case

**Alternative: REPEATABLE READ** for critical operations requiring strict consistency.

### Indexing Strategy

```sql
-- Primary performance indexes
CREATE INDEX idx_registrations_student_section ON registrations(student_id, section_id);
CREATE INDEX idx_registrations_section_status ON registrations(section_id, status);
CREATE INDEX idx_sections_course_semester ON sections(course_id, semester_id);
CREATE INDEX idx_waitlist_section_position ON waitlist(section_id, position);
CREATE INDEX idx_students_student_number ON students(student_number);
CREATE INDEX idx_students_email ON students(email);

-- Composite indexes for complex queries
CREATE INDEX idx_sections_semester_active ON sections(semester_id, is_active) 
WHERE is_active = true;

-- Partial indexes for active records
CREATE INDEX idx_active_sections_availability ON sections(available_seats, total_seats) 
WHERE is_active = true AND available_seats > 0;
```

**B+ Tree Usage**:
- Logarithmic lookup time O(log n)
- Efficient range queries for pagination
- Supports ORDER BY operations without sorting

### Scalability Analysis

#### Horizontal Partitioning

**By Semester** (Recommended):
```sql
-- Partition by semester for time-based access patterns
CREATE TABLE registrations_fall2024 PARTITION OF registrations
FOR VALUES FROM ('2024-08-01') TO ('2024-12-31');

CREATE TABLE registrations_spring2025 PARTITION OF registrations  
FOR VALUES FROM ('2025-01-01') TO ('2025-05-31');
```

**Benefits**:
- Natural data lifecycle (archive old semesters)
- Improved query performance for current semester
- Simplified backup/restore operations

#### Database Sharding (Massive Scale)

**Shard Key Options**:

1. **student_id** (Recommended for global scalability):
   - Pros: Even distribution, student-centric queries efficient
   - Cons: Cross-shard queries for course availability

2. **course_id**: 
   - Pros: Course-centric operations efficient
   - Cons: Uneven distribution (popular courses)

**Sharding Implementation**:
```
Shard 1: student_id hash % 4 = 0
Shard 2: student_id hash % 4 = 1  
Shard 3: student_id hash % 4 = 2
Shard 4: student_id hash % 4 = 3
```

### Performance Benchmarks

**Target Metrics**:
- Registration API: < 200ms P95 latency
- Throughput: 1000+ registrations/second
- Cache hit ratio: > 95% for seat availability
- Queue processing: < 5 seconds end-to-end

**Monitoring**:
- Database connection pool usage
- Redis memory usage and hit rates
- Queue length and processing times
- Failed registration rates

## Implementation Status

✅ **Completed**:
- Domain models and interfaces
- Repository implementations (mock and real)
- Service layer with business logic
- Cache layer (Redis + Mock)
- Queue system (In-memory + interfaces for production)
- REST API handlers
- Database schema with optimistic locking

🔄 **Next Steps**:
- Integration tests
- Load testing
- Production queue implementation (RabbitMQ/Kafka)
- Monitoring and alerting
- API documentation
- Deployment configuration

This architecture provides a solid foundation for a scalable course registration system that can handle high concurrency while maintaining data consistency.
