package adapters

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresTaxonomyRepository struct {
	repo repository.TaxonomyRepository
}

func NewPostgresTaxonomyRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TaxonomyRepository {
	return &PostgresTaxonomyRepository{
		repo: repository.NewTaxonomyRepository(db, redisClient),
	}
}

func (r *PostgresTaxonomyRepository) Create(ctx context.Context, taxonomy *entities.Taxonomy) error {
	taxonomyModel := &model.Taxonomy{
		ID:          taxonomy.ID.String(),
		Name:        taxonomy.Name,
		Slug:        taxonomy.Slug,
		Code:        taxonomy.Code,
		Description: taxonomy.Description,
		ParentID:    nil, // Will be set below
		RecordLeft:  taxonomy.RecordLeft,
		RecordRight: taxonomy.RecordRight,
		RecordDepth: taxonomy.RecordDepth,
		CreatedAt:   taxonomy.CreatedAt,
		UpdatedAt:   taxonomy.UpdatedAt,
		CreatedBy:   "",
		UpdatedBy:   "",
	}

	// Handle parent ID
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		taxonomyModel.ParentID = &parentIDStr
	}

	// Handle deleted at
	if taxonomy.DeletedAt != nil {
		taxonomyModel.DeletedAt = taxonomy.DeletedAt
	}

	return r.repo.Create(ctx, taxonomyModel)
}

func (r *PostgresTaxonomyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Taxonomy, error) {
	taxonomyModel, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertModelToEntity(taxonomyModel), nil
}

func (r *PostgresTaxonomyRepository) GetBySlug(ctx context.Context, slug string) (*entities.Taxonomy, error) {
	taxonomyModel, err := r.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return r.convertModelToEntity(taxonomyModel), nil
}

func (r *PostgresTaxonomyRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetRootTaxonomies(ctx context.Context) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetRootTaxonomies(ctx)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetHierarchy(ctx context.Context) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetDescendants(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetDescendants(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetAncestors(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetAncestors(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) GetSiblings(ctx context.Context, id uuid.UUID) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.GetSiblings(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Taxonomy, error) {
	taxonomyModels, err := r.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(taxonomyModels), nil
}

func (r *PostgresTaxonomyRepository) Update(ctx context.Context, taxonomy *entities.Taxonomy) error {
	taxonomyModel := &model.Taxonomy{
		ID:          taxonomy.ID.String(),
		Name:        taxonomy.Name,
		Slug:        taxonomy.Slug,
		Code:        taxonomy.Code,
		Description: taxonomy.Description,
		ParentID:    nil, // Will be set below
		RecordLeft:  taxonomy.RecordLeft,
		RecordRight: taxonomy.RecordRight,
		RecordDepth: taxonomy.RecordDepth,
		CreatedAt:   taxonomy.CreatedAt,
		UpdatedAt:   taxonomy.UpdatedAt,
		CreatedBy:   "",
		UpdatedBy:   "",
	}

	// Handle parent ID
	if taxonomy.ParentID != nil {
		parentIDStr := taxonomy.ParentID.String()
		taxonomyModel.ParentID = &parentIDStr
	}

	// Handle deleted at
	if taxonomy.DeletedAt != nil {
		taxonomyModel.DeletedAt = taxonomy.DeletedAt
	}

	return r.repo.Update(ctx, taxonomyModel)
}

func (r *PostgresTaxonomyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresTaxonomyRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.repo.ExistsBySlug(ctx, slug)
}

func (r *PostgresTaxonomyRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

// convertModelToEntity converts a taxonomy model to entity
func (r *PostgresTaxonomyRepository) convertModelToEntity(taxonomyModel *model.Taxonomy) *entities.Taxonomy {
	taxonomyID, _ := uuid.Parse(taxonomyModel.ID)

	taxonomy := &entities.Taxonomy{
		ID:          taxonomyID,
		Name:        taxonomyModel.Name,
		Slug:        taxonomyModel.Slug,
		Code:        taxonomyModel.Code,
		Description: taxonomyModel.Description,
		ParentID:    nil, // Will be set below
		RecordLeft:  taxonomyModel.RecordLeft,
		RecordRight: taxonomyModel.RecordRight,
		RecordDepth: taxonomyModel.RecordDepth,
		CreatedAt:   taxonomyModel.CreatedAt,
		UpdatedAt:   taxonomyModel.UpdatedAt,
		Children:    []*entities.Taxonomy{},
	}

	// Handle parent ID
	if taxonomyModel.ParentID != nil {
		if parentID, err := uuid.Parse(*taxonomyModel.ParentID); err == nil {
			taxonomy.ParentID = &parentID
		}
	}

	// Handle deleted at
	if taxonomyModel.DeletedAt != nil {
		taxonomy.DeletedAt = taxonomyModel.DeletedAt
	}

	return taxonomy
}

// convertModelsToEntities converts taxonomy models to entities
func (r *PostgresTaxonomyRepository) convertModelsToEntities(taxonomyModels []*model.Taxonomy) []*entities.Taxonomy {
	var result []*entities.Taxonomy

	for _, taxonomyModel := range taxonomyModels {
		result = append(result, r.convertModelToEntity(taxonomyModel))
	}

	return result
}
