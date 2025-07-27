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

type PostgresTagRepository struct {
	repo repository.TagRepository
}

func NewPostgresTagRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TagRepository {
	return &PostgresTagRepository{
		repo: repository.NewTagRepository(db, redisClient),
	}
}

func (r *PostgresTagRepository) Create(ctx context.Context, tag *entities.Tag) error {
	// Convert domain entity to model
	tagModel := &model.Tag{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedBy: "",
		UpdatedBy: "",
		DeletedBy: "",
		DeletedAt: time.Time{},
	}

	return r.repo.Create(ctx, tagModel)
}

func (r *PostgresTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	tagModel, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert model to domain entity
	tagID, _ := uuid.Parse(tagModel.ID)
	var deletedAt *time.Time
	if !tagModel.DeletedAt.IsZero() {
		deletedAt = &tagModel.DeletedAt
	}

	tag := &entities.Tag{
		ID:          tagID,
		Name:        tagModel.Name,
		Slug:        tagModel.Slug,
		Description: "",
		Color:       "",
		CreatedAt:   time.Now(), // Not available in model
		UpdatedAt:   time.Now(), // Not available in model
		DeletedAt:   deletedAt,
	}

	return tag, nil
}

func (r *PostgresTagRepository) GetBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning nil as the existing repository doesn't have this method
	return nil, nil
}

func (r *PostgresTagRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	tagModels, err := r.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Convert models to domain entities
	var tags []*entities.Tag
	for _, tagModel := range tagModels {
		tagID, _ := uuid.Parse(tagModel.ID)
		var deletedAt *time.Time
		if !tagModel.DeletedAt.IsZero() {
			deletedAt = &tagModel.DeletedAt
		}

		tags = append(tags, &entities.Tag{
			ID:          tagID,
			Name:        tagModel.Name,
			Slug:        tagModel.Slug,
			Description: "",
			Color:       "",
			CreatedAt:   time.Now(), // Not available in model
			UpdatedAt:   time.Now(), // Not available in model
			DeletedAt:   deletedAt,
		})
	}

	return tags, nil
}

func (r *PostgresTagRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning empty slice as the existing repository doesn't have this method
	return []*entities.Tag{}, nil
}

func (r *PostgresTagRepository) Update(ctx context.Context, tag *entities.Tag) error {
	// Convert domain entity to model
	var deletedAt time.Time
	if tag.DeletedAt != nil {
		deletedAt = *tag.DeletedAt
	}

	tagModel := &model.Tag{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedBy: "",
		UpdatedBy: "",
		DeletedBy: "",
		DeletedAt: deletedAt,
	}

	return r.repo.Update(ctx, tagModel)
}

func (r *PostgresTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresTagRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning false as the existing repository doesn't have this method
	return false, nil
}

func (r *PostgresTagRepository) Count(ctx context.Context) (int64, error) {
	// This would need to be implemented based on your specific requirements
	// For now, returning 0 as the existing repository doesn't have this method
	return 0, nil
}
