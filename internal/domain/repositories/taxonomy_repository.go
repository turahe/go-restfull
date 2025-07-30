package repositories

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// TaxonomyRepository defines the interface for taxonomy data access
type TaxonomyRepository interface {
	Create(ctx context.Context, taxonomy *entities.Taxonomy) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error)
	GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error)
	GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error)
	GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error)
	Update(ctx context.Context, taxonomy *entities.Taxonomy) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountWithSearch(ctx context.Context, query string) (int64, error)
}
