// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
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

// RegisterRequest represents the request for user registration.
// This struct defines all required fields for creating a new user account,
// including comprehensive validation for username, email, phone, and password
// with confirmation. The request supports international phone numbers and
// follows security best practices for password requirements.
type RegisterRequest struct {
	// Username is the unique identifier for the user account (required, 3-32 characters)
	Username string `json:"username" validate:"required,min=3,max=32"`
	// Email is the user's email address for authentication and communication (required, must be valid email format)
	Email string `json:"email" validate:"required,email"`
	// Phone is the user's phone number with country code (required, e.g., +1234567890)
	Phone string `json:"phone" validate:"required"` // Full phone number with country code (e.g., +1234567890)
	// Password is the user's secure password for account access (required, 8-32 characters)
	Password string `json:"password" validate:"required,min=8,max=32"`
	// ConfirmPassword is the password confirmation to prevent typos (required, must match Password)
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the RegisterRequest with Laravel-style error responses.
// This method performs comprehensive validation including field requirements,
// format validation, and business rule enforcement using a custom validator.
//
// Validation Rules:
// - Username: required, 3-32 characters, valid username format
// - Email: required, valid email format
// - Phone: required, valid international phone format
// - Password: required, 8-32 characters
// - ConfirmPassword: required, must match Password
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *RegisterRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate username with length and format requirements
	validator.ValidateRequired("username", r.Username)
	if r.Username != "" {
		validator.ValidateBetween("username", r.Username, 3, 32)
		validator.ValidateUsername("username", r.Username)
	}

	// Validate email format
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate phone number with country code parsing
	validator.ValidateRequired("phone", r.Phone)
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			validator.ValidateCustom("phone", "invalid phone number: "+err.Error())
		}
	}

	// Validate password length requirements
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation matches
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	// Return validation results
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ValidateWithDatabase validates the RegisterRequest and checks database uniqueness with Laravel-style errors.
// This method extends basic validation by performing database checks to ensure
// username, email, and phone number uniqueness across the system.
//
// Parameters:
//   - ctx: Context for database operations
//   - userRepo: User repository interface for database queries
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *RegisterRequest) ValidateWithDatabase(ctx context.Context, userRepo repositories.UserRepository) (*responses.ValidationErrorBuilder, error) {
	// First, do basic validation
	validator := validation.NewValidator()

	// Validate username with length and format requirements
	validator.ValidateRequired("username", r.Username)
	if r.Username != "" {
		validator.ValidateBetween("username", r.Username, 3, 32)
		validator.ValidateUsername("username", r.Username)
	}

	// Validate email format
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate phone number with country code parsing
	validator.ValidateRequired("phone", r.Phone)
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			validator.ValidateCustom("phone", "invalid phone number: "+err.Error())
		}
	}

	// Validate password length requirements
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation matches
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	// Check database uniqueness if basic validation passes
	if !validator.HasErrors() {
		// Check if username already exists in the system
		exists, err := userRepo.ExistsByUsername(ctx, r.Username)
		validator.ValidateUnique("username", r.Username, exists, err)

		// Check if email already exists in the system
		exists, err = userRepo.ExistsByEmail(ctx, r.Email)
		validator.ValidateUnique("email", r.Email, exists, err)

		// Check if phone already exists using normalized phone number
		normalizedPhone, err := r.GetNormalizedPhone()
		if err == nil {
			exists, err = userRepo.ExistsByPhone(ctx, normalizedPhone)
			validator.ValidateUnique("phone", r.Phone, exists, err)
		}
	}

	// Return validation results
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ParsePhone parses the phone number string and returns a Phone value object.
// This method validates the phone number format and creates a structured
// representation for further processing and validation.
//
// Returns:
//   - *valueobjects.Phone: The parsed phone value object
//   - error: Any error that occurred during phone parsing
func (r *RegisterRequest) ParsePhone() (*valueobjects.Phone, error) {
	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// GetNormalizedPhone returns the normalized phone number string.
// This method parses the phone number and returns it in a standardized format
// suitable for storage, comparison, and uniqueness checks.
//
// Returns:
//   - string: The normalized phone number string
//   - error: Any error that occurred during phone parsing
func (r *RegisterRequest) GetNormalizedPhone() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.String(), nil
}

