package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/interfaces/http/controllers"

	"github.com/gofiber/fiber/v2"
)

// RegisterSearchRoutes registers search-related routes
func RegisterSearchRoutes(router fiber.Router, container *container.Container) {
	searchController := controllers.NewSearchController(container.HybridSearchService)

	// Search endpoints
	router.Get("/search", searchController.GetSearchStatus)
	router.Post("/search", searchController.Search)
	router.Get("/search/:type", searchController.SearchByType)
}
