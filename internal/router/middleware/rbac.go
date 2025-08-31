package middleware

import (
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RBACMiddleware creates a middleware that checks RBAC permissions
func RBACMiddleware(rbacService services.RBACService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("DEBUG: RBAC middleware called for %s %s\n", c.Method(), c.Path())

		// If RBAC service is not initialized, skip checks
		if rbacService == nil {
			fmt.Printf("DEBUG: RBAC service is nil, skipping checks\n")
			return c.Next()
		}

		// Get user from context (set by JWT middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			fmt.Println("ERROR: User not authenticated in RBAC middleware")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "User not authenticated",
			})
		}

		// Convert userID to string for RBAC service
		var userIDStr string
		if uuid, ok := userID.(uuid.UUID); ok {
			userIDStr = uuid.String()
		} else if str, ok := userID.(string); ok {
			userIDStr = str
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Invalid user ID format",
			})
		}

		// Get user roles from context or fetch from service
		userRoles := c.Locals("user_roles")
		if userRoles == nil {
			// Fetch roles from RBAC service
			roles, err := rbacService.GetRolesForUser(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to get user roles",
					"error":   err.Error(),
				})
			}
			userRoles = roles
			c.Locals("user_roles", roles)
		}

		// Get request path and method
		path := c.Path()
		method := c.Method()

		// Check if any role has permission
		hasPermission := false
		roles := userRoles.([]string)

		for _, role := range roles {
			allowed, err := rbacService.CheckPermission(role, path, method)
			if err != nil {
				continue // Skip this role if there's an error
			}
			if allowed {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Access denied",
				"path":    path,
				"method":  method,
				"roles":   roles,
			})
		}

		return c.Next()
	}
}

// OptionalRBACMiddleware creates a middleware that checks RBAC permissions but doesn't fail if no user is authenticated
func OptionalRBACMiddleware(rbacService services.RBACService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If RBAC service is not initialized, skip checks
		if rbacService == nil {
			return c.Next()
		}

		// Get user from context (set by JWT middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			// No user authenticated, continue without RBAC check
			return c.Next()
		}

		// Convert userID to string for RBAC service
		var userIDStr string
		if uuid, ok := userID.(uuid.UUID); ok {
			userIDStr = uuid.String()
		} else if str, ok := userID.(string); ok {
			userIDStr = str
		} else {
			// Continue without RBAC check if user ID format is invalid
			return c.Next()
		}

		// Get user roles from context or fetch from service
		userRoles := c.Locals("user_roles")
		if userRoles == nil {
			// Fetch roles from RBAC service
			roles, err := rbacService.GetRolesForUser(userIDStr)
			if err != nil {
				// Continue without RBAC check if there's an error
				return c.Next()
			}
			userRoles = roles
			c.Locals("user_roles", roles)
		}

		// Get request path and method
		path := c.Path()
		method := c.Method()

		// Check if any role has permission
		hasPermission := false
		roles := userRoles.([]string)

		for _, role := range roles {
			allowed, err := rbacService.CheckPermission(role, path, method)
			if err != nil {
				continue // Skip this role if there's an error
			}
			if allowed {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Access denied",
				"path":    path,
				"method":  method,
				"roles":   roles,
			})
		}

		return c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(rbacService services.RBACService, requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If RBAC service is not initialized, skip checks
		if rbacService == nil {
			return c.Next()
		}

		// Get user from context (set by JWT middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "User not authenticated",
			})
		}

		// Convert userID to string for RBAC service
		var userIDStr string
		if uuid, ok := userID.(uuid.UUID); ok {
			userIDStr = uuid.String()
		} else if str, ok := userID.(string); ok {
			userIDStr = str
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Invalid user ID format",
			})
		}

		// Get user roles from context or fetch from service
		userRoles := c.Locals("user_roles")
		if userRoles == nil {
			// Fetch roles from RBAC service
			roles, err := rbacService.GetRolesForUser(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to get user roles",
					"error":   err.Error(),
				})
			}
			userRoles = roles
			c.Locals("user_roles", roles)
		}

		// Check if user has the required role
		roles := userRoles.([]string)
		hasRole := false

		for _, role := range roles {
			if strings.EqualFold(role, requiredRole) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message":       "Insufficient permissions",
				"required_role": requiredRole,
				"user_roles":    roles,
			})
		}

		return c.Next()
	}
}
