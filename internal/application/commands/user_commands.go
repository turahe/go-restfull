package commands

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserCommand represents a command to update a user
type UpdateUserCommand struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Username *string   `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string   `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string   `json:"phone,omitempty"`
}

// ChangePasswordCommand represents a command to change user password
type ChangePasswordCommand struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	OldPassword string    `json:"old_password" validate:"required"`
	NewPassword string    `json:"new_password" validate:"required,min=8"`
}

// VerifyEmailCommand represents a command to verify user email
type VerifyEmailCommand struct {
	UserID           uuid.UUID `json:"user_id" validate:"required"`
	VerificationCode string    `json:"verification_code" validate:"required"`
}

// VerifyPhoneCommand represents a command to verify user phone
type VerifyPhoneCommand struct {
	UserID           uuid.UUID `json:"user_id" validate:"required"`
	VerificationCode string    `json:"verification_code" validate:"required"`
}

// AssignRoleCommand represents a command to assign a role to a user
type AssignRoleCommand struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// RemoveRoleCommand represents a command to remove a role from a user
type RemoveRoleCommand struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// UpdateUserProfileCommand represents a command to update user profile
type UpdateUserProfileCommand struct {
	UserID      uuid.UUID  `json:"user_id" validate:"required"`
	FirstName   string     `json:"first_name" validate:"required,min=1,max=50"`
	LastName    string     `json:"last_name" validate:"required,min=1,max=50"`
	Avatar      *string    `json:"avatar,omitempty"`
	Bio         *string    `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website     *string    `json:"website,omitempty" validate:"omitempty,url,max=255"`
	Location    *string    `json:"location,omitempty" validate:"omitempty,max=100"`
	Gender      *string    `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
}

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}