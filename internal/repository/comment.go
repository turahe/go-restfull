package repository

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

type CommentRepositoryInterface interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	CreateCommentWithContent(ctx context.Context, comment *model.Comment, content *model.Content) error
	CreateCommentWithContents(ctx context.Context, comment *model.Comment, contents []model.Content) error
	GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	GetAllComments(ctx context.Context) ([]*model.Comment, error)
	UpdateComment(ctx context.Context, comment *model.Comment) error
	DeleteComment(ctx context.Context, id uuid.UUID) error
	GetCommentWithContents(ctx context.Context, id uuid.UUID) (*model.Comment, error)
}

// Minimal DB interface for mocking and production
// (In production, use pgxpool.Pool which implements this)
type DBIface interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type CommentRepositoryImpl struct {
	db          DBIface
	redisClient redis.Cmdable
}

func NewCommentRepository(db DBIface, redisClient redis.Cmdable) CommentRepositoryInterface {
	return &CommentRepositoryImpl{
		db:          db,
		redisClient: redisClient,
	}
}

func (r *CommentRepositoryImpl) CreateComment(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO comments (id, model_type, model_id, title, status, parent_id, record_left, record_right, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, comment.ID, comment.ModelType, comment.ModelID, comment.Title, comment.Status, comment.ParentID, comment.RecordLeft, comment.RecordRight, comment.RecordOrdering, comment.CreatedBy, comment.UpdatedBy, comment.DeletedBy, comment.DeletedAt, comment.CreatedAt, comment.UpdatedAt)
	return err
}

func (r *CommentRepositoryImpl) CreateCommentWithContent(ctx context.Context, comment *model.Comment, content *model.Content) error {
	tx, err := r.db.(interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	}).Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO comments (id, model_type, model_id, title, status, parent_id, record_left, record_right, record_ordering, created_by, updated_by, deleted_by, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, comment.ID, comment.ModelType, comment.ModelID, comment.Title, comment.Status, comment.ParentID, comment.RecordLeft, comment.RecordRight, comment.RecordOrdering, comment.CreatedBy, comment.UpdatedBy, comment.DeletedBy, comment.DeletedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, content.ID, content.ModelType, content.ModelID, content.ContentRaw, content.ContentHTML, content.CreatedBy, content.UpdatedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *CommentRepositoryImpl) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	row := r.db.QueryRow(ctx, `SELECT id, model_type, model_id, title, status, parent_id, record_left, record_right, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at FROM comments WHERE id = $1`, id)
	c := &model.Comment{}
	err := row.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.Title, &c.Status, &c.ParentID, &c.RecordLeft, &c.RecordRight, &c.RecordOrdering, &c.CreatedBy, &c.UpdatedBy, &c.DeletedBy, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CommentRepositoryImpl) GetAllComments(ctx context.Context) ([]*model.Comment, error) {
	rows, err := r.db.Query(ctx, `SELECT id, model_type, model_id, title, status, parent_id, record_left, record_right, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at FROM comments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.Title, &c.Status, &c.ParentID, &c.RecordLeft, &c.RecordRight, &c.RecordOrdering, &c.CreatedBy, &c.UpdatedBy, &c.DeletedBy, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *CommentRepositoryImpl) UpdateComment(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.Exec(ctx, `
		UPDATE comments SET model_type=$2, model_id=$3, title=$4, status=$5, parent_id=$6, record_left=$7, record_right=$8, record_ordering=$9, updated_by=$10, updated_at=$11 WHERE id=$1
	`, comment.ID, comment.ModelType, comment.ModelID, comment.Title, comment.Status, comment.ParentID, comment.RecordLeft, comment.RecordRight, comment.RecordOrdering, comment.UpdatedBy, comment.UpdatedAt)
	return err
}

func (r *CommentRepositoryImpl) DeleteComment(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM comments WHERE id = $1`, id)
	return err
}

