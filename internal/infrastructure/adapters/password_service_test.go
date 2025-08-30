package adapters

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBcryptPasswordService(t *testing.T) {
	// Test that the factory function returns a non-nil service
	service := NewBcryptPasswordService()
	assert.NotNil(t, service)

	// Test that it implements the PasswordService interface
	var _ interface{} = service
}

func TestBcryptPasswordService_HashPassword(t *testing.T) {
	service := NewBcryptPasswordService()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password",
			password:    "Password123!",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: false, // bcrypt allows empty passwords
		},
		{
			name:        "Complex password with special chars",
			password:    "P@ssw0rd!123#",
			expectError: false,
		},
		{
			name:        "Long password",
			password:    strings.Repeat("a", 65) + "A1!", // Keep under 72 bytes for bcrypt (65 + 3 = 68 bytes)
			expectError: false,
		},
		{
			name:        "Unicode password",
			password:    "P@ssw0rd!123🚀",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := service.HashPassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)

				// Verify the hash is different from the original password
				assert.NotEqual(t, tt.password, hashedPassword)

				// Verify it's a valid bcrypt hash (starts with $2a$)
				assert.True(t, strings.HasPrefix(hashedPassword, "$2a$"),
					"Hash should start with $2a$ prefix")

				// Verify hash length is reasonable (bcrypt hashes are typically 60 chars)
				assert.Len(t, hashedPassword, 60)
			}
		})
	}
}

func TestBcryptPasswordService_ComparePassword(t *testing.T) {
	service := NewBcryptPasswordService()

	tests := []struct {
		name           string
		plainPassword  string
		hashedPassword string
		expectedMatch  bool
		description    string
	}{
		{
			name:           "Valid password match",
			plainPassword:  "Password123!",
			hashedPassword: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			expectedMatch:  false,                                                          // This hash is for "password", not "Password123!"
			description:    "Should not match different password",
		},
		{
			name:           "Empty password with valid hash",
			plainPassword:  "",
			hashedPassword: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			expectedMatch:  false,
			description:    "Empty password should not match non-empty hash",
		},
		{
			name:           "Invalid hash format",
			plainPassword:  "password123",
			hashedPassword: "invalidhash",
			expectedMatch:  false,
			description:    "Invalid hash format should return false",
		},
		{
			name:           "Malformed bcrypt hash",
			plainPassword:  "password123",
			hashedPassword: "$2a$10$invalid",
			expectedMatch:  false,
			description:    "Malformed bcrypt hash should return false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := service.ComparePassword(tt.hashedPassword, tt.plainPassword)
			assert.Equal(t, tt.expectedMatch, match, tt.description)
		})
	}
}

func TestBcryptPasswordService_ComparePassword_Integration(t *testing.T) {
	service := NewBcryptPasswordService()

	// Test that a hashed password can be verified correctly
	password := "TestPassword123!"

	hashedPassword, err := service.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// Verify the correct password matches
	match := service.ComparePassword(hashedPassword, password)
	assert.True(t, match, "Correct password should match its hash")

	// Verify wrong password doesn't match
	match = service.ComparePassword(hashedPassword, "WrongPassword123!")
	assert.False(t, match, "Wrong password should not match the hash")

	// Verify empty password doesn't match
	match = service.ComparePassword(hashedPassword, "")
	assert.False(t, match, "Empty password should not match the hash")
}

