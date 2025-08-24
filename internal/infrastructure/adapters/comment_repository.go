package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresCommentRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresCommentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.CommentRepository {
	return &PostgresCommentRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
	}
}

func (r *PostgresCommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	query := `
		INSERT INTO comments (
			id, model_type, model_id, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`
	parentIDStr := ""
	if comment.ParentID != nil {
		parentIDStr = comment.ParentID.String()
	}
	_, err := r.db.Exec(ctx, query,
		comment.ID, comment.ModelType, comment.ModelID, comment.Status,
		parentIDStr, comment.RecordLeft, comment.RecordRight,
		comment.RecordDepth, comment.RecordOrdering,
		comment.CreatedBy, comment.UpdatedBy, comment.CreatedAt, comment.UpdatedAt,
	)
	return err
}

func (r *PostgresCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL`
	var comment entities.Comment
	var parentIDStr *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
		&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
		&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
		&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}
	return &comment, nil
}

func (r *PostgresCommentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE model_type = 'post' AND model_id = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_ordering ASC, record_left ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, parentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &p
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetApproved(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE status = 'approved' AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE status = 'pending' AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresCommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	query := `
		UPDATE comments
		SET status = $2, updated_by = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query,
		comment.ID, comment.Status, comment.UpdatedBy, comment.UpdatedAt,
	)
	return err
}

func (r *PostgresCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresCommentRepository) Approve(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'approved', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresCommentRepository) Reject(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresCommentRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *PostgresCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE model_type = 'post' AND model_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	return count, err
}

func (r *PostgresCommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE created_by = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresCommentRepository) CountPending(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE status = 'pending' AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *PostgresCommentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error) {
	q := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE deleted_at IS NULL
		  AND (model_type ILIKE $1 OR status ILIKE $1)
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		var parentIDStr *string
		err := rows.Scan(
			&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
			&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
			&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}
