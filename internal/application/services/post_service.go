// Package services provides application-level business logic for post management.
// This package contains the post service implementation that handles content creation,
// publication workflow, content lifecycle, and post management while ensuring proper
// data integrity and business rules.
package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// postService implements the PostService interface and provides comprehensive
// post management functionality. It handles content creation, publication workflow,
// content lifecycle, and post management while ensuring proper data integrity
// and business rules.
type postService struct {
	postRepo repositories.PostRepository
}

// NewPostService creates a new post service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - postRepo: Repository interface for post data access operations
//
// Returns:
//   - ports.PostService: The post service interface implementation
func NewPostService(postRepo repositories.PostRepository) ports.PostService {
	return &postService{
		postRepo: postRepo,
	}
}

// CreatePost creates a new post with comprehensive validation and content management.
// This method enforces business rules for post creation and supports various
// content states and author attribution.
//
// Business Rules:
//   - Post title is required and validated
//   - Content must be provided and validated
//   - Slug must be unique if provided
//   - Author must exist and be valid
//   - Post validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - title: Title of the post
//   - content: Content body of the post
//   - slug: Optional unique slug for URL-friendly routing
//   - status: Initial status of the post (draft, published, etc.)
//   - authorID: UUID of the post author
//
// Returns:
//   - *entities.Post: The created post entity
//   - error: Any error that occurred during the operation
func (s *postService) CreatePost(ctx context.Context, title, content, slug, status string, authorID uuid.UUID) (*entities.Post, error) {
	// Create post entity with the provided parameters
	post, err := entities.NewPost(title, content, slug, status, authorID)
	if err != nil {
		return nil, err
	}

	// Persist the post to the repository
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// GetPostByID retrieves a post by its unique identifier.
// This method includes soft delete checking to ensure deleted posts
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the post to retrieve
//
// Returns:
//   - *entities.Post: The post entity if found
//   - error: Error if post not found or other issues occur
func (s *postService) GetPostByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Check if the post has been soft deleted
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}
	return post, nil
}

// GetPostBySlug retrieves a post by its unique slug identifier.
// This method is useful for URL-friendly routing and public post access.
//
// Parameters:
//   - ctx: Context for the operation
//   - slug: Slug identifier of the post to retrieve
//
// Returns:
//   - *entities.Post: The post entity if found
//   - error: Error if post not found or other issues occur
func (s *postService) GetPostBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	// Check if the post has been soft deleted
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}
	return post, nil
}

// GetPostsByAuthor retrieves all posts by a specific author with pagination.
// This method is useful for author profile pages and content management.
//
// Parameters:
//   - ctx: Context for the operation
//   - authorID: UUID of the author to get posts for
//   - limit: Maximum number of posts to return
//   - offset: Number of posts to skip for pagination
//
// Returns:
//   - []*entities.Post: List of posts by the author
//   - error: Any error that occurred during the operation
func (s *postService) GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetByAuthor(ctx, authorID, limit, offset)
}

// GetAllPosts retrieves all posts in the system with pagination.
// This method is useful for administrative purposes and content management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of posts to return
//   - offset: Number of posts to skip for pagination
//
// Returns:
//   - []*entities.Post: List of all posts
//   - error: Any error that occurred during the operation
func (s *postService) GetAllPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetAll(ctx, limit, offset)
}

// GetPublishedPosts retrieves only published posts with pagination.
// This method is useful for public-facing content display and blog listings.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of published posts to return
//   - offset: Number of published posts to skip for pagination
//
// Returns:
//   - []*entities.Post: List of published posts
//   - error: Any error that occurred during the operation
func (s *postService) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetPublished(ctx, limit, offset)
}

// SearchPosts searches for posts based on a query string.
// This method supports full-text search capabilities for finding posts
// by title, content, or other attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Post: List of matching posts
//   - error: Any error that occurred during the operation
func (s *postService) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.Search(ctx, query, limit, offset)
}