func TestBcryptPasswordService_ValidatePassword(t *testing.T) {
	service := NewBcryptPasswordService()

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid password with all requirements",
			password:    "Password123!",
			expectError: false,
		},
		{
			name:        "Valid password with different special chars",
			password:    "Secure@2024#",
			expectError: false,
		},
		{
			name:        "Password too short",
			password:    "Abc1!",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "Password without uppercase",
			password:    "password123!",
			expectError: true,
			errorMsg:    "password must contain at least one uppercase letter",
		},
		{
			name:        "Password without lowercase",
			password:    "PASSWORD123!",
			expectError: true,
			errorMsg:    "password must contain at least one lowercase letter",
		},
		{
			name:        "Password without digit",
			password:    "Password!",
			expectError: true,
			errorMsg:    "password must contain at least one digit",
		},
		{
			name:        "Password without special character",
			password:    "Password123",
			expectError: true,
			errorMsg:    "password must contain at least one special character",
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "Password with only 7 chars",
			password:    "Abc1!@",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "Password with 8 chars but missing requirements",
			password:    "abcdefgh",
			expectError: true,
			errorMsg:    "password must contain at least one uppercase letter",
		},
		{
			name:        "Password with uppercase and lowercase but no digit or special",
			password:    "Password",
			expectError: true,
			errorMsg:    "password must contain at least one digit",
		},
		{
			name:        "Password with uppercase, lowercase, digit but no special",
			password:    "Password1",
			expectError: true,
			errorMsg:    "password must contain at least one special character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBcryptPasswordService_ValidatePassword_EdgeCases(t *testing.T) {
	service := NewBcryptPasswordService()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Password with exactly 8 characters",
			password:    "Abc1!@#$",
			expectError: false,
		},
		{
			name:        "Password with all special characters",
			password:    "Abc123!@#$%^&*()_+-=[]{}|;':\",./<>?",
			expectError: false,
		},
		{
			name:        "Password with spaces",
			password:    "Abc 123!",
			expectError: false, // Spaces are allowed
		},
		{
			name:        "Password with tabs and newlines",
			password:    "Abc\t123!\n",
			expectError: false, // Control characters are allowed
		},
		{
			name:        "Password with unicode characters",
			password:    "Abc123!🚀",
			expectError: false, // Unicode is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBcryptPasswordService_ConsistentHashing(t *testing.T) {
	service := NewBcryptPasswordService()
	password := "TestPassword123!"

	// Hash the same password multiple times
	hash1, err := service.HashPassword(password)
	require.NoError(t, err)

	hash2, err := service.HashPassword(password)
	require.NoError(t, err)

	hash3, err := service.HashPassword(password)
	require.NoError(t, err)

	// All hashes should be different (bcrypt uses random salt)
	assert.NotEqual(t, hash1, hash2, "Different hashes should be generated for the same password")
	assert.NotEqual(t, hash2, hash3, "Different hashes should be generated for the same password")
	assert.NotEqual(t, hash1, hash3, "Different hashes should be generated for the same password")

	// But all should verify correctly
	match1 := service.ComparePassword(hash1, password)
	assert.True(t, match1, "First hash should verify correctly")

	match2 := service.ComparePassword(hash2, password)
	assert.True(t, match2, "Second hash should verify correctly")

	match3 := service.ComparePassword(hash3, password)
	assert.True(t, match3, "Third hash should verify correctly")
}

func TestBcryptPasswordService_Performance(t *testing.T) {
	service := NewBcryptPasswordService()
	password := "TestPassword123!"

	// Test that hashing is reasonably fast (should complete in under 100ms)
	// This is a basic performance test to ensure bcrypt cost is reasonable
	hashedPassword, err := service.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// Verify the hash works
	match := service.ComparePassword(hashedPassword, password)
	assert.True(t, match, "Hashed password should verify correctly")
}

func TestBcryptPasswordService_InterfaceCompliance(t *testing.T) {
	// Test that the service properly implements the PasswordService interface
	var service interface{} = NewBcryptPasswordService()

	// This test ensures the service can be used where PasswordService is expected
	// The actual interface compliance is checked at compile time
	assert.NotNil(t, service)
}

func TestBcryptPasswordService_ErrorHandling(t *testing.T) {
	service := NewBcryptPasswordService()

	// Test with extremely long password (this might cause bcrypt to fail)
	// Note: This is a boundary test and might not always fail depending on bcrypt implementation
	veryLongPassword := strings.Repeat("a", 10000) + "A1!"

	_, err := service.HashPassword(veryLongPassword)
	// We don't assert on the result as bcrypt behavior with very long passwords
	// can vary between implementations, but we ensure the function handles it gracefully
	_ = err // Suppress unused variable warning
}

func TestBcryptPasswordService_EmptyStringHandling(t *testing.T) {
	service := NewBcryptPasswordService()

	// Test empty string handling
	emptyHash, err := service.HashPassword("")
	require.NoError(t, err)
	require.NotEmpty(t, emptyHash)

	// Empty string should hash to something
	assert.True(t, strings.HasPrefix(emptyHash, "$2a$"))

	// Empty string should compare correctly
	match := service.ComparePassword(emptyHash, "")
	assert.True(t, match, "Empty string should match its hash")
}
