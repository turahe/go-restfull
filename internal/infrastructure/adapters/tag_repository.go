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

type PostgresTagRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresTagRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.TagRepository {
	return &PostgresTagRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

func (r *PostgresTagRepository) Create(ctx context.Context, tag *entities.Tag) error {
	query := `INSERT INTO tags (id, name, slug, color, created_by, updated_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.Exec(ctx, query, tag.ID, tag.Name, tag.Slug, tag.Color, tag.CreatedBy, tag.UpdatedBy, tag.CreatedAt, tag.UpdatedAt)
	return err
}

func (r *PostgresTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE id = $1 AND deleted_at IS NULL`
	var tag entities.Tag
	if err := r.db.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *PostgresTagRepository) GetBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE slug = $1 AND deleted_at IS NULL`
	var tag entities.Tag
	if err := r.db.QueryRow(ctx, query, slug).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *PostgresTagRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	query := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*entities.Tag
	for rows.Next() {
		var tag entities.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (r *PostgresTagRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	q := `SELECT id, name, slug, color, created_by, updated_by, created_at, updated_at, deleted_at FROM tags WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*entities.Tag
	for rows.Next() {
		var tag entities.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Color, &tag.CreatedBy, &tag.UpdatedBy, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (r *PostgresTagRepository) Update(ctx context.Context, tag *entities.Tag) error {
	query := `UPDATE tags SET name=$1, slug=$2, color=$3, updated_at=$4, updated_by=$5 WHERE id=$6 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, tag.Name, tag.Slug, tag.Color, tag.UpdatedAt, tag.UpdatedBy, tag.ID)
	return err
}

func (r *PostgresTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTagRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tags WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresTagRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tags WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
