package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new email value object with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	if !isValidEmail(email) {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: email}, nil
}

// String returns the string representation of the email
func (e Email) String() string {
	return e.value
}

// Value returns the email value
func (e Email) Value() string {
	return e.value
}

// Domain returns the domain part of the email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part of the email
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}