package ports

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// RoleService defines the interface for role business operations
type RoleService interface {
	CreateRole(ctx context.Context, name, slug, description string) (*entities.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetRoleBySlug(ctx context.Context, slug string) (*entities.Role, error)
	GetAllRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	GetActiveRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	SearchRoles(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, name, slug, description string) (*entities.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ActivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	DeactivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetRoleCount(ctx context.Context) (int64, error)
	GetActiveRoleCount(ctx context.Context) (int64, error)
} 