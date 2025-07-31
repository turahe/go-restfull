// Package services provides application-level business logic for content management.
// This package contains the content service implementation that handles content creation,
// retrieval, updates, and polymorphic content relationships while enforcing business rules.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// ContentService implements the ContentService interface and provides comprehensive
// content management functionality. It handles polymorphic content relationships,
// content versioning, soft delete operations, and content search capabilities
// while enforcing business rules and data integrity.
type ContentService struct {
	contentRepository repositories.ContentRepository
}

// NewContentService creates a new content service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - contentRepository: Repository interface for content data access operations
//
// Returns:
//   - ports.ContentService: The content service interface implementation
func NewContentService(contentRepository repositories.ContentRepository) ports.ContentService {
	return &ContentService{
		contentRepository: contentRepository,
	}
}

// CreateContent creates a new content entry for a polymorphic model relationship.
// This method enforces business rules for content creation and supports various
// content types (posts, pages, etc.) through polymorphic associations.
//
// Business Rules:
//   - Model type must be provided and validated
//   - Content raw text must be provided and validated
//   - Model ID must reference an existing entity
//   - Created by user must be specified for audit trails
//   - HTML content is optional but recommended for display
//
// Parameters:
//   - ctx: Context for the operation
//   - modelType: Type of the model this content belongs to (e.g., "post", "page")
//   - modelID: UUID of the model entity this content belongs to
//   - contentRaw: Raw text content
//   - contentHTML: HTML formatted content for display
//   - createdBy: UUID of the user creating the content
//
// Returns:
//   - *entities.Content: The created content entity
//   - error: Any error that occurred during the operation
func (s *ContentService) CreateContent(ctx context.Context, modelType string, modelID uuid.UUID, contentRaw, contentHTML string, createdBy uuid.UUID) (*entities.Content, error) {
	// Validate model type to ensure data integrity
	if strings.TrimSpace(modelType) == "" {
		return nil, errors.New("invalid model type")
	}

	// Validate content to ensure meaningful data
	if strings.TrimSpace(contentRaw) == "" {
		return nil, errors.New("invalid content")
	}

	// Create new content entity with the provided data
	content := entities.NewContent(modelType, modelID, contentRaw, contentHTML, createdBy)

	// Persist the content to the repository
	err := s.contentRepository.Create(ctx, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// GetContentByID retrieves a content entry by its unique identifier.
// This method includes soft delete checking to ensure deleted content
// is not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the content to retrieve
//
// Returns:
//   - *entities.Content: The content entity if found
//   - error: Error if content not found or other issues occur
func (s *ContentService) GetContentByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the content has been soft deleted
	if content.IsDeleted() {
		return nil, errors.New("content not found")
	}

	return content, nil
}

// GetContentByModelTypeAndID retrieves all content for a specific model type and model ID.
// This method supports polymorphic content relationships and filters out deleted content.
//
// Parameters:
//   - ctx: Context for the operation
//   - modelType: Type of the model to get content for
//   - modelID: UUID of the model entity
//
// Returns:
//   - []*entities.Content: List of active content for the model
//   - error: Any error that occurred during the operation
func (s *ContentService) GetContentByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	// Validate model type to ensure data integrity
	if strings.TrimSpace(modelType) == "" {
		return nil, errors.New("invalid model type")
	}

	contents, err := s.contentRepository.GetByModelTypeAndID(ctx, modelType, modelID)
	if err != nil {
		return nil, err
	}

	// Filter out deleted content to maintain data integrity
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}

// GetAllContent retrieves all content in the system with pagination.
// This method is useful for administrative purposes and content management.
//
// Business Rules:
//   - Default limit of 10 items if not specified
//   - Offset must be non-negative
//   - Deleted content is automatically filtered out
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of content items to return
//   - offset: Number of content items to skip for pagination
//
// Returns:
//   - []*entities.Content: List of active content items
//   - error: Any error that occurred during the operation
func (s *ContentService) GetAllContent(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
	// Set default pagination values for better user experience
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	contents, err := s.contentRepository.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter out deleted content to maintain data integrity
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}

