package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// tagService implements TagService interface
type tagService struct {
	tagRepository repositories.TagRepository
}

// NewTagService creates a new tag service
func NewTagService(tagRepository repositories.TagRepository) ports.TagService {
	return &tagService{
		tagRepository: tagRepository,
	}
}

// CreateTag creates a new tag
func (s *tagService) CreateTag(ctx context.Context, name, slug, description, color string) (*entities.Tag, error) {
	// Generate slug if not provided
	if slug == "" {
		slug = s.generateSlug(name)
	}

	// Check if slug already exists
	exists, err := s.tagRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("tag with slug '%s' already exists", slug)
	}

	// Create tag entity
	tag, err := entities.NewTag(name, slug, description, color)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.tagRepository.Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetTagByID retrieves tag by ID
func (s *tagService) GetTagByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	return s.tagRepository.GetByID(ctx, id)
}

// GetTagBySlug retrieves tag by slug
func (s *tagService) GetTagBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	return s.tagRepository.GetBySlug(ctx, slug)
}

// GetAllTags retrieves all tags with pagination
func (s *tagService) GetAllTags(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	return s.tagRepository.GetAll(ctx, limit, offset)
}

// SearchTags searches tags by query
func (s *tagService) SearchTags(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	return s.tagRepository.Search(ctx, query, limit, offset)
}

// UpdateTag updates tag information
func (s *tagService) UpdateTag(ctx context.Context, id uuid.UUID, name, slug, description, color string) (*entities.Tag, error) {
	tag, err := s.tagRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tag.UpdateTag(name, slug, description, color)
	if err != nil {
		return nil, err
	}

	err = s.tagRepository.Update(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// DeleteTag deletes tag
func (s *tagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return s.tagRepository.Delete(ctx, id)
}

// GetTagCount returns the total number of tags
func (s *tagService) GetTagCount(ctx context.Context) (int64, error) {
	return s.tagRepository.Count(ctx)
}

// generateSlug generates a URL-friendly slug from a name
func (s *tagService) generateSlug(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Remove multiple consecutive hyphens
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
