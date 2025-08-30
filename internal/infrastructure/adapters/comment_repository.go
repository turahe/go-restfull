// Package adapters provides infrastructure layer implementations that adapt external systems
// and frameworks to the domain layer interfaces. This package contains repository implementations,
// external service adapters, and infrastructure-specific services.
package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresCommentRepository implements the CommentRepository interface using PostgreSQL as the primary
// data store and Redis for caching. This repository provides CRUD operations for comment entities
// with support for hierarchical comment structures (nested comments), comment moderation, soft deletes,
// and polymorphic relationships. It embeds the BaseTransactionalRepository to inherit transaction
// management capabilities.
type PostgresCommentRepository struct {
	// BaseTransactionalRepository provides transaction management functionality
	*BaseTransactionalRepository
	// db holds the PostgreSQL connection pool for database operations
	db *pgxpool.Pool
	// redisClient holds the Redis client for caching operations
	redisClient redis.Cmdable
	// nestedSetManager handles nested set operations for hierarchical comment structures
	nestedSetManager *nestedset.NestedSetManager
}

// NewPostgresCommentRepository creates a new PostgreSQL comment repository instance.
// This factory function initializes the repository with database and Redis connections,
// and sets up the base transactional repository for transaction management.
//
// Parameters:
//   - db: The PostgreSQL connection pool for database operations
//   - redisClient: The Redis client for caching operations
//
// Returns:
//   - repositories.CommentRepository: A new comment repository instance
func NewPostgresCommentRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.CommentRepository {
	return &PostgresCommentRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
		redisClient:                 redisClient,
		nestedSetManager:            nestedset.NewNestedSetManager(db),
	}
}

// Create persists a new comment entity to the database.
// This method inserts a new comment record with all required fields including
// hierarchical structure information (parent_id, nested set coordinates), audit
// information (created_by, updated_by, timestamps), and polymorphic relationships.
// The method handles optional parent_id conversion from UUID to string for storage.
//
// Parameters:
//   - ctx: Context for the database operation
//   - comment: The comment entity to create
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	// Calculate nested set values using the nested set manager
	nestedSetValues, err := r.nestedSetManager.CreateNode(ctx, "comments", comment.ParentID, 0)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Convert markdown to HTML
	contentHTML := string(markdown.ToHTML([]byte(comment.Content), nil, nil))

	// Use a transaction to ensure both comment and content are inserted atomically
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			// Rollback on error
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Log rollback error but return the original error
				fmt.Printf("failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	// Insert comment with nested set values
	commentQuery := `
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

	_, err = tx.Exec(ctx, commentQuery,
		comment.ID, comment.ModelType, comment.ModelID, comment.Status,
		parentIDStr, nestedSetValues.Left, nestedSetValues.Right,
		nestedSetValues.Depth, nestedSetValues.Ordering,
		comment.CreatedBy, comment.UpdatedBy, comment.CreatedAt, comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	// Insert content
	contentQuery := `
		INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	contentID := uuid.New()
	now := time.Now()
	_, err = tx.Exec(ctx, contentQuery,
		contentID,
		"comment",
		comment.ID,
		comment.Content,
		contentHTML,
		comment.CreatedBy,
		comment.UpdatedBy,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert content: %w", err)
	}

	// Update the comment entity with the calculated nested set values
	comment.RecordLeft = &nestedSetValues.Left
	comment.RecordRight = &nestedSetValues.Right
	comment.RecordDepth = &nestedSetValues.Depth
	comment.RecordOrdering = &nestedSetValues.Ordering

	// Commit the transaction
	return tx.Commit(ctx)
}

// GetByID retrieves a comment entity by its unique identifier.
// This method performs a soft-delete aware query, excluding records that have been
// marked as deleted. It returns the complete comment entity with all fields populated,
// including hierarchical structure information and properly parsed parent_id.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the comment to retrieve
//
// Returns:
//   - *entities.Comment: The found comment entity, or nil if not found
//   - error: Any error that occurred during the database operation
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
	// Parse parent_id string back to UUID if it exists
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}
	return &comment, nil
}

