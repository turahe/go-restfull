package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TagRepository interface {
	Create(ctx context.Context, tag *entities.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error)
	GetAll(ctx context.Context) ([]*entities.Tag, error)
	Update(ctx context.Context, tag *entities.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error
	DetachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error
	GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*entities.Tag, error)
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

func (r *TagRepositoryImpl) Create(ctx context.Context, tag *entities.Tag) error {
	query := `INSERT INTO tags (id, name, slug, color, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pgxPool.Exec(ctx, query, tag.ID.String(), tag.Name, tag.Slug, tag.Color, tag.CreatedAt, tag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_at, updated_at, deleted_at FROM tags WHERE id = $1 AND deleted_at IS NULL`

	var tag entities.Tag
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return &tag, nil
}

func (r *TagRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL`
	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []*entities.Tag
	for rows.Next() {
		tag, err := r.scanTagRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *TagRepositoryImpl) Update(ctx context.Context, tag *entities.Tag) error {
	query := `UPDATE tags SET name = $1, slug = $2, color = $3, updated_at = $4 WHERE id = $5 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, tag.Name, tag.Slug, tag.Color, tag.UpdatedAt, tag.ID.String())
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

func (r *TagRepositoryImpl) AttachTag(ctx context.Context, tagID, taggableID uuid.UUID, taggableType string) error {
	query := `INSERT INTO taggables (id, tag_id, taggable_id, taggable_type, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pgxPool.Exec(ctx, query, uuid.New().String(), tagID.String(), taggableID.String(), taggableType, time.Now())
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

func (r *TagRepositoryImpl) GetTagsForEntity(ctx context.Context, taggableID uuid.UUID, taggableType string) ([]*entities.Tag, error) {
	query := `SELECT t.id, t.name, t.slug, t.color, t.created_at, t.updated_at, t.deleted_at FROM tags t JOIN taggables tg ON t.id = tg.tag_id WHERE tg.taggable_id = $1 AND tg.taggable_type = $2 AND t.deleted_at IS NULL`
	rows, err := r.pgxPool.Query(ctx, query, taggableID.String(), taggableType)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for entity: %w", err)
	}
	defer rows.Close()

	var tags []*entities.Tag
	for rows.Next() {
		tag, err := r.scanTagRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// scanTagRow is a helper function to scan a tag row from database
func (r *TagRepositoryImpl) scanTagRow(rows pgx.Rows) (*entities.Tag, error) {
	var tag entities.Tag
	err := rows.Scan(
		&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}
