package validation

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator initializes the validator
func InitValidator() {
	validate = validator.New()
}

// GetValidator returns the validator instance
func GetValidator() (*validator.Validate, error) {
	if validate == nil {
		InitValidator()
	}
	return validate, nil
}

// Translate translates validation errors to human-readable messages
func Translate(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return err.Field() + " must be a valid email address"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	case "uuid":
		return err.Field() + " must be a valid UUID"
	case "url":
		return err.Field() + " must be a valid URL"
	case "numeric":
		return err.Field() + " must be numeric"
	case "alpha":
		return err.Field() + " must contain only alphabetic characters"
	case "alphanum":
		return err.Field() + " must contain only alphanumeric characters"
	default:
		return err.Field() + " failed validation: " + err.Tag()
	}
}
