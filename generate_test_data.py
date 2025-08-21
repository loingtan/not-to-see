#!/usr/bin/env python3
"""
Test Data Generator for Course Registration System

Generates realistic test data based on system assumptions:
- 300+ courses per semester (30 slots each)
- 8000+ concurrent students  
- Each student registers for at least 3 subjects
"""

import json
import csv
import random
import string
from datetime import datetime, timedelta
from typing import List, Dict, Any
import uuid

class CourseRegistrationDataGenerator:
    def __init__(self):
        self.courses = []
        self.students = []
        self.registrations = []
        
        # Configuration based on system assumptions
        self.num_courses = 320  # Over 300 courses
        self.slots_per_course = 30
        self.num_students = 8200  # Over 8000 students
        self.min_courses_per_student = 3
        self.max_courses_per_student = 6
        
        # Sample data for realistic generation
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
            "Statistics": ["STAT", "STA"],
            "Philosophy": ["PHIL", "PHI"]
        }
        
        self.course_types = ["Lecture", "Lab", "Seminar", "Workshop", "Tutorial"]
        self.time_slots = [
            "08:00-09:30", "09:30-11:00", "11:00-12:30", 
            "13:30-15:00", "15:00-16:30", "16:30-18:00"
        ]
        self.days = ["MWF", "TTH", "MW", "WF", "TH"]
        
        # Student data
        self.first_names = [
            "John", "Jane", "Michael", "Sarah", "David", "Emily", "Robert", "Jessica",
            "William", "Ashley", "James", "Amanda", "Christopher", "Stephanie", "Daniel",
            "Jennifer", "Matthew", "Elizabeth", "Anthony", "Lauren", "Mark", "Megan",
            "Steven", "Nicole", "Paul", "Samantha", "Andrew", "Rachel", "Joshua", "Amy"
        ]
        
        self.last_names = [
            "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
            "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
            "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
            "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson"
        ]

    def generate_course_code(self, department: str, course_number: int) -> str:
        """Generate a realistic course code"""
        prefixes = self.course_prefixes.get(department, ["GEN"])
        prefix = random.choice(prefixes)
        return f"{prefix}{course_number:03d}"

    def generate_courses(self) -> None:
        """Generate course data"""
        print(f"Generating {self.num_courses} courses...")
        
        for i in range(self.num_courses):
            department = random.choice(self.departments)
            course_number = random.randint(100, 599)
            course_code = self.generate_course_code(department, course_number)
            
            # Generate course name
            topics = [
                "Introduction to", "Advanced", "Fundamentals of", "Applied", 
                "Theory of", "Practical", "Modern", "Classical", "Contemporary"
            ]
            subjects = [
                "Programming", "Algorithms", "Data Structures", "Database Systems",
                "Machine Learning", "Statistics", "Calculus", "Linear Algebra",
                "Research Methods", "Analysis", "Design", "Systems", "Networks"
            ]
            
            course_name = f"{random.choice(topics)} {random.choice(subjects)}"
            if random.random() < 0.3:  # 30% chance of having a roman numeral
                course_name += f" {random.choice(['I', 'II', 'III'])}"
            
            course = {
                "course_id": str(uuid.uuid4()),
                "course_code": course_code,
                "course_name": course_name,
                "department": department,
                "credits": random.choice([3, 4, 6]),
                "total_slots": self.slots_per_course,
                "available_slots": self.slots_per_course,
                "instructor": f"Prof. {random.choice(self.first_names)} {random.choice(self.last_names)}",
                "course_type": random.choice(self.course_types),
                "schedule": {
                    "days": random.choice(self.days),
                    "time": random.choice(self.time_slots),
                    "room": f"{random.choice(['A', 'B', 'C', 'D'])}{random.randint(100, 999)}"
                },
                "prerequisites": [] if random.random() < 0.6 else [f"{random.choice(list(self.course_prefixes.keys()))}{random.randint(100, 299):03d}"],
                "semester": "Fall 2024"
            }
            
            self.courses.append(course)

    def generate_students(self) -> None:
        """Generate student data"""
        print(f"Generating {self.num_students} students...")
        
        for i in range(self.num_students):
            student_id = f"STU{i+1:06d}"
            first_name = random.choice(self.first_names)
            last_name = random.choice(self.last_names)
            
            student = {
                "student_id": student_id,
                "first_name": first_name,
                "last_name": last_name,
                "email": f"{first_name.lower()}.{last_name.lower()}{random.randint(1, 999):03d}@university.edu",
                "year": random.choice(["Freshman", "Sophomore", "Junior", "Senior"]),
                "major": random.choice(self.departments),
                "gpa": round(random.uniform(2.0, 4.0), 2),
                "registration_date": (datetime.now() - timedelta(days=random.randint(1, 30))).isoformat(),
                "status": random.choice(["Active", "Active", "Active", "Hold"]),  # 75% active
                "phone": f"+1-{random.randint(200, 999)}-{random.randint(200, 999)}-{random.randint(1000, 9999)}"
            }
            
            self.students.append(student)

    def generate_empty_registrations_table(self) -> None:
        """Create empty registrations structure for load testing"""
        print("Preparing empty registrations table for load testing...")
        
        # Create empty registrations list - will be populated during load testing
        self.registrations = []
        
        # Reset all course slots to full capacity for load testing
        for course in self.courses:
            course["available_slots"] = course["total_slots"]

    def export_to_json(self) -> None:
        """Export all data to JSON files"""
        print("Exporting data to JSON files...")
        
        data = {
            "courses": self.courses,
            "students": self.students,
            "registrations": self.registrations,
            "metadata": {
                "generated_at": datetime.now().isoformat(),
                "total_courses": len(self.courses),
                "total_students": len(self.students),
                "total_registrations": len(self.registrations)
            }
        }
        
        with open("course_registration_data.json", "w") as f:
            json.dump(data, f, indent=2)
        
        # Export individual collections
        with open("courses.json", "w") as f:
            json.dump(self.courses, f, indent=2)
        
        with open("students.json", "w") as f:
            json.dump(self.students, f, indent=2)
        
        with open("registrations.json", "w") as f:
            json.dump(self.registrations, f, indent=2)

    def export_to_csv(self) -> None:
        """Export data to CSV files"""
        print("Exporting data to CSV files...")
        
        # Export courses
        with open("courses.csv", "w", newline="") as f:
            if self.courses:
                writer = csv.DictWriter(f, fieldnames=[
                    "course_id", "course_code", "course_name", "department", "credits",
                    "total_slots", "available_slots", "instructor", "course_type",
                    "schedule_days", "schedule_time", "schedule_room", "semester"
                ])
                writer.writeheader()
                for course in self.courses:
                    row = course.copy()
                    row["schedule_days"] = course["schedule"]["days"]
                    row["schedule_time"] = course["schedule"]["time"]
                    row["schedule_room"] = course["schedule"]["room"]
                    del row["schedule"]
                    del row["prerequisites"]
                    writer.writerow(row)
        
        # Export students
        with open("students.csv", "w", newline="") as f:
            if self.students:
                writer = csv.DictWriter(f, fieldnames=list(self.students[0].keys()))
                writer.writeheader()
                writer.writerows(self.students)
        
        # Export registrations
        with open("registrations.csv", "w", newline="") as f:
            if self.registrations:
                writer = csv.DictWriter(f, fieldnames=list(self.registrations[0].keys()))
                writer.writeheader()
                writer.writerows(self.registrations)

    def export_to_sql(self) -> None:
        """Export data as SQL INSERT statements"""
        print("Exporting data to SQL file...")
        
        with open("course_registration_data.sql", "w") as f:
            f.write("-- Course Registration System Test Data\n")
            f.write(f"-- Generated on {datetime.now().isoformat()}\n\n")
            
            # Create tables
            f.write("""-- Create tables
CREATE TABLE IF NOT EXISTS courses (
    course_id VARCHAR(36) PRIMARY KEY,
    course_code VARCHAR(10) NOT NULL,
    course_name VARCHAR(255) NOT NULL,
    department VARCHAR(100) NOT NULL,
    credits INT NOT NULL,
    total_slots INT NOT NULL,
    available_slots INT NOT NULL,
    instructor VARCHAR(255) NOT NULL,
    course_type VARCHAR(50) NOT NULL,
    schedule_days VARCHAR(10),
    schedule_time VARCHAR(20),
    schedule_room VARCHAR(20),
    semester VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS students (
    student_id VARCHAR(10) PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    year VARCHAR(20) NOT NULL,
    major VARCHAR(100) NOT NULL,
    gpa DECIMAL(3,2),
    registration_date DATETIME,
    status VARCHAR(20) NOT NULL,
    phone VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS registrations (
    registration_id VARCHAR(12) PRIMARY KEY,
    student_id VARCHAR(10) NOT NULL,
    course_id VARCHAR(36) NOT NULL,
    course_code VARCHAR(10) NOT NULL,
    registration_timestamp DATETIME NOT NULL,
    status VARCHAR(20) NOT NULL,
    grade VARCHAR(2),
    semester VARCHAR(20) NOT NULL,
    FOREIGN KEY (student_id) REFERENCES students(student_id),
    FOREIGN KEY (course_id) REFERENCES courses(course_id)
);

-- Insert data
""")
            
            # Insert courses
            f.write("INSERT INTO courses VALUES\n")
            for i, course in enumerate(self.courses):
                schedule = course["schedule"]
                values = (
                    course["course_id"], course["course_code"], course["course_name"],
                    course["department"], course["credits"], course["total_slots"],
                    course["available_slots"], course["instructor"], course["course_type"],
                    schedule["days"], schedule["time"], schedule["room"], course["semester"]
                )
                f.write(f"({', '.join(repr(v) for v in values)})")
                f.write(",\n" if i < len(self.courses) - 1 else ";\n\n")
            
            # Insert students
            f.write("INSERT INTO students VALUES\n")
            for i, student in enumerate(self.students):
                values = tuple(student.values())
                f.write(f"({', '.join(repr(v) for v in values)})")
                f.write(",\n" if i < len(self.students) - 1 else ";\n\n")
            
            # Insert registrations (empty for load testing)
            if self.registrations:
                f.write("INSERT INTO registrations VALUES\n")
                for i, registration in enumerate(self.registrations):
                    values = tuple(registration.values())
                    f.write(f"({', '.join(repr(v) for v in values)})")
                    f.write(",\n" if i < len(self.registrations) - 1 else ";\n\n")
            else:
                f.write("-- No registrations data - empty table ready for load testing\n\n")

    def generate_summary_report(self) -> None:
        """Generate a summary report of the generated data"""
        print("Generating summary report...")
        
        # Calculate statistics
        enrolled_registrations = len([r for r in self.registrations if r["status"] == "Enrolled"])
        waitlisted_registrations = len([r for r in self.registrations if r["status"] == "Waitlisted"])
        
        courses_by_dept = {}
        for course in self.courses:
            dept = course["department"]
            courses_by_dept[dept] = courses_by_dept.get(dept, 0) + 1
        
        students_by_year = {}
        for student in self.students:
            year = student["year"]
            students_by_year[year] = students_by_year.get(year, 0) + 1
        
        avg_registrations_per_student = 0 if len(self.students) == 0 else len(self.registrations) / len(self.students)
        
        total_slots = sum(c["total_slots"] for c in self.courses)
        remaining_slots = sum(c["available_slots"] for c in self.courses)
        utilization_rate = ((total_slots - remaining_slots) / total_slots * 100) if total_slots > 0 else 0
        
        enrolled_pct = (enrolled_registrations / len(self.registrations) * 100) if len(self.registrations) > 0 else 0
        waitlisted_pct = (waitlisted_registrations / len(self.registrations) * 100) if len(self.registrations) > 0 else 0
        
        report = f"""
Course Registration System - Test Data Summary Report
Generated on: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}

SYSTEM ASSUMPTIONS VERIFICATION:
✓ Over 300 courses: {len(self.courses)} courses generated
✓ 30 slots per course: All courses have 30 slots
✓ Over 8000 students: {len(self.students)} students generated  
✓ At least 3 subjects per student: Will be enforced during load testing

DETAILED STATISTICS:
Courses:
- Total courses: {len(self.courses)}
- Total available slots: {total_slots:,}
- Remaining slots: {remaining_slots:,}
- Utilization rate: {utilization_rate:.1f}%

Students:
- Total students: {len(self.students):,}
- Active students: {len([s for s in self.students if s["status"] == "Active"]):,}

Registrations:
- Total registrations: {len(self.registrations):,} (empty - for load testing)
- Ready for concurrent registration simulation

Courses by Department:
"""
        
        for dept, count in sorted(courses_by_dept.items()):
            report += f"- {dept}: {count} courses\n"
        
        report += f"\nStudents by Year:\n"
        for year, count in sorted(students_by_year.items()):
            report += f"- {year}: {count:,} students\n"
        
        report += f"""
FILES GENERATED:
- course_registration_data.json (complete dataset)
- courses.json, students.json, registrations.json (individual collections)
- courses.csv, students.csv, registrations.csv (CSV format)
- course_registration_data.sql (SQL INSERT statements)
- data_summary_report.txt (this report)
"""
        
        with open("data_summary_report.txt", "w") as f:
            f.write(report)
        
        print(report)

    def generate_all_data(self) -> None:
        """Generate all test data"""
        print("Starting test data generation for Course Registration System...")
        print("=" * 60)
        
        self.generate_courses()
        self.generate_students()
        self.generate_empty_registrations_table()
        
        print("\nExporting data...")
        self.export_to_json()
        self.export_to_csv()
        self.export_to_sql()
        self.generate_summary_report()
        
        print("\n" + "=" * 60)
        print("Test data generation completed successfully!")

if __name__ == "__main__":
    generator = CourseRegistrationDataGenerator()
    generator.generate_all_data()
