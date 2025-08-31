package routes

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"
	v1 "github.com/turahe/go-restfull/internal/interfaces/http/routes/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterPublicRoutes registers public endpoints (no authentication required)
func RegisterPublicRoutes(app *fiber.App) {
	// Comprehensive health check endpoint - MUST BE PUBLIC
	healthzHandler := controllers.NewHealthzHTTPHandler()
	app.Get("/healthz", healthzHandler.Healthz)
}

// RegisterProtectedRoutes registers protected endpoints (require authentication)
func RegisterProtectedRoutes(app *fiber.App, container *container.Container) {
	// API v1 routes (protected by authentication)
	api := app.Group("/api")
	v1Group := api.Group("/v1")

	// Register v1 routes
	v1.RegisterV1Routes(v1Group, container)
}

// RegisterRoutes registers all routes using the Hexagonal Architecture (legacy function)
func RegisterRoutes(app *fiber.App, container *container.Container) {
	// Public endpoints (no authentication required)
	RegisterPublicRoutes(app)

	// Protected endpoints (require authentication)
	RegisterProtectedRoutes(app, container)
}
