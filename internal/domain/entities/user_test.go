package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser_Success(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	phone := "+1234567890"
	password := "password123"

	user, err := entities.NewUser(username, email, phone, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.UserName)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, phone, user.Phone)
	assert.Equal(t, password, user.Password)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
	assert.Nil(t, user.DeletedAt)
	assert.Nil(t, user.EmailVerifiedAt)
	assert.Nil(t, user.PhoneVerifiedAt)
	assert.Empty(t, user.Roles)
}

func TestNewUser_EmptyUsername(t *testing.T) {
	email := "test@example.com"
	phone := "+1234567890"
	password := "password123"

	user, err := entities.NewUser("", email, phone, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "username is required", err.Error())
}

func TestNewUser_EmptyEmail(t *testing.T) {
	username := "testuser"
	phone := "+1234567890"
	password := "password123"

	user, err := entities.NewUser(username, "", phone, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is required", err.Error())
}

func TestNewUser_EmptyPhone(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := entities.NewUser(username, email, "", password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "phone is required", err.Error())
}

func TestNewUser_EmptyPassword(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	phone := "+1234567890"

	user, err := entities.NewUser(username, email, phone, "")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password is required", err.Error())
}

func TestUser_UpdateUser(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := user.UpdateUser("newuser", "new@example.com", "newphone")

	assert.NoError(t, err)
	assert.Equal(t, "newuser", user.UserName)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "newphone", user.Phone)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
}

func TestUser_UpdateUser_PartialUpdate(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalUsername := user.UserName

	err := user.UpdateUser("", "new@example.com", "")

	assert.NoError(t, err)
	assert.Equal(t, originalUsername, user.UserName) // Should remain unchanged
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "oldphone", user.Phone) // Should remain unchanged
}

func TestUser_UpdateUser_EmptyStrings(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalUsername := user.UserName
	originalEmail := user.Email
	originalPhone := user.Phone

	err := user.UpdateUser("", "", "")

	assert.NoError(t, err)
	assert.Equal(t, originalUsername, user.UserName)
	assert.Equal(t, originalEmail, user.Email)
	assert.Equal(t, originalPhone, user.Phone)
}

func TestUser_ChangePassword(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "oldpassword")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := user.ChangePassword("newpassword")

	assert.NoError(t, err)
	assert.Equal(t, "newpassword", user.Password)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
}

func TestUser_ChangePassword_EmptyPassword(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "oldpassword")
	originalPassword := user.Password

	err := user.ChangePassword("")

	assert.Error(t, err)
	assert.Equal(t, "new password is required", err.Error())
	assert.Equal(t, originalPassword, user.Password) // Should remain unchanged
}

func TestUser_VerifyEmail(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	user.VerifyEmail()

	assert.NotNil(t, user.EmailVerifiedAt)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsEmailVerified())
}

func TestUser_VerifyPhone(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	user.VerifyPhone()

	assert.NotNil(t, user.PhoneVerifiedAt)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsPhoneVerified())
}

func TestUser_SoftDelete(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")
	originalUpdatedAt := user.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	user.SoftDelete()

	assert.NotNil(t, user.DeletedAt)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, user.IsDeleted())
}

func TestUser_IsDeleted(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")

	// Initially not deleted
	assert.False(t, user.IsDeleted())

	// After soft delete
	user.SoftDelete()
	assert.True(t, user.IsDeleted())
}

func TestUser_IsEmailVerified(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")

	// Initially not verified
	assert.False(t, user.IsEmailVerified())

	// After verification
	user.VerifyEmail()
	assert.True(t, user.IsEmailVerified())
}

func TestUser_IsPhoneVerified(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")

	// Initially not verified
	assert.False(t, user.IsPhoneVerified())

	// After verification
	user.VerifyPhone()
	assert.True(t, user.IsPhoneVerified())
}

func TestUser_VerificationFlow(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")

	// Initially not verified
	assert.False(t, user.IsEmailVerified())
	assert.False(t, user.IsPhoneVerified())

	// Verify email
	user.VerifyEmail()
	assert.True(t, user.IsEmailVerified())
	assert.False(t, user.IsPhoneVerified())

	// Verify phone
	user.VerifyPhone()
	assert.True(t, user.IsEmailVerified())
	assert.True(t, user.IsPhoneVerified())
}

func TestUser_SoftDelete_MultipleCalls(t *testing.T) {
	user, _ := entities.NewUser("testuser", "test@example.com", "phone", "password")

	// First soft delete
	user.SoftDelete()
	firstDeletedAt := user.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	user.SoftDelete()
	secondDeletedAt := user.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, user.IsDeleted())
}

func TestUser_UpdateUser_OnlyUsername(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalEmail := user.Email
	originalPhone := user.Phone

	err := user.UpdateUser("newuser", "", "")

	assert.NoError(t, err)
	assert.Equal(t, "newuser", user.UserName)
	assert.Equal(t, originalEmail, user.Email) // Should remain unchanged
	assert.Equal(t, originalPhone, user.Phone) // Should remain unchanged
}

func TestUser_UpdateUser_OnlyEmail(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalUsername := user.UserName
	originalPhone := user.Phone

	err := user.UpdateUser("", "new@example.com", "")

	assert.NoError(t, err)
	assert.Equal(t, originalUsername, user.UserName) // Should remain unchanged
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, originalPhone, user.Phone) // Should remain unchanged
}

func TestUser_UpdateUser_OnlyPhone(t *testing.T) {
	user, _ := entities.NewUser("olduser", "old@example.com", "oldphone", "password")
	originalUsername := user.UserName
	originalEmail := user.Email

	err := user.UpdateUser("", "", "newphone")

	assert.NoError(t, err)
	assert.Equal(t, originalUsername, user.UserName) // Should remain unchanged
	assert.Equal(t, originalEmail, user.Email)       // Should remain unchanged
	assert.Equal(t, "newphone", user.Phone)
}
