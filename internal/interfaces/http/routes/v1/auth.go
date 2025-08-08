package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(public fiber.Router, protected fiber.Router, container *container.Container) {
	authController := container.GetAuthController()

	// Public auth routes
	auth := public.Group("/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/register", authController.Register)
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/forget-password", authController.ForgetPassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// Protected auth routes (only if protected router is provided)
	if protected != nil {
		authProtected := protected.Group("/auth")
		authProtected.Post("/logout", authController.Logout)
	}
}
