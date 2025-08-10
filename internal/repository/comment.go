package repository

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/nestedset"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// commentRepository implements the CommentRepository interface using nested set model
type commentRepository struct {
	db        *pgxpool.Pool
	nestedSet *nestedset.NestedSetManager
}

// NewCommentRepository creates a new comment repository instance
func NewCommentRepository(db *pgxpool.Pool) repositories.CommentRepository {
	return &commentRepository{
		db:        db,
		nestedSet: nestedset.NewNestedSetManager(db),
	}
}

func (r *commentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	// Calculate nested set values using the shared manager
	values, err := r.nestedSet.CreateNode(ctx, "comments", comment.ParentID, 1)
	if err != nil {
		return fmt.Errorf("failed to calculate nested set values: %w", err)
	}

	// Assign computed nested set values to the entity
	comment.RecordLeft = &values.Left
	comment.RecordRight = &values.Right
	comment.RecordDepth = &values.Depth
	comment.RecordOrdering = &values.Ordering

	// Insert the new comment
	query := `
		INSERT INTO comments (
			id, model_type, model_id, status, parent_id, 
			record_left, record_right, record_depth, record_ordering,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	parentIDStr := ""
	if comment.ParentID != nil {
		parentIDStr = comment.ParentID.String()
	}

	_, err = r.db.Exec(ctx, query,
		comment.ID, comment.ModelType, comment.ModelID, comment.Status,
		parentIDStr, comment.RecordLeft, comment.RecordRight,
		comment.RecordDepth, comment.RecordOrdering,
		comment.CreatedBy, comment.UpdatedBy, comment.CreatedAt, comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var comment entities.Comment
	var parentIDStr *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.ModelType, &comment.ModelID, &comment.Status,
		&parentIDStr, &comment.RecordLeft, &comment.RecordRight, &comment.RecordDepth,
		&comment.RecordOrdering, &comment.CreatedBy, &comment.UpdatedBy, &comment.DeletedBy,
		&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}

	return &comment, nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	return r.GetByModel(ctx, "post", postID, limit, offset)
}

func (r *commentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by user: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY record_ordering ASC, record_left ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, parentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) GetApproved(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE status = 'approved' AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved comments: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE status = 'pending' AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending comments: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	// For nested set, we need to handle parent changes carefully
	// This is a simplified update that doesn't change the tree structure
	query := `
		UPDATE comments
		SET status = $2, updated_by = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		comment.ID, comment.Status, comment.UpdatedBy, comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft delete - mark as deleted
	query := `
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

func (r *commentRepository) Approve(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'approved', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to approve comment: %w", err)
	}

	return nil
}

func (r *commentRepository) Reject(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments
		SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reject comment: %w", err)
	}

	return nil
}

func (r *commentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Comment, error) {
	searchQuery := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE deleted_at IS NULL
		  AND (model_type ILIKE $1 OR status ILIKE $1)
		ORDER BY record_left ASC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search comments: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *commentRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	return count, nil
}

func (r *commentRepository) CountByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE model_type = 'post' AND model_id = $1 AND deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by post: %w", err)
	}

	return count, nil
}

func (r *commentRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE created_by = $1 AND deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments by user: %w", err)
	}

	return count, nil
}

func (r *commentRepository) CountPending(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE status = 'pending' AND deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending comments: %w", err)
	}

	return count, nil
}

// Helper method for nested set operations
func (r *commentRepository) GetByModel(ctx context.Context, modelType string, modelID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE model_type = $1 AND model_id = $2 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, modelType, modelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by model: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// scanCommentRow is a helper function to scan a comment row from database
func (r *commentRepository) scanCommentRow(rows pgx.Rows) (*entities.Comment, error) {
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

	// Convert parent ID string to UUID
	if parentIDStr != nil {
		if parentID, err := uuid.Parse(*parentIDStr); err == nil {
			comment.ParentID = &parentID
		}
	}

	return &comment, nil
}

// GetThread retrieves the complete thread (comment + all replies) for a given comment
func (r *commentRepository) GetThread(ctx context.Context, commentID uuid.UUID) ([]*entities.Comment, error) {
	// Get the comment and all its descendants
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE record_left >= (
			SELECT record_left FROM comments WHERE id = $1 AND deleted_at IS NULL
		) 
		AND record_right <= (
			SELECT record_right FROM comments WHERE id = $1 AND deleted_at IS NULL
		) 
		AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment thread: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetCommentTree retrieves comments in a hierarchical tree structure
func (r *commentRepository) GetCommentTree(ctx context.Context, modelType string, modelID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	query := `
		WITH RECURSIVE comment_tree AS (
			-- Get root comments (no parent or parent is null)
			SELECT id, model_type, model_id, status, parent_id, 
			       record_left, record_right, record_depth, record_ordering,
			       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at,
			       0 as level
			FROM comments
			WHERE model_type = $1 AND model_id = $2 
			  AND (parent_id IS NULL OR parent_id = '')
			  AND deleted_at IS NULL
			ORDER BY record_left ASC
			LIMIT $3 OFFSET $4
			
			UNION ALL
			
			-- Get child comments
			SELECT c.id, c.model_type, c.model_id, c.status, c.parent_id,
			       c.record_left, c.record_right, c.record_depth, c.record_ordering,
			       c.created_by, c.updated_by, c.deleted_by, c.created_at, c.updated_at, c.deleted_at,
			       ct.level + 1
			FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id
			WHERE c.deleted_at IS NULL
		)
		SELECT * FROM comment_tree
		ORDER BY record_left ASC
	`

	rows, err := r.db.Query(ctx, query, modelType, modelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment tree: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// MoveComment moves a comment to a new parent (changes the thread structure)
func (r *commentRepository) MoveComment(ctx context.Context, commentID, newParentID uuid.UUID) error {
	// Use the nested set manager to move the subtree
	return r.nestedSet.MoveSubtree(ctx, "comments", commentID, newParentID)
}

// GetCommentDepth retrieves the depth of a comment in the thread
func (r *commentRepository) GetCommentDepth(ctx context.Context, commentID uuid.UUID) (uint64, error) {
	query := `
		SELECT record_depth 
		FROM comments 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var depth uint64
	err := r.db.QueryRow(ctx, query, commentID).Scan(&depth)
	if err != nil {
		return 0, fmt.Errorf("failed to get comment depth: %w", err)
	}

	return depth, nil
}

