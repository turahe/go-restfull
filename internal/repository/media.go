package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MediaRepository interface {
	Create(ctx context.Context, media *entities.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error)
	Update(ctx context.Context, media *entities.Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByFilename(ctx context.Context, filename string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error)
}

type MediaRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewMediaRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MediaRepository {
	return &MediaRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *MediaRepositoryImpl) Create(ctx context.Context, media *entities.Media) error {
	query := `INSERT INTO media (id, file_name, name, mime_type, size, disk, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pgxPool.Exec(ctx, query,
		media.ID.String(), media.FileName, media.Name, media.MimeType, media.Size,
		media.Disk, media.CreatedBy, media.UpdatedBy, media.CreatedAt, media.UpdatedAt)
	return err
}

func (r *MediaRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	query := `SELECT id, file_name, name, mime_type, size, disk, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM media WHERE id = $1 AND deleted_at IS NULL`

	var media entities.Media
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&media.ID, &media.FileName, &media.Name, &media.MimeType, &media.Size,
		&media.Disk, &media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &media, nil
}

func (r *MediaRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	query := `SELECT id, file_name, name, mime_type, size, disk, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM media WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

func (r *MediaRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	query := `SELECT id, file_name, name, mime_type, size, disk, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM media WHERE user_id = $1 AND deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pgxPool.Query(ctx, query, userID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

func (r *MediaRepositoryImpl) Update(ctx context.Context, media *entities.Media) error {
	query := `UPDATE media SET file_name = $1, name = $2, mime_type = $3, size = $4, disk = $5, updated_by = $6, updated_at = $7
			  WHERE id = $8 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, media.FileName, media.Name, media.MimeType, media.Size,
		media.Disk, media.UpdatedBy, media.UpdatedAt, media.ID.String())
	return err
}

func (r *MediaRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE media SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *MediaRepositoryImpl) ExistsByFilename(ctx context.Context, filename string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM media WHERE file_name = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, filename).Scan(&exists)
	return exists, err
}

func (r *MediaRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *MediaRepositoryImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE user_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, userID.String()).Scan(&count)
	return count, err
}

func (r *MediaRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	searchQuery := `SELECT id, file_name, name, mime_type, size, disk, created_by, updated_by, created_at, updated_at, deleted_at
					FROM media WHERE deleted_at IS NULL 
					AND (file_name ILIKE $1 OR name ILIKE $1 OR mime_type ILIKE $1)
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	searchPattern := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*entities.Media
	for rows.Next() {
		media, err := r.scanMediaRow(rows)
		if err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// scanMediaRow is a helper function to scan a media row from database
func (r *MediaRepositoryImpl) scanMediaRow(rows pgx.Rows) (*entities.Media, error) {
	var media entities.Media
	err := rows.Scan(
		&media.ID, &media.FileName, &media.Name, &media.MimeType, &media.Size,
		&media.Disk, &media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &media, nil
}
