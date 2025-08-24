package repositories

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// SettingRepository defines the contract for setting data operations.
// This port is implemented by infrastructure adapters.
type SettingRepository interface {
	Create(ctx context.Context, setting *entities.Setting) error
	BatchCreate(ctx context.Context, settings []*entities.Setting) error

	GetByID(ctx context.Context, id uuid.UUID) (*entities.Setting, error)
	GetByKey(ctx context.Context, key string) (*entities.Setting, error)
	GetAll(ctx context.Context) ([]*entities.Setting, error)

	Update(ctx context.Context, setting *entities.Setting) error
	Delete(ctx context.Context, id uuid.UUID) error

	ExistsByKey(ctx context.Context, key string) (bool, error)
	Count(ctx context.Context) (int64, error)
}
