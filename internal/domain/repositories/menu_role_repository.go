package repositories

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// MenuRoleRepository defines the interface for menu-role relationship data access
type MenuRoleRepository interface {
	AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error
	RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error
	GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error)
	GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error)
	HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error)
	HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
	GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error)
	CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
}
