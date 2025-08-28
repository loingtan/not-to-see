# Course Registration System - Architecture Documentation

## System Architecture Overview

The system follows a **microservices architecture** with **asynchronous processing** and **Redis caching** for high throughput. It's built using **Clean Architecture** principles with clear separation of concerns across layers.

### High-Level System Architecture

```mermaid
graph TB
    Client[Web Client<br/>React/Vue/Mobile] 
    LB[Load Balancer<br/>Nginx/HAProxy]
    API[API Gateway<br/>Gin Router]
    
    subgraph "Application Layer"
        RS[Registration Service<br/>Core Business Logic]
        Handler[HTTP Handlers<br/>Request/Response]
        Middleware[Middleware<br/>Auth/Logging/CORS]
    end
    
    subgraph "Data Layer"
        Redis[(Redis Cache<br/>Seat Counts)]
        Queue[In-Memory Queue<br/>Async Jobs]
        DB[(PostgreSQL<br/>Primary Database)]
    end
    
    subgraph "Infrastructure"
        Health[Health Checks<br/>/health /ready /live]
        Monitor[Monitoring<br/>Metrics/Logs]
    end
    
    Client --> LB
    LB --> API
    API --> Middleware
    Middleware --> Handler
    Handler --> RS
    
    RS --> Redis
    RS --> Queue
    RS --> DB
    
    API --> Health
    RS --> Monitor
    
    style RS fill:#e1f5fe
    style Redis fill:#ffecb3
    style DB fill:#f3e5f5
    style Queue fill:#e8f5e8
```

### Registration Flow Sequence

```mermaid
sequenceDiagram
    participant Client
    participant API as API Gateway
    participant Service as Registration Service
    participant Cache as Redis Cache
    participant Queue as Message Queue
    participant DB as PostgreSQL
    
    Client->>API: POST /api/v1/register
    API->>Service: Register(studentID, sectionIDs)
    
    loop For each section
        Service->>Cache: DecrementAvailableSeats(sectionID)
        
        alt Seats Available
            Cache-->>Service: Success (new count)
            Service->>Queue: Enqueue DatabaseSyncJob
            Service-->>API: Status: "enrolled"
        else No Seats Available
            Cache-->>Service: Error: No seats
            Service->>Queue: Enqueue WaitlistJob
            Service-->>API: Status: "waitlisted"
        end
    end
    
    API-->>Client: Registration Results
    
    Note over Queue,DB: Asynchronous Processing
    Queue->>Service: Process DatabaseSyncJob
    Service->>DB: Create Registration Record
    
    Queue->>Service: Process WaitlistJob
    Service->>DB: Create Waitlist Entry
```

### Queue Processing Architecture

```mermaid
graph TB
    subgraph "Registration Flow"
        RegReq[Registration Request] --> Cache{Redis Cache<br/>Check Seats}
        Cache -->|Seats Available| Reserve[Reserve Seat<br/>Atomic Decrement]
        Cache -->|No Seats| Waitlist[Add to Waitlist]
        Reserve --> DBSync[Enqueue DB Sync Job]
        Waitlist --> WLJob[Enqueue Waitlist Job]
    end
    
    subgraph "Queue System"
        DBSync --> DBQueue[(Database Sync Queue)]
        WLJob --> WLQueue[(Waitlist Queue)]
        WLProc[Waitlist Processing] --> WLEntryQueue[(Waitlist Entry Queue)]
        
        DBQueue --> DBWorker[DB Sync Workers<br/>Pool of 10]
        WLQueue --> WLWorker[Waitlist Workers<br/>Pool of 10]
        WLEntryQueue --> WLEntryWorker[Waitlist Entry Workers<br/>Pool of 10]
    end
    
    subgraph "Database Operations"
        DBWorker --> CreateReg[Create Registration]
        DBWorker --> UpdateSeats[Update Section Seats]
        WLWorker --> ProcessNext[Create Wailist]
        WLEntryWorker --> CreateWL[Create Waitlist Entry]
        
        CreateReg --> DB[(PostgreSQL)]
        UpdateSeats --> DB
        ProcessNext --> DB
        CreateWL --> DB
    end
    
    style Cache fill:#ffecb3
    style DBQueue fill:#e8f5e8
    style WLQueue fill:#e8f5e8
    style WLEntryQueue fill:#e8f5e8
    style DB fill:#f3e5f5
```

### Database Schema Relationships

