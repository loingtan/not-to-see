#!/usr/bin/env python3
"""
Registration Database Reset Script for Course Registration System

This script resets the registration database to its original state for load testing:
1. Removes all registrations from the registrations table
2. Removes all waitlist entries from the waitlist table
3. Resets all section available_seats to match total_seats
4. Clears Redis cache to ensure consistency
5. Clears Redis queues to prevent stale jobs
6. Generates a reset summary report

Usage:
    python reset_registration_database.py
    
Prerequisites:
    - PostgreSQL database running (docker-compose up postgres)
    - Redis running (docker-compose up redis-master)
    - psycopg2-binary, redis packages installed
"""

import psycopg2
import psycopg2.extras
import redis
import sys
from datetime import datetime
from typing import Dict, Any
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class RegistrationDatabaseReset:
    def __init__(self):
        # Database configuration
        self.db_config = {
            'host': 'localhost',
            'port': 5432,
            'database': 'course_registration',
            'user': 'postgres',
            'password': 'password123'
        }

        # Redis configuration
        self.redis_config = {
            'host': 'localhost',
            'port': 6379,
            'db': 0,
            'decode_responses': True
        }

        self.db_conn = None
        self.db_cursor = None
        self.redis_client = None

        # Statistics for reporting
        self.reset_stats = {
            'registrations_deleted': 0,
            'waitlist_deleted': 0,
            'sections_reset': 0,
            'redis_keys_cleared': 0,
            'redis_queues_cleared': 0,
            'total_students': 0,
            'total_courses': 0,
            'total_sections': 0,
            'total_available_seats': 0
        }

    def connect_to_database(self):
        """Establish database connection with error handling"""
        try:
            self.db_conn = psycopg2.connect(**self.db_config)
            self.db_cursor = self.db_conn.cursor(
                cursor_factory=psycopg2.extras.RealDictCursor)
            logger.info("✅ Successfully connected to PostgreSQL database")
            return True
        except psycopg2.Error as e:
            logger.error(f"❌ Failed to connect to database: {e}")
            return False

    def connect_to_redis(self):
        """Establish Redis connection with error handling"""
        try:
            self.redis_client = redis.Redis(**self.redis_config)
            # Test the connection
            self.redis_client.ping()
            logger.info("✅ Successfully connected to Redis")
            return True
        except redis.RedisError as e:
            logger.error(f"❌ Failed to connect to Redis: {e}")
            return False

    def get_current_state(self):
        """Get current database state before reset"""
        logger.info("📊 Analyzing current database state...")

        try:
            # Count registrations
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM registrations;")
            current_registrations = self.db_cursor.fetchone()['count']

            # Count waitlist entries
            self.db_cursor.execute("SELECT COUNT(*) as count FROM waitlist;")
            current_waitlist = self.db_cursor.fetchone()['count']

            # Count students
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM students WHERE enrollment_status = 'active';")
            self.reset_stats['total_students'] = self.db_cursor.fetchone()[
                'count']

            # Count courses
            self.db_cursor.execute("SELECT COUNT(*) as count FROM courses;")
            self.reset_stats['total_courses'] = self.db_cursor.fetchone()[
                'count']

            # Count sections
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM sections WHERE is_active = true;")
            self.reset_stats['total_sections'] = self.db_cursor.fetchone()[
                'count']

            # Get sections with reduced availability
            self.db_cursor.execute("""
                SELECT COUNT(*) as sections_with_registrations
                FROM sections 
                WHERE available_seats < total_seats AND is_active = true;
            """)
            sections_with_registrations = self.db_cursor.fetchone()[
                'sections_with_registrations']

            logger.info(f"📈 Current state:")
            logger.info(
                f"   - Active students: {self.reset_stats['total_students']:,}")
            logger.info(
                f"   - Total courses: {self.reset_stats['total_courses']:,}")
            logger.info(
                f"   - Active sections: {self.reset_stats['total_sections']:,}")
            logger.info(
                f"   - Current registrations: {current_registrations:,}")
            logger.info(f"   - Current waitlist entries: {current_waitlist:,}")
            logger.info(
                f"   - Sections with reduced availability: {sections_with_registrations:,}")

            return True

        except psycopg2.Error as e:
            logger.error(f"❌ Error analyzing current state: {e}")
            return False

    def clear_registrations(self):
        """Remove all registrations from the database"""
        logger.info("🧹 Clearing all registrations...")

        try:
            # Count before deletion
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM registrations;")
            before_count = self.db_cursor.fetchone()['count']

            # Delete all registrations
            self.db_cursor.execute("DELETE FROM registrations;")
            self.reset_stats['registrations_deleted'] = before_count

            # Verify deletion
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM registrations;")
            after_count = self.db_cursor.fetchone()['count']

            if after_count == 0:
                logger.info(
                    f"✅ Successfully deleted {before_count:,} registrations")
                return True
            else:
                logger.error(
                    f"❌ Failed to delete all registrations. {after_count} remain")
                return False

        except psycopg2.Error as e:
            logger.error(f"❌ Error clearing registrations: {e}")
            self.db_conn.rollback()
            return False

    def clear_waitlist(self):
        """Remove all waitlist entries from the database"""
        logger.info("🧹 Clearing all waitlist entries...")

        try:
            # Count before deletion
            self.db_cursor.execute("SELECT COUNT(*) as count FROM waitlist;")
            before_count = self.db_cursor.fetchone()['count']

            # Delete all waitlist entries
            self.db_cursor.execute("DELETE FROM waitlist;")
            self.reset_stats['waitlist_deleted'] = before_count

            # Verify deletion
            self.db_cursor.execute("SELECT COUNT(*) as count FROM waitlist;")
            after_count = self.db_cursor.fetchone()['count']

            if after_count == 0:
                logger.info(
                    f"✅ Successfully deleted {before_count:,} waitlist entries")
                return True
            else:
                logger.error(
                    f"❌ Failed to delete all waitlist entries. {after_count} remain")
                return False

        except psycopg2.Error as e:
            logger.error(f"❌ Error clearing waitlist: {e}")
            self.db_conn.rollback()
            return False

    def reset_section_availability(self):
        """Reset all section available_seats to match total_seats"""
        logger.info("🔄 Resetting section seat availability...")

        try:
            # Get sections that need to be reset
            self.db_cursor.execute("""
                SELECT section_id, total_seats, available_seats,
                       (total_seats - available_seats) as occupied_seats
                FROM sections 
                WHERE available_seats < total_seats AND is_active = true;
            """)
            sections_to_reset = self.db_cursor.fetchall()

            if not sections_to_reset:
                logger.info("✅ All sections already have full availability")
                return True

            logger.info(f"📝 Found {len(sections_to_reset)} sections to reset")

            # Reset all sections to full availability
            self.db_cursor.execute("""
                UPDATE sections 
                SET available_seats = total_seats,
                    updated_at = NOW()
                WHERE is_active = true;
            """)

            self.reset_stats['sections_reset'] = self.db_cursor.rowcount

            # Calculate total available seats after reset
            self.db_cursor.execute("""
                SELECT SUM(total_seats) as total_seats
                FROM sections 
                WHERE is_active = true;
            """)
            self.reset_stats['total_available_seats'] = self.db_cursor.fetchone()[
                'total_seats']

            # Verify reset
            self.db_cursor.execute("""
                SELECT COUNT(*) as count
                FROM sections 
                WHERE available_seats < total_seats AND is_active = true;
            """)
            remaining_incomplete = self.db_cursor.fetchone()['count']

            if remaining_incomplete == 0:
                logger.info(
                    f"✅ Successfully reset {self.reset_stats['sections_reset']:,} sections to full availability")
                logger.info(
                    f"📊 Total available seats: {self.reset_stats['total_available_seats']:,}")
                return True
            else:
                logger.error(
                    f"❌ Failed to reset all sections. {remaining_incomplete} sections still have reduced availability")
                return False

        except psycopg2.Error as e:
            logger.error(f"❌ Error resetting section availability: {e}")
            self.db_conn.rollback()
            return False

    def clear_redis_cache(self):
        """Clear Redis cache to ensure consistency"""
        logger.info("🗑️ Clearing Redis cache...")

        try:
            # Patterns of keys to clear
            cache_patterns = [
                'section:seats:*',      # Section seat counts
                'student:registrations:*',  # Student registration cache
                'student:waitlist:*',   # Student waitlist cache
                'sections:available:*',  # Available sections cache
                'section:details:*',    # Section details cache
                'course:details:*',     # Course details cache
                'idempotency:*'         # Idempotency keys
            ]

            total_cleared = 0

            for pattern in cache_patterns:
                keys = self.redis_client.keys(pattern)
                if keys:
                    deleted = self.redis_client.delete(*keys)
                    total_cleared += deleted
                    logger.info(
                        f"   Cleared {deleted:,} keys matching pattern: {pattern}")

            self.reset_stats['redis_keys_cleared'] = total_cleared

            if total_cleared > 0:
                logger.info(
                    f"✅ Successfully cleared {total_cleared:,} Redis cache keys")
            else:
                logger.info("✅ Redis cache was already clean")

            return True

        except redis.RedisError as e:
            logger.error(f"❌ Error clearing Redis cache: {e}")
            return False

    def clear_redis_queues(self):
        """Clear Redis queues to prevent stale job processing"""
        logger.info("🗑️ Clearing Redis queues...")

        try:
            # Queue names to clear
            queues = [
                'queue:database_sync',
                'queue:waitlist',
                'queue:waitlist_entry'
            ]

            total_cleared = 0

            for queue in queues:
                queue_length = self.redis_client.llen(queue)
                if queue_length > 0:
                    self.redis_client.delete(queue)
                    total_cleared += queue_length
                    logger.info(
                        f"   Cleared queue {queue}: {queue_length:,} jobs")
                else:
                    logger.info(f"   Queue {queue}: already empty")

            self.reset_stats['redis_queues_cleared'] = total_cleared

            if total_cleared > 0:
                logger.info(
                    f"✅ Successfully cleared {total_cleared:,} queued jobs")
            else:
                logger.info("✅ All Redis queues were already empty")

            return True

        except redis.RedisError as e:
            logger.error(f"❌ Error clearing Redis queues: {e}")
            return False

    def commit_changes(self):
        """Commit all database changes"""
        try:
            self.db_conn.commit()
            logger.info("✅ All database changes committed successfully")
            return True
        except psycopg2.Error as e:
            logger.error(f"❌ Failed to commit changes: {e}")
            self.db_conn.rollback()
            return False

    def verify_reset(self):
        """Verify that the reset was successful"""
        logger.info("🔍 Verifying reset completion...")

        try:
            # Check registrations
            self.db_cursor.execute(
                "SELECT COUNT(*) as count FROM registrations;")
            reg_count = self.db_cursor.fetchone()['count']

            # Check waitlist
            self.db_cursor.execute("SELECT COUNT(*) as count FROM waitlist;")
            waitlist_count = self.db_cursor.fetchone()['count']

            # Check section availability
            self.db_cursor.execute("""
                SELECT COUNT(*) as incomplete_sections
                FROM sections 
                WHERE available_seats < total_seats AND is_active = true;
            """)
            incomplete_sections = self.db_cursor.fetchone()[
                'incomplete_sections']

            # Check Redis queue status
            total_queue_items = 0
            for queue in ['queue:database_sync', 'queue:waitlist', 'queue:waitlist_entry']:
                total_queue_items += self.redis_client.llen(queue)

            # Verify complete reset
            if reg_count == 0 and waitlist_count == 0 and incomplete_sections == 0 and total_queue_items == 0:
                logger.info("✅ Reset verification successful!")
                logger.info("   - Registrations: 0")
                logger.info("   - Waitlist entries: 0")
                logger.info("   - Sections with reduced availability: 0")
                logger.info("   - Pending queue jobs: 0")
                return True
            else:
                logger.error("❌ Reset verification failed!")
                logger.error(f"   - Registrations remaining: {reg_count}")
                logger.error(
                    f"   - Waitlist entries remaining: {waitlist_count}")
                logger.error(
                    f"   - Incomplete sections: {incomplete_sections}")
                logger.error(f"   - Pending queue jobs: {total_queue_items}")
                return False

        except (psycopg2.Error, redis.RedisError) as e:
            logger.error(f"❌ Error during verification: {e}")
            return False

    def generate_reset_report(self):
        """Generate and display a comprehensive reset summary"""
        logger.info("📋 Generating reset summary report...")

        print("\n" + "="*80)
        print("🎯 COURSE REGISTRATION DATABASE RESET SUMMARY")
        print("="*80)
        print(
            f"Reset completed at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print()

        print("📊 SYSTEM STATISTICS:")
        print(
            f"   Total Students (Active):    {self.reset_stats['total_students']:,}")
        print(
            f"   Total Courses:              {self.reset_stats['total_courses']:,}")
        print(
            f"   Total Sections (Active):    {self.reset_stats['total_sections']:,}")
        print(
            f"   Total Available Seats:      {self.reset_stats['total_available_seats']:,}")
        print()

        print("🧹 RESET OPERATIONS:")
        print(
            f"   Registrations Deleted:      {self.reset_stats['registrations_deleted']:,}")
        print(
            f"   Waitlist Entries Deleted:   {self.reset_stats['waitlist_deleted']:,}")
        print(
            f"   Sections Reset:             {self.reset_stats['sections_reset']:,}")
        print(
            f"   Redis Cache Keys Cleared:   {self.reset_stats['redis_keys_cleared']:,}")
        print(
            f"   Redis Queue Jobs Cleared:   {self.reset_stats['redis_queues_cleared']:,}")
        print()

        print("✅ SYSTEM STATUS:")
        print("   - All registrations removed")
        print("   - All waitlist entries cleared")
        print("   - All sections restored to full availability")
        print("   - Redis cache completely cleared")
        print("   - All pending queue jobs removed")
        print("   - Database ready for fresh load testing")
        print()

        print("🚀 READY FOR LOAD TESTING!")
        print("   You can now run your performance tests with a clean slate.")
        print("   All students can register for any course without conflicts.")
        print("="*80)

    def close_connections(self):
        """Close database and Redis connections"""
        try:
            if self.db_cursor:
                self.db_cursor.close()
            if self.db_conn:
                self.db_conn.close()
            if self.redis_client:
                self.redis_client.close()
            logger.info("✅ All connections closed successfully")
        except Exception as e:
            logger.error(f"❌ Error closing connections: {e}")

    def run_complete_reset(self):
        """Execute the complete reset process"""
        logger.info("🚀 Starting Course Registration Database Reset...")
        print("\n" + "="*80)
        print("🔄 COURSE REGISTRATION DATABASE RESET")
        print("="*80)

        success = True

        # Step 1: Connect to services
        if not self.connect_to_database():
            return False
        if not self.connect_to_redis():
            return False

        # Step 2: Analyze current state
        if not self.get_current_state():
            return False

        # Step 3: Perform reset operations
        operations = [
            ("Clear Registrations", self.clear_registrations),
            ("Clear Waitlist", self.clear_waitlist),
            ("Reset Section Availability", self.reset_section_availability),
            ("Clear Redis Cache", self.clear_redis_cache),
            ("Clear Redis Queues", self.clear_redis_queues),
            ("Commit Changes", self.commit_changes),
            ("Verify Reset", self.verify_reset)
        ]

        for operation_name, operation_func in operations:
            logger.info(f"🔄 {operation_name}...")
            if not operation_func():
                logger.error(f"❌ {operation_name} failed!")
                success = False
                break
            logger.info(f"✅ {operation_name} completed")

        if success:
            self.generate_reset_report()
            logger.info("🎉 Database reset completed successfully!")
        else:
            logger.error("❌ Database reset failed!")

        return success


def main():
    """Main function to run the reset script"""
    print("🔄 Course Registration Database Reset Tool")
    print("This script will reset the database to its original state for load testing.")
    print()

    # Confirm reset operation
    confirm = input(
        "⚠️  This will DELETE ALL registrations and waitlist data. Continue? (yes/no): ").strip().lower()
    if confirm not in ['yes', 'y']:
        print("❌ Reset cancelled by user.")
        return False

    # Initialize and run reset
    reset_tool = RegistrationDatabaseReset()

    try:
        success = reset_tool.run_complete_reset()
        return success
    except KeyboardInterrupt:
        logger.error("❌ Reset interrupted by user")
        return False
    except Exception as e:
        logger.error(f"❌ Unexpected error during reset: {e}")
        return False
    finally:
        reset_tool.close_connections()


if __name__ == "__main__":
    exit_code = 0 if main() else 1
    sys.exit(exit_code)
