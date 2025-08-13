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
	"cobra-template/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	port string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the HTTP server with the configured routes and middleware.
The server will handle REST API requests and provide mock endpoints
for development and testing purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Flags for server command
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port for the server to listen on")
}

func startServer() {
	cfg := config.Get()
	
	// Override port if flag is provided
	if port != "8080" {
		cfg.Server.Port = port
	}

	// Create router
	r := router.NewRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give server 5 seconds to finish current requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