// UpdateContent updates an existing content entry with new content and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Content must exist and not be deleted
//   - Updated content must be provided and validated
//   - Updated by user must be specified for audit trails
//   - HTML content is optional but recommended for display
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the content to update
//   - contentRaw: Updated raw text content
//   - contentHTML: Updated HTML formatted content
//   - updatedBy: UUID of the user updating the content
//
// Returns:
//   - *entities.Content: The updated content entity
//   - error: Any error that occurred during the operation
func (s *ContentService) UpdateContent(ctx context.Context, id uuid.UUID, contentRaw, contentHTML string, updatedBy uuid.UUID) (*entities.Content, error) {
	// Retrieve existing content to ensure it exists and is not deleted
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the content has been soft deleted
	if content.IsDeleted() {
		return nil, errors.New("content not found")
	}

	// Validate updated content to ensure meaningful data
	if strings.TrimSpace(contentRaw) == "" {
		return nil, errors.New("invalid content")
	}

	// Update the content entity with new data
	content.UpdateContent(contentRaw, contentHTML, updatedBy)

	// Persist the updated content to the repository
	err = s.contentRepository.Update(ctx, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// DeleteContent performs a soft delete of content by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - Content must exist and not be already deleted
//   - Deleted by user must be specified for audit trails
//   - Soft delete preserves data for potential restoration
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the content to delete
//   - deletedBy: UUID of the user performing the deletion
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *ContentService) DeleteContent(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Retrieve existing content to ensure it exists and is not already deleted
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the content has already been soft deleted
	if content.IsDeleted() {
		return errors.New("content not found")
	}

	// Perform soft delete by marking the content as deleted
	return s.contentRepository.Delete(ctx, id, deletedBy)
}

// HardDeleteContent permanently removes content from the database.
// This operation is irreversible and should be used with extreme caution.
//
// Security Considerations:
//   - This operation is irreversible
//   - Should only be used for data cleanup or compliance requirements
//   - Consider backup before execution
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the content to permanently delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *ContentService) HardDeleteContent(ctx context.Context, id uuid.UUID) error {
	return s.contentRepository.HardDelete(ctx, id)
}

// RestoreContent restores a previously soft-deleted content entry.
// This method is useful for content recovery and administrative operations.
//
// Business Rules:
//   - Content must exist and be in deleted state
//   - Restored by user must be specified for audit trails
//   - Restoration makes content available again
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the content to restore
//   - updatedBy: UUID of the user performing the restoration
//
// Returns:
//   - *entities.Content: The restored content entity
//   - error: Any error that occurred during the operation
func (s *ContentService) RestoreContent(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*entities.Content, error) {
	// Retrieve existing content to ensure it exists and is deleted
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the content is not deleted (cannot restore active content)
	if !content.IsDeleted() {
		return nil, errors.New("content is not deleted")
	}

	// Restore the content by removing the deleted flag
	err = s.contentRepository.Restore(ctx, id, updatedBy)
	if err != nil {
		return nil, err
	}

	// Retrieve and return the restored content
	return s.contentRepository.GetByID(ctx, id)
}

// GetContentCount returns the total number of content items in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of content items
//   - error: Any error that occurred during the operation
func (s *ContentService) GetContentCount(ctx context.Context) (int64, error) {
	return s.contentRepository.Count(ctx)
}

// GetContentCountByModelType returns the total number of content items
// for a specific model type. This method is useful for content type statistics.
//
// Parameters:
//   - ctx: Context for the operation
//   - modelType: Type of the model to count content for
//
// Returns:
//   - int64: Total count of content items for the model type
//   - error: Any error that occurred during the operation
func (s *ContentService) GetContentCountByModelType(ctx context.Context, modelType string) (int64, error) {
	// Validate model type to ensure data integrity
	if strings.TrimSpace(modelType) == "" {
		return 0, errors.New("invalid model type")
	}

	return s.contentRepository.CountByModelType(ctx, modelType)
}

// SearchContent searches for content items based on a query string.
// This method supports full-text search capabilities with pagination.
//
// Business Rules:
//   - Search query must be provided and validated
//   - Default limit of 10 items if not specified
//   - Offset must be non-negative
//   - Deleted content is automatically filtered out
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Content: List of matching content items
//   - error: Any error that occurred during the operation
func (s *ContentService) SearchContent(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	// Validate search query to ensure meaningful search
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("invalid search query")
	}

	// Set default pagination values for better user experience
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	contents, err := s.contentRepository.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter out deleted content to maintain data integrity
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}
