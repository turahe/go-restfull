package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterTaxonomyRoutes registers all taxonomy-related routes
func RegisterTaxonomyRoutes(protected fiber.Router, container *container.Container) {
	taxonomyController := container.GetTaxonomyController()

	// Taxonomy routes (protected)
	taxonomies := protected.Group("/taxonomies")
	taxonomies.Post("/", taxonomyController.CreateTaxonomy)
	taxonomies.Get("/", taxonomyController.GetTaxonomies)
	taxonomies.Get("/root", taxonomyController.GetRootTaxonomies)
	taxonomies.Get("/hierarchy", taxonomyController.GetTaxonomyHierarchy)
	taxonomies.Get("/search", taxonomyController.SearchTaxonomies)
	taxonomies.Get("/slug/:slug", taxonomyController.GetTaxonomyBySlug)
	taxonomies.Get("/:id", taxonomyController.GetTaxonomyByID)
	taxonomies.Put("/:id", taxonomyController.UpdateTaxonomy)
	taxonomies.Delete("/:id", taxonomyController.DeleteTaxonomy)
	taxonomies.Get("/:id/children", taxonomyController.GetTaxonomyChildren)
	taxonomies.Get("/:id/descendants", taxonomyController.GetTaxonomyDescendants)
	taxonomies.Get("/:id/ancestors", taxonomyController.GetTaxonomyAncestors)
	taxonomies.Get("/:id/siblings", taxonomyController.GetTaxonomySiblings)
}
