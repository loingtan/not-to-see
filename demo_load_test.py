#!/usr/bin/env python3
"""
Demo Load Test for Course Registration System

This is a scaled-down demonstration of the load testing simulator
to show how the "at least 3 subjects per student" requirement works.

Demo scale: 50 students, 20 courses for faster execution
"""

import json
import random
import time
import threading
from datetime import datetime
from typing import List, Dict, Any
from collections import defaultdict
from concurrent.futures import ThreadPoolExecutor, as_completed


class DemoRegistrationLoadTester:
    def __init__(self):
        """Initialize demo with scaled-down data"""
        self.courses = []
        self.students = []
        self.registrations = []
        
        # Demo configuration (reduced scale)
        self.demo_students = 50
        self.demo_courses = 20
        self.min_courses_per_student = 3
        self.max_courses_per_student = 6
        
        # Load original data and sample it
        self.load_demo_data()
        
        # Thread-safe tracking
        self.course_slots = {}
        self.lock = threading.Lock()
        
        # Initialize course slots
        for course in self.courses:
            self.course_slots[course["course_id"]] = course["total_slots"]

    def load_demo_data(self):
        """Load and sample data for demo"""
        try:
            with open("course_registration_data.json", 'r') as f:
                data = json.load(f)
                all_courses = data.get("courses", [])
                all_students = data.get("students", [])
            
            # Sample data for demo
            self.courses = random.sample(all_courses, min(self.demo_courses, len(all_courses)))
            self.students = random.sample(all_students, min(self.demo_students, len(all_students)))
            
            print(f"Demo loaded: {len(self.courses)} courses, {len(self.students)} students")
            
        except FileNotFoundError:
            print("Error: course_registration_data.json not found.")
            print("Please run 'python3 generate_test_data.py' first.")
            raise

    def simulate_registration_request(self, student: Dict, course: Dict) -> Dict[str, Any]:
        """Simulate a single registration request"""
        start_time = time.time()
        
        # Simulate processing time
        processing_delay = random.uniform(0.01, 0.05)  # Faster for demo
        time.sleep(processing_delay)
        
        # Thread-safe slot checking
        with self.lock:
            available_slots = self.course_slots.get(course["course_id"], 0)
            
            if available_slots > 0:
                self.course_slots[course["course_id"]] -= 1
                status = "Enrolled"
            elif random.random() < 0.7:  # 70% chance to waitlist
                status = "Waitlisted"
            else:
                status = "Failed"
        
        response_time = (time.time() - start_time) * 1000
        
        return {
            "registration_id": f"REG{int(time.time() * 1000000)}_{random.randint(1000, 9999)}",
            "student_id": student["student_id"],
            "course_id": course["course_id"],
            "course_code": course["course_code"],
            "status": status,
            "response_time_ms": response_time,
            "timestamp": datetime.now().isoformat()
        }

    def student_registration_session(self, student: Dict) -> List[Dict[str, Any]]:
        """Simulate complete registration session for one student"""
        print(f"Student {student['student_id']} starting registration...")
        
        student_registrations = []
        target_courses = random.randint(self.min_courses_per_student, self.max_courses_per_student)
        
        # Select courses to attempt
        courses_to_try = random.sample(self.courses, min(target_courses, len(self.courses)))
        
        enrolled_count = 0
        
        for course in courses_to_try:
            # Small delay between attempts
            time.sleep(random.uniform(0.1, 0.3))
            
            # Attempt registration
            registration = self.simulate_registration_request(student, course)
            student_registrations.append(registration)
            
            if registration["status"] == "Enrolled":
                enrolled_count += 1
                print(f"  ✓ {student['student_id']} enrolled in {course['course_code']}")
            elif registration["status"] == "Waitlisted":
                print(f"  ~ {student['student_id']} waitlisted for {course['course_code']}")
            else:
                print(f"  ✗ {student['student_id']} failed to register for {course['course_code']}")
        
        min_met = "✓" if enrolled_count >= self.min_courses_per_student else "✗"
        print(f"  {min_met} {student['student_id']} final: {enrolled_count} courses enrolled (min: {self.min_courses_per_student})")
        
        return student_registrations

    def run_demo_load_test(self):
        """Run demo load test"""
        print("COURSE REGISTRATION DEMO LOAD TEST")
        print("=" * 50)
        print(f"Demo Configuration:")
        print(f"- Students: {len(self.students)}")
        print(f"- Courses: {len(self.courses)}")
        print(f"- Min courses per student: {self.min_courses_per_student}")
        print(f"- Each course has 30 slots")
        print("")
        
        start_time = time.time()
        
        # Run concurrent registration simulation
        with ThreadPoolExecutor(max_workers=10) as executor:
            future_to_student = {
                executor.submit(self.student_registration_session, student): student 
                for student in self.students
            }
            
            for future in as_completed(future_to_student):
                student_registrations = future.result()
                self.registrations.extend(student_registrations)
        
        end_time = time.time()
        
        # Generate results
        self.generate_demo_results(end_time - start_time)

    def generate_demo_results(self, duration: float):
        """Generate demo results summary"""
        print("\n" + "=" * 50)
        print("DEMO LOAD TEST RESULTS")
        print("=" * 50)
        
        # Calculate metrics
        total_attempts = len(self.registrations)
        enrolled = len([r for r in self.registrations if r["status"] == "Enrolled"])
        waitlisted = len([r for r in self.registrations if r["status"] == "Waitlisted"])
        failed = len([r for r in self.registrations if r["status"] == "Failed"])
        
        # Count students who met minimum requirement
        student_enrolled_counts = defaultdict(int)
        for reg in self.registrations:
            if reg["status"] == "Enrolled":
                student_enrolled_counts[reg["student_id"]] += 1
        
        students_with_min = sum(1 for count in student_enrolled_counts.values() 
                               if count >= self.min_courses_per_student)
        
        # Course utilization
        course_enrollment = defaultdict(int)
        for reg in self.registrations:
            if reg["status"] == "Enrolled":
                course_enrollment[reg["course_code"]] += 1
        
        print(f"Test Duration: {duration:.2f} seconds")
        print(f"Total Registration Attempts: {total_attempts}")
        print(f"")
        print(f"Registration Results:")
        print(f"- Enrolled: {enrolled} ({enrolled/total_attempts*100:.1f}%)")
        print(f"- Waitlisted: {waitlisted} ({waitlisted/total_attempts*100:.1f}%)")
        print(f"- Failed: {failed} ({failed/total_attempts*100:.1f}%)")
        print(f"")
        print(f"Student Success Metrics:")
        print(f"- Students processed: {len(self.students)}")
        print(f"- Students with ≥3 courses: {students_with_min} ({students_with_min/len(self.students)*100:.1f}%)")
        print(f"- Average courses per student: {total_attempts/len(self.students):.1f}")
        print(f"")
        print(f"Top Course Enrollments:")
        sorted_courses = sorted(course_enrollment.items(), key=lambda x: x[1], reverse=True)
        for course_code, count in sorted_courses[:10]:
            course_info = next(c for c in self.courses if c["course_code"] == course_code)
            utilization = (count / course_info["total_slots"]) * 100
            print(f"- {course_code}: {count}/30 slots ({utilization:.1f}%)")
        
        print(f"")
        print(f"SYSTEM ASSUMPTION VERIFICATION:")
        print(f"✓ Concurrent students: {len(self.students)} students processed simultaneously")
        print(f"✓ Multiple courses: {len(self.courses)} courses available")
        print(f"✓ Minimum 3 subjects: {students_with_min}/{len(self.students)} students achieved requirement")
        
        # Show individual student results
        print(f"\nINDIVIDUAL STUDENT RESULTS (showing first 10):")
        for student_id, enrolled_count in list(student_enrolled_counts.items())[:10]:
            student_name = next(s["first_name"] + " " + s["last_name"] for s in self.students if s["student_id"] == student_id)
            status = "✓ PASSED" if enrolled_count >= self.min_courses_per_student else "✗ FAILED"
            print(f"- {student_id} ({student_name}): {enrolled_count} courses enrolled {status}")


def main():
    """Run demo load test"""
    try:
        tester = DemoRegistrationLoadTester()
        tester.run_demo_load_test()
        
    except Exception as e:
        print(f"Demo failed: {e}")
        raise


if __name__ == "__main__":
    main()
