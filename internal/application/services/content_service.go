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

// ContentService implements the ContentService interface
type ContentService struct {
	contentRepository repositories.ContentRepository
}

// NewContentService creates a new content service
func NewContentService(contentRepository repositories.ContentRepository) ports.ContentService {
	return &ContentService{
		contentRepository: contentRepository,
	}
}

// CreateContent creates a new content
func (s *ContentService) CreateContent(ctx context.Context, modelType string, modelID uuid.UUID, contentRaw, contentHTML string, createdBy uuid.UUID) (*entities.Content, error) {
	// Validate inputs
	if strings.TrimSpace(modelType) == "" {
		return nil, errors.New("invalid model type")
	}

	if strings.TrimSpace(contentRaw) == "" {
		return nil, errors.New("invalid content")
	}

	// Create new content entity
	content := entities.NewContent(modelType, modelID, contentRaw, contentHTML, createdBy)

	// Save to repository
	err := s.contentRepository.Create(ctx, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// GetContentByID retrieves content by ID
func (s *ContentService) GetContentByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if content.IsDeleted() {
		return nil, errors.New("content not found")
	}

	return content, nil
}

// GetContentByModelTypeAndID retrieves content by model type and model ID
func (s *ContentService) GetContentByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error) {
	if strings.TrimSpace(modelType) == "" {
		return nil, errors.New("invalid model type")
	}

	contents, err := s.contentRepository.GetByModelTypeAndID(ctx, modelType, modelID)
	if err != nil {
		return nil, err
	}

	// Filter out deleted content
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}

// GetAllContent retrieves all content with pagination
func (s *ContentService) GetAllContent(ctx context.Context, limit, offset int) ([]*entities.Content, error) {
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

	// Filter out deleted content
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}

// UpdateContent updates an existing content
func (s *ContentService) UpdateContent(ctx context.Context, id uuid.UUID, contentRaw, contentHTML string, updatedBy uuid.UUID) (*entities.Content, error) {
	// Get existing content
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if content.IsDeleted() {
		return nil, errors.New("content not found")
	}

	if strings.TrimSpace(contentRaw) == "" {
		return nil, errors.New("invalid content")
	}

	// Update content
	content.UpdateContent(contentRaw, contentHTML, updatedBy)

	// Save to repository
	err = s.contentRepository.Update(ctx, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// DeleteContent soft deletes content by ID
func (s *ContentService) DeleteContent(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Get existing content
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if content.IsDeleted() {
		return errors.New("content not found")
	}

	// Soft delete content
	return s.contentRepository.Delete(ctx, id, deletedBy)
}

// HardDeleteContent permanently deletes content by ID
func (s *ContentService) HardDeleteContent(ctx context.Context, id uuid.UUID) error {
	return s.contentRepository.HardDelete(ctx, id)
}

// RestoreContent restores soft deleted content
func (s *ContentService) RestoreContent(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*entities.Content, error) {
	// Get existing content
	content, err := s.contentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !content.IsDeleted() {
		return nil, errors.New("content is not deleted")
	}

	// Restore content
	err = s.contentRepository.Restore(ctx, id, updatedBy)
	if err != nil {
		return nil, err
	}

	// Get updated content
	return s.contentRepository.GetByID(ctx, id)
}

// GetContentCount returns the total number of content
func (s *ContentService) GetContentCount(ctx context.Context) (int64, error) {
	return s.contentRepository.Count(ctx)
}

// GetContentCountByModelType returns the total number of content by model type
func (s *ContentService) GetContentCountByModelType(ctx context.Context, modelType string) (int64, error) {
	if strings.TrimSpace(modelType) == "" {
		return 0, errors.New("invalid model type")
	}

	return s.contentRepository.CountByModelType(ctx, modelType)
}

// SearchContent searches content by query
func (s *ContentService) SearchContent(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("invalid search query")
	}

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

	// Filter out deleted content
	var activeContents []*entities.Content
	for _, content := range contents {
		if !content.IsDeleted() {
			activeContents = append(activeContents, content)
		}
	}

	return activeContents, nil
}
