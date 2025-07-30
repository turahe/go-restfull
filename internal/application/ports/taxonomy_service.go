package ports

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/pagination"

	"github.com/google/uuid"
)

// TaxonomyService defines the interface for taxonomy business operations
type TaxonomyService interface {
	CreateTaxonomy(ctx context.Context, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error)
	GetTaxonomyByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error)
	GetTaxonomyBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error)
	GetAllTaxonomies(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error)
	GetAllTaxonomiesWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error)
	GetTaxonomyChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error)
	GetTaxonomyHierarchy(ctx context.Context) ([]*entities.Taxonomy, error)
	GetTaxonomyDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetTaxonomyAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetTaxonomySiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	SearchTaxonomies(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	UpdateTaxonomy(ctx context.Context, id uuid.UUID, name, slug, code, description string, parentID *uuid.UUID) (*entities.Taxonomy, error)
	DeleteTaxonomy(ctx context.Context, id uuid.UUID) error
	GetTaxonomyCount(ctx context.Context) (int64, error)
	GetTaxonomyCountWithSearch(ctx context.Context, query string) (int64, error)
	SearchTaxonomiesWithPagination(ctx context.Context, request *pagination.TaxonomySearchRequest) (*pagination.TaxonomySearchResponse, error)
}