// GetPhoneCountryCode returns the country code from the phone number.
// This method extracts the international country code (e.g., "+1" for US/Canada)
// from the provided phone number for regional processing.
//
// Returns:
//   - string: The country code (e.g., "+1", "+44", "+62")
//   - error: Any error that occurred during phone parsing
func (r *RegisterRequest) GetPhoneCountryCode() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.CountryCode(), nil
}

// GetPhoneNationalNumber returns the national number from the phone number.
// This method extracts the local phone number without the country code
// (e.g., "555-1234" from "+1-555-1234") for local processing.
//
// Returns:
//   - string: The national phone number without country code
//   - error: Any error that occurred during phone parsing
func (r *RegisterRequest) GetPhoneNationalNumber() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.NationalNumber(), nil
}

// RefreshTokenRequest represents the request for refreshing access tokens.
// This struct handles the refresh token flow for maintaining user sessions
// without requiring re-authentication.
type RefreshTokenRequest struct {
	// RefreshToken is the refresh token used to obtain a new access token (required)
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Validate validates the RefreshTokenRequest with Laravel-style error responses.
// This method ensures the refresh token is provided for the token refresh operation.
//
// Validation Rules:
// - RefreshToken is required
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *RefreshTokenRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate refresh token presence
	validator.ValidateRequired("refresh_token", r.RefreshToken)

	// Return validation results
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ForgetPasswordRequest represents the request for initiating password reset.
// This struct supports flexible identification using username, email, or phone number,
// allowing users to reset passwords through their preferred contact method.
type ForgetPasswordRequest struct {
	// Identifier can be username, email, or phone number for password reset (required)
	Identifier string `json:"identifier" validate:"required"` // Can be username, email, or phone
}

// Validate validates the ForgetPasswordRequest with Laravel-style error responses.
// This method validates the identifier and performs format validation for phone numbers
// when the identifier appears to be a phone number.
//
// Validation Rules:
// - Identifier is required
// - If identifier looks like a phone number, it must be valid format
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *ForgetPasswordRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate identifier presence
	validator.ValidateRequired("identifier", r.Identifier)

	// If identifier looks like a phone number, validate its format
	if r.Identifier != "" && (strings.HasPrefix(r.Identifier, "+") || regexp.MustCompile(`^\d{10,15}$`).MatchString(r.Identifier)) {
		if _, err := valueobjects.NewPhone(r.Identifier); err != nil {
			validator.ValidateCustom("identifier", "invalid phone number format")
		}
	}

	// Return validation results
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}

// ResetPasswordRequest represents the request for completing password reset.
// This struct handles the final step of password reset using OTP verification
// and new password confirmation for security.
type ResetPasswordRequest struct {
	// Email is the user's email address for verification (required, must be valid email format)
	Email string `json:"email" validate:"required,email"`
	// OTP is the one-time password for verification (required)
	OTP string `json:"otp" validate:"required"`
	// Password is the new password to set (required, 8-32 characters)
	Password string `json:"password" validate:"required,min=8,max=32"`
	// ConfirmPassword is the password confirmation to prevent typos (required, must match Password)
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate validates the ResetPasswordRequest with Laravel-style error responses.
// This method ensures all required fields are provided and validates the
// password requirements and confirmation matching.
//
// Validation Rules:
// - Email is required and must be valid format
// - OTP is required for verification
// - Password is required, 8-32 characters
// - ConfirmPassword is required and must match Password
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *ResetPasswordRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate email format
	validator.ValidateRequired("email", r.Email)
	validator.ValidateEmail("email", r.Email)

	// Validate OTP presence
	validator.ValidateRequired("otp", r.OTP)

	// Validate password length requirements
	validator.ValidateRequired("password", r.Password)
	if r.Password != "" {
		validator.ValidateBetween("password", r.Password, 8, 32)
	}

	// Validate password confirmation matches
	validator.ValidateRequired("confirm_password", r.ConfirmPassword)
	validator.ValidateConfirmed("password", r.Password, r.ConfirmPassword)

	// Return validation results
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}
