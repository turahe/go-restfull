package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/turahe/go-restfull/internal/shared/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom tag name function to use json tag names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct using the validator package
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var validationErrors []string
		
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, formatValidationError(err))
		}
		
		return errors.NewDomainError(
			errors.ValidationErrorCode,
			fmt.Sprintf("Validation failed: %s", strings.Join(validationErrors, ", ")),
		).WithDetails("validation_errors", validationErrors)
	}
	
	return nil
}

// formatValidationError formats a validation error into a human-readable message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be no more than %s characters long", field, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}