package database

import (
	"fmt"
	"log"
	"time"

	domain "cobra-template/internal/domain/registration"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config represents database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection creates a new GORM database connection
func NewConnection(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Auto migrate all models
	err := db.AutoMigrate(
		&domain.Student{},
		&domain.Course{},
		&domain.Semester{},
		&domain.Section{},
		&domain.Registration{},
		&domain.WaitlistEntry{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Add custom constraints and indexes
	if err := addConstraintsAndIndexes(db); err != nil {
		return fmt.Errorf("failed to add constraints and indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// addConstraintsAndIndexes adds custom constraints and indexes
func addConstraintsAndIndexes(db *gorm.DB) error {
	// Add unique constraint for student-section registration
	if err := db.Exec(`
		ALTER TABLE registrations 
		ADD CONSTRAINT IF NOT EXISTS unique_student_section 
		UNIQUE (student_id, section_id)
	`).Error; err != nil {
		log.Printf("Warning: failed to add unique constraint: %v", err)
	}

	// Add unique constraint for course-semester-section
	if err := db.Exec(`
		ALTER TABLE sections 
		ADD CONSTRAINT IF NOT EXISTS unique_course_semester_section 
		UNIQUE (course_id, semester_id, section_number)
	`).Error; err != nil {
		log.Printf("Warning: failed to add section unique constraint: %v", err)
	}

	// Add seat consistency check constraint
	if err := db.Exec(`
		ALTER TABLE sections 
		ADD CONSTRAINT IF NOT EXISTS check_seat_consistency 
		CHECK (available_seats + reserved_seats <= total_seats)
	`).Error; err != nil {
		log.Printf("Warning: failed to add seat consistency constraint: %v", err)
	}

	// Add indexes for frequently queried columns
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_registrations_student_id ON registrations(student_id)",
		"CREATE INDEX IF NOT EXISTS idx_registrations_section_id ON registrations(section_id)",
		"CREATE INDEX IF NOT EXISTS idx_registrations_status ON registrations(status)",
		"CREATE INDEX IF NOT EXISTS idx_waitlist_section_id ON waitlist(section_id)",
		"CREATE INDEX IF NOT EXISTS idx_waitlist_student_id ON waitlist(student_id)",
		"CREATE INDEX IF NOT EXISTS idx_waitlist_position ON waitlist(section_id, position)",
		"CREATE INDEX IF NOT EXISTS idx_sections_course_semester ON sections(course_id, semester_id)",
		"CREATE INDEX IF NOT EXISTS idx_sections_available_seats ON sections(available_seats)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	return nil
}

// HealthCheck checks database connectivity
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
