package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ContentRepository interface {
	Create(ctx context.Context, content *entities.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error)
	GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error)
	Update(ctx context.Context, content *entities.Content) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) (int64, error)
}

type ContentRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewContentRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) ContentRepository {
	return &ContentRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *ContentRepositoryImpl) Create(ctx context.Context, content *entities.Content) error {
	query := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pgxPool.Exec(ctx, query,
		content.ID.String(), content.ModelType, content.ModelID.String(), content.ContentRaw, content.ContentHTML,
		content.CreatedBy, content.UpdatedBy, content.CreatedAt, content.UpdatedAt)
	return err
}

func (r *ContentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents WHERE id = $1`

	var content entities.Content
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML,
		&content.CreatedBy, &content.UpdatedBy, &content.CreatedAt, &content.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (r *ContentRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContentRow(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (r *ContentRepositoryImpl) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at
			  FROM contents WHERE model_type = $1 AND model_id = $2
			  ORDER BY created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, modelType, modelID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContentRow(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (r *ContentRepositoryImpl) Update(ctx context.Context, content *entities.Content) error {
	query := `UPDATE contents SET model_type = $1, model_id = $2, content_raw = $3, content_html = $4, updated_at = $5, updated_by = $6
			  WHERE id = $7`

	_, err := r.pgxPool.Exec(ctx, query, content.ModelType, content.ModelID.String(), content.ContentRaw, content.ContentHTML,
		content.UpdatedAt, content.UpdatedBy, content.ID.String())
	return err
}

func (r *ContentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contents WHERE id = $1`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *ContentRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM contents`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *ContentRepositoryImpl) CountByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE model_type = $1 AND model_id = $2`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, modelType, modelID.String()).Scan(&count)
	return count, err
}

// scanContentRow is a helper function to scan a content row from database
func (r *ContentRepositoryImpl) scanContentRow(rows pgx.Rows) (*entities.Content, error) {
	var content entities.Content
	err := rows.Scan(
		&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML,
		&content.CreatedBy, &content.UpdatedBy, &content.CreatedAt, &content.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
