package handlers

import (
	"net/http"
	"time"

	"cobra-template/internal/config"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	cfg := config.Get()

	services := make(map[string]string)

	// Check database (mock for now)
	services["database"] = "healthy"

	// Check cache (mock for now)
	services["cache"] = "healthy"

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   cfg.App.Version,
		Services:  services,
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck handles GET /ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// In a real application, you would check if all dependencies are ready
	// For now, we'll just return ready

	response := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// LivenessCheck handles GET /live
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// In a real application, you would check if the application is alive
	// For now, we'll just return alive

	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
