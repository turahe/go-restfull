package requests

import (
	"errors"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/interfaces/http/validation"
)

// CreateUserRequest represents the request for creating a user
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// Validate validates the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
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

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	// Validate phone number with country code parsing
	if _, err := valueobjects.NewPhone(r.Phone); err != nil {
		return errors.New("invalid phone number: " + err.Error())
	}

	// Validate password strength
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}

// ToEntity transforms CreateUserRequest to a User entity
func (r *CreateUserRequest) ToEntity() (*entities.User, error) {
	// Create user entity using the existing constructor
	user, err := entities.NewUser(r.Username, r.Email, r.Phone, r.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ParsePhone parses the phone number and returns the phone value object
func (r *CreateUserRequest) ParsePhone() (*valueobjects.Phone, error) {
	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// GetNormalizedPhone returns the normalized phone number string
func (r *CreateUserRequest) GetNormalizedPhone() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.String(), nil
}

// GetPhoneCountryCode returns the country code from the phone number
func (r *CreateUserRequest) GetPhoneCountryCode() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.CountryCode(), nil
}

// GetPhoneNationalNumber returns the national number from the phone number
func (r *CreateUserRequest) GetPhoneNationalNumber() (string, error) {
	phone, err := r.ParsePhone()
	if err != nil {
		return "", err
	}
	return phone.NationalNumber(), nil
}

// UpdateUserRequest represents the request for updating a user
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// Validate validates the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(r.Email) {
			return errors.New("invalid email format")
		}
	}

	// Validate phone number if provided
	if r.Phone != "" {
		if _, err := valueobjects.NewPhone(r.Phone); err != nil {
			return errors.New("invalid phone number: " + err.Error())
		}
	}

	return nil
}

// ToEntity transforms UpdateUserRequest to update an existing User entity
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

// ParsePhone parses the phone number and returns the phone value object
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

// ChangePasswordRequest represents the request for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Validate validates the ChangePasswordRequest
func (r *ChangePasswordRequest) Validate() error {
	if r.OldPassword == "" {
		return errors.New("old password is required")
	}
	if r.NewPassword == "" {
		return errors.New("new password is required")
	}

	if len(r.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}

	return nil
}

// LoginRequest represents the request for user login
type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

// Validate validates the LoginRequest with Laravel-style error responses
func (r *LoginRequest) Validate() (*responses.ValidationErrorBuilder, error) {
	validator := validation.NewValidator()

	// Validate identity
	validator.ValidateRequired("identity", r.Identity)

	// Validate password
	validator.ValidateRequired("password", r.Password)

	if validator.HasErrors() {
		return validator.GetErrorBuilder(), errors.New("validation failed")
	}

	return validator.GetErrorBuilder(), nil
}
