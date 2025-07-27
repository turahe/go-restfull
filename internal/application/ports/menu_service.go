package ports

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// MenuService defines the interface for menu business operations
type MenuService interface {
	CreateMenu(ctx context.Context, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error)
	GetMenuByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	GetMenuBySlug(ctx context.Context, slug string) (*entities.Menu, error)
	GetAllMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetActiveMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetVisibleMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	GetRootMenus(ctx context.Context) ([]*entities.Menu, error)
	GetMenuHierarchy(ctx context.Context) ([]*entities.Menu, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error)
	SearchMenus(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error)
	UpdateMenu(ctx context.Context, id uuid.UUID, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error)
	DeleteMenu(ctx context.Context, id uuid.UUID) error
	ActivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	DeactivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	ShowMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	HideMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	GetMenuCount(ctx context.Context) (int64, error)
	GetActiveMenuCount(ctx context.Context) (int64, error)
	GetVisibleMenuCount(ctx context.Context) (int64, error)
}
