package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConnection(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s connect_timeout=10",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	log.Printf("DEBUG: Attempting to connect with DSN: %s", dsn)
	log.Printf("DEBUG: Config values - Host: %s, User: %s, Password: %s, DBName: %s, Port: %d, SSLMode: %s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	log.Println("Running SQL migrations...")

	migrationRunner := NewMigrationRunner(db, "migrations")
	err := migrationRunner.RunMigrations()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
