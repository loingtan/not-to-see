package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cobra-template/internal/api/router"
	"cobra-template/internal/config"
	"cobra-template/internal/infrastructure/database"
	"cobra-template/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	registrationPort string
)

// registrationCmd represents the registration server command
var registrationCmd = &cobra.Command{
	Use:   "registration",
	Short: "Start the Course Registration HTTP server",
	Long: `Start the Course Registration HTTP server with full registration system functionality.
This includes:
- Course registration endpoints
- Waitlist management
- Section availability queries
- Async processing with queue workers
- Redis caching for performance`,
	Run: func(cmd *cobra.Command, args []string) {
		startRegistrationServer()
	},
}

func init() {
	rootCmd.AddCommand(registrationCmd)

	// Flags for registration server command
	registrationCmd.Flags().StringVarP(&registrationPort, "port", "p", "8080", "Port for the registration server to listen on")
}

func startRegistrationServer() {
	cfg := config.Get()

	// Override port if flag is provided
	if registrationPort != "8080" {
		cfg.Server.Port = registrationPort
	}

	// Initialize database connection
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

	// Create registration router with full system
	r := router.NewRegistrationRouter(db)

	// Create HTTP server
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🎓 Starting Course Registration Server on port %s", cfg.Server.Port)
		logger.Info("📚 Available endpoints:")
		logger.Info("  POST /api/v1/register - Register for courses")
		logger.Info("  POST /api/v1/register/drop - Drop a course")
		logger.Info("  GET  /api/v1/students/{id}/registrations - Get student registrations")
		logger.Info("  GET  /api/v1/students/{id}/waitlist - Get waitlist status")
		logger.Info("  GET  /api/v1/sections/available - Get available sections")
		logger.Info("  GET  /health - Health check")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start registration server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down Course Registration Server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("✅ Course Registration Server exited")
}
