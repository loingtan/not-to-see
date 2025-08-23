package cmd

import (
	"fmt"
	"os"

	"cobra-template/internal/config"
	"cobra-template/internal/infrastructure/database"
	"cobra-template/pkg/logger"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management",
	Long:  "Manage database migrations for the course registration system",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run pending migrations",
	Long:  "Execute all pending database migrations",
	Run:   runMigrateUp,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Display the status of all migrations",
	Run:   runMigrateStatus,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg := config.Get()

	// Connect to database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.Username,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}

	// Run migrations
	migrationRunner := database.NewMigrationRunner(db, "migrations")
	if err := migrationRunner.RunMigrations(); err != nil {
		logger.Error("Migration failed: %v", err)
		os.Exit(1)
	}

	fmt.Println("Migrations completed successfully!")
}

func runMigrateStatus(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg := config.Get()

	// Connect to database
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.Username,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}

	// Get migration status
	migrationRunner := database.NewMigrationRunner(db, "migrations")
	migrations, err := migrationRunner.GetMigrationStatus()
	if err != nil {
		logger.Error("Failed to get migration status: %v", err)
		os.Exit(1)
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")
	for _, migration := range migrations {
		status := "Pending"
		if migration.AppliedAt != nil {
			status = fmt.Sprintf("Applied at %s", migration.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("%s - %s [%s]\n", migration.ID, migration.Description, status)
	}
}
