package handlers

import (
	"net/http"
	"time"

	"cobra-template/internal/config"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	cfg := config.Get()

	services := make(map[string]string)

	services["database"] = "healthy"

	services["cache"] = "healthy"

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   cfg.App.Version,
		Services:  services,
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) ReadinessCheck(c *gin.Context) {

	response := map[string]any{
		"ready":     true,
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {

	response := map[string]any{
		"alive":     true,
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
