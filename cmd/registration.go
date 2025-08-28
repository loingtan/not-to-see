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
	registrationPort    string
	enableLoadTestCache bool
)

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
	registrationCmd.Flags().StringVarP(&registrationPort, "port", "p", "8080", "Port for the registration server to listen on")
	registrationCmd.Flags().BoolVar(&enableLoadTestCache, "load-test-cache", false, "Enable enhanced pre-caching for load testing")
}

func startRegistrationServer() {
	cfg := config.Get()
	if registrationPort != "8080" {
		cfg.Server.Port = registrationPort
	}

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

	if err := database.RunMigrations(db); err != nil {
		logger.Error("Failed to run database migrations: %v", err)
		os.Exit(1)
	}

	if err := database.HealthCheck(db); err != nil {
		logger.Error("Database health check failed: %v", err)
		os.Exit(1)
	}

	routerComponents := router.NewRegistrationRouterWithQueue(db)
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        routerComponents.Router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	go func() {
		logger.Info("🎓 Starting Course Registration Server on port %s", cfg.Server.Port)
		logger.Info("📚 Available endpoints:")
		logger.Info("  POST /api/v1/register - Register for courses")
		logger.Info("  POST /api/v1/register/drop - Drop a course")
		logger.Info("  GET  /api/v1/students/{id}/registrations - Get student registrations")
		logger.Info("  GET  /api/v1/students/{id}/waitlist - Get waitlist status")
		logger.Info("  GET  /api/v1/sections/available - Get available sections")
		logger.Info("  POST /api/v1/cache/warmup - Manual cache warmup")
		logger.Info("  POST /api/v1/cache/warmup/loadtest - Enhanced load test cache warmup")
		logger.Info("  GET  /api/v1/cache/stats - Cache statistics")
		logger.Info("  GET  /api/v1/cache/loadtest/status - Load test readiness status")
		logger.Info("  GET  /health - Health check")

		if enableLoadTestCache {
			logger.Info("🚀 Load test cache optimization enabled")
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start registration server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Course Registration Server...")
	logger.Info("Stopping queue workers...")
	routerComponents.QueueService.StopWorkers()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("✅ Course Registration Server exited")
}
