package routes

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"
	v1 "github.com/turahe/go-restfull/internal/interfaces/http/routes/v1"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all routes using the Hexagonal Architecture
func RegisterRoutes(app *fiber.App, container *container.Container) {
	// Comprehensive health check endpoint
	healthzHandler := controllers.NewHealthzHTTPHandler()
	app.Get("/healthz", healthzHandler.Healthz)

	// API v1 routes
	api := app.Group("/api")
	v1Group := api.Group("/v1")

	// Register v1 routes
	v1.RegisterV1Routes(v1Group, container)
}
