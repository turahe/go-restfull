package responses

import (
	"strings"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// ValidationErrorResponse represents a Laravel-style validation error response
type ValidationErrorResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Errors  ValidationErrors `json:"errors"`
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(message string, errors ValidationErrors) ValidationErrorResponse {
	return ValidationErrorResponse{
		Status:  "error",
		Message: message,
		Errors:  errors,
	}
}

// AddError adds a validation error to the collection
func (v *ValidationErrors) AddError(field, message string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
	})
}

// AddErrorWithValue adds a validation error with the invalid value
func (v *ValidationErrors) AddErrorWithValue(field, message, value string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors checks if there are any validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// GetFirstError returns the first error message
func (v ValidationErrors) GetFirstError() string {
	if len(v) > 0 {
		return v[0].Message
	}
	return ""
}

// GetErrorsByField returns all errors for a specific field
func (v ValidationErrors) GetErrorsByField(field string) []string {
	var errors []string
	for _, err := range v {
		if err.Field == field {
			errors = append(errors, err.Message)
		}
	}
	return errors
}

// ToMap converts validation errors to a map format (Laravel style)
func (v ValidationErrors) ToMap() map[string][]string {
	errors := make(map[string][]string)
	for _, err := range v {
		errors[err.Field] = append(errors[err.Field], err.Message)
	}
	return errors
}

// Common validation error messages
const (
	ErrRequired           = "The :field field is required."
	ErrEmail              = "The :field must be a valid email address."
	ErrMin                = "The :field must be at least :min characters."
	ErrMax                = "The :field may not be greater than :max characters."
	ErrBetween            = "The :field must be between :min and :max characters."
	ErrNumeric            = "The :field must be a number."
	ErrInteger            = "The :field must be an integer."
	ErrString             = "The :field must be a string."
	ErrBoolean            = "The :field must be true or false."
	ErrDate               = "The :field is not a valid date."
	ErrDateFormat         = "The :field does not match the format :format."
	ErrBefore             = "The :field must be a date before :date."
	ErrAfter              = "The :field must be a date after :date."
	ErrBeforeOrEqual      = "The :field must be a date before or equal to :date."
	ErrAfterOrEqual       = "The :field must be a date after or equal to :date."
	ErrConfirmed          = "The :field confirmation does not match."
	ErrDifferent          = "The :field and :other must be different."
	ErrDigits             = "The :field must be :digits digits."
	ErrDigitsBetween      = "The :field must be between :min and :max digits."
	ErrDimensions         = "The :field has invalid image dimensions."
	ErrDistinct           = "The :field field has a duplicate value."
	ErrExists             = "The selected :field is invalid."
	ErrFile               = "The :field must be a file."
	ErrFilled             = "The :field field must have a value."
	ErrImage              = "The :field must be an image."
	ErrIn                 = "The selected :field is invalid."
	ErrInArray            = "The :field field does not exist in :other."
	ErrIP                 = "The :field must be a valid IP address."
	ErrIPv4               = "The :field must be a valid IPv4 address."
	ErrIPv6               = "The :field must be a valid IPv6 address."
	ErrJSON               = "The :field must be a valid JSON string."
	ErrMimes              = "The :field must be a file of type: :values."
	ErrMimetypes          = "The :field must be a file of type: :values."
	ErrNotIn              = "The selected :field is invalid."
	ErrNotRegex           = "The :field format is invalid."
	ErrNullable           = "The :field field must be null or present."
	ErrPresent            = "The :field field must be present."
	ErrRegex              = "The :field format is invalid."
	ErrRequiredIf         = "The :field field is required when :other is :value."
	ErrRequiredUnless     = "The :field field is required unless :other is in :values."
	ErrRequiredWith       = "The :field field is required when :values is present."
	ErrRequiredWithAll    = "The :field field is required when :values are present."
	ErrRequiredWithout    = "The :field field is required when :values is not present."
	ErrRequiredWithoutAll = "The :field field is required when none of :values are present."
	ErrSame               = "The :field and :other must match."
	ErrSize               = "The :field must be :size."
	ErrTimezone           = "The :field must be a valid timezone."
	ErrUnique             = "The :field has already been taken."
	ErrURL                = "The :field format is invalid."
	ErrUUID               = "The :field must be a valid UUID."
)

// ValidationErrorBuilder helps build validation error responses
type ValidationErrorBuilder struct {
	errors ValidationErrors
}

// NewValidationErrorBuilder creates a new validation error builder
func NewValidationErrorBuilder() *ValidationErrorBuilder {
	return &ValidationErrorBuilder{
		errors: make(ValidationErrors, 0),
	}
}

// AddRequired adds a required field error
func (b *ValidationErrorBuilder) AddRequired(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrRequired, ":field", field))
	return b
}

// AddEmail adds an email validation error
func (b *ValidationErrorBuilder) AddEmail(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrEmail, ":field", field))
	return b
}

// AddMin adds a minimum length error
func (b *ValidationErrorBuilder) AddMin(field string, min int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrMin, ":field", field)
	message = strings.ReplaceAll(message, ":min", string(rune(min)))
	b.errors.AddError(field, message)
	return b
}

// AddMax adds a maximum length error
func (b *ValidationErrorBuilder) AddMax(field string, max int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrMax, ":field", field)
	message = strings.ReplaceAll(message, ":max", string(rune(max)))
	b.errors.AddError(field, message)
	return b
}

// AddBetween adds a between length error
func (b *ValidationErrorBuilder) AddBetween(field string, min, max int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrBetween, ":field", field)
	message = strings.ReplaceAll(message, ":min", string(rune(min)))
	message = strings.ReplaceAll(message, ":max", string(rune(max)))
	b.errors.AddError(field, message)
	return b
}

// AddUnique adds a unique constraint error
func (b *ValidationErrorBuilder) AddUnique(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrUnique, ":field", field))
	return b
}

// AddConfirmed adds a confirmation mismatch error
func (b *ValidationErrorBuilder) AddConfirmed(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrConfirmed, ":field", field))
	return b
}

// AddCustom adds a custom validation error
func (b *ValidationErrorBuilder) AddCustom(field, message string) *ValidationErrorBuilder {
	b.errors.AddError(field, message)
	return b
}

// AddCustomWithValue adds a custom validation error with the invalid value
func (b *ValidationErrorBuilder) AddCustomWithValue(field, message, value string) *ValidationErrorBuilder {
	b.errors.AddErrorWithValue(field, message, value)
	return b
}

// Build creates the final validation error response
func (b *ValidationErrorBuilder) Build(message string) ValidationErrorResponse {
	return NewValidationErrorResponse(message, b.errors)
}

// BuildDefault creates a validation error response with default message
func (b *ValidationErrorBuilder) BuildDefault() ValidationErrorResponse {
	return NewValidationErrorResponse("The given data was invalid.", b.errors)
}

// HasErrors checks if the builder has any errors
func (b *ValidationErrorBuilder) HasErrors() bool {
	return len(b.errors) > 0
}

// GetErrors returns the validation errors
func (b *ValidationErrorBuilder) GetErrors() ValidationErrors {
	return b.errors
}
