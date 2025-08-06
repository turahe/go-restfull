package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Phone represents a phone number value object
type Phone struct {
	value string
}

// NewPhone creates a new phone value object with validation
func NewPhone(phone string) (Phone, error) {
	phone = strings.TrimSpace(phone)
	
	if phone == "" {
		return Phone{}, errors.New("phone cannot be empty")
	}

	if !isValidPhone(phone) {
		return Phone{}, errors.New("invalid phone format")
	}

	// Normalize phone number (remove spaces, dashes, etc.)
	normalized := normalizePhone(phone)
	
	return Phone{value: normalized}, nil
}

// String returns the string representation of the phone
func (p Phone) String() string {
	return p.value
}

// Value returns the phone value
func (p Phone) Value() string {
	return p.value
}

// Equals checks if two phones are equal
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// isValidPhone validates phone format
func isValidPhone(phone string) bool {
	// Accept various international phone formats
	phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

// normalizePhone normalizes the phone number
func normalizePhone(phone string) string {
	// Remove all non-digit characters except +
	reg := regexp.MustCompile(`[^\d+]`)
	return reg.ReplaceAllString(phone, "")
}