package repositories

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
)

// ContentRepository defines the interface for content operations
type ContentRepository interface {
	// Create creates a new content
	Create(ctx context.Context, content *entities.Content) error

	// GetByID retrieves content by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)

	// GetByModelTypeAndID retrieves content by model type and model ID
	GetByModelTypeAndID(ctx context.Context, modelType string, modelID uuid.UUID) ([]*entities.Content, error)

	// GetAll retrieves all content with pagination
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Content, error)

	// Update updates an existing content
	Update(ctx context.Context, content *entities.Content) error

	// Delete soft deletes content by ID
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// HardDelete permanently deletes content by ID
	HardDelete(ctx context.Context, id uuid.UUID) error

	// Restore restores soft deleted content
	Restore(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error

	// Count returns the total number of content
	Count(ctx context.Context) (int64, error)

	// CountByModelType returns the total number of content by model type
	CountByModelType(ctx context.Context, modelType string) (int64, error)

	// Search searches content by query
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Content, error)
}
