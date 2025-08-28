-- Migration: 001_initial_schema
-- Description: Create initial course registration system schema
-- Created: 2024-01-01

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create students table
CREATE TABLE IF NOT EXISTS students (
    student_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_number TEXT UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    enrollment_status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Create courses table
CREATE TABLE IF NOT EXISTS courses (
    course_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_code TEXT UNIQUE NOT NULL,
    course_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Create semesters table
CREATE TABLE IF NOT EXISTS semesters (
    semester_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    semester_code TEXT UNIQUE NOT NULL,
    semester_name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    registration_start TIMESTAMP WITH TIME ZONE NOT NULL,
    registration_end TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create sections table
CREATE TABLE IF NOT EXISTS sections (
    section_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id UUID NOT NULL,
    semester_id UUID NOT NULL,
    section_number VARCHAR(10) NOT NULL,
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    available_seats INTEGER NOT NULL DEFAULT 0 CHECK (available_seats >= 0),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    UNIQUE(course_id, semester_id, section_number)
);

-- Add foreign key constraints separately to avoid dependency issues
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'sections_course_id_fkey'
    ) THEN
        ALTER TABLE sections ADD CONSTRAINT sections_course_id_fkey 
        FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'sections_semester_id_fkey'
    ) THEN
        ALTER TABLE sections ADD CONSTRAINT sections_semester_id_fkey 
        FOREIGN KEY (semester_id) REFERENCES semesters(semester_id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create registration status enum
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'registration_status') THEN
        CREATE TYPE registration_status AS ENUM ('enrolled', 'waitlisted', 'dropped', 'failed');
    END IF;
END $$;

-- Create registrations table
CREATE TABLE IF NOT EXISTS registrations (
    registration_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL,
    section_id UUID NOT NULL,
    status registration_status NOT NULL DEFAULT 'enrolled',
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,
    UNIQUE(student_id, section_id)
);

-- Add foreign key constraints for registrations
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'registrations_student_id_fkey'
    ) THEN
        ALTER TABLE registrations ADD CONSTRAINT registrations_student_id_fkey 
        FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'registrations_section_id_fkey'
    ) THEN
        ALTER TABLE registrations ADD CONSTRAINT registrations_section_id_fkey 
        FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create waitlist table
CREATE TABLE IF NOT EXISTS waitlist (
    waitlist_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL,
    section_id UUID NOT NULL,
    position INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_waitlist_student FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    CONSTRAINT fk_waitlist_section FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE,
    CONSTRAINT unique_waitlist_student_section UNIQUE (student_id, section_id)
);

-- Create idempotency_keys table for duplicate request prevention
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    student_id UUID NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response_data TEXT,
    status_code INTEGER NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_idempotency_student FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE
);

-- Add foreign key constraints for waitlist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'waitlist_student_id_fkey'
    ) THEN
        ALTER TABLE waitlist ADD CONSTRAINT waitlist_student_id_fkey 
        FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'waitlist_section_id_fkey'
    ) THEN
        ALTER TABLE waitlist ADD CONSTRAINT waitlist_section_id_fkey 
        FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create indexes for performance optimization (using IF NOT EXISTS for PostgreSQL 9.5+)
DO $$
BEGIN
    -- Students indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_students_student_number') THEN
        CREATE INDEX idx_students_student_number ON students(student_number);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_students_enrollment_status') THEN
        CREATE INDEX idx_students_enrollment_status ON students(enrollment_status);
    END IF;
    
    -- Courses indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_courses_course_code') THEN
        CREATE INDEX idx_courses_course_code ON courses(course_code);
    END IF;
    
    -- Semesters indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_semesters_semester_code') THEN
        CREATE INDEX idx_semesters_semester_code ON semesters(semester_code);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_semesters_is_active') THEN
        CREATE INDEX idx_semesters_is_active ON semesters(is_active);
    END IF;
    
    -- Sections indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_course_id') THEN
        CREATE INDEX idx_sections_course_id ON sections(course_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_semester_id') THEN
        CREATE INDEX idx_sections_semester_id ON sections(semester_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_available_seats') THEN
        CREATE INDEX idx_sections_available_seats ON sections(available_seats);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_is_active') THEN
        CREATE INDEX idx_sections_is_active ON sections(is_active);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_course_semester') THEN
        CREATE INDEX idx_sections_course_semester ON sections(course_id, semester_id);
    END IF;
    
    -- Registrations indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_registrations_student_id') THEN
        CREATE INDEX idx_registrations_student_id ON registrations(student_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_registrations_section_id') THEN
        CREATE INDEX idx_registrations_section_id ON registrations(section_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_registrations_status') THEN
        CREATE INDEX idx_registrations_status ON registrations(status);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_registrations_student_status') THEN
        CREATE INDEX idx_registrations_student_status ON registrations(student_id, status);
    END IF;
    
    -- Waitlist indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_waitlist_section_id') THEN
        CREATE INDEX idx_waitlist_section_id ON waitlist(section_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_waitlist_student_id') THEN
        CREATE INDEX idx_waitlist_student_id ON waitlist(student_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_waitlist_position') THEN
        CREATE INDEX idx_waitlist_position ON waitlist(section_id, position);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_waitlist_timestamp') THEN
        CREATE INDEX idx_waitlist_timestamp ON waitlist(timestamp);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_waitlist_section_position') THEN
        CREATE INDEX idx_waitlist_section_position ON waitlist(section_id, position);
    END IF;
    
    -- Idempotency indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_idempotency_student_id') THEN
        CREATE INDEX idx_idempotency_student_id ON idempotency_keys(student_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_idempotency_expires_at') THEN
        CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);
    END IF;
    
    -- Composite indexes
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_sections_course_semester_active') THEN
        CREATE INDEX idx_sections_course_semester_active ON sections(course_id, semester_id, is_active);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_registrations_section_student_status') THEN
        CREATE INDEX idx_registrations_section_student_status ON registrations(section_id, student_id, status);
    END IF;
END $$;
