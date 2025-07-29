package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterRBACRoutes registers all RBAC-related routes
func RegisterRBACRoutes(protected fiber.Router, container *container.Container) {
	rbacController := container.GetRBACController()

	// RBAC Management routes (admin only)
	rbac := protected.Group("/rbac")
	rbac.Get("/policies", rbacController.GetPolicy)
	rbac.Post("/policies", rbacController.AddPolicy)
	rbac.Delete("/policies", rbacController.RemovePolicy)
	rbac.Get("/users/:user_id/roles", rbacController.GetRolesForUser)
	rbac.Post("/users/:user_id/roles", rbacController.AddRoleForUser)
	rbac.Delete("/users/:user_id/roles", rbacController.RemoveRoleForUser)
	rbac.Get("/roles/:role_id/users", rbacController.GetUsersForRole)
}