// GetPostsWithPagination retrieves posts with pagination and returns total count.
// This method provides a comprehensive pagination solution with search and status filtering.
//
// Business Rules:
//   - Page and perPage parameters are properly handled
//   - Search and status filtering are supported
//   - Total count is calculated for pagination metadata
//   - Offset is calculated based on page and perPage
//   - Published search combines both search and published status
//
// Parameters:
//   - ctx: Context for the operation
//   - page: Current page number (1-based)
//   - perPage: Number of posts per page
//   - search: Optional search query for filtering
//   - status: Optional status filter for posts
//
// Returns:
//   - []*entities.Post: List of posts for the current page
//   - int64: Total count of posts for pagination
//   - error: Any error that occurred during the operation
func (s *postService) GetPostsWithPagination(ctx context.Context, page, perPage int, search, status string) ([]*entities.Post, int64, error) {
	// Calculate offset based on page and perPage for pagination
	offset := (page - 1) * perPage

	var posts []*entities.Post
	var err error

	// Get posts based on search and status parameters
	if search != "" && status == "published" {
		// Search within published posts only
		posts, err = s.postRepo.SearchPublished(ctx, search, perPage, offset)
	} else if search != "" {
		// Search all posts
		posts, err = s.postRepo.Search(ctx, search, perPage, offset)
	} else if status == "published" {
		// Get only published posts
		posts, err = s.postRepo.GetPublished(ctx, perPage, offset)
	} else {
		// Get all posts
		posts, err = s.postRepo.GetAll(ctx, perPage, offset)
	}

	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination metadata
	total, err := s.GetPostsCount(ctx, search, status)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsCount returns total count of posts for pagination calculations.
// This method supports filtering by search query and status.
//
// Parameters:
//   - ctx: Context for the operation
//   - search: Optional search query for filtered count
//   - status: Optional status filter for count
//
// Returns:
//   - int64: Total count of posts
//   - error: Any error that occurred during the operation
func (s *postService) GetPostsCount(ctx context.Context, search, status string) (int64, error) {
	if search != "" && status == "published" {
		return s.postRepo.CountBySearchPublished(ctx, search)
	} else if search != "" {
		return s.postRepo.CountBySearch(ctx, search)
	} else if status == "published" {
		return s.postRepo.CountPublished(ctx)
	}
	return s.postRepo.Count(ctx)
}

// UpdatePost updates an existing post's content and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Post must exist and not be deleted
//   - Post validation ensures proper structure
//   - Content and metadata are updated atomically
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the post to update
//   - title: Updated title of the post
//   - content: Updated content body of the post
//   - slug: Updated slug for URL-friendly routing
//   - status: Updated status of the post
//
// Returns:
//   - *entities.Post: The updated post entity
//   - error: Any error that occurred during the operation
func (s *postService) UpdatePost(ctx context.Context, id uuid.UUID, title, content, slug, status string) (*entities.Post, error) {
	// Retrieve existing post to ensure it exists and is not deleted
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}

	// Update the post entity with new information
	if err := post.UpdatePost(title, content, slug, status); err != nil {
		return nil, err
	}

	// Persist the updated post to the repository
	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost performs a soft delete of a post by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - Post must exist before deletion
//   - Soft delete preserves post data
//   - Deleted posts are not returned in queries
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the post to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *postService) DeletePost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	// Perform soft delete by marking the post as deleted
	post.SoftDelete()
	return s.postRepo.Update(ctx, post)
}

// PublishPost changes a post's status to published, making it publicly visible.
// This method is part of the content publication workflow.
//
// Business Rules:
//   - Post must exist and not be deleted
//   - Publication status is updated atomically
//   - Published posts become publicly visible
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the post to publish
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *postService) PublishPost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	// Mark post as published
	post.Publish()
	return s.postRepo.Update(ctx, post)
}

// UnpublishPost changes a post's status to unpublished, making it private.
// This method is part of the content publication workflow.
//
// Business Rules:
//   - Post must exist and not be deleted
//   - Unpublication status is updated atomically
//   - Unpublished posts become private
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the post to unpublish
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *postService) UnpublishPost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	// Mark post as unpublished
	post.Unpublish()
	return s.postRepo.Update(ctx, post)
}
