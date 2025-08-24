package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoleRoutes registers all role-related routes
func RegisterRoleRoutes(protected fiber.Router, container *container.Container) {
	roleController := container.GetRoleController()

	roles := protected.Group("/roles")
	roles.Get("/", roleController.GetRoles)
	roles.Get("/search", roleController.SearchRoles)
	roles.Get("/slug/:slug", roleController.GetRoleBySlug)
	roles.Get("/:id", roleController.GetRoleByID)
	roles.Post("/", roleController.CreateRole)
	roles.Put("/:id", roleController.UpdateRole)
	roles.Put("/:id/activate", roleController.ActivateRole)
	roles.Put("/:id/deactivate", roleController.DeactivateRole)
	roles.Delete("/:id", roleController.DeleteRole)
}


