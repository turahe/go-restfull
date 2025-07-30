package services_test

import (
	"testing"

	"github.com/turahe/go-restfull/internal/infrastructure/adapters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptPasswordService_HashPassword(t *testing.T) {
	service := adapters.NewBcryptPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "Complex password",
			password: "P@ssw0rd!123",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "verylongpasswordwithlotsofcharacters123456789",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := service.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
				assert.NotEqual(t, tt.password, hashedPassword)
				assert.Contains(t, hashedPassword, "$2a$") // bcrypt prefix
			}
		})
	}
}

func TestBcryptPasswordService_CheckPassword(t *testing.T) {
	service := adapters.NewBcryptPasswordService()

	tests := []struct {
		name           string
		password       string
		hashedPassword string
		wantMatch      bool
		wantErr        bool
	}{
		{
			name:           "Valid password match",
			password:       "password",
			hashedPassword: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			wantMatch:      true,
			wantErr:        false,
		},
		{
			name:           "Invalid password",
			password:       "wrongpassword",
			hashedPassword: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
			wantMatch:      false,
			wantErr:        false,
		},
		{
			name:           "Empty password",
			password:       "",
			hashedPassword: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			wantMatch:      false,
			wantErr:        false,
		},
		{
			name:           "Invalid hash format",
			password:       "password123",
			hashedPassword: "invalidhash",
			wantMatch:      false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := service.ComparePassword(tt.hashedPassword, tt.password)

			if tt.wantErr {
				assert.False(t, match)
			} else {
				assert.Equal(t, tt.wantMatch, match)
			}
		})
	}
}

func TestBcryptPasswordService_Integration(t *testing.T) {
	service := adapters.NewBcryptPasswordService()

	// Test that hashed password can be verified
	password := "testpassword123"

	hashedPassword, err := service.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// Verify the password
	match := service.ComparePassword(hashedPassword, password)
	assert.True(t, match)

	// Verify wrong password doesn't match
	match = service.ComparePassword(hashedPassword, "wrongpassword")
	assert.False(t, match)
}

func TestBcryptPasswordService_ConsistentHashing(t *testing.T) {
	service := adapters.NewBcryptPasswordService()
	password := "testpassword123"

	// Hash the same password multiple times
	hash1, err := service.HashPassword(password)
	require.NoError(t, err)

	hash2, err := service.HashPassword(password)
	require.NoError(t, err)

	// Hashes should be different (bcrypt uses random salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should verify correctly
	match1 := service.ComparePassword(hash1, password)
	assert.True(t, match1)

	match2 := service.ComparePassword(hash2, password)
	assert.True(t, match2)
}

func TestBcryptPasswordService_ValidatePassword(t *testing.T) {
	service := adapters.NewBcryptPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "Too short password",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "Password without uppercase",
			password: "password123!",
			wantErr:  true,
		},
		{
			name:     "Password without lowercase",
			password: "PASSWORD123!",
			wantErr:  true,
		},
		{
			name:     "Password without number",
			password: "Password!",
			wantErr:  true,
		},
		{
			name:     "Password without special character",
			password: "Password123",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
