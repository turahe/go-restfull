package v1

import (
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterV1Routes registers all v1 API routes
func RegisterV1Routes(v1Group fiber.Router, container *container.Container) {
	// Register public routes first (no middleware)
	RegisterAuthRoutes(v1Group, nil, container)
	RegisterOrganizationRoutes(v1Group, nil, container)

	// Protected routes (require JWT + RBAC)
	protected := v1Group.Group("/", middleware.JWTAuth(), middleware.RBACMiddleware(container.RBACService))
	RegisterUserRoutes(protected, container)
	RegisterPostRoutes(protected, container)
	RegisterMenuRoutes(protected, container)
	RegisterTaxonomyRoutes(protected, container)
	RegisterAddressRoutes(protected, container)
	RegisterMediaRoutes(protected, container)
	RegisterTagRoutes(protected, container)
	RegisterCommentRoutes(protected, container)
	RegisterSettingRoutes(protected, container) // TODO: Implement when SettingController is created
	RegisterRBACRoutes(protected, container)
	RegisterJobRoutes(protected, container)
	RegisterBackupRoutes(protected, container)
}
