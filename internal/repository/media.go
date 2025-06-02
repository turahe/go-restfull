package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"webapi/internal/db/model"
	"webapi/internal/dto"
	"webapi/internal/http/requests"
)

type MediaRepository interface {
	GetMedia(ctx context.Context) ([]model.Media, error)
	GetMediaByID(ctx context.Context, id uuid.UUID) (model.Media, error)
	GetMediaByHash(ctx context.Context, hash string) (model.Media, error)
	GetMediaByFileName(ctx context.Context, fileName string) (model.Media, error)
	UpdateMedia(ctx context.Context, media model.Media) (model.Media, error)
	GetMediaByParentID(ctx context.Context, parentID uuid.UUID) ([]model.Media, error)
	GetMediaWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (dto.DataWithPaginationDTO, error)
	GetMediaByParentIDWithPagination(ctx context.Context, parentID uuid.UUID, page int, limit int) ([]model.Media, error)
	CreateMedia(ctx context.Context, media model.Media) (model.Media, error)
	DeleteMedia(ctx context.Context, media model.Media) (bool, error)
	AttachMedia(ctx context.Context, media dto.MediaRelation) error
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
func (m *MediaRepositoryImpl) GetMedia(ctx context.Context) ([]model.Media, error) {
	var media []model.Media
	rows, err := m.pgxPool.Query(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth FROM media")
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
		media = append(media, mediaModel)
	}
	return media, nil
}
func (m *MediaRepositoryImpl) GetMediaByID(ctx context.Context, id uuid.UUID) (model.Media, error) {
	var mediaModel model.Media
	err := m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth FROM media WHERE id = $1", id).Scan(
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
		&mediaModel.RecordDepth)
	if err != nil {
		return model.Media{}, err
	}
	return mediaModel, nil
}
func (m *MediaRepositoryImpl) GetMediaByHash(ctx context.Context, hash string) (model.Media, error) {
	var mediaModel model.Media
	err := m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth FROM media WHERE hash = $1", hash).Scan(
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
		&mediaModel.RecordDepth)
	if err != nil {
		return model.Media{}, err
	}
	return mediaModel, nil
}

func (m *MediaRepositoryImpl) GetMediaByFileName(ctx context.Context, fileName string) (model.Media, error) {
	var mediaModel model.Media
	err := m.pgxPool.QueryRow(ctx, "SELECT id, name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth FROM media WHERE file_name = $1", fileName).Scan(
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
		&mediaModel.RecordDepth)
	if err != nil {
		return model.Media{}, err
	}
	return mediaModel, nil
}

func (m *MediaRepositoryImpl) UpdateMedia(ctx context.Context, media model.Media) (model.Media, error) {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return model.Media{}, err
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
		return model.Media{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Media{}, err
	}

	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaByParentID(ctx context.Context, parentID uuid.UUID) ([]model.Media, error) {
	var media []model.Media
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
		media = append(media, mediaModel)
	}
	return media, nil
}

func (m *MediaRepositoryImpl) GetMediaWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (dto.DataWithPaginationDTO, error) {
	var media []model.Media
	var totalMedia int
	var query = input.Query
	var limit = input.Limit
	var page = input.Page

	rows, err := m.pgxPool.Query(ctx, `
	SELECT id, name, hash, file_name, disk, size, mime_type, record_left, record_right FROM media
	WHERE name ILIKE $1 OR file_name ILIKE $1
	LIMIT $2 OFFSET $3`, fmt.Sprintf("%%%s%%", query), limit, page)
	if err != nil {
		return dto.DataWithPaginationDTO{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var mediaModel model.Media
		err = rows.Scan(&mediaModel.ID, &mediaModel.Name, &mediaModel.Hash, &mediaModel.FileName,
			&mediaModel.Disk, &mediaModel.Size, &mediaModel.MimeType,
			&mediaModel.RecordLeft, &mediaModel.RecordRight)
		if err != nil {
			return dto.DataWithPaginationDTO{}, err
		}
		media = append(media, mediaModel)
	}
	// Query to get total user count with search functionality
	err = m.pgxPool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM users 
		WHERE username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1`, fmt.Sprintf("%%%s%%", query)).Scan(&totalMedia)
	if err != nil {
		return dto.DataWithPaginationDTO{}, err
	}

	// Iterate through rows and append to users slice
	var mediaDto []interface{}
	for _, u := range media {
		mediaDto = append(mediaDto, dto.GetMediaDTO{
			ID:        u.ID,
			FileName:  u.FileName,
			Name:      u.Name,
			Size:      u.Size,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
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

	return responseMedia, nil
}

func (m *MediaRepositoryImpl) GetMediaByParentIDWithPagination(ctx context.Context, parentID uuid.UUID, page int, limit int) ([]model.Media, error) {
	var media []model.Media
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
		media = append(media, mediaModel)
	}
	return media, nil
}
func (m *MediaRepositoryImpl) CreateMedia(ctx context.Context, media model.Media) (model.Media, error) {
	tx, err := m.pgxPool.Begin(ctx)
	if err != nil {
		return model.Media{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, "INSERT INTO media (name, hash, file_name, disk, size, mime_type, custom_attributes, record_left, record_right, record_depth) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		media.Name,
		media.Hash,
		media.FileName,
		media.Disk,
		media.Size,
		media.MimeType,
		media.CustomAttributes,
		media.RecordLeft,
		media.RecordRight,
		media.RecordDepth).Scan(&media.ID)
	if err != nil {
		return model.Media{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Media{}, err
	}

	return media, nil
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
