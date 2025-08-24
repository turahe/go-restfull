package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresMediaRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresMediaRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MediaRepository {
	return &PostgresMediaRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

func (r *PostgresMediaRepository) Create(ctx context.Context, media *entities.Media) error {
	query := `
		INSERT INTO media (
			id, name, file_name, hash, disk, mime_type, size,
			record_left, record_right, record_depth, record_ordering,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`
	_, err := r.db.Exec(ctx, query,
		media.ID, media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size,
		media.RecordLeft, media.RecordRight, media.RecordDepth, media.RecordOrdering,
		media.CreatedBy, media.UpdatedBy, media.CreatedAt, media.UpdatedAt)
	return err
}

func (r *PostgresMediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Media, error) {
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE id = $1 AND deleted_at IS NULL
	`
	var media entities.Media
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *PostgresMediaRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Media, error) {
	query := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Media
	for rows.Next() {
		var media entities.Media
		if err := rows.Scan(
			&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
			&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
			&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &media)
	}
	return list, nil
}

func (r *PostgresMediaRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Media, error) {
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
	var list []*entities.Media
	for rows.Next() {
		var media entities.Media
		if err := rows.Scan(
			&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
			&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
			&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &media)
	}
	return list, nil
}

func (r *PostgresMediaRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Media, error) {
	q := `
		SELECT id, name, file_name, hash, disk, mime_type, size,
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM media WHERE deleted_at IS NULL AND (file_name ILIKE $1 OR name ILIKE $1 OR mime_type ILIKE $1)
		ORDER BY record_left ASC LIMIT $2 OFFSET $3
	`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entities.Media
	for rows.Next() {
		var media entities.Media
		if err := rows.Scan(
			&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
			&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
			&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &media)
	}
	return list, nil
}

func (r *PostgresMediaRepository) Update(ctx context.Context, media *entities.Media) error {
	query := `UPDATE media SET name=$1, file_name=$2, hash=$3, disk=$4, mime_type=$5, size=$6, updated_by=$7, updated_at=$8 WHERE id = $9 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, media.Name, media.FileName, media.Hash, media.Disk, media.MimeType, media.Size, media.UpdatedBy, media.UpdatedAt, media.ID)
	return err
}

func (r *PostgresMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE media SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresMediaRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresMediaRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM media WHERE created_by = $1 AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresMediaRepository) GetAvatarByUserID(ctx context.Context, userID uuid.UUID) (*entities.Media, error) {
	return r.GetByGroup(ctx, userID, "User", "avatar")
}

func (r *PostgresMediaRepository) GetByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string) (*entities.Media, error) {
	query := `SELECT m.id, m.name, m.file_name, m.hash, m.disk, m.mime_type, m.size,
		m.record_left, m.record_right, m.record_depth, m.record_ordering,
		m.created_by, m.updated_by, m.created_at, m.updated_at, m.deleted_at
	FROM media m
	INNER JOIN mediables mb ON m.id = mb.media_id
	WHERE mb.mediable_id = $1 
		AND mb.mediable_type = $2 
		AND mb.group = $3
		AND m.deleted_at IS NULL
	ORDER BY m.created_at DESC
	LIMIT 1`

	var media entities.Media
	if err := r.db.QueryRow(ctx, query, mediableID, mediableType, group).Scan(
		&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
		&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
		&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *PostgresMediaRepository) GetAllByGroup(ctx context.Context, mediableID uuid.UUID, mediableType, group string, limit, offset int) ([]*entities.Media, error) {
	query := `SELECT m.id, m.name, m.file_name, m.hash, m.disk, m.mime_type, m.size,
		m.record_left, m.record_right, m.record_depth, m.record_ordering,
		m.created_by, m.updated_by, m.created_at, m.updated_at, m.deleted_at
	FROM media m
	INNER JOIN mediables mb ON m.id = mb.media_id
	WHERE mb.mediable_id = $1 
		AND mb.mediable_type = $2 
		AND mb.group = $3
		AND m.deleted_at IS NULL
	ORDER BY m.created_at DESC
	LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, query, mediableID, mediableType, group, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []*entities.Media
	for rows.Next() {
		var media entities.Media
		if err := rows.Scan(
			&media.ID, &media.Name, &media.FileName, &media.Hash, &media.Disk, &media.MimeType, &media.Size,
			&media.RecordLeft, &media.RecordRight, &media.RecordDepth, &media.RecordOrdering,
			&media.CreatedBy, &media.UpdatedBy, &media.CreatedAt, &media.UpdatedAt, &media.DeletedAt,
		); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, &media)
	}

	return mediaList, nil
}
