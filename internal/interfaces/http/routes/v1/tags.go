package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterTagRoutes registers all tag-related routes
func RegisterTagRoutes(protected fiber.Router, container *container.Container) {
	tagController := container.GetTagController()

	// Tag routes (protected)
	tags := protected.Group("/tags")
	tags.Get("/", tagController.GetTags)
	tags.Get("/search", tagController.SearchTags)
	tags.Get("/:id", tagController.GetTagByID)
	tags.Post("/", tagController.CreateTag)
	tags.Put("/:id", tagController.UpdateTag)
	tags.Delete("/:id", tagController.DeleteTag)
}
