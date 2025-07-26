package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents the core user domain entity
type User struct {
	ID              uuid.UUID  `json:"id"`
	UserName        string     `json:"username"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	Password        string     `json:"-"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new user with validation
func NewUser(username, email, phone, password string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if phone == "" {
		return nil, errors.New("phone is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	now := time.Now()
	return &User{
		ID:        uuid.New(),
		UserName:  username,
		Email:     email,
		Phone:     phone,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateUser updates user information
func (u *User) UpdateUser(username, email, phone string) error {
	if username != "" {
		u.UserName = username
	}
	if email != "" {
		u.Email = email
	}
	if phone != "" {
		u.Phone = phone
	}
	u.UpdatedAt = time.Now()
	return nil
}

// ChangePassword updates the user's password
func (u *User) ChangePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("new password is required")
	}
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	return nil
}

// VerifyEmail marks the user's email as verified
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// VerifyPhone marks the user's phone as verified
func (u *User) VerifyPhone() {
	now := time.Now()
	u.PhoneVerifiedAt = &now
	u.UpdatedAt = now
}

// SoftDelete marks the user as deleted
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
}

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsEmailVerified checks if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsPhoneVerified checks if the user's phone is verified
func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil
} 