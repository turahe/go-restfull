package requests

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/interfaces/http/validation"
)

// RegisterRequest represents the request for user registration
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=32"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"required"` // Full phone number with country code (e.g., +1234567890)
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

	// Validate phone number with country code parsing
	validator.ValidateRequired("phone", r.Phone)
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			validator.ValidateCustom("phone", "invalid phone number: "+err.Error())
		}
	}

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

	// Validate phone number with country code parsing
	validator.ValidateRequired("phone", r.Phone)
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			validator.ValidateCustom("phone", "invalid phone number: "+err.Error())
		}
	}

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

		// Check if phone already exists (using normalized phone number)
		normalizedPhone, err := r.GetNormalizedPhone()
		if err == nil {
			exists, err = userRepo.ExistsByPhone(ctx, normalizedPhone)
			validator.ValidateUnique("phone", r.Phone, exists, err)
		}
	}

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ParsePhone parses the phone number and returns the phone value object
func (r *RegisterRequest) ParsePhone() (*valueobjects.Phone, error) {
	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// GetNormalizedPhone returns the normalized phone number string
func (r *RegisterRequest) GetNormalizedPhone() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.String(), nil
}

// GetPhoneCountryCode returns the country code from the phone number
func (r *RegisterRequest) GetPhoneCountryCode() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.CountryCode(), nil
}

// GetPhoneNationalNumber returns the national number from the phone number
func (r *RegisterRequest) GetPhoneNationalNumber() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.NationalNumber(), nil
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

// ForgetPasswordRequest represents the request for forgetting password
type ForgetPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Can be username, email, or phone
}

// Validate validates the ForgetPasswordRequest with Laravel-style error responses
func (r *ForgetPasswordRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate identifier
	validator.ValidateRequired("identifier", r.Identifier)

	// If identifier looks like a phone number, validate it
	if r.Identifier != "" && (strings.HasPrefix(r.Identifier, "+") || regexp.MustCompile(`^\d{10,15}$`).MatchString(r.Identifier)) {
		if _, err := valueobjects.NewPhone(r.Identifier); err != nil {
			validator.ValidateCustom("identifier", "invalid phone number format")
		}
	}

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ResetPasswordRequest represents the request for resetting password
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
