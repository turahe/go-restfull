package repository

import (
	"context"
	"fmt"
	"time"
	"webapi/internal/db/model"
	"webapi/internal/dto"
	"webapi/internal/helper/cache"
	"webapi/internal/http/requests"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MediaRepository interface {
	GetMedia(ctx context.Context) ([]*model.Media, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
	GetMediaByHash(ctx context.Context, hash string) (*model.Media, error)
	GetMediaByFileName(ctx context.Context, fileName string) (*model.Media, error)
	UpdateMedia(ctx context.Context, media model.Media) (*model.Media, error)
	GetMediaByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.Media, error)
	GetMediaWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (*dto.DataWithPaginationDTO, error)
	GetMediaByParentIDWithPagination(ctx context.Context, parentID uuid.UUID, page int, limit int) ([]*model.Media, error)
	CreateMedia(ctx context.Context, media model.Media) (*model.Media, error)
	DeleteMedia(ctx context.Context, media model.Media) (bool, error)
	AttachMedia(ctx context.Context, media dto.MediaRelation) error
	GetMediaDescendants(ctx context.Context, id uuid.UUID) ([]*model.Media, error)
	GetMediaAncestors(ctx context.Context, id uuid.UUID) ([]*model.Media, error)
	GetMediaSiblings(ctx context.Context, id uuid.UUID) ([]*model.Media, error)
	GetMediaRootNodes(ctx context.Context) ([]*model.Media, error)
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

func (m *MediaRepositoryImpl) GetMedia(ctx context.Context) ([]*model.Media, error) {
	var media []*model.Media

	// Try to get from cache first
	err := cache.GetJSON(ctx, cache.KEY_MEDIA_ALL, &media)
	if err == nil {
		return media, nil
	}

	rows, err := m.pgxPool.Query(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at FROM media ORDER BY record_left")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth, &mediaModel.ParentID,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt)
		if err != nil {
			return nil, err
		}
		media = append(media, &mediaModel)
	}

	// Cache the result
	cache.SetJSON(ctx, cache.KEY_MEDIA_ALL, media, cache.DefaultCacheDuration)

	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	var mediaModel model.Media

	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_MEDIA_BY_ID, id.String())
	err := cache.GetJSON(ctx, cacheKey, &mediaModel)
	if err == nil {
		return &mediaModel, nil
	}

	err = m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at FROM media WHERE id = $1", id).Scan(
		&mediaModel.ID,
		&mediaModel.Name,
		&mediaModel.Hash,
		&mediaModel.FileName,
		&mediaModel.Disk,
		&mediaModel.Size,
		&mediaModel.MimeType,
		&mediaModel.CustomAttributes,
		&mediaModel.RecordLeft,
		&mediaModel.RecordRight,
		&mediaModel.RecordDepth,
		&mediaModel.ParentID,
		&mediaModel.CreatedAt,
		&mediaModel.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &mediaModel, cache.DefaultCacheDuration)

	return &mediaModel, nil
}

func (m *MediaRepositoryImpl) GetMediaByHash(ctx context.Context, hash string) (*model.Media, error) {
	var mediaModel model.Media

	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_MEDIA_BY_HASH, hash)
	err := cache.GetJSON(ctx, cacheKey, &mediaModel)
	if err == nil {
		return &mediaModel, nil
	}

	err = m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at FROM media WHERE hash = $1", hash).Scan(
		&mediaModel.ID,
		&mediaModel.Name,
		&mediaModel.Hash,
		&mediaModel.FileName,
		&mediaModel.Disk,
		&mediaModel.Size,
		&mediaModel.MimeType,
		&mediaModel.CustomAttributes,
		&mediaModel.RecordLeft,
		&mediaModel.RecordRight,
		&mediaModel.RecordDepth,
		&mediaModel.ParentID,
		&mediaModel.CreatedAt,
		&mediaModel.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &mediaModel, cache.DefaultCacheDuration)

	return &mediaModel, nil
}

func (m *MediaRepositoryImpl) GetMediaByFileName(ctx context.Context, fileName string) (*model.Media, error) {
	var mediaModel model.Media

	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_MEDIA_BY_FILENAME, fileName)
	err := cache.GetJSON(ctx, cacheKey, &mediaModel)
	if err == nil {
		return &mediaModel, nil
	}

	err = m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at FROM media WHERE file_name = $1", fileName).Scan(
		&mediaModel.ID,
		&mediaModel.Name,
		&mediaModel.Hash,
		&mediaModel.FileName,
		&mediaModel.Disk,
		&mediaModel.Size,
		&mediaModel.MimeType,
		&mediaModel.CustomAttributes,
		&mediaModel.RecordLeft,
		&mediaModel.RecordRight,
		&mediaModel.RecordDepth,
		&mediaModel.ParentID,
		&mediaModel.CreatedAt,
		&mediaModel.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &mediaModel, cache.DefaultCacheDuration)

	return &mediaModel, nil
}

