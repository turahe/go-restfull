package v1

import (
	"webapi/internal/infrastructure/container"
	"webapi/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterV1Routes registers all v1 API routes
func RegisterV1Routes(v1Group fiber.Router, container *container.Container) {
	// Public routes (no authentication required)
	public := v1Group.Group("/")

	// Protected routes (require JWT + RBAC)
	rbacProtected := v1Group.Group("/", middleware.JWTAuth(), middleware.RBACMiddleware(container.RBACService))

	// Register route groups
	RegisterAuthRoutes(public, rbacProtected, container)
	RegisterUserRoutes(rbacProtected, container)
	RegisterPostRoutes(rbacProtected, container)
	RegisterMenuRoutes(rbacProtected, container)
	RegisterTaxonomyRoutes(rbacProtected, container)
	RegisterAddressRoutes(rbacProtected, container)
	RegisterOrganizationRoutes(public, rbacProtected, container) // Organization routes
	RegisterMediaRoutes(rbacProtected, container)
	RegisterTagRoutes(rbacProtected, container)
	RegisterCommentRoutes(rbacProtected, container)
	RegisterSettingRoutes(rbacProtected, container) // TODO: Implement when SettingController is created
	RegisterRBACRoutes(rbacProtected, container)
	RegisterJobRoutes(rbacProtected, container)
}
