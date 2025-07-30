package ports

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRoleService defines the interface for user-role relationship business operations
type UserRoleService interface {
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error)
	HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
	HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
	GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetUserRoleCount(ctx context.Context, roleID uuid.UUID) (int64, error)
} 