func (r *CommentRepositoryImpl) GetCommentWithContents(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	comment, err := r.GetCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by FROM contents WHERE model_type = $1 AND model_id = $2`, "comment", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contents []model.Content
	for rows.Next() {
		var c model.Content
		err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.ContentRaw, &c.ContentHTML, &c.CreatedBy, &c.UpdatedBy)
		if err != nil {
			return nil, err
		}
		contents = append(contents, c)
	}
	comment.Contents = contents
	return comment, nil
}

func (r *CommentRepositoryImpl) CreateCommentWithContents(ctx context.Context, comment *model.Comment, contents []model.Content) error {
	tx, err := r.db.(interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	}).Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO comments (id, model_type, model_id, title, status, parent_id, record_left, record_right, record_ordering, created_by, updated_by, deleted_by, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, comment.ID, comment.ModelType, comment.ModelID, comment.Title, comment.Status, comment.ParentID, comment.RecordLeft, comment.RecordRight, comment.RecordOrdering, comment.CreatedBy, comment.UpdatedBy, comment.DeletedBy, comment.DeletedAt)
	if err != nil {
		return err
	}

	for i := range contents {
		contents[i].ModelType = "comment"
		contents[i].ModelID = comment.ID.String()
		_, err = tx.Exec(ctx, `
			INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, contents[i].ID, contents[i].ModelType, contents[i].ModelID, contents[i].ContentRaw, contents[i].ContentHTML, contents[i].CreatedBy, contents[i].UpdatedBy)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Insert a nested comment as a child of parentID
func (r *CommentRepositoryImpl) InsertNestedComment(ctx context.Context, parentID uuid.UUID, comment *model.Comment) error {
	tx, err := r.db.(interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	}).Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Get parent's right and depth
	var parentRight, parentDepth int64
	err = tx.QueryRow(ctx, `SELECT record_right, record_depth FROM comments WHERE id = $1`, parentID).Scan(&parentRight, &parentDepth)
	if err != nil {
		return err
	}

	// Shift right values
	_, err = tx.Exec(ctx, `UPDATE comments SET record_right = record_right + 2 WHERE record_right >= $1`, parentRight)
	if err != nil {
		return err
	}
	// Shift left values
	_, err = tx.Exec(ctx, `UPDATE comments SET record_left = record_left + 2 WHERE record_left > $1`, parentRight)
	if err != nil {
		return err
	}

	comment.RecordLeft = parentRight
	comment.RecordRight = parentRight + 1
	comment.RecordDepth = parentDepth + 1
	comment.ParentID = &parentID

	_, err = tx.Exec(ctx, `
		INSERT INTO comments (id, model_type, model_id, title, status, parent_id, record_left, record_right, record_depth, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`, comment.ID, comment.ModelType, comment.ModelID, comment.Title, comment.Status, comment.ParentID, comment.RecordLeft, comment.RecordRight, comment.RecordDepth, comment.RecordOrdering, comment.CreatedBy, comment.UpdatedBy, comment.DeletedBy, comment.DeletedAt, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Get all descendants (subtree) of a comment
func (r *CommentRepositoryImpl) GetDescendants(ctx context.Context, id uuid.UUID) ([]*model.Comment, error) {
	query := `
		SELECT id, model_type, model_id, title, status, parent_id, record_left, record_right, record_depth, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at
		FROM comments
		WHERE record_left > (SELECT record_left FROM comments WHERE id = $1)
		  AND record_right < (SELECT record_right FROM comments WHERE id = $1)
		ORDER BY record_left
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.Title, &c.Status, &c.ParentID, &c.RecordLeft, &c.RecordRight, &c.RecordDepth, &c.RecordOrdering, &c.CreatedBy, &c.UpdatedBy, &c.DeletedBy, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// Recursively build the comment tree
func (r *CommentRepositoryImpl) GetCommentTree(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	root, err := r.GetCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Remove unused variable 'children'
	_, err = r.getChildrenRecursive(ctx, root.ID)
	if err != nil {
		return nil, err
	}
	root.Contents = nil // or use a Children field if you add one
	// If you want to use Contents for children, uncomment:
	// root.Contents = children
	return root, nil
}

func (r *CommentRepositoryImpl) getChildrenRecursive(ctx context.Context, parentID uuid.UUID) ([]model.Comment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, model_type, model_id, title, status, parent_id, record_left, record_right, record_depth, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at
		FROM comments WHERE parent_id = $1 ORDER BY record_left
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []model.Comment
	for rows.Next() {
		c := model.Comment{}
		err := rows.Scan(&c.ID, &c.ModelType, &c.ModelID, &c.Title, &c.Status, &c.ParentID, &c.RecordLeft, &c.RecordRight, &c.RecordDepth, &c.RecordOrdering, &c.CreatedBy, &c.UpdatedBy, &c.DeletedBy, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Recursively get children, but don't assign unused variable
		_, err = r.getChildrenRecursive(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Contents = nil // or set to grandChildren if you want
		// c.Contents = grandChildren
		children = append(children, c)
	}
	return children, nil
}

var _ CommentRepositoryInterface = (*CommentRepositoryImpl)(nil)
