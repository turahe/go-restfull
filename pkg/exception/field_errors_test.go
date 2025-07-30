package exception_test

import (
	"errors"
	"testing"

	"github.com/turahe/go-restfull/internal/http/validation"
	. "github.com/turahe/go-restfull/pkg/exception"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func Test_NewValidationFailedErrors(t *testing.T) {
	type Student struct {
		Name  string `json:"name" validate:"required"`
		Age   string `json:"age" validate:"required"`
		Class string `json:"class" validate:"-"`
		Grade string `json:"grade" validate:"-"`
	}

	type testCase struct {
		name     string
		student  Student
		expected *ExceptionErrors
	}

	testCases := []testCase{
		{
			name: "test missing age which is required field",
			student: Student{
				Name:  "Youth",
				Age:   "",
				Class: "1/6",
				Grade: "A++",
			},
			expected: &ExceptionErrors{
				HttpStatusCode: 422,
				GlobalMessage:  "validation failed",
				ErrItems: []*ExceptionError{
					{
						Message:      "Age is required",
						Type:         ERROR_TYPE_VALIDATION_ERROR,
						ErrorSubcode: SUBCODE_VALIDATION_FAILED,
					},
				},
			},
		},
		{
			name: "test missing name, age which is required field",
			student: Student{
				Name:  "",
				Age:   "",
				Class: "1/6",
				Grade: "A++",
			},
			expected: &ExceptionErrors{
				HttpStatusCode: 422,
				GlobalMessage:  "validation failed",
				ErrItems: []*ExceptionError{
					{
						Message:      "Name is required",
						Type:         ERROR_TYPE_VALIDATION_ERROR,
						ErrorSubcode: SUBCODE_VALIDATION_FAILED,
					},
					{
						Message:      "Age is required",
						Type:         ERROR_TYPE_VALIDATION_ERROR,
						ErrorSubcode: SUBCODE_VALIDATION_FAILED,
					},
				},
			},
		},
	}

	validate, _ := validation.GetValidator()

	for _, tc := range testCases {
		if err := validate.Struct(tc.student); err != nil {
			var vErrs validator.ValidationErrors
			if errors.As(err, &vErrs) {
				actual := NewValidationFailedErrors(vErrs)
				assert.Equal(t, tc.expected, actual)
			}
		}
	}

}

func TestAppendFieldErrors(t *testing.T) {
	type Student struct {
		Name  string `json:"name" validate:"required"`
		Age   string `json:"age" validate:"required"`
		Class string `json:"class" validate:"-"`
		Grade string `json:"grade" validate:"-"`
	}

	student := Student{
		Name:  "Youth",
		Age:   "",
		Class: "1/6",
		Grade: "A++",
	}

	expected := &ExceptionErrors{
		HttpStatusCode: 422,
		GlobalMessage:  "validation failed",
		ErrItems: []*ExceptionError{
			{
				Message:      "Age is required",
				Type:         ERROR_TYPE_VALIDATION_ERROR,
				ErrorSubcode: SUBCODE_VALIDATION_FAILED,
			},
			{
				Message:      "a",
				Type:         "b",
				ErrorSubcode: 1,
			},
		},
	}

	validate, _ := validation.GetValidator()
	if err := validate.Struct(student); err != nil {
		if vErrs, ok := err.(validator.ValidationErrors); ok {
			actual := NewValidationFailedErrors(vErrs)
			actual = actual.Append(&ExceptionError{
				Message:      "a",
				Type:         "b",
				ErrorSubcode: 1,
			})

			assert.Equal(t, expected, actual)
		}
	}
}
