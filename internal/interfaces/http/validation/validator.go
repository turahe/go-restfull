package validation

import (
	"context"
	"regexp"
	"strings"

	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
)

// Validator provides Laravel-style validation functionality
type Validator struct {
	errors *responses.ValidationErrorBuilder
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: responses.NewValidationErrorBuilder(),
	}
}

// ValidateRequired validates that a field is required
func (v *Validator) ValidateRequired(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors.AddRequired(field)
	}
	return v
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(field, value string) *Validator {
	if value != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			v.errors.AddEmail(field)
		}
	}
	return v
}

// ValidateMin validates minimum length
func (v *Validator) ValidateMin(field, value string, min int) *Validator {
	if value != "" && len(value) < min {
		v.errors.AddMin(field, min)
	}
	return v
}

// ValidateMax validates maximum length
func (v *Validator) ValidateMax(field, value string, max int) *Validator {
	if value != "" && len(value) > max {
		v.errors.AddMax(field, max)
	}
	return v
}

// ValidateBetween validates length between min and max
func (v *Validator) ValidateBetween(field, value string, min, max int) *Validator {
	if value != "" {
		length := len(value)
		if length < min || length > max {
			v.errors.AddBetween(field, min, max)
		}
	}
	return v
}

// ValidateRegex validates against a regex pattern
func (v *Validator) ValidateRegex(field, value, pattern string) *Validator {
	if value != "" {
		regex := regexp.MustCompile(pattern)
		if !regex.MatchString(value) {
			v.errors.AddCustom(field, "The "+field+" format is invalid.")
		}
	}
	return v
}

// ValidateUsername validates username format (alphanumeric and underscore only)
func (v *Validator) ValidateUsername(field, value string) *Validator {
	if value != "" {
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		if !usernameRegex.MatchString(value) {
			v.errors.AddCustom(field, "The "+field+" can only contain letters, numbers, and underscores.")
		}
	}
	return v
}

// ValidatePhone validates international phone number format
func (v *Validator) ValidatePhone(field, value string) *Validator {
	if value != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(value) {
			v.errors.AddCustom(field, "The "+field+" format is invalid.")
		}
	}
	return v
}

// ValidateConfirmed validates that two fields match (for password confirmation)
func (v *Validator) ValidateConfirmed(field, value, confirmation string) *Validator {
	if value != "" && confirmation != "" && value != confirmation {
		v.errors.AddConfirmed(field)
	}
	return v
}

// ValidateUnique validates uniqueness against database
func (v *Validator) ValidateUnique(field, value string, exists bool, err error) *Validator {
	if value != "" {
		if err != nil {
			v.errors.AddCustom(field, "Failed to validate "+field+" uniqueness.")
		} else if exists {
			v.errors.AddUnique(field)
		}
	}
	return v
}

// ValidateCustom adds a custom validation error
func (v *Validator) ValidateCustom(field, message string) *Validator {
	v.errors.AddCustom(field, message)
	return v
}

// ValidateCustomWithValue adds a custom validation error with the invalid value
func (v *Validator) ValidateCustomWithValue(field, message, value string) *Validator {
	v.errors.AddCustomWithValue(field, message, value)
	return v
}

// HasErrors checks if the validator has any errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// GetErrors returns the validation errors
func (v *Validator) GetErrors() responses.ValidationErrors {
	return v.errors.GetErrors()
}

// GetErrorBuilder returns the validation error builder
func (v *Validator) GetErrorBuilder() *responses.ValidationErrorBuilder {
	return v.errors
}

// BuildResponse builds a validation error response
func (v *Validator) BuildResponse(message string) responses.ValidationErrorResponse {
	return v.errors.Build(message)
}

// BuildDefaultResponse builds a validation error response with default message
func (v *Validator) BuildDefaultResponse() responses.ValidationErrorResponse {
	return v.errors.BuildDefault()
}

// ValidateRegisterRequest validates a registration request with Laravel-style errors
func ValidateRegisterRequest(ctx context.Context, req interface{}, userRepo interface{}) (*responses.ValidationErrorBuilder, error) {
	// This is a placeholder for the actual validation logic
	// You would implement this based on your specific request structure
	return responses.NewValidationErrorBuilder(), nil
}

// ValidateLoginRequest validates a login request with Laravel-style errors
func ValidateLoginRequest(req interface{}) (*responses.ValidationErrorBuilder, error) {
	// This is a placeholder for the actual validation logic
	// You would implement this based on your specific request structure
	return responses.NewValidationErrorBuilder(), nil
}

// ValidateRefreshTokenRequest validates a refresh token request with Laravel-style errors
func ValidateRefreshTokenRequest(req interface{}) (*responses.ValidationErrorBuilder, error) {
	// This is a placeholder for the actual validation logic
	// You would implement this based on your specific request structure
	return responses.NewValidationErrorBuilder(), nil
}

// ValidateForgetPasswordRequest validates a forget password request with Laravel-style errors
func ValidateForgetPasswordRequest(req interface{}) (*responses.ValidationErrorBuilder, error) {
	// This is a placeholder for the actual validation logic
	// You would implement this based on your specific request structure
	return responses.NewValidationErrorBuilder(), nil
}

// ValidateResetPasswordRequest validates a reset password request with Laravel-style errors
func ValidateResetPasswordRequest(req interface{}) (*responses.ValidationErrorBuilder, error) {
	// This is a placeholder for the actual validation logic
	// You would implement this based on your specific request structure
	return responses.NewValidationErrorBuilder(), nil
}
