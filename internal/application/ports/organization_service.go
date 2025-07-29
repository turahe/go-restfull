package ports

import (
	"context"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// OrganizationService defines the interface for organization business logic
type OrganizationService interface {
	// Basic CRUD operations
	CreateOrganization(ctx context.Context, name, description, code, email, phone, address, website, logoURL string, parentID *uuid.UUID) (*entities.Organization, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error)
	GetOrganizationByCode(ctx context.Context, code string) (*entities.Organization, error)
	GetAllOrganizations(ctx context.Context, limit, offset int) ([]*entities.Organization, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, name, description, code, email, phone, address, website, logoURL string) (*entities.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error

	// Hierarchy operations
	GetRootOrganizations(ctx context.Context) ([]*entities.Organization, error)
	GetChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error)
	GetDescendantOrganizations(ctx context.Context, parentID uuid.UUID) ([]*entities.Organization, error)
	GetAncestorOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)
	GetSiblingOrganizations(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)
	GetOrganizationPath(ctx context.Context, organizationID uuid.UUID) ([]*entities.Organization, error)

	// Tree operations
	GetOrganizationTree(ctx context.Context) ([]*entities.Organization, error)
	GetOrganizationSubtree(ctx context.Context, rootID uuid.UUID) ([]*entities.Organization, error)

	// Hierarchy management
	AddChildOrganization(ctx context.Context, parentID, childID uuid.UUID) error
	MoveOrganization(ctx context.Context, organizationID, newParentID uuid.UUID) error
	DeleteOrganizationSubtree(ctx context.Context, organizationID uuid.UUID) error

	// Status management
	SetOrganizationStatus(ctx context.Context, id uuid.UUID, status entities.OrganizationStatus) error
	GetOrganizationsByStatus(ctx context.Context, status entities.OrganizationStatus, limit, offset int) ([]*entities.Organization, error)

	// Search operations
	SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*entities.Organization, error)
	GetOrganizationsWithPagination(ctx context.Context, page, perPage int, search string, status entities.OrganizationStatus) ([]*entities.Organization, int64, error)

	// Validation operations
	ValidateOrganizationHierarchy(ctx context.Context, parentID, childID uuid.UUID) error
	CheckOrganizationExists(ctx context.Context, id uuid.UUID) (bool, error)
	CheckOrganizationCodeExists(ctx context.Context, code string) (bool, error)

	// Count operations
	GetOrganizationsCount(ctx context.Context, search string, status entities.OrganizationStatus) (int64, error)
	GetChildrenCount(ctx context.Context, parentID uuid.UUID) (int64, error)
	GetDescendantsCount(ctx context.Context, parentID uuid.UUID) (int64, error)
}
