package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID          string
	Description string
	SQL         string
	AppliedAt   *time.Time
}

type MigrationRunner struct {
	db            *gorm.DB
	migrationsDir string
}

func NewMigrationRunner(db *gorm.DB, migrationsDir string) *MigrationRunner {
	return &MigrationRunner{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

func (mr *MigrationRunner) createMigrationsTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id VARCHAR(255) PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	return mr.db.Exec(sql).Error
}

func (mr *MigrationRunner) getAppliedMigrations() (map[string]bool, error) {
	var migrations []Migration
	err := mr.db.Raw("SELECT id FROM schema_migrations ORDER BY id").Scan(&migrations).Error
	if err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, migration := range migrations {
		applied[migration.ID] = true
	}

	return applied, nil
}

func (mr *MigrationRunner) getMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(mr.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func (mr *MigrationRunner) readMigrationFile(filePath string) (*Migration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	id := parts[0]
	description := strings.TrimSuffix(parts[1], ".sql")
	description = strings.ReplaceAll(description, "_", " ")

	return &Migration{
		ID:          id,
		Description: description,
		SQL:         string(content),
	}, nil
}

func (mr *MigrationRunner) RunMigrations() error {

	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	files, err := mr.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	pendingCount := 0
	for _, file := range files {
		migration, err := mr.readMigrationFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if applied[migration.ID] {
			continue
		}

		err = mr.db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Exec(migration.SQL).Error; err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
			}

			if err := tx.Exec("INSERT INTO schema_migrations (id, description) VALUES (?, ?)",
				migration.ID, migration.Description).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		fmt.Printf("Applied migration: %s - %s\n", migration.ID, migration.Description)
		pendingCount++
	}

	if pendingCount == 0 {
		fmt.Println("No pending migrations to apply")
	} else {
		fmt.Printf("Successfully applied %d migrations\n", pendingCount)
	}

	return nil
}

func (mr *MigrationRunner) GetMigrationStatus() ([]Migration, error) {

	applied, err := mr.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	files, err := mr.getMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		migration, err := mr.readMigrationFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if applied[migration.ID] {

			var appliedAt time.Time
			err := mr.db.Raw("SELECT applied_at FROM schema_migrations WHERE id = ?", migration.ID).Scan(&appliedAt).Error
			if err == nil {
				migration.AppliedAt = &appliedAt
			}
		}

		migrations = append(migrations, *migration)
	}

	return migrations, nil
}

func RunSQLMigrations(db *gorm.DB) error {

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	migrationsDir := "scripts/migrations"
	runner := NewMigrationRunner(db, migrationsDir)
	return runner.RunMigrations()
}
