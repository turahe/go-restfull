package ports

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// ContentService defines the interface for content business operations
type ContentService interface {
	// CreateContent creates a new content
	CreateContent(ctx context.Context, modelType string, modelID uuid.UUID, contentRaw, contentHTML string, createdBy uuid.UUID) (*entities.Content, error)

	// GetContentByID retrieves content by ID
	GetContentByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)

	// GetContentByModelTypeAndID retrieves content by model type and model ID
	GetContentByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error)

	// GetAllContent retrieves all content with pagination
	GetAllContent(ctx context.Context, limit, offset int) ([]*entities.Content, error)

	// UpdateContent updates an existing content
	UpdateContent(ctx context.Context, id uuid.UUID, contentRaw, contentHTML string, updatedBy uuid.UUID) (*entities.Content, error)

	// DeleteContent soft deletes content by ID
	DeleteContent(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// HardDeleteContent permanently deletes content by ID
	HardDeleteContent(ctx context.Context, id uuid.UUID) error

	// RestoreContent restores soft deleted content
	RestoreContent(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*entities.Content, error)

	// GetContentCount returns the total number of content
	GetContentCount(ctx context.Context) (int64, error)

	// GetContentCountByModelType returns the total number of content by model type
	GetContentCountByModelType(ctx context.Context, modelType string) (int64, error)

	// SearchContent searches content by query
	SearchContent(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error)
}
