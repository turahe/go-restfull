package responses

import (
	"time"

	"webapi/internal/domain/entities"
)

// PostResponse represents a post in API responses
type PostResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Slug        string     `json:"slug"`
	Status      string     `json:"status"`
	AuthorID    string     `json:"author_id"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ContentResponse represents content in API responses
type ContentResponse struct {
	ID          string    `json:"id"`
	PostID      string    `json:"post_id"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PostListResponse represents a list of posts with pagination
type PostListResponse struct {
	Posts []PostResponse `json:"posts"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Page  int            `json:"page"`
}

// NewPostResponse creates a new PostResponse from post entity
func NewPostResponse(post *entities.Post) *PostResponse {
	return &PostResponse{
		ID:          post.ID.String(),
		Title:       post.Title,
		Content:     post.Content,
		Slug:        post.Slug,
		Status:      post.Status,
		AuthorID:    post.AuthorID.String(),
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		DeletedAt:   post.DeletedAt,
	}
}

// NewPostListResponse creates a new PostListResponse from post entities
func NewPostListResponse(posts []*entities.Post, total int64, limit, page int) *PostListResponse {
	postResponses := make([]PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = *NewPostResponse(post)
	}

	return &PostListResponse{
		Posts: postResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}
