// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"errors"
	"regexp"

	"github.com/turahe/go-restfull/internal/domain/services"

	"golang.org/x/crypto/bcrypt"
)

// bcryptPasswordService implements the PasswordService interface using bcrypt for secure
// password hashing and validation. This service provides password security features
// including hashing, comparison, and strength validation according to security best practices.
type bcryptPasswordService struct{}

// NewBcryptPasswordService creates a new bcrypt password service instance.
// This factory function returns a concrete implementation of the PasswordService interface
// that uses bcrypt for cryptographic operations.
//
// Returns:
//   - services.PasswordService: A new password service instance
func NewBcryptPasswordService() services.PasswordService {
	return &bcryptPasswordService{}
}

// HashPassword creates a secure hash of the provided password using bcrypt.
// The function uses bcrypt.DefaultCost (10) which provides a good balance between
// security and performance. The resulting hash is safe to store in databases
// and includes the salt automatically.
//
// Parameters:
//   - password: The plain text password to hash
//
// Returns:
//   - string: The bcrypt hash of the password
//   - error: Any error that occurred during the hashing process
func (s *bcryptPasswordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a plain text password with a stored hash to verify authenticity.
// This function uses bcrypt's constant-time comparison to prevent timing attacks.
// The function returns true if the password matches the hash, false otherwise.
//
// Parameters:
//   - hashedPassword: The stored bcrypt hash to compare against
//   - plainPassword: The plain text password to verify
//
// Returns:
//   - bool: True if the password matches the hash, false otherwise
func (s *bcryptPasswordService) ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// ValidatePassword checks if a password meets security requirements.
// The validation enforces the following security policies:
// - Minimum length of 8 characters
// - At least one uppercase letter (A-Z)
// - At least one lowercase letter (a-z)
// - At least one digit (0-9)
// - At least one special character from a predefined set
//
// Parameters:
//   - password: The password string to validate
//
// Returns:
//   - error: Validation error if the password doesn't meet requirements, nil if valid
func (s *bcryptPasswordService) ValidatePassword(password string) error {
	// Check minimum length requirement
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter using regex
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter using regex
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit using regex
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	// Check for at least one special character using regex
	// This includes common special characters that are typically allowed in passwords
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
} 
