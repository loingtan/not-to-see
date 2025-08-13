-- Course Registration System Database Schema
-- PostgreSQL Database Design with Optimistic Concurrency Control

-- Enable UUID extension for unique identifiers
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Students table
CREATE TABLE students (
    student_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_number VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    enrollment_status VARCHAR(20) DEFAULT 'active' CHECK (enrollment_status IN ('active', 'inactive', 'suspended')),
    academic_level VARCHAR(20) DEFAULT 'undergraduate' CHECK (academic_level IN ('undergraduate', 'graduate', 'doctoral')),
    major VARCHAR(100),
    gpa DECIMAL(3,2) CHECK (gpa >= 0.0 AND gpa <= 4.0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1 -- For optimistic locking
);

-- Courses table
CREATE TABLE courses (
    course_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_code VARCHAR(20) UNIQUE NOT NULL, -- e.g., 'CS101'
    course_name VARCHAR(200) NOT NULL,
    description TEXT,
    credits INTEGER NOT NULL CHECK (credits > 0),
    department VARCHAR(100) NOT NULL,
    prerequisites TEXT[], -- Array of prerequisite course codes
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Semesters table for academic terms
CREATE TABLE semesters (
    semester_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    semester_code VARCHAR(20) UNIQUE NOT NULL, -- e.g., 'FALL2024'
    semester_name VARCHAR(50) NOT NULL, -- e.g., 'Fall 2024'
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    registration_start TIMESTAMP WITH TIME ZONE NOT NULL,
    registration_end TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Sections table (course offerings in specific semesters)
CREATE TABLE sections (
    section_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id UUID NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
    semester_id UUID NOT NULL REFERENCES semesters(semester_id) ON DELETE CASCADE,
    section_number VARCHAR(10) NOT NULL, -- e.g., '001', '002'
    instructor_name VARCHAR(200),
    meeting_times JSONB, -- Flexible format for class schedule
    location VARCHAR(200),
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    available_seats INTEGER NOT NULL CHECK (available_seats >= 0),
    reserved_seats INTEGER DEFAULT 0 CHECK (reserved_seats >= 0), -- For temporary reservations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1, -- Critical for optimistic locking
    UNIQUE(course_id, semester_id, section_number),
    CONSTRAINT check_seat_consistency CHECK (available_seats + reserved_seats <= total_seats)
);

-- Registration status enum
CREATE TYPE registration_status AS ENUM ('enrolled', 'waitlisted', 'dropped', 'failed');

-- Registrations table (many-to-many relationship between students and sections)
CREATE TABLE registrations (
    registration_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    section_id UUID NOT NULL REFERENCES sections(section_id) ON DELETE CASCADE,
    status registration_status NOT NULL DEFAULT 'enrolled',
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    grade VARCHAR(5), -- Final grade (A, B, C, D, F, W, I, etc.)
    credits_earned INTEGER,
    is_audit BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    UNIQUE(student_id, section_id) -- Prevent duplicate registrations
);

-- Waitlist table for managing waiting students
CREATE TABLE waitlist (
    waitlist_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    section_id UUID NOT NULL REFERENCES sections(section_id) ON DELETE CASCADE,
    position INTEGER NOT NULL, -- Position in waitlist
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notification_sent BOOLEAN DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE, -- Optional: auto-remove after expiry
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(student_id, section_id) -- Prevent duplicate waitlist entries
);

-- Audit table for tracking all registration changes
CREATE TABLE registration_audit (
    audit_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_id UUID,
    student_id UUID NOT NULL,
    section_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'register', 'drop', 'waitlist', 'promote'
    old_status registration_status,
    new_status registration_status,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET,
    processing_time_ms INTEGER, -- Performance tracking
    error_message TEXT -- For failed attempts
);

-- Temporary seat reservations (for handling concurrent requests)
CREATE TABLE seat_reservations (
    reservation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    section_id UUID NOT NULL REFERENCES sections(section_id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance optimization

-- Students indexes
CREATE INDEX idx_students_student_number ON students(student_number);
CREATE INDEX idx_students_email ON students(email);
CREATE INDEX idx_students_enrollment_status ON students(enrollment_status);

-- Courses indexes
CREATE INDEX idx_courses_course_code ON courses(course_code);
CREATE INDEX idx_courses_department ON courses(department);
CREATE INDEX idx_courses_is_active ON courses(is_active);

-- Sections indexes
CREATE INDEX idx_sections_course_id ON sections(course_id);
CREATE INDEX idx_sections_semester_id ON sections(semester_id);
CREATE INDEX idx_sections_available_seats ON sections(available_seats);
CREATE INDEX idx_sections_course_semester ON sections(course_id, semester_id);

-- Registrations indexes
CREATE INDEX idx_registrations_student_id ON registrations(student_id);
CREATE INDEX idx_registrations_section_id ON registrations(section_id);
CREATE INDEX idx_registrations_status ON registrations(status);
CREATE INDEX idx_registrations_student_status ON registrations(student_id, status);

-- Waitlist indexes
CREATE INDEX idx_waitlist_section_id ON waitlist(section_id);
CREATE INDEX idx_waitlist_student_id ON waitlist(student_id);
CREATE INDEX idx_waitlist_position ON waitlist(section_id, position);
CREATE INDEX idx_waitlist_timestamp ON waitlist(timestamp);

-- Audit indexes
CREATE INDEX idx_audit_student_id ON registration_audit(student_id);
CREATE INDEX idx_audit_section_id ON registration_audit(section_id);
CREATE INDEX idx_audit_timestamp ON registration_audit(timestamp);
CREATE INDEX idx_audit_action ON registration_audit(action);

-- Seat reservations indexes
CREATE INDEX idx_reservations_section_id ON seat_reservations(section_id);
CREATE INDEX idx_reservations_expires_at ON seat_reservations(expires_at);

-- Composite indexes for complex queries
CREATE INDEX idx_sections_course_semester_active ON sections(course_id, semester_id, is_active);
CREATE INDEX idx_registrations_section_student_status ON registrations(section_id, student_id, status);

-- Functions and triggers for maintaining data consistency

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for automatic timestamp and version updates
CREATE TRIGGER update_students_updated_at BEFORE UPDATE ON students FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_courses_updated_at BEFORE UPDATE ON courses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sections_updated_at BEFORE UPDATE ON sections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_registrations_updated_at BEFORE UPDATE ON registrations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_waitlist_updated_at BEFORE UPDATE ON waitlist FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to maintain waitlist positions
CREATE OR REPLACE FUNCTION reorder_waitlist()
RETURNS TRIGGER AS $$
BEGIN
    -- Reorder positions when a waitlist entry is removed
    IF TG_OP = 'DELETE' THEN
        UPDATE waitlist 
        SET position = position - 1
        WHERE section_id = OLD.section_id 
        AND position > OLD.position;
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER reorder_waitlist_trigger AFTER DELETE ON waitlist FOR EACH ROW EXECUTE FUNCTION reorder_waitlist();

-- Function to automatically clean up expired reservations
CREATE OR REPLACE FUNCTION cleanup_expired_reservations()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM seat_reservations WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Views for common queries

-- View for section availability with course information
CREATE VIEW section_availability AS
SELECT 
    s.section_id,
    s.section_number,
    c.course_code,
    c.course_name,
    sem.semester_name,
    s.total_seats,
    s.available_seats,
    s.reserved_seats,
    (s.total_seats - s.available_seats - s.reserved_seats) as enrolled_count,
    COALESCE(w.waitlist_count, 0) as waitlist_count,
    s.instructor_name,
    s.meeting_times,
    s.location
FROM sections s
JOIN courses c ON s.course_id = c.course_id
JOIN semesters sem ON s.semester_id = sem.semester_id
LEFT JOIN (
    SELECT section_id, COUNT(*) as waitlist_count
    FROM waitlist
    GROUP BY section_id
) w ON s.section_id = w.section_id
WHERE s.is_active = true AND c.is_active = true;

-- View for student registration summary
CREATE VIEW student_registrations AS
SELECT 
    r.student_id,
    s.first_name,
    s.last_name,
    s.student_number,
    COUNT(CASE WHEN r.status = 'enrolled' THEN 1 END) as enrolled_courses,
    COUNT(CASE WHEN r.status = 'waitlisted' THEN 1 END) as waitlisted_courses,
    SUM(CASE WHEN r.status = 'enrolled' THEN c.credits ELSE 0 END) as total_credits
FROM students s
LEFT JOIN registrations r ON s.student_id = r.student_id
LEFT JOIN sections sec ON r.section_id = sec.section_id
LEFT JOIN courses c ON sec.course_id = c.course_id
GROUP BY r.student_id, s.first_name, s.last_name, s.student_number;

-- Sample data inserts for testing

-- Insert sample semester
INSERT INTO semesters (semester_code, semester_name, start_date, end_date, registration_start, registration_end)
VALUES ('FALL2024', 'Fall 2024', '2024-08-26', '2024-12-15', '2024-03-01 08:00:00-05', '2024-08-15 23:59:59-05');

-- Insert sample courses
INSERT INTO courses (course_code, course_name, description, credits, department)
VALUES 
    ('CS101', 'Introduction to Computer Science', 'Fundamentals of programming and problem solving', 3, 'Computer Science'),
    ('CS201', 'Data Structures', 'Advanced data structures and algorithms', 4, 'Computer Science'),
    ('MATH101', 'Calculus I', 'Differential and integral calculus', 4, 'Mathematics'),
    ('ENG101', 'English Composition', 'Academic writing and critical thinking', 3, 'English');

-- Insert sample students
INSERT INTO students (student_number, first_name, last_name, email, academic_level, major)
VALUES 
    ('S001001', 'John', 'Doe', 'john.doe@university.edu', 'undergraduate', 'Computer Science'),
    ('S001002', 'Jane', 'Smith', 'jane.smith@university.edu', 'undergraduate', 'Computer Science'),
    ('S001003', 'Bob', 'Johnson', 'bob.johnson@university.edu', 'undergraduate', 'Mathematics'),
    ('S001004', 'Alice', 'Brown', 'alice.brown@university.edu', 'graduate', 'Computer Science');

-- Insert sample sections
INSERT INTO sections (course_id, semester_id, section_number, instructor_name, total_seats, available_seats, meeting_times, location)
SELECT 
    c.course_id,
    s.semester_id,
    '001',
    CASE c.course_code
        WHEN 'CS101' THEN 'Dr. Smith'
        WHEN 'CS201' THEN 'Dr. Johnson'
        WHEN 'MATH101' THEN 'Dr. Williams'
        WHEN 'ENG101' THEN 'Prof. Davis'
    END,
    CASE c.course_code
        WHEN 'CS101' THEN 30
        WHEN 'CS201' THEN 25
        WHEN 'MATH101' THEN 35
        WHEN 'ENG101' THEN 20
    END,
    CASE c.course_code
        WHEN 'CS101' THEN 30
        WHEN 'CS201' THEN 25
        WHEN 'MATH101' THEN 35
        WHEN 'ENG101' THEN 20
    END,
    '{"days": ["Monday", "Wednesday", "Friday"], "time": "10:00-11:00"}'::jsonb,
    CASE c.course_code
        WHEN 'CS101' THEN 'Room 101'
        WHEN 'CS201' THEN 'Room 102'
        WHEN 'MATH101' THEN 'Room 201'
        WHEN 'ENG101' THEN 'Room 301'
    END
FROM courses c
CROSS JOIN semesters s
WHERE s.semester_code = 'FALL2024';
