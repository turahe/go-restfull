package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(public fiber.Router, protected fiber.Router, container *container.Container) {
	authController := container.GetAuthController()

	// Public auth routes
	auth := public.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/forget-password", authController.ForgetPassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// Protected auth routes
	authProtected := protected.Group("/auth")
	authProtected.Post("/logout", authController.Logout)
}
