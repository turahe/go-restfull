package adapters

import (
	"context"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresContentRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresContentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.ContentRepository {
	return &PostgresContentRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
	}
}

func (r *PostgresContentRepository) Create(ctx context.Context, content *entities.Content) error {
	query := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.Exec(ctx, query,
		content.ID, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML,
		content.CreatedBy, content.UpdatedBy, content.CreatedAt, content.UpdatedAt,
	)
	return err
}

func (r *PostgresContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE id = $1`
	var c entities.Content
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML,
		&c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *PostgresContentRepository) GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE model_type = $1 AND model_id = $2 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, modelType, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

func (r *PostgresContentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

func (r *PostgresContentRepository) Update(ctx context.Context, content *entities.Content) error {
	query := `UPDATE contents SET model_type=$1, model_id=$2, content_raw=$3, content_html=$4, updated_by=$5, updated_at=$6 WHERE id=$7`
	_, err := r.db.Exec(ctx, query, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML, content.UpdatedBy, content.UpdatedAt, content.ID)
	return err
}

func (r *PostgresContentRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_by=$1, deleted_at=NOW(), updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, deletedBy, id)
	return err
}

func (r *PostgresContentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM contents WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresContentRepository) Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	query := `UPDATE contents SET deleted_by=NULL, deleted_at=NULL, updated_by=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, query, updatedBy, id)
	return err
}

func (r *PostgresContentRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresContentRepository) CountByModelType(ctx context.Context, modelType string) (int64, error) {
	query := `SELECT COUNT(*) FROM contents WHERE model_type = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, modelType).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresContentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	q := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at, deleted_by, deleted_at
		FROM contents WHERE deleted_at IS NULL AND (content_raw ILIKE $1 OR content_html ILIKE $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Content
	for rows.Next() {
		var c entities.Content
		if err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedBy, &c.DeletedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}
