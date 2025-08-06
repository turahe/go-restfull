package valueobjects

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashedPassword represents a hashed password value object
type HashedPassword struct {
	hash string
}

// NewHashedPasswordFromPlaintext creates a new hashed password from plaintext
func NewHashedPasswordFromPlaintext(plaintext string) (HashedPassword, error) {
	if err := validatePassword(plaintext); err != nil {
		return HashedPassword{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return HashedPassword{}, errors.New("failed to hash password")
	}

	return HashedPassword{hash: string(hash)}, nil
}

// NewHashedPasswordFromHash creates a new hashed password from an existing hash
func NewHashedPasswordFromHash(hash string) (HashedPassword, error) {
	if hash == "" {
		return HashedPassword{}, errors.New("hash cannot be empty")
	}

	return HashedPassword{hash: hash}, nil
}

// Hash returns the password hash
func (p HashedPassword) Hash() string {
	return p.hash
}

// Verify verifies if the plaintext matches the hashed password
func (p HashedPassword) Verify(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext))
	return err == nil
}

// Equals checks if two hashed passwords are equal
func (p HashedPassword) Equals(other HashedPassword) bool {
	return p.hash == other.hash
}

// validatePassword validates password requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be no more than 128 characters long")
	}

	// Must contain at least one uppercase letter
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Must contain at least one lowercase letter
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Must contain at least one digit
	if matched, _ := regexp.MatchString(`\d`, password); !matched {
		return errors.New("password must contain at least one digit")
	}

	// Must contain at least one special character
	if matched, _ := regexp.MatchString(`[!@#$%^&*(),.?":{}|<>]`, password); !matched {
		return errors.New("password must contain at least one special character")
	}

	return nil
}