// GetCommentPath retrieves the path from root to a specific comment
func (r *commentRepository) GetCommentPath(ctx context.Context, commentID uuid.UUID) ([]*entities.Comment, error) {
	// Use the nested set manager to get the path
	pathIDs, err := r.nestedSet.GetPath(ctx, "comments", commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment path: %w", err)
	}

	if len(pathIDs) == 0 {
		return []*entities.Comment{}, nil
	}

	// Build the query for multiple IDs
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(pathIDs))
	for i, id := range pathIDs {
		idStrings[i] = id.String()
	}

	rows, err := r.db.Query(ctx, query, idStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment path: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// BulkApprove approves multiple comments at once
func (r *commentRepository) BulkApprove(ctx context.Context, commentIDs []uuid.UUID) error {
	if len(commentIDs) == 0 {
		return nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(commentIDs))
	for i, id := range commentIDs {
		idStrings[i] = id.String()
	}

	query := `
		UPDATE comments
		SET status = 'approved', updated_at = NOW()
		WHERE id = ANY($1) AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, idStrings)
	if err != nil {
		return fmt.Errorf("failed to bulk approve comments: %w", err)
	}

	return nil
}

// BulkReject rejects multiple comments at once
func (r *commentRepository) BulkReject(ctx context.Context, commentIDs []uuid.UUID) error {
	if len(commentIDs) == 0 {
		return nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(commentIDs))
	for i, id := range commentIDs {
		idStrings[i] = id.String()
	}

	query := `
		UPDATE comments
		SET status = 'rejected', updated_at = NOW()
		WHERE id = ANY($1) AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, idStrings)
	if err != nil {
		return fmt.Errorf("failed to bulk reject comments: %w", err)
	}

	return nil
}

// BulkDelete soft deletes multiple comments at once
func (r *commentRepository) BulkDelete(ctx context.Context, commentIDs []uuid.UUID) error {
	if len(commentIDs) == 0 {
		return nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(commentIDs))
	for i, id := range commentIDs {
		idStrings[i] = id.String()
	}

	query := `
		UPDATE comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = ANY($1) AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, idStrings)
	if err != nil {
		return fmt.Errorf("failed to bulk delete comments: %w", err)
	}

	return nil
}

// GetCommentStats retrieves statistics about comments
func (r *commentRepository) GetCommentStats(ctx context.Context, modelType string, modelID uuid.UUID) (map[string]int64, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected,
			COUNT(CASE WHEN parent_id IS NOT NULL AND parent_id != '' THEN 1 END) as replies
		FROM comments
		WHERE model_type = $1 AND model_id = $2 AND deleted_at IS NULL
	`

	var total, approved, pending, rejected, replies int64
	err := r.db.QueryRow(ctx, query, modelType, modelID).Scan(&total, &approved, &pending, &rejected, &replies)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment stats: %w", err)
	}

	return map[string]int64{
		"total":    total,
		"approved": approved,
		"pending":  pending,
		"rejected": rejected,
		"replies":  replies,
	}, nil
}

