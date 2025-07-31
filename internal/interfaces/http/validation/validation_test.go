package validation

import (
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func TestInitValidator(t *testing.T) {
	// Test that InitValidator creates a new validator instance
	InitValidator()

	// Verify validator is not nil
	if validate == nil {
		t.Error("InitValidator() should create a validator instance")
	}
}

func TestGetValidator(t *testing.T) {
	// Reset validator for testing
	validate = nil

	// Test GetValidator when validator is nil
	validator, err := GetValidator()
	if err != nil {
		t.Errorf("GetValidator() error = %v", err)
	}

	if validator == nil {
		t.Error("GetValidator() should return a validator instance")
	}

	// Test GetValidator when validator is already initialized
	validator2, err := GetValidator()
	if err != nil {
		t.Errorf("GetValidator() error = %v", err)
	}

	if validator2 == nil {
		t.Error("GetValidator() should return a validator instance")
	}

	// Verify both instances are the same
	if validator != validator2 {
		t.Error("GetValidator() should return the same validator instance")
	}
}

func TestTranslate(t *testing.T) {
	// Test cases for different validation tags
	tests := []struct {
		name     string
		field    string
		tag      string
		param    string
		expected string
	}{
		{
			name:     "required field",
			field:    "RequiredField",
			tag:      "required",
			param:    "",
			expected: "RequiredField is required",
		},
		{
			name:     "email field",
			field:    "EmailField",
			tag:      "email",
			param:    "",
			expected: "EmailField must be a valid email address",
		},
		{
			name:     "min field",
			field:    "MinField",
			tag:      "min",
			param:    "3",
			expected: "MinField must be at least 3 characters",
		},
		{
			name:     "max field",
			field:    "MaxField",
			tag:      "max",
			param:    "10",
			expected: "MaxField must be at most 10 characters",
		},
		{
			name:     "uuid field",
			field:    "UUIDField",
			tag:      "uuid",
			param:    "",
			expected: "UUIDField must be a valid UUID",
		},
		{
			name:     "url field",
			field:    "URLField",
			tag:      "url",
			param:    "",
			expected: "URLField must be a valid URL",
		},
		{
			name:     "numeric field",
			field:    "NumericField",
			tag:      "numeric",
			param:    "",
			expected: "NumericField must be numeric",
		},
		{
			name:     "alpha field",
			field:    "AlphaField",
			tag:      "alpha",
			param:    "",
			expected: "AlphaField must contain only alphabetic characters",
		},
		{
			name:     "alphanum field",
			field:    "AlphanumField",
			tag:      "alphanum",
			param:    "",
			expected: "AlphanumField must contain only alphanumeric characters",
		},
		{
			name:     "unknown field",
			field:    "UnknownField",
			tag:      "unknown",
			param:    "",
			expected: "UnknownField failed validation: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock field error
			fieldError := &mockFieldError{
				field: tt.field,
				tag:   tt.tag,
				param: tt.param,
			}

			result := Translate(fieldError)
			if result != tt.expected {
				t.Errorf("Translate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTranslate_EdgeCases(t *testing.T) {
	// Test with empty field name
	fieldError := &mockFieldError{
		field: "",
		tag:   "required",
		param: "",
	}

	result := Translate(fieldError)
	expected := " is required"
	if result != expected {
		t.Errorf("Translate() = %v, want %v", result, expected)
	}

	// Test with empty tag
	fieldError2 := &mockFieldError{
		field: "TestField",
		tag:   "",
		param: "",
	}

	result2 := Translate(fieldError2)
	expected2 := "TestField failed validation: "
	if result2 != expected2 {
		t.Errorf("Translate() = %v, want %v", result2, expected2)
	}
}

func TestValidatorIntegration(t *testing.T) {
	// Initialize validator
	InitValidator()

	// Test struct for validation
	type TestStruct struct {
		RequiredField string `validate:"required"`
		EmailField    string `validate:"email"`
		MinField      string `validate:"min=3"`
	}

	// Test valid struct
	validStruct := TestStruct{
		RequiredField: "test",
		EmailField:    "test@example.com",
		MinField:      "abc",
	}

	err := validate.Struct(validStruct)
	if err != nil {
		t.Errorf("Valid struct should not have validation errors: %v", err)
	}

	// Test invalid struct
	invalidStruct := TestStruct{
		RequiredField: "",
		EmailField:    "invalid-email",
		MinField:      "ab",
	}

	err = validate.Struct(invalidStruct)
	if err == nil {
		t.Error("Invalid struct should have validation errors")
	}

	// Test that we can get validation errors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrors) == 0 {
			t.Error("Should have validation errors")
		}

		// Test translation of first error
		firstError := validationErrors[0]
		translated := Translate(firstError)
		if translated == "" {
			t.Error("Translation should not be empty")
		}
	}
}

// Mock field error for testing
type mockFieldError struct {
	field string
	tag   string
	param string
}

func (e *mockFieldError) Tag() string {
	return e.tag
}

func (e *mockFieldError) Field() string {
	return e.field
}

func (e *mockFieldError) Param() string {
	return e.param
}

func (e *mockFieldError) Error() string {
	return e.field + " failed validation: " + e.tag
}

func (e *mockFieldError) Type() reflect.Type {
	return reflect.TypeOf("")
}

func (e *mockFieldError) Value() interface{} {
	return nil
}

func (e *mockFieldError) Kind() reflect.Kind {
	return reflect.String
}

func (e *mockFieldError) Namespace() string {
	return ""
}

func (e *mockFieldError) StructNamespace() string {
	return ""
}

func (e *mockFieldError) StructField() string {
	return ""
}

func (e *mockFieldError) ActualTag() string {
	return ""
}

func (e *mockFieldError) IsZero() bool {
	return false
}

func (e *mockFieldError) Translate(ut ut.Translator) string {
	return e.Error()
}

func BenchmarkGetValidator(b *testing.B) {
	// Reset validator for benchmarking
	validate = nil

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetValidator()
	}
}

func BenchmarkTranslate(b *testing.B) {
	fieldError := &mockFieldError{
		field: "TestField",
		tag:   "required",
		param: "",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Translate(fieldError)
	}
}
