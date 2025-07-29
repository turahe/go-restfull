package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterMenuRoutes registers all menu-related routes
func RegisterMenuRoutes(protected fiber.Router, container *container.Container) {
	menuController := container.GetMenuController()
	menuRoleController := container.GetMenuRoleController()

	// Menu routes (protected)
	menus := protected.Group("/menus")
	menus.Post("/", menuController.CreateMenu)
	menus.Get("/", menuController.GetMenus)
	menus.Get("/root", menuController.GetRootMenus)
	menus.Get("/hierarchy", menuController.GetMenuHierarchy)
	menus.Get("/search", menuController.SearchMenus)
	menus.Get("/slug/:slug", menuController.GetMenuBySlug)
	menus.Get("/:id", menuController.GetMenuByID)
	menus.Put("/:id", menuController.UpdateMenu)
	menus.Delete("/:id", menuController.DeleteMenu)
	menus.Patch("/:id/activate", menuController.ActivateMenu)
	menus.Patch("/:id/deactivate", menuController.DeactivateMenu)
	menus.Patch("/:id/show", menuController.ShowMenu)
	menus.Patch("/:id/hide", menuController.HideMenu)

	// Menu-Role routes (protected)
	menus.Post("/:menu_id/roles/:role_id", menuRoleController.AssignRoleToMenu)
	menus.Delete("/:menu_id/roles/:role_id", menuRoleController.RemoveRoleFromMenu)
	menus.Get("/:menu_id/roles", menuRoleController.GetMenuRoles)
	menus.Get("/:menu_id/roles/:role_id/check", menuRoleController.HasRole)

	// Role-Menu routes (protected)
	roles := protected.Group("/roles")
	roles.Get("/:role_id/menus", menuRoleController.GetRoleMenus)
	roles.Get("/:role_id/menus/count", menuRoleController.GetMenuRoleCount)
}
