package seeders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CommentSeeder seeds initial comments
type CommentSeeder struct{}

// NewCommentSeeder creates a new comment seeder
func NewCommentSeeder() *CommentSeeder {
	return &CommentSeeder{}
}

// GetName returns the seeder name
func (cs *CommentSeeder) GetName() string {
	return "CommentSeeder"
}

// Run executes the comment seeder
func (cs *CommentSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
	// Get user IDs
	var userID, authorID uuid.UUID

	err := db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'user'`).Scan(&userID)
	if err != nil {
		return err
	}

	err = db.QueryRow(ctx, `SELECT id FROM users WHERE username = 'author'`).Scan(&authorID)
	if err != nil {
		return err
	}

	// Get post ID
	var postID uuid.UUID
	err = db.QueryRow(ctx, `SELECT id FROM posts WHERE slug = 'getting-started-with-go-programming'`).Scan(&postID)
	if err != nil {
		return err
	}

	comments := []struct {
		ID        uuid.UUID
		Content   string
		PostID    uuid.UUID
		UserID    uuid.UUID
		ParentID  *uuid.UUID
		Status    string
		CreatedAt time.Time
		UpdatedAt time.Time
	}{
		{
			ID:        uuid.New(),
			Content:   "Great article! I've been learning Go and this really helped me understand the basics. Looking forward to more content like this.",
			PostID:    postID,
			UserID:    userID,
			ParentID:  nil,
			Status:    "approved",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Content:   "Thanks for the feedback! I'm glad you found it helpful. Go is indeed a fantastic language to learn.",
			PostID:    postID,
			UserID:    authorID,
			ParentID:  nil,
			Status:    "approved",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Content:   "Do you have any recommendations for Go learning resources? I'm particularly interested in web development.",
			PostID:    postID,
			UserID:    userID,
			ParentID:  nil,
			Status:    "approved",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, comment := range comments {
		_, err := db.Exec(ctx, `
			INSERT INTO comments (id, content, post_id, user_id, parent_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT DO NOTHING
		`, comment.ID, comment.Content, comment.PostID, comment.UserID, comment.ParentID, comment.Status, comment.CreatedAt, comment.UpdatedAt)

		if err != nil {
			return err
		}
	}

	return nil
}
