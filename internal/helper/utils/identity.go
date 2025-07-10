package utils

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseIdentity detects if identity is email or phone
func ParseIdentity(identity string) (email, phone string) {
	if strings.Contains(identity, "@") {
		return identity, ""
	}
	return "", identity
}

// IsEmail checks if the given string is an email
func IsEmail(identity string) bool {
	return strings.Contains(identity, "@")
}

// IsPhone checks if the given string is a phone number
func IsPhone(identity string) bool {
	return !strings.Contains(identity, "@")
}

// GetUserID extracts the authenticated user ID from Fiber context locals.
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user_id not found or invalid")
	}
	return userID, nil
}
