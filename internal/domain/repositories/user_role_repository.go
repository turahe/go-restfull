package repositories

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRoleRepository defines the interface for user-role relationship data access
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error)
	HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
	HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
	GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	CountUsersByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
} 