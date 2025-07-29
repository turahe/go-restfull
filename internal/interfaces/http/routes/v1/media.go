package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterMediaRoutes registers all media-related routes
func RegisterMediaRoutes(protected fiber.Router, container *container.Container) {
	mediaController := container.GetMediaController()

	// Media routes (protected)
	media := protected.Group("/media")
	media.Get("/", mediaController.GetMedia)
	media.Get("/:id", mediaController.GetMediaByID)
	media.Post("/", mediaController.CreateMedia)
	media.Put("/:id", mediaController.UpdateMedia)
	media.Delete("/:id", mediaController.DeleteMedia)
}
