package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *entities.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error)
	GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error)
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error)
}

type CommentRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewCommentRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) CommentRepository {
	return &CommentRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *CommentRepositoryImpl) Create(ctx context.Context, comment *entities.Comment) error {
	query := `INSERT INTO comments (id, model_type, model_id, parent_id, status, created_by, updated_by, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	parentIDStr := ""
	if comment.ParentID != nil {
		parentIDStr = comment.ParentID.String()
	}

	_, err := r.pgxPool.Exec(ctx, query,
		comment.ID.String(), comment.ModelType, comment.ModelID.String(), parentIDStr,
		comment.Status, comment.CreatedBy, comment.UpdatedBy, comment.CreatedAt, comment.UpdatedAt)

	return err
}

func (r *CommentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	query := `SELECT id, model_type, model_id, parent_id, status, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM comments WHERE id = $1 AND deleted_at IS NULL`

	var comment entities.Comment
	var parentIDStr *string

	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&comment.ID, &comment.ModelType, &comment.ModelID, &parentIDStr, &comment.Status,
		&comment.CreatedBy, &comment.UpdatedBy, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}

	return &comment, nil
}

func (r *CommentRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `SELECT id, model_type, model_id, parent_id, status, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM comments WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *CommentRepositoryImpl) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `SELECT id, model_type, model_id, parent_id, status, created_by, updated_by, created_at, updated_at, deleted_at
			  FROM comments WHERE post_id = $1 AND deleted_at IS NULL
			  ORDER BY created_at ASC LIMIT $2 OFFSET $3`

	rows, err := r.pgxPool.Query(ctx, query, postID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *CommentRepositoryImpl) Update(ctx context.Context, comment *entities.Comment) error {
	query := `UPDATE comments SET status = $1, updated_by = $2, updated_at = $3
			  WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, comment.Status, comment.UpdatedBy, comment.UpdatedAt, comment.ID.String())
	return err
}

func (r *CommentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *CommentRepositoryImpl) Approve(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET status = 'approved', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *CommentRepositoryImpl) Reject(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET status = 'rejected', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *CommentRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *CommentRepositoryImpl) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, postID.String()).Scan(&count)
	return count, err
}

// scanCommentRow is a helper function to scan a comment row from database
func (r *CommentRepositoryImpl) scanCommentRow(rows pgx.Rows) (*entities.Comment, error) {
	var comment entities.Comment
	var parentIDStr *string

	err := rows.Scan(
		&comment.ID, &comment.ModelType, &comment.ModelID, &parentIDStr, &comment.Status,
		&comment.CreatedBy, &comment.UpdatedBy, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt)
	if err != nil {
		return nil, err
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}

	return &comment, nil
}
