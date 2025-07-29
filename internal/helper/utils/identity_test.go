package utils

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func TestParseIdentity(t *testing.T) {
	tests := []struct {
		name          string
		identity      string
		expectedEmail string
		expectedPhone string
	}{
		{
			name:          "valid email",
			identity:      "user@example.com",
			expectedEmail: "user@example.com",
			expectedPhone: "",
		},
		{
			name:          "phone number",
			identity:      "08123456789",
			expectedEmail: "",
			expectedPhone: "08123456789",
		},
		{
			name:          "empty string",
			identity:      "",
			expectedEmail: "",
			expectedPhone: "",
		},
		{
			name:          "complex email",
			identity:      "user.name+tag@example-domain.co.uk",
			expectedEmail: "user.name+tag@example-domain.co.uk",
			expectedPhone: "",
		},
		{
			name:          "phone with plus",
			identity:      "+628123456789",
			expectedEmail: "",
			expectedPhone: "+628123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, phone := ParseIdentity(tt.identity)
			if email != tt.expectedEmail {
				t.Errorf("ParseIdentity() email = %v, want %v", email, tt.expectedEmail)
			}
			if phone != tt.expectedPhone {
				t.Errorf("ParseIdentity() phone = %v, want %v", phone, tt.expectedPhone)
			}
		})
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name     string
		identity string
		expected bool
	}{
		{
			name:     "valid email",
			identity: "user@example.com",
			expected: true,
		},
		{
			name:     "phone number",
			identity: "08123456789",
			expected: false,
		},
		{
			name:     "empty string",
			identity: "",
			expected: false,
		},
		{
			name:     "complex email",
			identity: "user.name+tag@example-domain.co.uk",
			expected: true,
		},
		{
			name:     "string with @ but not email",
			identity: "user@",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmail(tt.identity)
			if result != tt.expected {
				t.Errorf("IsEmail(%q) = %v, want %v", tt.identity, result, tt.expected)
			}
		})
	}
}

func TestIsPhone(t *testing.T) {
	tests := []struct {
		name     string
		identity string
		expected bool
	}{
		{
			name:     "phone number",
			identity: "08123456789",
			expected: true,
		},
		{
			name:     "valid email",
			identity: "user@example.com",
			expected: false,
		},
		{
			name:     "empty string",
			identity: "",
			expected: true,
		},
		{
			name:     "phone with plus",
			identity: "+628123456789",
			expected: true,
		},
		{
			name:     "string with @",
			identity: "user@",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPhone(tt.identity)
			if result != tt.expected {
				t.Errorf("IsPhone(%q) = %v, want %v", tt.identity, result, tt.expected)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	// Test with valid user ID
	validUserID := uuid.New()
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Locals("user_id", validUserID)

	result, err := GetUserID(ctx)
	if err != nil {
		t.Errorf("GetUserID() error = %v, want nil", err)
	}
	if result != validUserID {
		t.Errorf("GetUserID() = %v, want %v", result, validUserID)
	}

	// Test with missing user ID
	ctx2 := app.AcquireCtx(&fasthttp.RequestCtx{})
	result2, err2 := GetUserID(ctx2)
	if err2 == nil {
		t.Error("GetUserID() should return error when user_id is missing")
	}
	if result2 != uuid.Nil {
		t.Errorf("GetUserID() = %v, want %v", result2, uuid.Nil)
	}

	// Test with invalid user ID type
	ctx3 := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx3.Locals("user_id", "invalid-uuid")
	result3, err3 := GetUserID(ctx3)
	if err3 == nil {
		t.Error("GetUserID() should return error when user_id is invalid type")
	}
	if result3 != uuid.Nil {
		t.Errorf("GetUserID() = %v, want %v", result3, uuid.Nil)
	}
}

func BenchmarkParseIdentity(b *testing.B) {
	identities := []string{
		"user@example.com",
		"08123456789",
		"user.name+tag@example-domain.co.uk",
		"+628123456789",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity := identities[i%len(identities)]
		ParseIdentity(identity)
	}
}

func BenchmarkIsEmail(b *testing.B) {
	identities := []string{
		"user@example.com",
		"08123456789",
		"user.name+tag@example-domain.co.uk",
		"+628123456789",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity := identities[i%len(identities)]
		IsEmail(identity)
	}
}

func BenchmarkIsPhone(b *testing.B) {
	identities := []string{
		"user@example.com",
		"08123456789",
		"user.name+tag@example-domain.co.uk",
		"+628123456789",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity := identities[i%len(identities)]
		IsPhone(identity)
	}
}
