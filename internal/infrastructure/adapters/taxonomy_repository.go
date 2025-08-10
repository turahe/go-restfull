package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresTaxonomyRepository is an adapter that implements the TaxonomyRepository interface
// by delegating calls to the concrete repository implementation
type PostgresTaxonomyRepository struct {
	repo repositories.TaxonomyRepository
}

// NewPostgresTaxonomyRepository creates a new PostgresTaxonomyRepository adapter
func NewPostgresTaxonomyRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TaxonomyRepository {
	return &PostgresTaxonomyRepository{
		repo: repository.NewTaxonomyRepository(db, redisClient),
	}
}

// Create delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) Create(ctx context.Context, taxonomy *entities.Taxonomy) error {
	return r.repo.Create(ctx, taxonomy)
}

// GetByID delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	return r.repo.GetByID(ctx, id)
}

// GetBySlug delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	return r.repo.GetBySlug(ctx, slug)
}

// GetAll delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	return r.repo.GetAll(ctx, limit, offset)
}

// GetAllWithSearch delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetAllWithSearch(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return r.repo.GetAllWithSearch(ctx, query, limit, offset)
}

// GetRootTaxonomies delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	return r.repo.GetRootTaxonomies(ctx)
}

// GetChildren delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	return r.repo.GetChildren(ctx, parentID)
}

// GetHierarchy delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	return r.repo.GetHierarchy(ctx)
}

// GetDescendants delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return r.repo.GetDescendants(ctx, id)
}

// GetAncestors delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return r.repo.GetAncestors(ctx, id)
}

// GetSiblings delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	return r.repo.GetSiblings(ctx, id)
}

// Search delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	return r.repo.Search(ctx, query, limit, offset)
}

// Update delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) Update(ctx context.Context, taxonomy *entities.Taxonomy) error {
	return r.repo.Update(ctx, taxonomy)
}

// Delete delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

// ExistsBySlug delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.repo.ExistsBySlug(ctx, slug)
}

// Count delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

// CountWithSearch delegates to the underlying repository implementation
func (r *PostgresTaxonomyRepository) CountWithSearch(ctx context.Context, query string) (int64, error) {
	return r.repo.CountWithSearch(ctx, query)
}
