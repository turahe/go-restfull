package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterHealthRoutes registers health check routes (public access)
func RegisterHealthRoutes(v1Group fiber.Router, container *container.Container) {
	// Health check endpoint for API monitoring
	healthzHandler := controllers.NewHealthzHTTPHandler(container.StorageService)
	v1Group.Get("/health", healthzHandler.Healthz)
	v1Group.Get("/healthz", healthzHandler.Healthz) // Alternative endpoint
}
