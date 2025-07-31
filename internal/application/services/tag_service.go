// Package services provides application-level business logic for tag management.
// This package contains the tag service implementation that handles tag creation,
// retrieval, search, and management while ensuring proper slug generation
// and uniqueness validation.
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

// tagService implements the TagService interface and provides comprehensive
// tag management functionality. It handles tag creation, retrieval, search,
// and management while ensuring proper slug generation and uniqueness validation.
type tagService struct {
	tagRepository repositories.TagRepository
}

// NewTagService creates a new tag service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - tagRepository: Repository interface for tag data access operations
//
// Returns:
//   - ports.TagService: The tag service interface implementation
func NewTagService(tagRepository repositories.TagRepository) ports.TagService {
	return &tagService{
		tagRepository: tagRepository,
	}
}

// CreateTag creates a new tag with comprehensive validation and slug generation.
// This method enforces business rules for tag creation and ensures slug uniqueness.
//
// Business Rules:
//   - Tag name is required and validated
//   - Slug is auto-generated if not provided
//   - Slug must be unique across the system
//   - Color is optional but used for visual identification
//   - Description is optional for additional context
//
// Parameters:
//   - ctx: Context for the operation
//   - name: Display name of the tag
//   - slug: URL-friendly identifier (auto-generated if empty)
//   - description: Optional description of the tag
//   - color: Optional color code for visual identification
//
// Returns:
//   - *entities.Tag: The created tag entity
//   - error: Any error that occurred during the operation
func (s *tagService) CreateTag(ctx context.Context, name, slug, description, color string) (*entities.Tag, error) {
	// Generate slug automatically if not provided
	if slug == "" {
		slug = s.generateSlug(name)
	}

	// Check if slug already exists to maintain uniqueness
	exists, err := s.tagRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("tag with slug '%s' already exists", slug)
	}

	// Create tag entity with the provided parameters
	tag, err := entities.NewTag(name, slug, description, color)
	if err != nil {
		return nil, err
	}

	// Persist the tag to the repository
	err = s.tagRepository.Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetTagByID retrieves a tag by its unique identifier.
// This method provides access to individual tag details and metadata.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the tag to retrieve
//
// Returns:
//   - *entities.Tag: The tag entity if found
//   - error: Error if tag not found or other issues occur
func (s *tagService) GetTagByID(ctx context.Context, id uuid.UUID) (*entities.Tag, error) {
	return s.tagRepository.GetByID(ctx, id)
}

// GetTagBySlug retrieves a tag by its unique slug identifier.
// This method is useful for URL-based tag lookups and routing.
//
// Parameters:
//   - ctx: Context for the operation
//   - slug: Slug identifier of the tag to retrieve
//
// Returns:
//   - *entities.Tag: The tag entity if found
//   - error: Error if tag not found or other issues occur
func (s *tagService) GetTagBySlug(ctx context.Context, slug string) (*entities.Tag, error) {
	return s.tagRepository.GetBySlug(ctx, slug)
}

// GetAllTags retrieves all tags in the system with pagination.
// This method is useful for administrative purposes and tag management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of tags to return
//   - offset: Number of tags to skip for pagination
//
// Returns:
//   - []*entities.Tag: List of all tags
//   - error: Any error that occurred during the operation
func (s *tagService) GetAllTags(ctx context.Context, limit, offset int) ([]*entities.Tag, error) {
	return s.tagRepository.GetAll(ctx, limit, offset)
}

// SearchTags searches for tags based on a query string.
// This method supports full-text search capabilities for finding tags
// by name, description, or other attributes.
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Tag: List of matching tags
//   - error: Any error that occurred during the operation
func (s *tagService) SearchTags(ctx context.Context, query string, limit, offset int) ([]*entities.Tag, error) {
	return s.tagRepository.Search(ctx, query, limit, offset)
}

// UpdateTag updates an existing tag's information and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Tag must exist and be accessible
//   - Updated information must be provided and validated
//   - Slug uniqueness is maintained during updates
//   - Tag validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the tag to update
//   - name: Updated display name of the tag
//   - slug: Updated URL-friendly identifier
//   - description: Updated description of the tag
//   - color: Updated color code for visual identification
//
// Returns:
//   - *entities.Tag: The updated tag entity
//   - error: Any error that occurred during the operation
func (s *tagService) UpdateTag(ctx context.Context, id uuid.UUID, name, slug, description, color string) (*entities.Tag, error) {
	// Retrieve existing tag to ensure it exists and is accessible
	tag, err := s.tagRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update the tag entity with new information
	err = tag.UpdateTag(name, slug, description, color)
	if err != nil {
		return nil, err
	}

	// Persist the updated tag to the repository
	err = s.tagRepository.Update(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// DeleteTag performs a soft delete of a tag by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the tag to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *tagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return s.tagRepository.Delete(ctx, id)
}

// GetTagCount returns the total number of tags in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of tags
//   - error: Any error that occurred during the operation
func (s *tagService) GetTagCount(ctx context.Context) (int64, error) {
	return s.tagRepository.Count(ctx)
}

// generateSlug generates a URL-friendly slug from a tag name.
// This method creates SEO-friendly URLs by converting the tag name
// to a lowercase, hyphenated format with special character removal.
//
// Business Rules:
//   - Converts to lowercase for consistency
//   - Replaces spaces and underscores with hyphens
//   - Removes special characters except alphanumeric and hyphens
//   - Removes multiple consecutive hyphens
//   - Trims leading and trailing hyphens
//
// Parameters:
//   - name: Original tag name to convert to slug
//
// Returns:
//   - string: URL-friendly slug
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

	// Remove multiple consecutive hyphens for cleaner URLs
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading and trailing hyphens for cleaner URLs
	slug = strings.Trim(slug, "-")

	return slug
}
