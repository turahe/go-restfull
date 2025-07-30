package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// postService implements the PostService interface
type postService struct {
	postRepo repositories.PostRepository
}

// NewPostService creates a new post service instance
func NewPostService(postRepo repositories.PostRepository) ports.PostService {
	return &postService{
		postRepo: postRepo,
	}
}

func (s *postService) CreatePost(ctx context.Context, title, content, slug, status string, authorID uuid.UUID) (*entities.Post, error) {
	// Create post entity
	post, err := entities.NewPost(title, content, slug, status, authorID)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) GetPostByID(ctx context.Context, id uuid.UUID) (*entities.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (s *postService) GetPostBySlug(ctx context.Context, slug string) (*entities.Post, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (s *postService) GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetByAuthor(ctx, authorID, limit, offset)
}

func (s *postService) GetAllPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetAll(ctx, limit, offset)
}

func (s *postService) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.GetPublished(ctx, limit, offset)
}

func (s *postService) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error) {
	return s.postRepo.Search(ctx, query, limit, offset)
}

// GetPostsWithPagination retrieves posts with pagination and returns total count
func (s *postService) GetPostsWithPagination(ctx context.Context, page, perPage int, search, status string) ([]*entities.Post, int64, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	var posts []*entities.Post
	var err error

	// Get posts based on search and status parameters
	if search != "" && status == "published" {
		posts, err = s.postRepo.SearchPublished(ctx, search, perPage, offset)
	} else if search != "" {
		posts, err = s.postRepo.Search(ctx, search, perPage, offset)
	} else if status == "published" {
		posts, err = s.postRepo.GetPublished(ctx, perPage, offset)
	} else {
		posts, err = s.postRepo.GetAll(ctx, perPage, offset)
	}

	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := s.GetPostsCount(ctx, search, status)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsCount returns total count of posts (for pagination)
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

func (s *postService) UpdatePost(ctx context.Context, id uuid.UUID, title, content, slug, status string) (*entities.Post, error) {
	// Get existing post
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.IsDeleted() {
		return nil, errors.New("post not found")
	}

	// Update post
	if err := post.UpdatePost(title, content, slug, status); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) DeletePost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	post.SoftDelete()
	return s.postRepo.Update(ctx, post)
}

func (s *postService) PublishPost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	post.Publish()
	return s.postRepo.Update(ctx, post)
}

func (s *postService) UnpublishPost(ctx context.Context, id uuid.UUID) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.IsDeleted() {
		return errors.New("post not found")
	}

	post.Unpublish()
	return s.postRepo.Update(ctx, post)
}
