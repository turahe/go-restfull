package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterOrganizationRoutes registers the organization routes
func RegisterOrganizationRoutes(public fiber.Router, protected fiber.Router, container *container.Container) {
	organizationController := container.GetOrganizationController()

	// Public organization routes
	organizations := public.Group("/organizations")
	organizations.Get("/tree", organizationController.GetOrganizationTree)
	organizations.Get("/roots", organizationController.GetRootOrganizations)

	// Protected organization routes (only if protected router is provided)
	if protected != nil {
		organizationsProtected := protected.Group("/organizations")
		organizationsProtected.Post("/", organizationController.CreateOrganization)
		organizationsProtected.Get("/", organizationController.GetAllOrganizations)
		organizationsProtected.Get("/:id", organizationController.GetOrganizationByID)
		organizationsProtected.Put("/:id", organizationController.UpdateOrganization)
		organizationsProtected.Delete("/:id", organizationController.DeleteOrganization)

		// Hierarchy operations
		organizationsProtected.Get("/:id/children", organizationController.GetOrganizationChildren)
		organizationsProtected.Get("/:id/descendants", organizationController.GetOrganizationDescendants)
		organizationsProtected.Get("/:id/ancestors", organizationController.GetOrganizationAncestors)
		organizationsProtected.Get("/:id/siblings", organizationController.GetOrganizationSiblings)
		organizationsProtected.Get("/:id/path", organizationController.GetOrganizationPath)
		organizationsProtected.Get("/:id/subtree", organizationController.GetOrganizationSubtree)

		// Organization management
		organizationsProtected.Post("/:id/children", organizationController.AddOrganizationChild)
		organizationsProtected.Put("/:id/move", organizationController.MoveOrganizationSubtree)
		organizationsProtected.Delete("/:id/subtree", organizationController.DeleteOrganizationSubtree)
		organizationsProtected.Put("/:id/status", organizationController.SetOrganizationStatus)

		// Search and statistics
		organizationsProtected.Get("/search", organizationController.SearchOrganizations)
		organizationsProtected.Get("/:id/stats", organizationController.GetOrganizationStats)
		organizationsProtected.Get("/validate", organizationController.ValidateOrganizationHierarchy)
	}
}
