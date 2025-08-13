package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init initializes the validator
func init() {
	validate = validator.New()
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct validates a struct
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// FormatValidationError formats validation errors into a readable format
func FormatValidationError(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   strings.ToLower(fieldError.Field()),
				Tag:     fieldError.Tag(),
				Message: getErrorMessage(fieldError),
			})
		}
	}

	return errors
}

// getErrorMessage returns a human-readable error message for validation errors
func getErrorMessage(fieldError validator.FieldError) string {
	field := strings.ToLower(fieldError.Field())
	
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fieldError.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, fieldError.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fieldError.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fieldError.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fieldError.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fieldError.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldError.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