```mermaid
erDiagram
    STUDENTS {
        UUID student_id PK
        VARCHAR student_number UK
        VARCHAR first_name
        VARCHAR last_name
        VARCHAR email UK
        VARCHAR enrollment_status
        TIMESTAMP created_at
        TIMESTAMP updated_at
        INTEGER version
    }
    
    COURSES {
        UUID course_id PK
        VARCHAR course_code UK
        VARCHAR course_name
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
        INTEGER version
    }
    
    SEMESTERS {
        UUID semester_id PK
        VARCHAR semester_name
        DATE start_date
        DATE end_date
        TIMESTAMP registration_start
        TIMESTAMP registration_end
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
        INTEGER version
    }
    
    SECTIONS {
        UUID section_id PK
        UUID course_id FK
        UUID semester_id FK
        VARCHAR section_number
        INTEGER total_seats
        BOOLEAN is_active
        TIMESTAMP created_at
        TIMESTAMP updated_at
        INTEGER version
    }
    
    REGISTRATIONS {
        UUID registration_id PK
        UUID student_id FK
        UUID section_id FK
        VARCHAR status
        TIMESTAMP registration_date
        TIMESTAMP dropped_date
        TIMESTAMP created_at
        TIMESTAMP updated_at
        INTEGER version
    }
    
    WAITLIST {
        UUID waitlist_id PK
        UUID student_id FK
        UUID section_id FK
        INTEGER position
        TIMESTAMP timestamp
        TIMESTAMP expires_at
        TIMESTAMP created_at
        TIMESTAMP updated_at
    }

    %% Idempotency keys are now stored in Redis, not in the database
    %% REDIS_IDEMPOTENCY_KEYS {
    %%     VARCHAR key PK
    %%     UUID student_id
    %%     VARCHAR request_hash
    %%     TEXT response_data
    %%     INTEGER status_code
    %%     TIMESTAMP processed_at
    %%     TIMESTAMP expires_at
    %%     TIMESTAMP created_at
    %% }

    STUDENTS ||--o{ REGISTRATIONS : enrolls
    STUDENTS ||--o{ WAITLIST : waits_for
    STUDENTS ||--o{ IDEMPOTENCY_KEYS : requests
    COURSES ||--o{ SECTIONS : has
    SEMESTERS ||--o{ SECTIONS : offered_in
    SECTIONS ||--o{ REGISTRATIONS : contains
    SECTIONS ||--o{ WAITLIST : has_waitlist
```

### Concurrency Control Flow

```mermaid
graph TB
    Start[Registration Request] --> Check{Check Existing<br/>Registration}
    Check -->|Exists| AlreadyReg[Return: Already Registered]
    Check -->|Not Exists| CacheCheck[Redis: Atomic Decrement]
    
    CacheCheck -->|Success| CacheSuccess[Seat Reserved<br/>in Cache]
    CacheCheck -->|Fail: No Seats| AddWaitlist[Add to Waitlist]
    
    CacheSuccess --> EnqueueDB[Enqueue DB Sync Job]
    EnqueueDB --> ReturnSuccess[Return: Enrolled]
    
    AddWaitlist --> EnqueueWL[Enqueue Waitlist Job]
    EnqueueWL --> ReturnWaitlist[Return: Waitlisted]
    
    subgraph "Async Processing"
        DBWorker[DB Sync Worker] --> DBOp{Job Type}
        DBOp -->|create_registration| CreateReg[Create Registration<br/>with Optimistic Lock]
        DBOp -->|update_seats| UpdateSeats[Update Section Seats<br/>with Version Check]
        
        CreateReg --> DBCheck{DB Success?}
        DBCheck -->|Success| Complete[Job Complete]
        DBCheck -->|Fail: Version Conflict| Retry[Retry with<br/>Exponential Backoff]
        Retry --> CreateReg
        
        UpdateSeats --> SeatCheck{Update Success?}
        SeatCheck -->|Success| SyncComplete[Sync Complete]
        SeatCheck -->|Fail| RollbackCache[Rollback Cache<br/>Increment Seats]
    end
    
    style CacheCheck fill:#ffecb3
    style CreateReg fill:#f3e5f5
    style UpdateSeats fill:#f3e5f5
    style RollbackCache fill:#ffcdd2
```

## Use cases of redis data structure




## Idempotency and Duplicate Request Handling

The system implements **idempotency keys** to prevent duplicate registrations and ensure data consistency during network issues or client retries.

### Key Features

1. **Optional Idempotency Keys**: Clients can provide an idempotency key in the request body or `Idempotency-Key` header
2. **Automatic Duplicate Detection**: System compares request fingerprints to detect true duplicates
3. **Response Caching**: Successful responses are cached in Redis and returned for duplicate requests
4. **Configurable TTL**: Idempotency keys expire after 24 hours by default (stored in Redis)
5. **Error Handling**: Different request data with the same key returns an error
6. **High Performance**: Redis-based storage provides faster lookups and automatic cleanup

### Implementation Details

#### Request Fingerprinting
- Combines student ID and request data (section IDs) into a SHA-256 hash
- Ensures only identical requests are considered duplicates
- Prevents key reuse with different data

#### Storage and Cleanup
- **Idempotency keys stored in Redis** with automatic TTL-based expiration
- Keys expire automatically after 24 hours (configurable)
- **No database overhead** - Redis handles all idempotency key storage and cleanup
- **High performance** - Redis provides faster lookups compared to database queries
- **Memory efficient** - Automatic cleanup by Redis TTL mechanism

#### Usage Examples

**With idempotency key in request body:**
```json
{
  "student_id": "123e4567-e89b-12d3-a456-426614174000",
  "section_ids": ["789e0123-e45b-67c8-d901-234567890123"],
  "idempotency_key": "reg-2025-08-24-unique-key-001"
}
```

**With idempotency key in header:**
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: reg-2025-08-24-unique-key-001" \
  -d '{"student_id": "123e4567-e89b-12d3-a456-426614174000", "section_ids": ["789e0123-e45b-67c8-d901-234567890123"]}'
```

