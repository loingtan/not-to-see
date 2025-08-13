package handlers

import (
	"net/http"
	"strconv"

	"cobra-template/internal/domain"
	"cobra-template/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService domain.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest

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

	// Create user
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	})
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	
	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	// Get user
	user, err := h.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    user,
	})
}

// GetUserByEmail handles GET /users/email/:email
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    user,
	})
}

// GetUserByUsername handles GET /users/username/:username
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	
	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	var req domain.UpdateUserRequest

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

	// Update user
	user, err := h.userService.UpdateUser(id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusConflict, APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	})
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	
	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(id); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// List users
	users, err := h.userService.ListUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    users,
	})
}
