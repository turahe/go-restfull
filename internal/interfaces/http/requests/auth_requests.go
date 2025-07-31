package requests

import (
	"context"
	"errors"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/repositories"
)

// RegisterRequest represents the request for user registration
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=32"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the RegisterRequest (basic validation only)
func (r *RegisterRequest) Validate() error {
	if r.Username == "" {
		return errors.New("username is required")
	}
	if len(r.Username) < 3 || len(r.Username) > 32 {
		return errors.New("username must be between 3 and 32 characters")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Phone == "" {
		return errors.New("phone is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 8 || len(r.Password) > 32 {
		return errors.New("password must be between 8 and 32 characters")
	}
	if r.Password != r.ConfirmPassword {
		return errors.New("password confirmation does not match")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// Validate username format (alphanumeric and underscore only)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(r.Username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	// Validate phone format (international format)
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(r.Phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

// ValidateWithDatabase validates the RegisterRequest and checks database uniqueness
func (r *RegisterRequest) ValidateWithDatabase(ctx context.Context, userRepo repositories.UserRepository) error {
	// First, do basic validation
	if err := r.Validate(); err != nil {
		return err
	}

	// Check if username already exists
	exists, err := userRepo.ExistsByUsername(ctx, r.Username)
	if err != nil {
		return errors.New("failed to validate username")
	}
	if exists {
		return errors.New("username already exists")
	}

	// Check if email already exists
	exists, err = userRepo.ExistsByEmail(ctx, r.Email)
	if err != nil {
		return errors.New("failed to validate email")
	}
	if exists {
		return errors.New("email already exists")
	}

	// Check if phone already exists
	exists, err = userRepo.ExistsByPhone(ctx, r.Phone)
	if err != nil {
		return errors.New("failed to validate phone")
	}
	if exists {
		return errors.New("phone number already exists")
	}

	return nil
}

// RefreshTokenRequest represents the request for refreshing access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Validate validates the RefreshTokenRequest
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.New("refresh token is required")
	}
	return nil
}

// ForgetPasswordRequest represents the request for password reset
type ForgetPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Can be username, email, or phone
}

// Validate validates the ForgetPasswordRequest
func (r *ForgetPasswordRequest) Validate() error {
	if r.Identifier == "" {
		return errors.New("identifier is required")
	}

	// Validate that the identifier is either a valid email, username, or phone number
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	isEmail := emailRegex.MatchString(r.Identifier)
	isUsername := usernameRegex.MatchString(r.Identifier)
	isPhone := phoneRegex.MatchString(r.Identifier)

	if !isEmail && !isUsername && !isPhone {
		return errors.New("identifier must be a valid email, username, or phone number")
	}

	return nil
}

// ResetPasswordRequest represents the request for password reset with OTP
type ResetPasswordRequest struct {
	Email           string `json:"email" validate:"required,email"`
	OTP             string `json:"otp" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the ResetPasswordRequest
func (r *ResetPasswordRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.OTP == "" {
		return errors.New("OTP is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 8 || len(r.Password) > 32 {
		return errors.New("password must be between 8 and 32 characters")
	}
	if r.Password != r.ConfirmPassword {
		return errors.New("password confirmation does not match")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	return nil
}
