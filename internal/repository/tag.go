package repository

import (
	"context"
	"fmt"
	"time"

	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Tag, error)
	GetAll(ctx context.Context) ([]*model.Tag, error)
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error
	DetachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error
	GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*model.Tag, error)
}

type TagRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewTagRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) TagRepository {
	return &TagRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *TagRepositoryImpl) Create(ctx context.Context, tag *model.Tag) error {
	tag.ID = uuid.New().String()
	tag.CreatedBy = ""
	tag.UpdatedBy = ""
	tag.DeletedBy = ""
	tag.DeletedAt = time.Time{}

	query := `INSERT INTO tags (id, name, slug, created_by, updated_by, deleted_by, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pgxPool.Exec(ctx, query, tag.ID, tag.Name, tag.Slug, tag.CreatedBy, tag.UpdatedBy, tag.DeletedBy, tag.DeletedAt)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Tag, error) {
	query := `SELECT id, name, slug, created_by, updated_by, deleted_by, deleted_at FROM tags WHERE id = $1`
	tag := &model.Tag{}
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedBy, &tag.UpdatedBy, &tag.DeletedBy, &tag.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

func (r *TagRepositoryImpl) GetAll(ctx context.Context) ([]*model.Tag, error) {
	query := `SELECT id, name, slug, created_by, updated_by, deleted_by, deleted_at FROM tags`
	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	tags := []*model.Tag{}
	for rows.Next() {
		tag := &model.Tag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedBy, &tag.UpdatedBy, &tag.DeletedBy, &tag.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *TagRepositoryImpl) Update(ctx context.Context, tag *model.Tag) error {
	query := `UPDATE tags SET name = $1, slug = $2, updated_by = $3, deleted_by = $4, deleted_at = $5 WHERE id = $6`
	_, err := r.pgxPool.Exec(ctx, query, tag.Name, tag.Slug, tag.UpdatedBy, tag.DeletedBy, tag.DeletedAt, tag.ID)
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error {
	query := `INSERT INTO taggables (id, tag_id, taggable_id, taggable_type, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pgxPool.Exec(ctx, query, uuid.New().String(), tagID.String(), taggableID.String(), taggableType, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("failed to attach tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) DetachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error {
	query := `DELETE FROM taggables WHERE tag_id = $1 AND taggable_id = $2 AND taggable_type = $3`
	_, err := r.pgxPool.Exec(ctx, query, tagID.String(), taggableID.String(), taggableType)
	if err != nil {
		return fmt.Errorf("failed to detach tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*model.Tag, error) {
	query := `SELECT t.id, t.name, t.slug, t.created_by, t.updated_by, t.deleted_by, t.deleted_at FROM tags t JOIN taggables tg ON t.id = tg.tag_id WHERE tg.taggable_id = $1 AND tg.taggable_type = $2`
	rows, err := r.pgxPool.Query(ctx, query, taggableID.String(), taggableType)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for entity: %w", err)
	}
	defer rows.Close()

	tags := []*model.Tag{}
	for rows.Next() {
		tag := &model.Tag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedBy, &tag.UpdatedBy, &tag.DeletedBy, &tag.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
