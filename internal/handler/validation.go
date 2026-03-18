package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}
		if name == "" {
			return fld.Name
		}
		if i := strings.IndexByte(name, ','); i >= 0 {
			name = name[:i]
		}
		return name
	})
	validate = v
}

// validateStructLaravel returns Laravel-like validation errors:
// { "field": ["The field ..."], "other": ["..."] }
// It returns nil when valid.
func validateStructLaravel(s any) map[string][]string {
	if err := validate.Struct(s); err != nil {
		if ves, ok := err.(validator.ValidationErrors); ok {
			out := make(map[string][]string, len(ves))
			for _, fe := range ves {
				field := fe.Field()
				if field == "" {
					field = strings.ToLower(fe.StructField())
				}
				msg := laravelMessage(field, fe)
				if msg == "" {
					continue
				}
				out[field] = append(out[field], msg)
			}
			return out
		}
		return map[string][]string{"_": {err.Error()}}
	}
	return nil
}

// bindErrorLaravel converts Gin bind/validator errors into Laravel-like error payloads.
// - validator.ValidationErrors => map[field][]messages
// - json.SyntaxError / json.UnmarshalTypeError => string message
func bindError(err error) (any, bool) {
	if err == nil {
		return nil, false
	}

	var ves validator.ValidationErrors
	if errors.As(err, &ves) {
		out := make(map[string][]string, len(ves))
		for _, fe := range ves {
			field := fe.Field()
			if field == "" {
				field = strings.ToLower(fe.StructField())
			}
			msg := laravelMessage(field, fe)
			if msg == "" {
				continue
			}
			out[field] = append(out[field], msg)
		}
		return out, true
	}

	var se *json.SyntaxError
	if errors.As(err, &se) {
		return "The request body must be valid JSON.", true
	}
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		return fmt.Sprintf("The %s field has an invalid type.", ute.Field), true
	}

	return nil, false
}

func laravelMessage(field string, fe validator.FieldError) string {
	// Field shown in message should be the json name (already), with underscores.
	pretty := strings.ReplaceAll(field, "-", "_")
	pretty = strings.ReplaceAll(pretty, ".", "_")

	switch fe.ActualTag() {
	case "required":
		return fmt.Sprintf("The %s field is required.", pretty)
	case "email":
		return fmt.Sprintf("The %s field must be a valid email address.", pretty)
	case "min":
		return fmt.Sprintf("The %s field must be at least %s.", pretty, fe.Param())
	case "max":
		return fmt.Sprintf("The %s field must not be greater than %s.", pretty, fe.Param())
	default:
		return fmt.Sprintf("The %s field is invalid.", pretty)
	}
}