func (m *MediaRepositoryImpl) UpdateMedia(ctx context.Context, media model.Media) (*model.Media, error) {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = m.pgxPool.Exec(ctx, "UPDATE media SET name = $2, hash = $3, file_name = $4, disk = $5, size = $6, mime_type = $7, custom_attributes = $8, record_left = $9, record_right = $10, record_depth = $11 WHERE id = $1", media.ID,
		media.Name,
		media.Hash,
		media.FileName,
		media.Disk,
		media.Size,
		media.MimeType,
		media.CustomAttributes,
		media.RecordLeft,
		media.RecordRight,
		media.RecordDepth)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// Invalidate media cache
	cache.InvalidatePattern(ctx, cache.PATTERN_MEDIA_CACHE)

	return &media, nil
}

func (m *MediaRepositoryImpl) GetMediaByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.Media, error) {
	var media []*model.Media
	rows, err := m.pgxPool.Query(ctx, "SELECT 'id', 'name', 'hash', 'file_name', 'disk', 'size', 'mime_type', 'custom_attributes', 'record_left', 'record_right', 'record_depth' FROM media WHERE 'parent_id' = $1", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth)
		if err != nil {
			return nil, err
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (*dto.DataWithPaginationDTO, error) {
	var media []*model.Media
	var totalMedia int
	var query = input.Query
	var limit = input.Limit
	var page = input.Page

	rows, err := m.pgxPool.Query(ctx, `
	SELECT id, name, hash, file_name, disk, size, mime_type, record_left, record_right FROM media
	WHERE name ILIKE $1 OR file_name ILIKE $1
	LIMIT $2 OFFSET $3`, fmt.Sprintf("%%%s%%", query), limit, page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.RecordLeft, &mediaModel.RecordRight)
		if err != nil {
			return nil, err
		}
		media = append(media, &mediaModel)
	}
	// Query to get total user count with search functionality
	err = m.pgxPool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM users 
		WHERE username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1`, fmt.Sprintf("%%%s%%", query)).Scan(&totalMedia)
	if err != nil {
		return nil, err
	}

	// Iterate through rows and append to users slice
	var mediaDto []interface{}
	for _, u := range media {
		mediaDto = append(mediaDto, dto.GetMediaDTO{
			ID:       u.ID,
			FileName: u.FileName,
			Name:     u.Name,
			Size:     u.Size,
		})
	}
	// Calculate pagination details
	currentPage := (page / limit) + 1
	lastPage := (totalMedia + limit - 1) / limit
	// Prepare response
	responseMedia := dto.DataWithPaginationDTO{
		Total:       totalMedia,
		Limit:       limit,
		Data:        mediaDto,
		CurrentPage: currentPage,
		LastPage:    lastPage,
	}

	return &responseMedia, nil
}

func (m *MediaRepositoryImpl) GetMediaByParentIDWithPagination(ctx context.Context, parentID uuid.UUID, page int, limit int) ([]*model.Media, error) {
	var media []*model.Media
	rows, err := m.pgxPool.Query(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth FROM media WHERE parent_id = $1 LIMIT $2 OFFSET $3", parentID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth)
		if err != nil {
			return nil, err
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}
func (m *MediaRepositoryImpl) CreateMedia(ctx context.Context, media model.Media) (*model.Media, error) {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Start a transaction for nested set operations
	now := time.Now()
	media.CreatedAt = now
	media.UpdatedAt = now

	if media.ParentID != uuid.Nil {
		// Insert as child of existing parent
		query := `
			INSERT INTO media (name, hash, file_name, disk, size, mime_type, custom_attributes, parent_id, record_left, record_right, record_depth, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 
				(SELECT record_right FROM media WHERE id = $8),
				(SELECT record_right + 1 FROM media WHERE id = $8),
				(SELECT record_depth + 1 FROM media WHERE id = $8),
				$9, $10)
			RETURNING id
		`
		err = tx.QueryRow(ctx, query,
			media.Name,
			media.Hash,
			media.FileName,
			media.Disk,
			media.Size,
			media.MimeType,
			media.CustomAttributes,
			media.ParentID,
			media.CreatedAt,
			media.UpdatedAt,
		).Scan(&media.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create media: %w", err)
		}

		// Update the record_left and record_right values for all nodes to the right of the parent
		updateQuery := `
			UPDATE media 
			SET record_left = CASE 
				WHEN record_left > (SELECT record_right FROM media WHERE id = $1) THEN record_left + 2
				ELSE record_left
			END,
			record_right = CASE 
				WHEN record_right >= (SELECT record_right FROM media WHERE id = $1) THEN record_right + 2
				ELSE record_right
			END
			WHERE record_right >= (SELECT record_right FROM media WHERE id = $1)
		`
		_, err = tx.Exec(ctx, updateQuery, media.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update nested set values: %w", err)
		}

		// Update the new node's record_left and record_right
		finalUpdateQuery := `
			UPDATE media 
			SET record_left = (SELECT record_right - 1 FROM media WHERE id = $1),
				record_right = (SELECT record_right FROM media WHERE id = $1)
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, finalUpdateQuery, media.ParentID, media.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update new node values: %w", err)
		}
	} else {
		// Insert as root node
		// Find the maximum record_right value
		var maxRight int64
		err = tx.QueryRow(ctx, "SELECT COALESCE(MAX(record_right), 0) FROM media").Scan(&maxRight)
		if err != nil {
			return nil, fmt.Errorf("failed to get max record_right: %w", err)
		}

		media.RecordLeft = uint64(maxRight + 1)
		media.RecordRight = uint64(maxRight + 2)
		media.RecordDepth = 0

		query := `
			INSERT INTO media (name, hash, file_name, disk, size, mime_type, custom_attributes, parent_id, record_left, record_right, record_depth, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id
		`
		err = tx.QueryRow(ctx, query,
			media.Name,
			media.Hash,
			media.FileName,
			media.Disk,
			media.Size,
			media.MimeType,
			media.CustomAttributes,
			media.ParentID,
			media.RecordLeft,
			media.RecordRight,
			media.RecordDepth,
			media.CreatedAt,
			media.UpdatedAt,
		).Scan(&media.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create media: %w", err)
		}
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate media cache
	cache.InvalidatePattern(ctx, cache.PATTERN_MEDIA_CACHE)

	return &media, nil
}
func (m *MediaRepositoryImpl) DeleteMedia(ctx context.Context, media model.Media) (bool, error) {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	_, err = m.pgxPool.Exec(ctx, "UPDATE media SET deleted_at = NOW() WHERE id = $1", media.ID)
	if err != nil {
		return false, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return false, err
	}

	// Invalidate media cache
	cache.InvalidatePattern(ctx, cache.PATTERN_MEDIA_CACHE)

	return true, nil
}

