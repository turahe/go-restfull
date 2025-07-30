package repositories

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	Create(ctx context.Context, role *entities.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Role, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error)
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
} 