// GetByPostID retrieves all comments associated with a specific post.
// This method returns comments ordered by their hierarchical position (record_left)
// to maintain the proper comment tree structure. It supports pagination and
// excludes soft-deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - postID: The unique identifier of the post to get comments for
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of comment entities for the post
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// GetByUserID retrieves all comments created by a specific user.
// This method returns comments ordered by their hierarchical position (record_left)
// and supports pagination. It excludes soft-deleted records and is useful for
// displaying a user's comment history.
//
// Parameters:
//   - ctx: Context for the database operation
//   - userID: The unique identifier of the user to get comments for
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of comment entities created by the user
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// GetReplies retrieves all direct replies to a specific comment.
// This method returns child comments ordered by their ordering and hierarchical position
// to maintain proper reply sequence. It supports pagination and excludes soft-deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - parentID: The unique identifier of the parent comment
//   - limit: Maximum number of replies to return
//   - offset: Number of replies to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of reply comment entities
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if p, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &p
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// GetAll retrieves a paginated list of all active comments.
// This method returns comments ordered by their hierarchical position (record_left)
// to maintain the proper comment tree structure. It supports pagination and
// excludes soft-deleted records.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of comments to return
//   - offset: Number of comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of comment entities
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// GetApproved retrieves a paginated list of all approved comments.
// This method filters comments by their approved status and returns them ordered
// by hierarchical position. It's useful for displaying only moderated content
// and supports pagination.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of approved comments to return
//   - offset: Number of approved comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of approved comment entities
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// GetPending retrieves a paginated list of all pending comments awaiting moderation.
// This method filters comments by their pending status and returns them ordered
// by hierarchical position. It's useful for moderation workflows and supports pagination.
//
// Parameters:
//   - ctx: Context for the database operation
//   - limit: Maximum number of pending comments to return
//   - offset: Number of pending comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of pending comment entities
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

// Update modifies an existing comment entity in the database.
// This method updates the comment's status and audit fields while preserving
// the original ID, creation information, and hierarchical structure. It only
// updates non-deleted comments.
//
// Parameters:
//   - ctx: Context for the database operation
//   - comment: The comment entity with updated values
//
// Returns:
//   - error: Any error that occurred during the database operation
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

// Delete performs a soft delete of a comment entity.
// This method marks the comment as deleted by setting the deleted_at timestamp
// and updates the updated_at timestamp. This preserves data integrity
// and allows for potential recovery while maintaining audit trails.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the comment to delete
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Approve changes a comment's status to approved for public display.
// This method is part of the comment moderation workflow and updates
// the updated_at timestamp to track when the approval occurred.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the comment to approve
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) Approve(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'approved', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Reject changes a comment's status to rejected, preventing public display.
// This method is part of the comment moderation workflow and updates
// the updated_at timestamp to track when the rejection occurred.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: The unique identifier of the comment to reject
//
// Returns:
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) Reject(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Count returns the total number of active comments in the system.
// This method excludes soft-deleted records and is useful for pagination
// calculations and system statistics.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of active comments
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByPostID returns the total number of active comments for a specific post.
// This method is useful for understanding comment activity on posts and for
// pagination calculations within post contexts.
//
// Parameters:
//   - ctx: Context for the database operation
//   - postID: The unique identifier of the post to count comments for
//
// Returns:
//   - int64: The total count of active comments for the post
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE model_type = 'post' AND model_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	return count, err
}

// CountByUserID returns the total number of active comments created by a specific user.
// This method is useful for user activity tracking and moderation workflows.
//
// Parameters:
//   - ctx: Context for the database operation
//   - userID: The unique identifier of the user to count comments for
//
// Returns:
//   - int64: The total count of active comments created by the user
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE created_by = $1 AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountPending returns the total number of comments awaiting moderation.
// This method is useful for moderation dashboard statistics and workflow management.
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - int64: The total count of pending comments
//   - error: Any error that occurred during the database operation
func (r *PostgresCommentRepository) CountPending(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE status = 'pending' AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// Search performs a case-insensitive search for comments by model type or status.
// This method uses ILIKE for pattern matching and supports partial matches.
// Results are ordered by hierarchical position (record_left) to maintain
// comment tree structure and support pagination.
//
// Parameters:
//   - ctx: Context for the database operation
//   - query: The search term to match against model_type and status fields
//   - limit: Maximum number of matching comments to return
//   - offset: Number of matching comments to skip for pagination
//
// Returns:
//   - []*entities.Comment: List of matching comment entities
//   - error: Any error that occurred during the database operation
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
		// Parse parent_id string back to UUID if it exists
		if parentIDStr != nil {
			if parentID, err := uuid.Parse(*parentIDStr); err == nil {
				comment.ParentID = &parentID
			}
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}
