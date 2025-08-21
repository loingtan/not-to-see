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

	if err := validator.ValidateStruct(&req); err != nil {
		validationErrors := validator.FormatValidationError(err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}
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

// GetAvailableSections handles GET /api/sections/available
func (h *RegistrationHandler) GetAvailableSections(c *gin.Context) {
	semesterIDStr := c.Query("semester_id")
	courseIDStr := c.Query("course_id")

	// Validate query parameters
	if semesterIDStr == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "semester_id is required",
		})
		return
	}

	semesterID, err := uuid.Parse(semesterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid semester_id format",
		})
		return
	}

	var courseID *uuid.UUID
	if courseIDStr != "" {
		parsedCourseID, err := uuid.Parse(courseIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid course_id format",
			})
			return
		}
		courseID = &parsedCourseID
	}

	sections, err := h.registrationService.GetAvailableSections(c.Request.Context(), semesterID, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to retrieve available sections",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Available sections retrieved successfully",
		Data:    map[string]interface{}{"sections": sections},
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

	waitlistEntries, err := h.registrationService.GetStudentWaitlistStatus(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to retrieve waitlist status",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Waitlist status retrieved successfully",
		Data:    map[string]interface{}{"waitlist_entries": waitlistEntries},
	})
}

// GetStudentRegistrations handles GET /api/students/:student_id/registrations
func (h *RegistrationHandler) GetStudentRegistrations(c *gin.Context) {
	studentIDStr := c.Param("student_id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student ID format",
		})
		return
	}

	registrations, err := h.registrationService.GetStudentRegistrations(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to retrieve student registrations",
			Errors:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Student registrations retrieved successfully",
		Data:    map[string]interface{}{"registrations": registrations},
	})
}
