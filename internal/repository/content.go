package repository

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ContentRepository interface {
	Create(ctx context.Context, content *model.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error)
	GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*model.Content, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Content, error)
	Update(ctx context.Context, content *model.Content) error
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByModelType(ctx context.Context, modelType string) (int64, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*model.Content, error)
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

func (r *ContentRepositoryImpl) Create(ctx context.Context, content *model.Content) error {
	query := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`

	_, err := r.pgxPool.Exec(ctx, query,
		content.ID, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML, content.CreatedBy, content.UpdatedBy)
	return err
}

func (r *ContentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by 
			  FROM contents WHERE id = $1 AND deleted_at IS NULL`

	var content model.Content
	err := r.pgxPool.QueryRow(ctx, query, id).Scan(
		&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML, &content.CreatedBy, &content.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (r *ContentRepositoryImpl) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*model.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by 
			  FROM contents WHERE model_type = $1 AND model_id = $2 AND deleted_at IS NULL 
			  ORDER BY created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, modelType, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*model.Content
	for rows.Next() {
		var content model.Content
		err := rows.Scan(&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML, &content.CreatedBy, &content.UpdatedBy)
		if err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, nil
}

func (r *ContentRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*model.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by 
			  FROM contents WHERE deleted_at IS NULL 
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*model.Content
	for rows.Next() {
		var content model.Content
		err := rows.Scan(&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML, &content.CreatedBy, &content.UpdatedBy)
		if err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, nil
}

func (r *ContentRepositoryImpl) Update(ctx context.Context, content *model.Content) error {
	query := `UPDATE contents SET content_raw = $1, content_html = $2, updated_by = $3, updated_at = NOW() 
			  WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, content.ContentRaw, content.ContentHTML, content.UpdatedBy, content.ID)
	return err
}

func (r *ContentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_at = NOW(), deleted_by = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.pgxPool.Exec(ctx, query, deletedBy, id)
	return err
}

func (r *ContentRepositoryImpl) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contents WHERE id = $1`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

func (r *ContentRepositoryImpl) Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_at = NULL, deleted_by = NULL, updated_by = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.pgxPool.Exec(ctx, query, updatedBy, id)
	return err
}

func (r *ContentRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *ContentRepositoryImpl) CountByModelType(ctx context.Context, modelType string) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE model_type = $1 AND deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query, modelType).Scan(&count)
	return count, err
}

func (r *ContentRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*model.Content, error) {
	searchQuery := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by 
					FROM contents WHERE deleted_at IS NULL AND 
					(content_raw ILIKE '%' || $1 || '%' OR content_html ILIKE '%' || $1 || '%') 
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pgxPool.Query(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*model.Content
	for rows.Next() {
		var content model.Content
		err := rows.Scan(&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML, &content.CreatedBy, &content.UpdatedBy)
		if err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, nil
}
