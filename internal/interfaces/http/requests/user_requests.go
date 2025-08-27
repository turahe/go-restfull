// Package requests provides HTTP request structures and validation logic for the REST API.
// This package contains request DTOs (Data Transfer Objects) that define the structure
// and validation rules for incoming HTTP requests. Each request type includes validation
// methods and transformation methods to convert requests to domain entities.
package requests

import (
	"errors"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/interfaces/http/validation"
)

// CreateUserRequest represents the request for creating a new user entity.
// This struct defines the required fields for user creation, including
// username, email, phone, and password. The request includes comprehensive
// validation for all fields to ensure data quality and security.
type CreateUserRequest struct {
	// Username is the unique identifier for the user account (required)
	Username string `json:"username"`
	// Email is the user's email address for authentication and communication (required)
	Email string `json:"email"`
	// Phone is the user's phone number with country code (required)
	Phone string `json:"phone"`
	// Password is the user's secure password for account access (required, minimum 8 characters)
	Password string `json:"password"`
}

// Validate performs comprehensive validation on the CreateUserRequest.
// This method checks all required fields, validates email format using regex,
// validates phone number format using the phone value object, and ensures
// password meets minimum security requirements.
//
// Validation Rules:
// - All fields are required
// - Email must match standard email format
// - Phone must be valid international format
// - Password must be at least 8 characters long
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *CreateUserRequest) Validate() error {
	// Check required fields
	if r.Username == "" {
		return errors.New("username is required")
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

	// Validate email format using regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// Validate phone number with country code parsing using value object
	if _, err := valueobjects.NewPhone(r.Phone); err != nil {
		return errors.New("invalid phone number: " + err.Error())
	}

	// Validate password strength (minimum length requirement)
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}

// ToEntity transforms the CreateUserRequest to a User domain entity.
// This method uses the domain entity's constructor to ensure proper
// initialization and validation of the user object.
//
// Returns:
//   - *entities.User: The created user entity
//   - error: Any error that occurred during entity creation
func (r *CreateUserRequest) ToEntity() (*entities.User, error) {
	// Create user entity using the existing constructor for proper initialization
	user, err := entities.NewUser(r.Username, r.Email, r.Phone, r.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ParsePhone parses the phone number string and returns a Phone value object.
// This method validates the phone number format and creates a structured
// representation for further processing.
//
// Returns:
//   - *valueobjects.Phone: The parsed phone value object
//   - error: Any error that occurred during phone parsing
func (r *CreateUserRequest) ParsePhone() (*valueobjects.Phone, error) {
	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// GetNormalizedPhone returns the normalized phone number string.
// This method parses the phone number and returns it in a standardized format
// suitable for storage and comparison.
//
// Returns:
//   - string: The normalized phone number string
//   - error: Any error that occurred during phone parsing
func (r *CreateUserRequest) GetNormalizedPhone() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.String(), nil
}

// GetPhoneCountryCode returns the country code from the phone number.
// This method extracts the international country code (e.g., "+1" for US/Canada)
// from the provided phone number.
//
// Returns:
//   - string: The country code (e.g., "+1", "+44", "+62")
//   - error: Any error that occurred during phone parsing
func (r *CreateUserRequest) GetPhoneCountryCode() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.CountryCode(), nil
}

// GetPhoneNationalNumber returns the national number from the phone number.
// This method extracts the local phone number without the country code
// (e.g., "555-1234" from "+1-555-1234").
//
// Returns:
//   - string: The national phone number without country code
//   - error: Any error that occurred during phone parsing
func (r *CreateUserRequest) GetPhoneNationalNumber() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.NationalNumber(), nil
}

// UpdateUserRequest represents the request for updating an existing user entity.
// This struct allows partial updates where only provided fields are modified.
// All fields are optional, enabling flexible update operations.
type UpdateUserRequest struct {
	// Username is the unique identifier for the user account (optional)
	Username string `json:"username"`
	// Email is the user's email address for authentication and communication (optional)
	Email string `json:"email"`
	// Phone is the user's phone number with country code (optional)
	Phone string `json:"phone"`
}

// Validate performs validation on the UpdateUserRequest for any provided fields.
// This method validates email format and phone number format only if the fields
// are provided, allowing for partial updates.
//
// Validation Rules:
// - Email must match standard email format if provided
// - Phone must be valid international format if provided
// - Username has no format restrictions
//
// Returns:
//   - error: Validation error if any provided field fails validation, nil if valid
func (r *UpdateUserRequest) Validate() error {
	// Validate email format if provided
	if r.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(r.Email) {
			return errors.New("invalid email format")
		}
	}

	// Validate phone number if provided using value object
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			return errors.New("invalid phone number: " + err.Error())
		}
	}

	return nil
}

