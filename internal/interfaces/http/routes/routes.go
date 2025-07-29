package routes

import (
	"webapi/internal/http/controllers/healthz"
	"webapi/internal/infrastructure/container"
	v1 "webapi/internal/interfaces/http/routes/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all routes using the Hexagonal Architecture
func RegisterRoutes(app *fiber.App, container *container.Container) {
	// Comprehensive health check endpoint
	healthzHandler := healthz.NewHealthzHTTPHandler()
	app.Get("/healthz", healthzHandler.Healthz)

	// API v1 routes
	api := app.Group("/api")
	v1Group := api.Group("/v1")

	// Register v1 routes
	v1.RegisterV1Routes(v1Group, container)
}
