package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// MenuRepository defines the interface for menu data access
type MenuRepository interface {
	TransactionalRepository // Embed transaction support

	// Basic CRUD operations
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

	// Nested Set Tree Traversal
	GetDescendants(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	GetAncestors(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	GetSiblings(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	GetPath(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	GetTree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	GetSubtree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)

	// Advanced Nested Set Operations
	AddChild(ctx context.Context, parentID, childID uuid.UUID) error
	MoveSubtree(ctx context.Context, menuID, newParentID uuid.UUID) error
	DeleteSubtree(ctx context.Context, menuID uuid.UUID) error
	InsertBetween(ctx context.Context, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error
	SwapPositions(ctx context.Context, menu1ID, menu2ID uuid.UUID) error
	GetLeafNodes(ctx context.Context) ([]*entities.Menu, error)
	GetInternalNodes(ctx context.Context) ([]*entities.Menu, error)

	// Batch Operations
	BatchMoveSubtrees(ctx context.Context, moves []struct {
		MenuID      uuid.UUID
		NewParentID uuid.UUID
	}) error
	BatchInsertBetween(ctx context.Context, insertions []struct {
		Menu           *entities.Menu
		LeftSiblingID  *uuid.UUID
		RightSiblingID *uuid.UUID
	}) error

	// Tree Maintenance and Optimization
	ValidateTree(ctx context.Context) ([]string, error)
	RebuildTree(ctx context.Context) error
	OptimizeTree(ctx context.Context) error
	GetTreeStatistics(ctx context.Context) (map[string]interface{}, error)
	GetTreeHeight(ctx context.Context) (uint64, error)
	GetLevelWidth(ctx context.Context, level uint64) (int64, error)
	GetSubtreeSize(ctx context.Context, menuID uuid.UUID) (int64, error)
	GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error)
	ValidateTreeIntegrity(ctx context.Context) ([]string, error)
}