// ToEntity transforms the UpdateUserRequest to update an existing User entity.
// This method applies only the provided fields to the existing user, preserving
// unchanged values. It's designed for partial updates where not all fields are provided.
//
// Parameters:
//   - existingUser: The existing user entity to update
//
// Returns:
//   - *entities.User: The updated user entity
//   - error: Any error that occurred during transformation
func (r *UpdateUserRequest) ToEntity(existingUser *entities.User) (*entities.User, error) {
	// Update username if provided
	if r.Username != "" {
		existingUser.UserName = r.Username
	}

	// Update email if provided
	if r.Email != "" {
		existingUser.Email = r.Email
	}

	// Update phone if provided
	if r.Phone != "" {
		existingUser.Phone = r.Phone
	}

	return existingUser, nil
}

// ParsePhone parses the phone number string and returns a Phone value object.
// This method validates the phone number format only if a phone number is provided.
// It's designed for update scenarios where phone may be optional.
//
// Returns:
//   - *valueobjects.Phone: The parsed phone value object
//   - error: Any error that occurred during phone parsing or if phone is empty
func (r *UpdateUserRequest) ParsePhone() (*valueobjects.Phone, error) {
	if r.Phone == "" {
		return nil, errors.New("phone number is empty")
	}
	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// ChangePasswordRequest represents the request for changing a user's password.
// This struct requires both the old password for verification and the new password
// for the update operation.
type ChangePasswordRequest struct {
	// OldPassword is the current password for verification (required)
	OldPassword string `json:"old_password"`
	// NewPassword is the new password to set (required, minimum 8 characters)
	NewPassword string `json:"new_password"`
}

// Validate performs validation on the ChangePasswordRequest.
// This method ensures both passwords are provided and the new password
// meets minimum security requirements.
//
// Validation Rules:
// - Both old and new passwords are required
// - New password must be at least 8 characters long
//
// Returns:
//   - error: Validation error if any field fails validation, nil if valid
func (r *ChangePasswordRequest) Validate() error {
	// Check required fields
	if r.OldPassword == "" {
		return errors.New("old password is required")
	}
	if r.NewPassword == "" {
		return errors.New("new password is required")
	}

	// Validate new password strength (minimum length requirement)
	if len(r.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}

	return nil
}

// LoginRequest represents the request for user authentication.
// This struct supports flexible login using either username or email
// as the identity field, along with the password for verification.
type LoginRequest struct {
	// Identity is the username or email for authentication (required)
	Identity string `json:"identity"`
	// Password is the user's password for verification (required)
	Password string `json:"password"`
}

// Validate validates the LoginRequest using Laravel-style validation responses.
// This method uses a custom validator to provide structured error responses
// that match the expected API format for validation failures.
//
// Validation Rules:
// - Identity (username or email) is required
// - Password is required
//
// Returns:
//   - *responses.ValidationErrorBuilder: Builder for structured validation errors
//   - error: General validation error if validation fails
func (r *LoginRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate identity field (username or email)
	validator.ValidateRequired("identity", r.Identity)

	// Validate password field
	validator.ValidateRequired("password", r.Password)

	// Check if validation produced any errors
	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}
