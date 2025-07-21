package repository

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/http/requests"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository interface {
	GetByIDWithContents(ctx context.Context, id uuid.UUID) (*model.Post, error)
	GetAllWithContents(ctx context.Context) ([]*model.Post, error)
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPostsWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) ([]*model.Post, int, error)
}

type PostRepositoryImpl struct {
	pgxPool *pgxpool.Pool
}

func NewPostRepository(pgxPool *pgxpool.Pool) PostRepository {
	return &PostRepositoryImpl{pgxPool: pgxPool}
}

func (r *PostRepositoryImpl) GetByIDWithContents(ctx context.Context, id uuid.UUID) (*model.Post, error) {
	post := &model.Post{}
	query := `SELECT id, slug, title, subtitle, description, type, is_sticky, published_at, language, layout, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at FROM posts WHERE id = $1`
	err := r.pgxPool.QueryRow(ctx, query, id).Scan(
		&post.ID, &post.Slug, &post.Title, &post.Subtitle, &post.Description, &post.Type, &post.IsSticky, &post.PublishedAt, &post.Language, &post.Layout, &post.RecordOrdering, &post.CreatedBy, &post.UpdatedBy, &post.DeletedBy, &post.DeletedAt, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	contents, err := r.getContentsForPost(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Contents = contents
	return post, nil
}

func (r *PostRepositoryImpl) GetAllWithContents(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT id, slug, title, subtitle, description, type, is_sticky, published_at, language, layout, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at FROM posts`
	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.ID, &post.Slug, &post.Title, &post.Subtitle, &post.Description, &post.Type, &post.IsSticky, &post.PublishedAt, &post.Language, &post.Layout, &post.RecordOrdering, &post.CreatedBy, &post.UpdatedBy, &post.DeletedBy, &post.DeletedAt, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contents, err := r.getContentsForPost(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Contents = contents
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *model.Post) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO posts (id, slug, title, subtitle, description, type, is_sticky, published_at, language, layout, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`
	_, err = tx.Exec(ctx, query,
		post.ID, post.Slug, post.Title, post.Subtitle, post.Description, post.Type, post.IsSticky, post.PublishedAt, post.Language, post.Layout, post.RecordOrdering, post.CreatedBy, post.UpdatedBy, post.DeletedBy, post.DeletedAt, post.CreatedAt, post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	for _, content := range post.Contents {
		contentQuery := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by) VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err := tx.Exec(ctx, contentQuery, content.ID, content.ModelType, post.ID, content.ContentRaw, content.ContentHTML, content.CreatedBy, content.UpdatedBy)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepositoryImpl) Update(ctx context.Context, post *model.Post) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE posts SET slug=$1, title=$2, subtitle=$3, description=$4, type=$5, is_sticky=$6, published_at=$7, language=$8, layout=$9, record_ordering=$10, updated_by=$11, updated_at=$12 WHERE id=$13`
	_, err = tx.Exec(ctx, query,
		post.Slug, post.Title, post.Subtitle, post.Description, post.Type, post.IsSticky, post.PublishedAt, post.Language, post.Layout, post.RecordOrdering, post.UpdatedBy, post.UpdatedAt, post.ID,
	)
	if err != nil {
		return err
	}
	// For simplicity, delete all old contents and re-insert
	_, err = tx.Exec(ctx, `DELETE FROM contents WHERE model_type='post' AND model_id=$1`, post.ID)
	if err != nil {
		return err
	}
	for _, content := range post.Contents {
		contentQuery := `INSERT INTO contents (id, model_type, model_id, content_raw, content_html, created_by, updated_by) VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err := tx.Exec(ctx, contentQuery, content.ID, content.ModelType, post.ID, content.ContentRaw, content.ContentHTML, content.CreatedBy, content.UpdatedBy)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pgxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := "now()"
	_, err = tx.Exec(ctx, `UPDATE posts SET deleted_at = `+now+` WHERE id = $1`, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE contents SET deleted_at = `+now+` WHERE model_type='post' AND model_id = $1`, id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostRepositoryImpl) GetPostsWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) ([]*model.Post, int, error) {
	var posts []*model.Post
	var total int
	query := `SELECT id, slug, title, subtitle, description, type, is_sticky, published_at, language, layout, record_ordering, created_by, updated_by, deleted_by, deleted_at, created_at, updated_at FROM posts WHERE title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%' ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pgxPool.Query(ctx, query, input.Query, input.Limit, input.Page*input.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.ID, &post.Slug, &post.Title, &post.Subtitle, &post.Description, &post.Type, &post.IsSticky, &post.PublishedAt, &post.Language, &post.Layout, &post.RecordOrdering, &post.CreatedBy, &post.UpdatedBy, &post.DeletedBy, &post.DeletedAt, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		contents, err := r.getContentsForPost(ctx, post.ID)
		if err != nil {
			return nil, 0, err
		}
		post.Contents = contents
		posts = append(posts, post)
	}
	totalQuery := `SELECT COUNT(*) FROM posts WHERE title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%';`
	err = r.pgxPool.QueryRow(ctx, totalQuery, input.Query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *PostRepositoryImpl) getContentsForPost(ctx context.Context, postID uuid.UUID) ([]model.Content, error) {
	query := `SELECT id, model_type, model_id, content_raw, content_html, created_by, updated_by FROM contents WHERE model_type = 'post' AND model_id = $1`
	rows, err := r.pgxPool.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contents []model.Content
	for rows.Next() {
		var content model.Content
		err := rows.Scan(&content.ID, &content.ModelType, &content.ModelID, &content.ContentRaw, &content.ContentHTML, &content.CreatedBy, &content.UpdatedBy)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	return contents, nil
}
