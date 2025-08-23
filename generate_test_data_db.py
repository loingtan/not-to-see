#!/usr/bin/env python3
"""
Test Data Generator with Database Insertion for Course Registration System

Generates realistic test data and inserts directly into PostgreSQL database:
- 300+ courses per semester (30 slots each)
- 8000+ students ready for registration
- All sections start with full availability for clean testing
- Registration table remains empty for system testing
"""

import json
import csv
import random
import string
import psycopg2
import psycopg2.extras
from datetime import datetime, timedelta
from typing import List, Dict, Any
import uuid
import sys


class CourseRegistrationDataInserter:
    def __init__(self):
        self.db_config = {
            'host': 'localhost',
            'port': 5432,
            'database': 'course_registration',
            'user': 'postgres',
            'password': 'password123'
        }

        self.conn = None
        self.cursor = None

        # Data containers
        self.courses = []
        self.students = []
        self.semesters = []
        self.sections = []
        self.registrations = []
        self.waitlist = []

        #
        self.num_courses = 320
        self.slots_per_course = 30
        self.num_students = 8200
        self.min_courses_per_student = 3
        self.max_courses_per_student = 6

        self.departments = [
            "Computer Science", "Mathematics", "Physics", "Chemistry",
            "Biology", "Economics", "Psychology", "History", "Literature",
            "Engineering", "Business Administration", "Statistics", "Philosophy"
        ]

        self.course_prefixes = {
            "Computer Science": ["CS", "CSE", "COMP"],
            "Mathematics": ["MATH", "STAT", "CALC"],
            "Physics": ["PHYS", "PHY"],
            "Chemistry": ["CHEM", "CHM"],
            "Biology": ["BIO", "BIOL"],
            "Economics": ["ECON", "ECO"],
            "Psychology": ["PSYC", "PSY"],
            "History": ["HIST", "HIS"],
            "Literature": ["LIT", "ENG"],
            "Engineering": ["ENG", "ENGR"],
            "Business Administration": ["BUS", "MGMT"],
            "Statistics": ["STAT", "DATA"],
            "Philosophy": ["PHIL", "PHI"]
        }

    def connect_to_database(self):
        """Establish connection to PostgreSQL database"""
        try:
            self.conn = psycopg2.connect(**self.db_config)
            self.cursor = self.conn.cursor(
                cursor_factory=psycopg2.extras.RealDictCursor)
            print("✅ Connected to PostgreSQL database")

            self.cursor.execute("SELECT version();")
            version = self.cursor.fetchone()
            print(f"PostgreSQL version: {version['version']}")

        except psycopg2.Error as e:
            print(f"❌ Error connecting to PostgreSQL database: {e}")
            sys.exit(1)

    def disconnect_from_database(self):
        """Close database connection"""
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()
        print("🔌 Disconnected from database")

    def clear_existing_data(self):
        """Clear existing test data from tables"""
        print("🧹 Clearing existing test data...")

        tables = ['waitlist', 'registrations', 'sections',
                  'courses', 'students', 'semesters']

        try:
            for table in tables:
                self.cursor.execute(f"DELETE FROM {table};")
                self.cursor.execute(f"SELECT COUNT(*) as count FROM {table};")
                count = self.cursor.fetchone()['count']
                print(f"   Cleared {table}: {count} rows remaining")

            self.conn.commit()
            print("✅ All test data cleared successfully")

        except psycopg2.Error as e:
            print(f"❌ Error clearing data: {e}")
            self.conn.rollback()
            raise

    def generate_semesters(self):
        """Generate semester data"""
        print("📅 Generating semesters...")

        current_year = datetime.now().year
        semesters_data = [
            {
                'semester_code': f'FALL{current_year}',
                'semester_name': f'Fall {current_year}',
                'start_date': f'{current_year}-08-15',
                'end_date': f'{current_year}-12-15',
                'registration_start': f'{current_year}-07-01 00:00:00+00',
                'registration_end': f'{current_year}-08-10 23:59:59+00',
                'is_active': True
            },
            {
                'semester_code': f'SPRING{current_year + 1}',
                'semester_name': f'Spring {current_year + 1}',
                'start_date': f'{current_year + 1}-01-15',
                'end_date': f'{current_year + 1}-05-15',
                'registration_start': f'{current_year}-11-01 00:00:00+00',
                'registration_end': f'{current_year + 1}-01-10 23:59:59+00',
                'is_active': True
            }
        ]

        for semester in semesters_data:
            semester['semester_id'] = str(uuid.uuid4())
            semester['created_at'] = datetime.now()
            semester['updated_at'] = datetime.now()
            self.semesters.append(semester)

    def insert_semesters(self):
        """Insert semesters into database"""
        print("📅 Inserting semesters into database...")

        insert_query = """
        INSERT INTO semesters (semester_id, semester_code, semester_name, start_date, end_date, 
                              registration_start, registration_end, is_active, created_at, updated_at)
        VALUES (%(semester_id)s, %(semester_code)s, %(semester_name)s, %(start_date)s, %(end_date)s,
                %(registration_start)s, %(registration_end)s, %(is_active)s, %(created_at)s, %(updated_at)s)
        """

        try:
            self.cursor.executemany(insert_query, self.semesters)
            self.conn.commit()
            print(f"✅ Inserted {len(self.semesters)} semesters")
        except psycopg2.Error as e:
            print(f"❌ Error inserting semesters: {e}")
            self.conn.rollback()
            raise

    def generate_courses(self):
        """Generate course data"""
        print("📚 Generating courses...")

        course_number = 100

        for dept in self.departments:
            courses_per_dept = self.num_courses // len(self.departments)
            prefixes = self.course_prefixes[dept]

            for i in range(courses_per_dept):
                prefix = random.choice(prefixes)

                course = {
                    'course_id': str(uuid.uuid4()),
                    'course_code': f"{prefix}{course_number}",
                    'course_name': self.generate_course_name(dept, course_number),
                    'created_at': datetime.now(),
                    'updated_at': datetime.now(),
                    'version': 1
                }

                self.courses.append(course)
                course_number += random.randint(1, 5)

    def insert_courses(self):
        """Insert courses into database"""
        print("📚 Inserting courses into database...")

        insert_query = """
        INSERT INTO courses (course_id, course_code, course_name, created_at, updated_at, version)
        VALUES (%(course_id)s, %(course_code)s, %(course_name)s, %(created_at)s, %(updated_at)s, %(version)s)
        """

        try:
            self.cursor.executemany(insert_query, self.courses)
            self.conn.commit()
            print(f"✅ Inserted {len(self.courses)} courses")
        except psycopg2.Error as e:
            print(f"❌ Error inserting courses: {e}")
            self.conn.rollback()
            raise

    def generate_course_name(self, department: str, course_number: int) -> str:
        """Generate realistic course names based on department"""
        course_names = {
            "Computer Science": [
                "Introduction to Programming", "Data Structures", "Algorithms", "Database Systems",
                "Software Engineering", "Computer Networks", "Machine Learning", "Artificial Intelligence",
                "Web Development", "Mobile App Development", "Cybersecurity", "Operating Systems"
            ],
            "Mathematics": [
                "Calculus I", "Calculus II", "Linear Algebra", "Differential Equations",
                "Abstract Algebra", "Real Analysis", "Number Theory", "Probability Theory",
                "Statistics", "Discrete Mathematics", "Geometry", "Topology"
            ],
            "Physics": [
                "Classical Mechanics", "Electromagnetic Theory", "Quantum Mechanics", "Thermodynamics",
                "Modern Physics", "Optics", "Relativity Theory", "Particle Physics",
                "Astrophysics", "Solid State Physics", "Nuclear Physics", "Fluid Dynamics"
            ],
            "Chemistry": [
                "General Chemistry", "Organic Chemistry", "Inorganic Chemistry", "Physical Chemistry",
                "Analytical Chemistry", "Biochemistry", "Quantum Chemistry", "Environmental Chemistry",
                "Polymer Chemistry", "Medicinal Chemistry", "Chemical Engineering", "Materials Science"
            ]
        }

        if department in course_names:
            base_name = random.choice(course_names[department])
        else:
            base_name = f"{department} Fundamentals"

        level = "I" if course_number < 200 else "II" if course_number < 300 else "III"

        if random.random() < 0.3:
            return f"{base_name} {level}"
        else:
            return base_name

    def generate_students(self):
        """Generate student data"""
        print("👥 Generating students...")

        first_names = [
            "James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
            "William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
            "Thomas", "Sarah", "Christopher", "Karen", "Charles", "Nancy", "Daniel", "Lisa",
            "Matthew", "Betty", "Anthony", "Helen", "Mark", "Sandra", "Donald", "Donna",
            "Steven", "Carol", "Paul", "Ruth", "Andrew", "Sharon", "Kenneth", "Michelle"
        ]

        last_names = [
            "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
            "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
            "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
            "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young"
        ]

        enrollment_statuses = ["active", "inactive", "graduated", "suspended"]

        for i in range(self.num_students):
            student = {
                'student_id': str(uuid.uuid4()),
                'student_number': f"S{2024000000 + i + 1}",
                'first_name': random.choice(first_names),
                'last_name': random.choice(last_names),
                'enrollment_status': random.choices(
                    enrollment_statuses,
                    weights=[85, 8, 5, 2],
                    k=1
                )[0],
                'created_at': datetime.now(),
                'updated_at': datetime.now(),
                'version': 1
            }

            self.students.append(student)

    def insert_students(self):
        """Insert students into database"""
        print("👥 Inserting students into database...")

        insert_query = """
        INSERT INTO students (student_id, student_number, first_name, last_name, enrollment_status, 
                             created_at, updated_at, version)
        VALUES (%(student_id)s, %(student_number)s, %(first_name)s, %(last_name)s, %(enrollment_status)s,
                %(created_at)s, %(updated_at)s, %(version)s)
        """

        try:
            self.cursor.executemany(insert_query, self.students)
            self.conn.commit()
            print(f"✅ Inserted {len(self.students)} students")
        except psycopg2.Error as e:
            print(f"❌ Error inserting students: {e}")
            self.conn.rollback()
            raise

    def generate_sections(self):
        """Generate section data for each course and semester"""
        print("🏛️ Generating sections...")

        for semester in self.semesters:
            for course in self.courses:

                num_sections = random.randint(1, 3)

                for section_num in range(1, num_sections + 1):
                    section = {
                        'section_id': str(uuid.uuid4()),
                        'course_id': course['course_id'],
                        'semester_id': semester['semester_id'],
                        'section_number': f"{section_num:03d}",
                        'total_seats': self.slots_per_course,
                        'available_seats': self.slots_per_course,
                        'is_active': True,
                        'created_at': datetime.now(),
                        'updated_at': datetime.now(),
                        'version': 1
                    }

                    self.sections.append(section)

    def insert_sections(self):
        """Insert sections into database"""
        print("🏛️ Inserting sections into database...")

        insert_query = """
        INSERT INTO sections (section_id, course_id, semester_id, section_number, total_seats, 
                             available_seats, is_active, created_at, updated_at, version)
        VALUES (%(section_id)s, %(course_id)s, %(semester_id)s, %(section_number)s, %(total_seats)s,
                %(available_seats)s, %(is_active)s, %(created_at)s, %(updated_at)s, %(version)s)
        """

        try:
            self.cursor.executemany(insert_query, self.sections)
            self.conn.commit()
            print(f"✅ Inserted {len(self.sections)} sections")
        except psycopg2.Error as e:
            print(f"❌ Error inserting sections: {e}")
            self.conn.rollback()
            raise

    def skip_initial_registrations(self):
        """Skip generating initial registrations - let the system handle this during testing"""
        print("⏭️  Skipping initial registrations generation")
        print("   Registrations will be created through the registration system during testing")
        self.registrations = []

    def generate_waitlist_entries(self):
        """Generate some waitlist entries"""
        print("⏳ Generating waitlist entries...")

        popular_sections = random.sample(
            self.sections, min(50, len(self.sections)))
        active_students = [
            s for s in self.students if s['enrollment_status'] == 'active']

        for section in popular_sections:
            # Each popular section gets 5-15 waitlist entries
            num_waitlist = random.randint(5, 15)
            selected_students = random.sample(
                active_students, min(num_waitlist, len(active_students)))

            for position, student in enumerate(selected_students, 1):
                waitlist_entry = {
                    'waitlist_id': str(uuid.uuid4()),
                    'student_id': student['student_id'],
                    'section_id': section['section_id'],
                    'position': position,
                    'timestamp': datetime.now() - timedelta(days=random.randint(1, 10)),
                    'expires_at': datetime.now() + timedelta(days=random.randint(7, 30)),
                    'created_at': datetime.now(),
                    'updated_at': datetime.now()
                }

                self.waitlist.append(waitlist_entry)

    def insert_waitlist_entries(self):
        """Insert waitlist entries into database"""
        if not self.waitlist:
            print("⏳ No waitlist entries to insert")
            return

        print("⏳ Inserting waitlist entries into database...")

        insert_query = """
        INSERT INTO waitlist (waitlist_id, student_id, section_id, position, timestamp, expires_at,
                             created_at, updated_at)
        VALUES (%(waitlist_id)s, %(student_id)s, %(section_id)s, %(position)s, %(timestamp)s, %(expires_at)s,
                %(created_at)s, %(updated_at)s)
        """

        try:
            self.cursor.executemany(insert_query, self.waitlist)
            self.conn.commit()
            print(f"✅ Inserted {len(self.waitlist)} waitlist entries")
        except psycopg2.Error as e:
            print(f"❌ Error inserting waitlist entries: {e}")
            self.conn.rollback()
            raise

    def update_section_availability(self):
        """Update section available seats based on registrations"""
        print("🔄 Updating section seat availability...")

        try:
            # Count enrolled students per section
            update_query = """
            UPDATE sections 
            SET available_seats = total_seats - COALESCE(enrolled_count, 0),
                updated_at = NOW()
            FROM (
                SELECT 
                    section_id,
                    COUNT(*) as enrolled_count
                FROM registrations 
                WHERE status = 'enrolled'
                GROUP BY section_id
            ) reg_counts
            WHERE sections.section_id = reg_counts.section_id
            """

            self.cursor.execute(update_query)
            updated_rows = self.cursor.rowcount
            self.conn.commit()

            print(f"✅ Updated seat availability for {updated_rows} sections")

        except psycopg2.Error as e:
            print(f"❌ Error updating section availability: {e}")
            self.conn.rollback()
            raise

    def generate_summary_report(self):
        """Generate and display summary report"""
        print("\n" + "="*60)
        print("📊 DATABASE INSERTION SUMMARY REPORT")
        print("="*60)

        try:
            # Get actual counts from database
            tables_info = []

            tables = ['semesters', 'courses', 'students',
                      'sections', 'registrations', 'waitlist']

            for table in tables:
                self.cursor.execute(f"SELECT COUNT(*) as count FROM {table}")
                count = self.cursor.fetchone()['count']
                tables_info.append((table.capitalize(), count))

            for table_name, count in tables_info:
                print(f"{table_name:15}: {count:,}")

            # Additional statistics
            print("\n📈 DETAILED STATISTICS:")

            # Active vs inactive students
            self.cursor.execute("""
                SELECT enrollment_status, COUNT(*) as count 
                FROM students 
                GROUP BY enrollment_status 
                ORDER BY count DESC
            """)

            print("\nStudent Status Distribution:")
            for row in self.cursor.fetchall():
                print(f"  {row['enrollment_status']:10}: {row['count']:,}")

            # Registration status distribution
            self.cursor.execute("""
                SELECT status, COUNT(*) as count 
                FROM registrations 
                GROUP BY status 
                ORDER BY count DESC
            """)

            if self.cursor.rowcount > 0:
                print("\nRegistration Status Distribution:")
                for row in self.cursor.fetchall():
                    print(f"  {row['status']:10}: {row['count']:,}")

            # Sections with highest enrollment
            self.cursor.execute("""
                SELECT 
                    s.section_number,
                    c.course_code,
                    c.course_name,
                    s.total_seats,
                    s.available_seats,
                    (s.total_seats - s.available_seats) as enrolled_count
                FROM sections s
                JOIN courses c ON s.course_id = c.course_id
                WHERE s.available_seats < s.total_seats
                ORDER BY enrolled_count DESC
                LIMIT 5
            """)

            if self.cursor.rowcount > 0:
                print("\nTop 5 Most Popular Sections:")
                for row in self.cursor.fetchall():
                    enrollment_pct = (
                        (row['total_seats'] - row['available_seats']) / row['total_seats']) * 100
                    print(
                        f"  {row['course_code']}-{row['section_number']}: {row['enrolled_count']}/{row['total_seats']} ({enrollment_pct:.1f}%)")

        except psycopg2.Error as e:
            print(f"❌ Error generating report: {e}")

    def run_complete_generation(self):
        """Run the complete data generation and insertion process"""
        print("🚀 Starting Complete Course Registration Data Generation")
        print("="*70)

        try:
            # Connect to database
            self.connect_to_database()

            # Clear existing data
            self.clear_existing_data()

            # Generate and insert data step by step
            self.generate_semesters()
            self.insert_semesters()

            self.generate_courses()
            self.insert_courses()

            self.generate_students()
            self.insert_students()

            self.generate_sections()
            self.insert_sections()

            # Skip generating registrations - let the system handle this during testing
            self.skip_initial_registrations()

            # Skip generating waitlist entries initially for clean testing
            # self.generate_waitlist_entries()
            # self.insert_waitlist_entries()

            # Skip section availability update since no registrations exist
            # self.update_section_availability()

            # Generate final report
            self.generate_summary_report()

            print("\n" + "="*70)
            print("✅ Course Registration Test Data Generation COMPLETED Successfully!")
            print("🎯 Database is now ready for testing:")
            print("   - All courses and sections created with full availability")
            print("   - Students ready for registration testing")
            print("   - Registration table empty for clean testing")
            print("="*70)

        except Exception as e:
            print(f"\n❌ Error during data generation: {e}")
            if self.conn:
                self.conn.rollback()
            raise
        finally:
            # Always disconnect
            self.disconnect_from_database()


if __name__ == "__main__":
    print("Course Registration System - Database Test Data Generator")
    print("This script will generate and insert test data directly into PostgreSQL")
    print("\nMake sure the database is running (docker-compose up -d postgres)")

    response = input("\nProceed with data generation? (y/N): ").strip().lower()

    if response in ['y', 'yes']:
        generator = CourseRegistrationDataInserter()
        generator.run_complete_generation()
    else:
        print("Data generation cancelled.")
