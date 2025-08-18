package handlers

import (
	"net/http"

	"cobra-template/internal/service"
	"cobra-template/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// RegistrationHandler handles registration-related HTTP requests
type RegistrationHandler struct {
	registrationService *service.RegistrationService
}

// NewRegistrationHandler creates a new registration handler
func NewRegistrationHandler(registrationService *service.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		registrationService: registrationService,
	}
}

// Register handles POST /api/register
func (h *RegistrationHandler) Register(c *gin.Context) {
	var req service.RegisterRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request format",
			Errors:  err.Error(),
		})
		return
	}

	// Validate request
	if err := validator.ValidateStruct(&req); err != nil {
		validationErrors := validator.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}

	// Process registration
	response, err := h.registrationService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Registration failed",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Registration processed successfully",
		Data:    response,
	})
}

// DropCourse handles POST /api/drop
func (h *RegistrationHandler) DropCourse(c *gin.Context) {
	type DropRequest struct {
		StudentID uuid.UUID `json:"student_id" validate:"required"`
		SectionID uuid.UUID `json:"section_id" validate:"required"`
	}

	var req DropRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request format",
			Errors:  err.Error(),
		})
		return
	}

	// Validate request
	if err := validator.ValidateStruct(&req); err != nil {
		validationErrors := validator.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}

	// Process course drop
	err := h.registrationService.DropCourse(c.Request.Context(), req.StudentID, req.SectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to drop course",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Course dropped successfully",
	})
}

// GetRegistrations handles GET /api/students/:student_id/registrations
func (h *RegistrationHandler) GetRegistrations(c *gin.Context) {
	studentIDStr := c.Param("student_id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student ID format",
		})
		return
	}

	// This would typically call a service method to get registrations
	// For now, returning a placeholder response
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Registrations retrieved successfully",
		Data:    map[string]interface{}{"student_id": studentID, "registrations": []interface{}{}},
	})
}

// GetAvailableSections handles GET /api/sections/available
func (h *RegistrationHandler) GetAvailableSections(c *gin.Context) {
	semesterID := c.Query("semester_id")
	courseID := c.Query("course_id")

	// Validate query parameters
	if semesterID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "semester_id is required",
		})
		return
	}

	// This would typically call a service method to get available sections
	// For now, returning a placeholder response
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Available sections retrieved successfully",
		Data: map[string]any{
			"semester_id": semesterID,
			"course_id":   courseID,
			"sections":    []any{},
		},
	})
}

// GetWaitlistStatus handles GET /api/students/:student_id/waitlist
func (h *RegistrationHandler) GetWaitlistStatus(c *gin.Context) {
	studentIDStr := c.Param("student_id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student ID format",
		})
		return
	}

	// This would typically call a service method to get waitlist status
	// For now, returning a placeholder response
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Waitlist status retrieved successfully",
		Data:    map[string]interface{}{"student_id": studentID, "waitlist_entries": []interface{}{}},
	})
}
