package middleware

import (
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/helper/utils"

	"github.com/gofiber/fiber/v2"
)

// JWTAuth middleware for protecting routes
func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("=== JWT middleware triggered for path: %s, method: %s ===\n", c.Path(), c.Method())
		fmt.Printf("Request URL: %s\n", c.OriginalURL())
		fmt.Printf("Request headers: %v\n", c.GetReqHeaders())

		// Get the Authorization header
		authHeader := c.Get("Authorization")
		fmt.Printf("Authorization header: '%s'\n", authHeader)

		if authHeader == "" {
			fmt.Println("ERROR: Authorization header is missing")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Authorization header is required",
			})
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("ERROR: Invalid authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the access token
		claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			fmt.Printf("ERROR: Token validation failed: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
		}

		// Store user information in context for later use
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("user_claims", claims)

		fmt.Println("JWT middleware: Authentication successful")
		return c.Next()
	}
}

// OptionalJWTAuth middleware that doesn't fail if no token is provided
func OptionalJWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the access token
		claims, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Next()
		}

		// Store user information in context for later use
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("user_claims", claims)

		return c.Next()
	}
}
