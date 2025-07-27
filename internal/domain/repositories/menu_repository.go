package repositories

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// MenuRepository defines the interface for menu data access
type MenuRepository interface {
	Create(ctx context.Context, menu *entities.Menu) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Menu, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetActive(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetVisible(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetRootMenus(ctx context.Context) ([]*entities.Menu, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error)
	GetHierarchy(ctx context.Context) ([]*entities.Menu, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error)
	Update(ctx context.Context, menu *entities.Menu) error
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	Show(ctx context.Context, id uuid.UUID) error
	Hide(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	CountVisible(ctx context.Context) (int64, error)
}
