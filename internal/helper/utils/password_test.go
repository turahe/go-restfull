package utils

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "MySecurePassword@2024!",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "very long password",
			password: strings.Repeat("a", 70), // Within bcrypt's 72-byte limit
		},
		{
			name:     "special characters",
			password: "!@#$%^&*()_+-=[]{}|;:,.<>?",
		},
		{
			name:     "unicode password",
			password: "CaféRésumé2024",
		},
		{
			name:     "numbers only",
			password: "123456789",
		},
		{
			name:     "single character",
			password: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GeneratePassword(tt.password)

			// Verify hash is not empty
			if hash == "" {
				t.Error("Generated hash should not be empty")
			}

			// Verify hash is different from original password
			if hash == tt.password {
				t.Error("Generated hash should not be the same as the original password")
			}

			// Verify hash starts with bcrypt identifier
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
				t.Error("Generated hash should be a valid bcrypt hash")
			}

			// Verify hash length is reasonable (bcrypt hashes are typically 60 characters)
			if len(hash) < 50 {
				t.Error("Generated hash should be at least 50 characters long")
			}
		})
	}
}

func TestComparePassword(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		hashedPassword string
		expected       bool
	}{
		{
			name:           "correct password",
			password:       "password123",
			hashedPassword: GeneratePassword("password123"),
			expected:       true,
		},
		{
			name:           "incorrect password",
			password:       "wrongpassword",
			hashedPassword: GeneratePassword("password123"),
			expected:       false,
		},
		{
			name:           "empty password with empty hash",
			password:       "",
			hashedPassword: GeneratePassword(""),
			expected:       true,
		},
		{
			name:           "empty password with non-empty hash",
			password:       "",
			hashedPassword: GeneratePassword("somepassword"),
			expected:       false,
		},
		{
			name:           "complex password",
			password:       "MySecurePassword@2024!",
			hashedPassword: GeneratePassword("MySecurePassword@2024!"),
			expected:       true,
		},
		{
			name:           "unicode password",
			password:       "CaféRésumé2024",
			hashedPassword: GeneratePassword("CaféRésumé2024"),
			expected:       true,
		},
		{
			name:           "very long password",
			password:       strings.Repeat("a", 70), // Within bcrypt's 72-byte limit
			hashedPassword: GeneratePassword(strings.Repeat("a", 70)),
			expected:       true,
		},
		{
			name:           "case sensitive",
			password:       "Password",
			hashedPassword: GeneratePassword("password"),
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePassword(tt.hashedPassword, tt.password)
			if result != tt.expected {
				t.Errorf("ComparePassword() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGeneratePassword_Consistency(t *testing.T) {
	password := "testpassword"

	// Generate multiple hashes for the same password
	hash1 := GeneratePassword(password)
	hash2 := GeneratePassword(password)
	hash3 := GeneratePassword(password)

	// Verify all hashes are different (bcrypt uses random salt)
	if hash1 == hash2 || hash1 == hash3 || hash2 == hash3 {
		t.Error("Generated hashes for the same password should be different due to random salt")
	}

	// Verify all hashes can be used to verify the original password
	if !ComparePassword(hash1, password) {
		t.Error("First generated hash should verify the original password")
	}
	if !ComparePassword(hash2, password) {
		t.Error("Second generated hash should verify the original password")
	}
	if !ComparePassword(hash3, password) {
		t.Error("Third generated hash should verify the original password")
	}
}

func TestComparePassword_InvalidHash(t *testing.T) {
	tests := []struct {
		name           string
		hashedPassword string
		password       string
		expected       bool
	}{
		{
			name:           "invalid hash format",
			hashedPassword: "invalidhash",
			password:       "password",
			expected:       false,
		},
		{
			name:           "malformed bcrypt hash",
			hashedPassword: "$2a$10$invalid",
			password:       "password",
			expected:       false,
		},
		{
			name:           "empty hash",
			hashedPassword: "",
			password:       "password",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePassword(tt.hashedPassword, tt.password)
			if result != tt.expected {
				t.Errorf("ComparePassword() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkGeneratePassword(b *testing.B) {
	password := "MySecurePassword@2024!"
	for i := 0; i < b.N; i++ {
		GeneratePassword(password)
	}
}

func BenchmarkComparePassword(b *testing.B) {
	password := "MySecurePassword@2024!"
	hash := GeneratePassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComparePassword(hash, password)
	}
}
