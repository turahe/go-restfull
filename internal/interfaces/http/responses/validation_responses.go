// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"strings"
)

// ValidationError represents a single field validation error.
// This struct provides detailed information about validation failures for specific
// form fields, including the field name, error message, and optionally the invalid value.
type ValidationError struct {
	// Field is the name of the form field that failed validation
	Field string `json:"field"`
	// Message is the human-readable error message describing the validation failure
	Message string `json:"message"`
	// Value is the optional invalid value that caused the validation failure
	Value string `json:"value,omitempty"`
}

// ValidationErrors represents a collection of validation errors.
// This type provides methods for managing and querying multiple validation errors
// across different form fields, following Laravel's validation error collection pattern.
type ValidationErrors []ValidationError

// ValidationErrorResponse represents a Laravel-style validation error response.
// This struct provides a standardized format for returning validation errors to clients,
// including HTTP status, general message, and detailed field-specific errors.
type ValidationErrorResponse struct {
	// Status indicates the response status (typically "error" for validation failures)
	Status string `json:"status"`
	// Message provides a general description of the validation failure
	Message string `json:"message"`
	// Errors contains the collection of field-specific validation errors
	Errors ValidationErrors `json:"errors"`
}

// NewValidationErrorResponse creates a new validation error response.
// This function creates a standardized validation error response with the specified
// message and collection of validation errors.
//
// Parameters:
//   - message: General description of the validation failure
//   - errors: Collection of field-specific validation errors
//
// Returns:
//   - A new ValidationErrorResponse with error status and validation details
func NewValidationErrorResponse(message string, errors ValidationErrors) ValidationErrorResponse {
	return ValidationErrorResponse{
		Status:  "error",
		Message: message,
		Errors:  errors,
	}
}

// AddError adds a validation error to the collection.
// This method appends a new validation error for a specific field without
// including the invalid value.
//
// Parameters:
//   - field: The name of the form field that failed validation
//   - message: The error message describing the validation failure
func (v *ValidationErrors) AddError(field, message string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
	})
}

