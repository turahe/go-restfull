package repository

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

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

type mediaRepository struct {
	db          *pgxpool.Pool
	redisClient redis.Cmdable
	nestedSet   *nestedset.NestedSetManager
}

func NewMediaRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) MediaRepository {
	return &mediaRepository{
		db:          pgxPool,
		redisClient: redisClient,
		nestedSet:   nestedset.NewNestedSetManager(pgxPool),
	}
}

func (r *mediaRepository) Create(ctx context.Context, media *entities.Media) error {
	// For media, we'll treat it as a flat structure initially
	// If we need to implement folder-like hierarchy later, we can add parent_id support
	values, err := r.nestedSet.CreateNode(ctx, "media", nil, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign computed nested set values to the entity
	media.RecordLeft = &values.Left
	media.RecordRight = &values.Right
	media.RecordDepth = &values.Depth
	media.RecordOrdering = &values.Ordering

	// Insert the new media
	query := `
		INSERT INTO media (
			id, name, file_name, hash, disk, mime_type, size,
			record_left, record_right, record_depth, record_ordering,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = r.db.Exec(ctx, query,
		media.ID, media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size,
		media.RecordLeft, media.RecordRight, media.RecordDepth, media.RecordOrdering,
		media.CreatedBy, media.UpdatedBy, media.CreatedAt, media.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	return nil
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE id = $1 AND deleted_at IS NULL
	`

	var media entities.Media
	err := r.db.QueryRow(ctx, query, id.String()).Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &media, nil
}

func (r *mediaRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
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

func (r *mediaRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID.String(), limit, offset)
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

func (r *mediaRepository) Update(ctx context.Context, media *entities.Media) error {
	// For updates, we only update basic fields, not the tree structure
	query := `
		UPDATE media SET name = $1, file_name = $2, hash = $3, disk = $4, 
		                mime_type = $5, size = $6, updated_by = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size,
		media.UpdatedBy, media.UpdatedAt, media.ID.String())
	return err
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete
	query := `UPDATE media SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id.String())
	return err
}

func (r *mediaRepository) ExistsByFilename(ctx context.Context, filename string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM media WHERE file_name = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, filename).Scan(&exists)
	return exists, err
}

func (r *mediaRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *mediaRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE created_by = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID.String()).Scan(&count)
	return count, err
}

func (r *mediaRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	searchQuery := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE deleted_at IS NULL 
		AND (file_name ILIKE $1 OR name ILIKE $1 OR mime_type ILIKE $1)
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchPattern, limit, offset)
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
func (r *mediaRepository) scanMediaRow(rows pgx.Rows) (*entities.Media, error) {
	var media entities.Media
	err := rows.Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &media, nil
}
