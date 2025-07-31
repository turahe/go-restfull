package requests

import (
	"context"
	"errors"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/interfaces/http/validation"
)

// RegisterRequest represents the request for user registration
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=32"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the RegisterRequest with Laravel-style error responses
func (r *RegisterRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate username
	validator.ValidateRequired("username", r.Username)
	if r.Username != "" {
		validator.ValidateBetween("username", r.Username, 3, 32)
		validator.ValidateUsername("username", r.Username)
	}

	// Validate email
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate phone
	validator.ValidateRequired("phone", r.Phone)
	validator.ValidatePhone("phone", r.Phone)

	// Validate password
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ValidateWithDatabase validates the RegisterRequest and checks database uniqueness with Laravel-style errors
func (r *RegisterRequest) ValidateWithDatabase(ctx context.Context, userRepo repositories.UserRepository) (*responses.ValidationErrorBuilder, error) {
	// First, do basic validation
	validator := validation.NewValidator()

	// Validate username
	validator.ValidateRequired("username", r.Username)
	if r.Username != "" {
		validator.ValidateBetween("username", r.Username, 3, 32)
		validator.ValidateUsername("username", r.Username)
	}

	// Validate email
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate phone
	validator.ValidateRequired("phone", r.Phone)
	validator.ValidatePhone("phone", r.Phone)

	// Validate password
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	// Check database uniqueness if basic validation passes
	if !validator.HasErrors() {
		// Check if username already exists
		exists, err := userRepo.ExistsByUsername(ctx, r.Username)
		validator.ValidateUnique("username", r.Username, exists, err)

		// Check if email already exists
		exists, err = userRepo.ExistsByEmail(ctx, r.Email)
		validator.ValidateUnique("email", r.Email, exists, err)

		// Check if phone already exists
		exists, err = userRepo.ExistsByPhone(ctx, r.Phone)
		validator.ValidateUnique("phone", r.Phone, exists, err)
	}

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// RefreshTokenRequest represents the request for refreshing access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Validate validates the RefreshTokenRequest with Laravel-style error responses
func (r *RefreshTokenRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate refresh token
	validator.ValidateRequired("refresh_token", r.RefreshToken)

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ForgetPasswordRequest represents the request for password reset
type ForgetPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Can be username, email, or phone
}

// Validate validates the ForgetPasswordRequest with Laravel-style error responses
func (r *ForgetPasswordRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate identifier is required
	validator.ValidateRequired("identifier", r.Identifier)

	// Validate that the identifier is either a valid email, username, or phone number
	if r.Identifier != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

		isEmail := emailRegex.MatchString(r.Identifier)
		isUsername := usernameRegex.MatchString(r.Identifier)
		isPhone := phoneRegex.MatchString(r.Identifier)

		if !isEmail && !isUsername && !isPhone {
			validator.ValidateCustom("identifier", "The identifier must be a valid email, username, or phone number.")
		}
	}

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ResetPasswordRequest represents the request for password reset with OTP
type ResetPasswordRequest struct {
	Email           string `json:"email" validate:"required,email"`
	OTP             string `json:"otp" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=32"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the ResetPasswordRequest with Laravel-style error responses
func (r *ResetPasswordRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate email
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate OTP
	validator.ValidateRequired("otp", r.OTP)

	// Validate password
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}
