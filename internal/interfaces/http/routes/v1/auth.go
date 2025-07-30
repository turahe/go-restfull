package v1

import (
	"fmt"
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(public fiber.Router, protected fiber.Router, container *container.Container) {
	authController := container.GetAuthController()

	// Public auth routes
	auth := public.Group("/auth")
	fmt.Printf("Registering public auth routes on router: %T\n", public)
	auth.Post("/login", authController.Login)
	fmt.Printf("Registered POST /auth/login\n")
	auth.Post("/register", authController.Register)
	fmt.Printf("Registered POST /auth/register\n")
	auth.Post("/refresh", authController.Refresh)
	auth.Post("/forget-password", authController.ForgetPassword)
	auth.Post("/reset-password", authController.ResetPassword)

	// Protected auth routes (only if protected router is provided)
	if protected != nil {
		authProtected := protected.Group("/auth")
		fmt.Printf("Registering protected auth routes on router: %T\n", protected)
		authProtected.Post("/logout", authController.Logout)
		fmt.Printf("Registered POST /auth/logout (protected)\n")
	} else {
		fmt.Printf("Skipping protected auth routes (protected router is nil)\n")
	}
}