// AddErrorWithValue adds a validation error with the invalid value.
// This method appends a new validation error including the field name,
// error message, and the invalid value that caused the failure.
//
// Parameters:
//   - field: The name of the form field that failed validation
//   - message: The error message describing the validation failure
//   - value: The invalid value that caused the validation failure
func (v *ValidationErrors) AddErrorWithValue(field, message, value string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors checks if there are any validation errors.
// This method returns true if the collection contains one or more validation errors.
//
// Returns:
//   - True if there are validation errors, false otherwise
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// GetFirstError returns the first error message.
// This method returns the message from the first validation error in the collection,
// useful for displaying a general error message to users.
//
// Returns:
//   - The first error message, or empty string if no errors exist
func (v ValidationErrors) GetFirstError() string {
	if len(v) > 0 {
		return v[0].Message
	}
	return ""
}

// GetErrorsByField returns all errors for a specific field.
// This method filters the validation errors to return only those associated
// with the specified field name.
//
// Parameters:
//   - field: The field name to filter errors by
//
// Returns:
//   - Slice of error messages for the specified field
func (v ValidationErrors) GetErrorsByField(field string) []string {
	var errors []string
	for _, err := range v {
		if err.Field == field {
			errors = append(errors, err.Message)
		}
	}
	return errors
}

// ToMap converts validation errors to a map format (Laravel style).
// This method transforms the validation errors into a map where keys are field names
// and values are slices of error messages, following Laravel's validation error format.
//
// Returns:
//   - Map with field names as keys and error message slices as values
func (v ValidationErrors) ToMap() map[string][]string {
	errors := make(map[string][]string)
	for _, err := range v {
		errors[err.Field] = append(errors[err.Field], err.Message)
	}
	return errors
}

// Common validation error messages provide standardized error messages
// that can be used across the application for consistent validation feedback.
// These messages follow Laravel's validation message format with placeholders
// for dynamic field names and values.
const (
	// ErrRequired indicates a required field is missing
	ErrRequired = "The :field field is required."
	// ErrEmail indicates an invalid email format
	ErrEmail = "The :field must be a valid email address."
	// ErrMin indicates a field value is below the minimum length
	ErrMin = "The :field must be at least :min characters."
	// ErrMax indicates a field value exceeds the maximum length
	ErrMax = "The :field may not be greater than :max characters."
	// ErrBetween indicates a field value is outside the allowed range
	ErrBetween = "The :field must be between :min and :max characters."
	// ErrNumeric indicates a field value is not numeric
	ErrNumeric = "The :field must be a number."
	// ErrInteger indicates a field value is not an integer
	ErrInteger = "The :field must be an integer."
	// ErrString indicates a field value is not a string
	ErrString = "The :field must be a string."
	// ErrBoolean indicates a field value is not a boolean
	ErrBoolean = "The :field must be true or false."
	// ErrDate indicates a field value is not a valid date
	ErrDate = "The :field is not a valid date."
	// ErrDateFormat indicates a field value doesn't match the expected date format
	ErrDateFormat = "The :field does not match the format :format."
	// ErrBefore indicates a field value must be before a specific date
	ErrBefore = "The :field must be a date before :date."
	// ErrAfter indicates a field value must be after a specific date
	ErrAfter = "The :field must be a date after :date."
	// ErrBeforeOrEqual indicates a field value must be before or equal to a specific date
	ErrBeforeOrEqual = "The :field must be a date before or equal to :date."
	// ErrAfterOrEqual indicates a field value must be after or equal to a specific date
	ErrAfterOrEqual = "The :field must be a date after or equal to :date."
	// ErrConfirmed indicates a field confirmation doesn't match
	ErrConfirmed = "The :field confirmation does not match."
	// ErrDifferent indicates two fields must have different values
	ErrDifferent = "The :field and :other must be different."
	// ErrDigits indicates a field must have a specific number of digits
	ErrDigits = "The :field must be :digits digits."
	// ErrDigitsBetween indicates a field must have digits within a specific range
	ErrDigitsBetween = "The :field must be between :min and :max digits."
	// ErrDimensions indicates an image has invalid dimensions
	ErrDimensions = "The :field has invalid image dimensions."
	// ErrDistinct indicates a field has duplicate values
	ErrDistinct = "The :field field has a duplicate value."
	// ErrExists indicates a selected value doesn't exist in the database
	ErrExists = "The selected :field is invalid."
	// ErrFile indicates a field must be a file
	ErrFile = "The :field must be a file."
	// ErrFilled indicates a field must have a value
	ErrFilled = "The :field field must have a value."
	// ErrImage indicates a field must be an image file
	ErrImage = "The :field must be an image."
	// ErrIn indicates a field value is not in the allowed list
	ErrIn = "The selected :field is invalid."
	// ErrInArray indicates a field value doesn't exist in another array
	ErrInArray = "The :field field does not exist in :other."
	// ErrIP indicates a field must be a valid IP address
	ErrIP = "The :field must be a valid IP address."
	// ErrIPv4 indicates a field must be a valid IPv4 address
	ErrIPv4 = "The :field must be a valid IPv4 address."
	// ErrIPv6 indicates a field must be a valid IPv6 address
	ErrIPv6 = "The :field must be a valid IPv6 address."
	// ErrJSON indicates a field must be valid JSON
	ErrJSON = "The :field must be a valid JSON string."
	// ErrMimes indicates a file must be of specific MIME types
	ErrMimes = "The :field must be a file of type: :values."
	// ErrMimetypes indicates a file must be of specific MIME types
	ErrMimetypes = "The :field must be a file of type: :values."
	// ErrNotIn indicates a field value is in the disallowed list
	ErrNotIn = "The selected :field is invalid."
	// ErrNotRegex indicates a field format doesn't match the regex pattern
	ErrNotRegex = "The :field format is invalid."
	// ErrNullable indicates a field can be null or present
	ErrNullable = "The :field field must be null or present."
	// ErrPresent indicates a field must be present (even if null)
	ErrPresent = "The :field field must be present."
	// ErrRegex indicates a field format doesn't match the regex pattern
	ErrRegex = "The :field format is invalid."
	// ErrRequiredIf indicates a field is required when another field has a specific value
	ErrRequiredIf = "The :field field is required when :other is :value."
	// ErrRequiredUnless indicates a field is required unless another field is in a list
	ErrRequiredUnless = "The :field field is required unless :other is in :values."
	// ErrRequiredWith indicates a field is required when another field is present
	ErrRequiredWith = "The :field field is required when :values is present."
	// ErrRequiredWithAll indicates a field is required when all specified fields are present
	ErrRequiredWithAll = "The :field field is required when :values are present."
	// ErrRequiredWithout indicates a field is required when another field is not present
	ErrRequiredWithout = "The :field field is required when :values is not present."
	// ErrRequiredWithoutAll indicates a field is required when none of the specified fields are present
	ErrRequiredWithoutAll = "The :field field is required when none of :values are present."
	// ErrSame indicates two fields must have the same value
	ErrSame = "The :field and :other must match."
	// ErrSize indicates a field must have a specific size
	ErrSize = "The :field must be :size."
	// ErrTimezone indicates a field must be a valid timezone
	ErrTimezone = "The :field must be a valid timezone."
	// ErrUnique indicates a field value must be unique in the database
	ErrUnique = "The :field has already been taken."
	// ErrURL indicates a field must be a valid URL
	ErrURL = "The :field format is invalid."
	// ErrUUID indicates a field must be a valid UUID
	ErrUUID = "The :field must be a valid UUID."
)

// ValidationErrorBuilder helps build validation error responses.
// This struct provides a fluent interface for building validation error responses
// with common validation error types and custom error messages.
type ValidationErrorBuilder struct {
	// errors contains the collection of validation errors being built
	errors ValidationErrors
}

// NewValidationErrorBuilder creates a new validation error builder.
// This function initializes a new builder instance with an empty error collection.
//
// Returns:
//   - A new ValidationErrorBuilder instance ready for building validation errors
func NewValidationErrorBuilder() *ValidationErrorBuilder {
	return &ValidationErrorBuilder{
		errors: make(ValidationErrors, 0),
	}
}

// AddRequired adds a required field error.
// This method adds a standard "required field" validation error for the specified field.
//
// Parameters:
//   - field: The name of the required field
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddRequired(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrRequired, ":field", field))
	return b
}

