package repositories

import (
	"context"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// OrganizationRepository defines the interface for organization data access with nested set hierarchy
type OrganizationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, organization *entities.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error)
	GetByCode(ctx context.Context, code string) (*entities.Organization, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Organization, error)
	Update(ctx context.Context, organization *entities.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Nested set hierarchy operations
	GetRoots(ctx context.Context) ([]*entities.Organization, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error)
	GetDescendants(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error)
	GetAncestors(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)
	GetSiblings(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)
	GetPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)

	// Tree operations
	GetTree(ctx context.Context) ([]*entities.Organization, error)
	GetSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error)

	// Hierarchy management
	AddChild(ctx context.Context, parentID, childID uuid.UUID) error
	MoveSubtree(ctx context.Context, organizationID, newParentID uuid.UUID) error
	DeleteSubtree(ctx context.Context, organizationID uuid.UUID) error

	// Search and filtering
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error)
	GetByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error)

	// Validation
	ExistsByCode(ctx context.Context, code string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	IsDescendant(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error)
	IsAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error)

	// Count operations
	Count(ctx context.Context) (int64, error)
	CountBySearch(ctx context.Context, query string) (int64, error)
	CountByStatus(ctx context.Context, status entities.OrganizationStatus) (int64, error)
	CountChildren(ctx context.Context, parentID uuid.UUID) (int64, error)
	CountDescendants(ctx context.Context, parentID uuid.UUID) (int64, error)
}
