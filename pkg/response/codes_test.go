package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildResponseCode_ParseRoundTrip(t *testing.T) {
	t.Parallel()

	code := BuildResponseCode(401, ServiceCodeAuth, CaseCodeUnauthorized)
	httpStatus, svc, ccase := ParseResponseCode(code)

	assert.Equal(t, 401, httpStatus)
	assert.Equal(t, ServiceCodeAuth, svc)
	assert.Equal(t, CaseCodeUnauthorized, ccase)
}

func TestBuildResponseCode_PanicsOnInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		httpStatus int
		service    string
		ccase      string
	}{
		{"invalid httpStatus low", 99, "01", "21"},
		{"invalid httpStatus high", 600, "01", "21"},
		{"invalid service len", 200, "1", "21"},
		{"invalid case len", 200, "01", "2"},
		{"non-numeric service", 200, "a1", "21"},
		{"non-numeric case", 200, "01", "2a"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Panics(t, func() {
				_ = BuildResponseCode(tc.httpStatus, tc.service, tc.ccase)
			})
		})
	}
}

func TestParseResponseCode_InvalidRange(t *testing.T) {
	t.Parallel()

	httpStatus, svc, ccase := ParseResponseCode(-1)
	assert.Equal(t, 0, httpStatus)
	assert.Equal(t, "", svc)
	assert.Equal(t, "", ccase)

	httpStatus, svc, ccase = ParseResponseCode(10000000) // 8 digits
	assert.Equal(t, 0, httpStatus)
	assert.Equal(t, "", svc)
	assert.Equal(t, "", ccase)
}