// GetCommentAncestors retrieves all ancestors of a comment
func (r *commentRepository) GetCommentAncestors(ctx context.Context, commentID uuid.UUID) ([]*entities.Comment, error) {
	// Use the nested set manager to get ancestor IDs
	ancestorIDs, err := r.nestedSet.GetAncestors(ctx, "comments", commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment ancestors: %w", err)
	}

	if len(ancestorIDs) == 0 {
		return []*entities.Comment{}, nil
	}

	// Build the query for multiple IDs
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(ancestorIDs))
	for i, id := range ancestorIDs {
		idStrings[i] = id.String()
	}

	rows, err := r.db.Query(ctx, query, idStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment ancestors: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetCommentSiblings retrieves all siblings of a comment
func (r *commentRepository) GetCommentSiblings(ctx context.Context, commentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// Use the nested set manager to get sibling IDs
	siblingIDs, err := r.nestedSet.GetSiblings(ctx, "comments", commentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment siblings: %w", err)
	}

	if len(siblingIDs) == 0 {
		return []*entities.Comment{}, nil
	}

	// Build the query for multiple IDs
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(siblingIDs))
	for i, id := range siblingIDs {
		idStrings[i] = id.String()
	}

	rows, err := r.db.Query(ctx, query, idStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment siblings: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetCommentDescendants retrieves all descendants of a comment
func (r *commentRepository) GetCommentDescendants(ctx context.Context, commentID uuid.UUID, limit, offset int) ([]*entities.Comment, error) {
	// Use the nested set manager to get descendant IDs
	descendantIDs, err := r.nestedSet.GetDescendants(ctx, "comments", commentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment descendants: %w", err)
	}

	if len(descendantIDs) == 0 {
		return []*entities.Comment{}, nil
	}

	// Build the query for multiple IDs
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY record_left ASC
	`

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(descendantIDs))
	for i, id := range descendantIDs {
		idStrings[i] = id.String()
	}

	rows, err := r.db.Query(ctx, query, idStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment descendants: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// IsCommentDescendant checks if one comment is a descendant of another
func (r *commentRepository) IsCommentDescendant(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	return r.nestedSet.IsDescendant(ctx, "comments", ancestorID, descendantID)
}

// IsCommentAncestor checks if one comment is an ancestor of another
func (r *commentRepository) IsCommentAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	return r.nestedSet.IsAncestor(ctx, "comments", ancestorID, descendantID)
}

// CountCommentDescendants counts the number of descendants of a comment
func (r *commentRepository) CountCommentDescendants(ctx context.Context, commentID uuid.UUID) (int64, error) {
	return r.nestedSet.CountDescendants(ctx, "comments", commentID)
}

// CountCommentChildren counts the number of direct children of a comment
func (r *commentRepository) CountCommentChildren(ctx context.Context, commentID uuid.UUID) (int64, error) {
	return r.nestedSet.CountChildren(ctx, "comments", commentID)
}

// GetCommentSubtreeSize calculates the size of a comment subtree
func (r *commentRepository) GetCommentSubtreeSize(ctx context.Context, commentID uuid.UUID) (int64, error) {
	return r.nestedSet.GetSubtreeSize(ctx, "comments", commentID)
}

// ValidateCommentTree validates the comment tree structure for integrity
func (r *commentRepository) ValidateCommentTree(ctx context.Context) ([]string, error) {
	return r.nestedSet.ValidateTree(ctx, "comments")
}

// RebuildCommentTree rebuilds the entire comment tree structure
func (r *commentRepository) RebuildCommentTree(ctx context.Context) error {
	return r.nestedSet.RebuildTree(ctx, "comments")
}

// GetCommentTreeStatistics returns comprehensive statistics about the comment tree
func (r *commentRepository) GetCommentTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	return r.nestedSet.GetTreeStatistics(ctx, "comments")
}

// GetCommentsByDepth retrieves comments at a specific depth level
func (r *commentRepository) GetCommentsByDepth(ctx context.Context, modelType string, modelID uuid.UUID, depth uint64, limit, offset int) ([]*entities.Comment, error) {
	query := `
		SELECT id, model_type, model_id, status, parent_id, 
		       record_left, record_right, record_depth, record_ordering,
		       created_by, updated_by, deleted_by, created_at, updated_at, deleted_at
		FROM comments
		WHERE model_type = $1 AND model_id = $2 
		  AND record_depth = $3 AND deleted_at IS NULL
		ORDER BY record_left ASC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, query, modelType, modelID, depth, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by depth: %w", err)
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		comment, err := r.scanCommentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetCommentTreeHeight calculates the maximum depth of the comment tree
func (r *commentRepository) GetCommentTreeHeight(ctx context.Context, modelType string, modelID uuid.UUID) (uint64, error) {
	query := `
		SELECT COALESCE(MAX(record_depth), 0)
		FROM comments
		WHERE model_type = $1 AND model_id = $2 AND deleted_at IS NULL
	`

	var height uint64
	err := r.db.QueryRow(ctx, query, modelType, modelID).Scan(&height)
	if err != nil {
		return 0, fmt.Errorf("failed to get comment tree height: %w", err)
	}

	return height, nil
}