func (m *MediaRepositoryImpl) AttachMedia(ctx context.Context, media dto.MediaRelation) error {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, "INSERT INTO mediables (media_id, mediable_id, mediable_type, group) VALUES ($1, $2, $3, $4) RETURNING media_id", media.MediaID, media.MediableId, media.MediableType, media.Group).Scan(&media.MediaID)
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaRepositoryImpl) GetMediaDescendants(ctx context.Context, id uuid.UUID) ([]*model.Media, error) {
	var media []*model.Media
	query := `
		SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at
		FROM media
		WHERE record_left > (SELECT record_left FROM media WHERE id = $1)
		AND record_right < (SELECT record_right FROM media WHERE id = $1)
		ORDER BY record_left
	`
	rows, err := m.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth, &mediaModel.ParentID,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaAncestors(ctx context.Context, id uuid.UUID) ([]*model.Media, error) {
	var media []*model.Media
	query := `
		SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at
		FROM media
		WHERE record_left < (SELECT record_left FROM media WHERE id = $1)
		AND record_right > (SELECT record_right FROM media WHERE id = $1)
		ORDER BY record_left
	`
	rows, err := m.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth, &mediaModel.ParentID,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaSiblings(ctx context.Context, id uuid.UUID) ([]*model.Media, error) {
	var media []*model.Media
	query := `
		SELECT m1.id, m1.name, m1.hash, m1.file_name, m1.disk, m1.size, m1.mime_type, m1.custom_attributes, m1.record_left, m1.record_right, m1.record_depth, m1.parent_id, m1.created_at, m1.updated_at
		FROM media m1
		JOIN media m2 ON m1.parent_id = m2.parent_id
		WHERE m2.id = $1 AND m1.id != $1
		ORDER BY m1.record_left
	`
	rows, err := m.pgxPool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get siblings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth, &mediaModel.ParentID,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaRootNodes(ctx context.Context) ([]*model.Media, error) {
	var media []*model.Media
	query := `
		SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth, parent_id, created_at, updated_at
		FROM media
		WHERE parent_id = $1
		ORDER BY record_left
	`
	rows, err := m.pgxPool.Query(ctx, query, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get root nodes: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.CustomAttributes, &mediaModel.RecordLeft,
			&mediaModel.RecordRight, &mediaModel.RecordDepth, &mediaModel.ParentID,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}
		media = append(media, &mediaModel)
	}
	return media, nil
}