// AddEmail adds an email validation error.
// This method adds a standard "invalid email" validation error for the specified field.
//
// Parameters:
//   - field: The name of the email field
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddEmail(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrEmail, ":field", field))
	return b
}

// AddMin adds a minimum length error.
// This method adds a standard "minimum length" validation error for the specified field.
//
// Parameters:
//   - field: The name of the field
//   - min: The minimum required length
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddMin(field string, min int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrMin, ":field", field)
	message = strings.ReplaceAll(message, ":min", string(rune(min)))
	b.errors.AddError(field, message)
	return b
}

// AddMax adds a maximum length error.
// This method adds a standard "maximum length" validation error for the specified field.
//
// Parameters:
//   - field: The name of the field
//   - max: The maximum allowed length
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddMax(field string, max int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrMax, ":field", field)
	message = strings.ReplaceAll(message, ":max", string(rune(max)))
	b.errors.AddError(field, message)
	return b
}

// AddBetween adds a between length error.
// This method adds a standard "between length" validation error for the specified field.
//
// Parameters:
//   - field: The name of the field
//   - min: The minimum required length
//   - max: The maximum allowed length
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddBetween(field string, min, max int) *ValidationErrorBuilder {
	message := strings.ReplaceAll(ErrBetween, ":field", field)
	message = strings.ReplaceAll(message, ":min", string(rune(min)))
	message = strings.ReplaceAll(message, ":max", string(rune(max)))
	b.errors.AddError(field, message)
	return b
}

// AddUnique adds a unique constraint error.
// This method adds a standard "unique constraint" validation error for the specified field.
//
// Parameters:
//   - field: The name of the field that must be unique
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddUnique(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrUnique, ":field", field))
	return b
}

// AddConfirmed adds a confirmation mismatch error.
// This method adds a standard "confirmation mismatch" validation error for the specified field.
//
// Parameters:
//   - field: The name of the field that failed confirmation
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddConfirmed(field string) *ValidationErrorBuilder {
	b.errors.AddError(field, strings.ReplaceAll(ErrConfirmed, ":field", field))
	return b
}

// AddCustom adds a custom validation error.
// This method adds a custom validation error message for the specified field.
//
// Parameters:
//   - field: The name of the field
//   - message: The custom error message
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddCustom(field, message string) *ValidationErrorBuilder {
	b.errors.AddError(field, message)
	return b
}

// AddCustomWithValue adds a custom validation error with the invalid value.
// This method adds a custom validation error including the invalid value that caused the failure.
//
// Parameters:
//   - field: The name of the field
//   - message: The custom error message
//   - value: The invalid value that caused the validation failure
//
// Returns:
//   - The builder instance for method chaining
func (b *ValidationErrorBuilder) AddCustomWithValue(field, message, value string) *ValidationErrorBuilder {
	b.errors.AddErrorWithValue(field, message, value)
	return b
}

// Build creates the final validation error response.
// This method constructs the final ValidationErrorResponse using the collected errors
// and the specified general message.
//
// Parameters:
//   - message: The general validation failure message
//
// Returns:
//   - A complete ValidationErrorResponse with all collected validation errors
func (b *ValidationErrorBuilder) Build(message string) ValidationErrorResponse {
	return NewValidationErrorResponse(message, b.errors)
}

// BuildDefault creates a validation error response with default message.
// This method constructs the final ValidationErrorResponse using the collected errors
// with a standard default validation failure message.
//
// Returns:
//   - A complete ValidationErrorResponse with default message and all collected errors
func (b *ValidationErrorBuilder) BuildDefault() ValidationErrorResponse {
	return NewValidationErrorResponse("The given data was invalid.", b.errors)
}

// HasErrors checks if the builder has any errors.
// This method returns true if the builder has collected one or more validation errors.
//
// Returns:
//   - True if there are validation errors, false otherwise
func (b *ValidationErrorBuilder) HasErrors() bool {
	return len(b.errors) > 0
}

// GetErrors returns the validation errors.
// This method returns the collection of validation errors that have been built.
//
// Returns:
//   - The collection of ValidationErrors that have been added to the builder
func (b *ValidationErrorBuilder) GetErrors() ValidationErrors {
	return b.errors
}
