// Package entities provides the core domain models and business logic entities
// for the application. This file contains the User entity for managing
// user accounts with authentication, verification, and role-based access control.
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents the core user domain entity that manages user accounts,
// authentication credentials, and access control within the system.
//
// The entity includes:
// - User identification (username, email, phone)
// - Authentication (password with JSON exclusion)
// - Verification status (email and phone verification)
// - Access control (roles and menu permissions)
// - Audit trail with creation, update, and deletion tracking
// - Soft delete functionality for account preservation
type User struct {
	ID              uuid.UUID  `json:"id"`                          // Unique identifier for the user
	UserName        string     `json:"username"`                    // Unique username for login and identification
	Email           string     `json:"email"`                       // User's email address for communication
	Phone           string     `json:"phone"`                       // User's phone number for contact
	Password        string     `json:"-"`                           // Hashed password (excluded from JSON for security)
	Avatar          string     `json:"avatar,omitempty"`            // User's avatar URL (optional)
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"` // Timestamp when email was verified
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"` // Timestamp when phone was verified
	Roles           []*Role    `json:"roles,omitempty"`             // Collection of roles assigned to the user
	Menus           []*Menu    `json:"menus,omitempty"`             // Collection of menus accessible to the user
	CreatedBy       uuid.UUID  `json:"created_by"`                  // ID of user who created this account
	UpdatedBy       uuid.UUID  `json:"updated_by"`                  // ID of user who last updated this account
	DeletedBy       *uuid.UUID `json:"deleted_by,omitempty"`        // ID of user who deleted this account (soft delete)
	CreatedAt       time.Time  `json:"created_at"`                  // Timestamp when user account was created
	UpdatedAt       time.Time  `json:"updated_at"`                  // Timestamp when user account was last updated
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`        // Timestamp when user account was soft deleted
}

// NewUser creates a new user with validation.
// This constructor validates required fields and initializes the user
// with generated UUID and timestamps.
//
// Parameters:
//   - username: Unique username for login (required)
//   - email: User's email address (required)
//   - phone: User's phone number (required)
//   - password: User's password (required)
//
// Returns:
//   - *User: Pointer to the newly created user entity
//   - error: Validation error if any required field is empty
//
// Validation rules:
// - username, email, phone, and password cannot be empty
//
// Security note: Password should be hashed before storage in production
func NewUser(username, email, phone, password string) (*User, error) {
	// Validate required fields
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

	// Create user with current timestamp
	now := time.Now()
	return &User{
		ID:        uuid.New(), // Generate new unique identifier
		UserName:  username,   // Set username
		Email:     email,      // Set email address
		Phone:     phone,      // Set phone number
		Password:  password,   // Set password (should be hashed before storage)
		CreatedAt: now,        // Set creation timestamp
		UpdatedAt: now,        // Set initial update timestamp
	}, nil
}

// UpdateUser updates user information with new values.
// This method allows partial updates - only non-empty values are updated.
// The UpdatedAt timestamp is automatically updated when any field changes.
//
// Parameters:
//   - username: New username (optional, only updated if not empty)
//   - email: New email address (optional, only updated if not empty)
//   - phone: New phone number (optional, only updated if not empty)
//
// Returns:
//   - error: Always nil, included for interface consistency
//
// Note: This method automatically updates the UpdatedAt timestamp
func (u *User) UpdateUser(username, email, phone string) error {
	// Update fields only if new values are provided
	if username != "" {
		u.UserName = username
	}
	if email != "" {
		u.Email = email
	}
	if phone != "" {
		u.Phone = phone
	}

	// Update modification timestamp
	u.UpdatedAt = time.Now()
	return nil
}

// ChangePassword updates the user's password.
// This method validates the new password and updates the UpdatedAt timestamp.
//
// Parameters:
//   - newPassword: New password for the user
//
// Returns:
//   - error: Validation error if new password is empty
//
// Security note: New password should be hashed before calling this method
func (u *User) ChangePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("new password is required")
	}

	u.Password = newPassword // Update password
	u.UpdatedAt = time.Now() // Update modification timestamp
	return nil
}

// VerifyEmail marks the user's email address as verified.
// This method sets the EmailVerifiedAt timestamp to the current time
// and updates the UpdatedAt timestamp.
//
// Note: This method automatically updates both EmailVerifiedAt and UpdatedAt timestamps
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerifiedAt = &now // Set email verification timestamp
	u.UpdatedAt = now        // Update modification timestamp
}

// VerifyPhone marks the user's phone number as verified.
// This method sets the PhoneVerifiedAt timestamp to the current time
// and updates the UpdatedAt timestamp.
//
// Note: This method automatically updates both PhoneVerifiedAt and UpdatedAt timestamps
func (u *User) VerifyPhone() {
	now := time.Now()
	u.PhoneVerifiedAt = &now // Set phone verification timestamp
	u.UpdatedAt = now        // Update modification timestamp
}

// SoftDelete marks the user account as deleted without removing it from the database.
// This sets the DeletedAt timestamp and updates the UpdatedAt timestamp.
// The user account will be excluded from normal queries but remains accessible
// for audit and recovery purposes.
//
// Note: This method automatically updates both DeletedAt and UpdatedAt timestamps
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now // Set deletion timestamp
	u.UpdatedAt = now  // Update modification timestamp
}

// IsDeleted checks if the user account is soft deleted
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
