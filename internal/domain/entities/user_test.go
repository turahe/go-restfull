package entities_test

import (
	"testing"
	"time"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		phone    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid user creation",
			username: "testuser",
			email:    "test@example.com",
			phone:    "+1234567890",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Empty username",
			username: "",
			email:    "test@example.com",
			phone:    "+1234567890",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "Empty email",
			username: "testuser",
			email:    "",
			phone:    "+1234567890",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "Empty phone",
			username: "testuser",
			email:    "test@example.com",
			phone:    "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			username: "testuser",
			email:    "test@example.com",
			phone:    "+1234567890",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := entities.NewUser(tt.username, tt.email, tt.phone, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.UserName)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.phone, user.Phone)
				assert.Equal(t, tt.password, user.Password)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
			}
		})
	}
}

func TestUser_UpdateUser(t *testing.T) {
	user, err := entities.NewUser("olduser", "old@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	tests := []struct {
		name     string
		username string
		email    string
		phone    string
	}{
		{
			name:     "Update all fields",
			username: "newuser",
			email:    "new@example.com",
			phone:    "+0987654321",
		},
		{
			name:     "Update only username",
			username: "newuser2",
			email:    "",
			phone:    "",
		},
		{
			name:     "Update only email",
			username: "",
			email:    "new2@example.com",
			phone:    "",
		},
		{
			name:     "Update only phone",
			username: "",
			email:    "",
			phone:    "+1111111111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.UpdateUser(tt.username, tt.email, tt.phone)
			assert.NoError(t, err)

			if tt.username != "" {
				assert.Equal(t, tt.username, user.UserName)
			}
			if tt.email != "" {
				assert.Equal(t, tt.email, user.Email)
			}
			if tt.phone != "" {
				assert.Equal(t, tt.phone, user.Phone)
			}

			assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
		})
	}
}

func TestUser_ChangePassword(t *testing.T) {

	tests := []struct {
		name        string
		newPassword string
		wantErr     bool
	}{
		{
			name:        "Valid password change",
			newPassword: "newpassword123",
			wantErr:     false,
		},
		{
			name:        "Empty password",
			newPassword: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh user for each test case
			testUser, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "oldpassword")
			require.NoError(t, err)
			require.NotNil(t, testUser)

			originalUpdatedAt := testUser.UpdatedAt

			// Wait a bit to ensure timestamp difference
			time.Sleep(1 * time.Millisecond)

			err = testUser.ChangePassword(tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "oldpassword", testUser.Password)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newPassword, testUser.Password)
				assert.True(t, testUser.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func TestUser_VerifyEmail(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	user.VerifyEmail()

	assert.NotNil(t, user.EmailVerifiedAt)
	assert.True(t, user.EmailVerifiedAt.After(originalUpdatedAt))
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsEmailVerified())
}

func TestUser_VerifyPhone(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	user.VerifyPhone()

	assert.NotNil(t, user.PhoneVerifiedAt)
	assert.True(t, user.PhoneVerifiedAt.After(originalUpdatedAt))
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsPhoneVerified())
}

func TestUser_SoftDelete(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	user.SoftDelete()

	assert.NotNil(t, user.DeletedAt)
	assert.True(t, user.DeletedAt.After(originalUpdatedAt))
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsDeleted())
}

func TestUser_IsDeleted(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Initially not deleted
	assert.False(t, user.IsDeleted())

	// After soft delete
	user.SoftDelete()
	assert.True(t, user.IsDeleted())
}

func TestUser_IsEmailVerified(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Initially not verified
	assert.False(t, user.IsEmailVerified())

	// After verification
	user.VerifyEmail()
	assert.True(t, user.IsEmailVerified())
}

func TestUser_IsPhoneVerified(t *testing.T) {
	user, err := entities.NewUser("testuser", "test@example.com", "+1234567890", "password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Initially not verified
	assert.False(t, user.IsPhoneVerified())

	// After verification
	user.VerifyPhone()
	assert.True(t, user.IsPhoneVerified())
}
