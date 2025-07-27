package adapters

import (
	"context"
	"time"
	"webapi/internal/db/model"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresContentRepository struct {
	repo repository.ContentRepository
}

func NewPostgresContentRepository(db *pgxpool.Pool) repositories.ContentRepository {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &PostgresContentRepository{
		repo: repository.NewContentRepository(db, redisClient),
	}
}

func (r *PostgresContentRepository) Create(ctx context.Context, content *entities.Content) error {
	// Convert domain entity to model
	contentModel := &model.Content{
		ID:          content.ID.String(),
		ModelType:   content.ModelType,
		ModelID:     content.ModelID.String(),
		ContentRaw:  content.ContentRaw,
		ContentHTML: content.ContentHTML,
		CreatedBy:   content.CreatedBy.String(),
		UpdatedBy:   content.UpdatedBy.String(),
	}

	return r.repo.Create(ctx, contentModel)
}

func (r *PostgresContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	contentModel, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert model to domain entity
	contentID, _ := uuid.Parse(contentModel.ID)
	modelID, _ := uuid.Parse(contentModel.ModelID)
	createdBy, _ := uuid.Parse(contentModel.CreatedBy)
	updatedBy, _ := uuid.Parse(contentModel.UpdatedBy)

	content := &entities.Content{
		ID:          contentID,
		ModelType:   contentModel.ModelType,
		ModelID:     modelID,
		ContentRaw:  contentModel.ContentRaw,
		ContentHTML: contentModel.ContentHTML,
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
		CreatedAt:   time.Now(), // Not available in model
		UpdatedAt:   time.Now(), // Not available in model
	}

	return content, nil
}

func (r *PostgresContentRepository) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	contentModels, err := r.repo.GetByModelTypeAndID(ctx, modelType, modelID)
	if err != nil {
		return nil, err
	}

	// Convert models to domain entities
	var contents []*entities.Content
	for _, contentModel := range contentModels {
		contentID, _ := uuid.Parse(contentModel.ID)
		modelID, _ := uuid.Parse(contentModel.ModelID)
		createdBy, _ := uuid.Parse(contentModel.CreatedBy)
		updatedBy, _ := uuid.Parse(contentModel.UpdatedBy)

		contents = append(contents, &entities.Content{
			ID:          contentID,
			ModelType:   contentModel.ModelType,
			ModelID:     modelID,
			ContentRaw:  contentModel.ContentRaw,
			ContentHTML: contentModel.ContentHTML,
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
			CreatedAt:   time.Now(), // Not available in model
			UpdatedAt:   time.Now(), // Not available in model
		})
	}

	return contents, nil
}

func (r *PostgresContentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	contentModels, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert models to domain entities
	var contents []*entities.Content
	for _, contentModel := range contentModels {
		contentID, _ := uuid.Parse(contentModel.ID)
		modelID, _ := uuid.Parse(contentModel.ModelID)
		createdBy, _ := uuid.Parse(contentModel.CreatedBy)
		updatedBy, _ := uuid.Parse(contentModel.UpdatedBy)

		contents = append(contents, &entities.Content{
			ID:          contentID,
			ModelType:   contentModel.ModelType,
			ModelID:     modelID,
			ContentRaw:  contentModel.ContentRaw,
			ContentHTML: contentModel.ContentHTML,
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
			CreatedAt:   time.Now(), // Not available in model
			UpdatedAt:   time.Now(), // Not available in model
		})
	}

	return contents, nil
}

func (r *PostgresContentRepository) Update(ctx context.Context, content *entities.Content) error {
	// Convert domain entity to model
	contentModel := &model.Content{
		ID:          content.ID.String(),
		ModelType:   content.ModelType,
		ModelID:     content.ModelID.String(),
		ContentRaw:  content.ContentRaw,
		ContentHTML: content.ContentHTML,
		CreatedBy:   content.CreatedBy.String(),
		UpdatedBy:   content.UpdatedBy.String(),
	}

	return r.repo.Update(ctx, contentModel)
}

func (r *PostgresContentRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	return r.repo.Delete(ctx, id, deletedBy)
}

func (r *PostgresContentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.repo.HardDelete(ctx, id)
}

func (r *PostgresContentRepository) Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	return r.repo.Restore(ctx, id, updatedBy)
}

func (r *PostgresContentRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

func (r *PostgresContentRepository) CountByModelType(ctx context.Context, modelType string) (int64, error) {
	return r.repo.CountByModelType(ctx, modelType)
}

func (r *PostgresContentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	contentModels, err := r.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert models to domain entities
	var contents []*entities.Content
	for _, contentModel := range contentModels {
		contentID, _ := uuid.Parse(contentModel.ID)
		modelID, _ := uuid.Parse(contentModel.ModelID)
		createdBy, _ := uuid.Parse(contentModel.CreatedBy)
		updatedBy, _ := uuid.Parse(contentModel.UpdatedBy)

		contents = append(contents, &entities.Content{
			ID:          contentID,
			ModelType:   contentModel.ModelType,
			ModelID:     modelID,
			ContentRaw:  contentModel.ContentRaw,
			ContentHTML: contentModel.ContentHTML,
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
			CreatedAt:   time.Now(), // Not available in model
			UpdatedAt:   time.Now(), // Not available in model
		})
	}

	return contents, nil
}
