package utils

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseIdentity detects if identity is email or phone
// It analyzes the input string and returns appropriate values based on content:
// - If the string contains "@", it's treated as an email
// - Otherwise, it's treated as a phone number
//
// Parameters:
//   - identity: string to be parsed (email or phone number)
//
// Returns:
//   - email: the email string if identity is an email, empty string otherwise
//   - phone: the phone string if identity is a phone number, empty string otherwise
func ParseIdentity(identity string) (email, phone string) {
	// Check if the identity string contains "@" symbol to determine if it's an email
	if strings.Contains(identity, "@") {
		return identity, "" // Return identity as email, empty phone
	}
	return "", identity // Return empty email, identity as phone
}

// IsEmail checks if the given string is an email address
// It performs a simple check by looking for the "@" symbol in the string.
// Note: This is a basic validation - for production use, consider more robust email validation.
//
// Parameters:
//   - identity: string to check for email format
//
// Returns:
//   - bool: true if the string contains "@" (likely an email), false otherwise
func IsEmail(identity string) bool {
	return strings.Contains(identity, "@")
}

// IsPhone checks if the given string is a phone number
// It determines this by checking if the string does NOT contain "@" symbol.
// Note: This is a basic validation - for production use, consider more robust phone validation.
//
// Parameters:
//   - identity: string to check for phone number format
//
// Returns:
//   - bool: true if the string does not contain "@" (likely a phone), false otherwise
func IsPhone(identity string) bool {
	return !strings.Contains(identity, "@")
}

// GetUserID extracts the authenticated user ID from Fiber context locals.
// This function is typically used in middleware or handlers to retrieve the user ID
// that was set during authentication (e.g., JWT token validation).
//
// Parameters:
//   - c: Fiber context containing the request information and locals
//
// Returns:
//   - uuid.UUID: the authenticated user's UUID if found and valid
//   - error: error message if user_id is not found or invalid
//
// Usage example:
//
//	userID, err := GetUserID(c)
//	if err != nil {
//	    return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
//	}
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	// Attempt to retrieve user_id from context locals and type assert to UUID
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		// Return nil UUID and error if user_id is not found or cannot be converted to UUID
		return uuid.Nil, errors.New("user_id not found or invalid")
	}
	return userID, nil
}
