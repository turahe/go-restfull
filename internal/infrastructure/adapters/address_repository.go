package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAddressRepository is an adapter that implements the AddressRepository interface
// by delegating calls to the concrete repository implementation
type PostgresAddressRepository struct {
	repo repositories.AddressRepository
}

// NewPostgresAddressRepository creates a new PostgresAddressRepository adapter
func NewPostgresAddressRepository(db *pgxpool.Pool) repositories.AddressRepository {
	return &PostgresAddressRepository{
		repo: repository.NewAddressRepository(db),
	}
}

// Create delegates to the underlying repository implementation
func (r *PostgresAddressRepository) Create(ctx context.Context, address *entities.Address) error {
	return r.repo.Create(ctx, address)
}

// GetByID delegates to the underlying repository implementation
func (r *PostgresAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	return r.repo.GetByID(ctx, id)
}

// Update delegates to the underlying repository implementation
func (r *PostgresAddressRepository) Update(ctx context.Context, address *entities.Address) error {
	return r.repo.Update(ctx, address)
}

// Delete delegates to the underlying repository implementation
func (r *PostgresAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

// GetByAddressable delegates to the underlying repository implementation
func (r *PostgresAddressRepository) GetByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error) {
	return r.repo.GetByAddressable(ctx, addressableID, addressableType)
}

// GetPrimaryByAddressable delegates to the underlying repository implementation
func (r *PostgresAddressRepository) GetPrimaryByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error) {
	return r.repo.GetPrimaryByAddressable(ctx, addressableID, addressableType)
}

// GetByAddressableAndType delegates to the underlying repository implementation
func (r *PostgresAddressRepository) GetByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error) {
	return r.repo.GetByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

// SetPrimary delegates to the underlying repository implementation
func (r *PostgresAddressRepository) SetPrimary(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error {
	return r.repo.SetPrimary(ctx, id, addressableID, addressableType)
}

// UnsetOtherPrimaries delegates to the underlying repository implementation
func (r *PostgresAddressRepository) UnsetOtherPrimaries(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, excludeID uuid.UUID) error {
	return r.repo.UnsetOtherPrimaries(ctx, addressableID, addressableType, excludeID)
}

// Search delegates to the underlying repository implementation
func (r *PostgresAddressRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Address, error) {
	return r.repo.Search(ctx, query, limit, offset)
}

// SearchByCity delegates to the underlying repository implementation
func (r *PostgresAddressRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error) {
	return r.repo.SearchByCity(ctx, city, limit, offset)
}

// SearchByState delegates to the underlying repository implementation
func (r *PostgresAddressRepository) SearchByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error) {
	return r.repo.SearchByState(ctx, state, limit, offset)
}

// SearchByCountry delegates to the underlying repository implementation
func (r *PostgresAddressRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error) {
	return r.repo.SearchByCountry(ctx, country, limit, offset)
}

// SearchByPostalCode delegates to the underlying repository implementation
func (r *PostgresAddressRepository) SearchByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error) {
	return r.repo.SearchByPostalCode(ctx, postalCode, limit, offset)
}

// CountByAddressable delegates to the underlying repository implementation
func (r *PostgresAddressRepository) CountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error) {
	return r.repo.CountByAddressable(ctx, addressableID, addressableType)
}

// CountByType delegates to the underlying repository implementation
func (r *PostgresAddressRepository) CountByType(ctx context.Context, addressType entities.AddressType) (int64, error) {
	return r.repo.CountByType(ctx, addressType)
}

// CountByAddressableAndType delegates to the underlying repository implementation
func (r *PostgresAddressRepository) CountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error) {
	return r.repo.CountByAddressableAndType(ctx, addressableID, addressableType, addressType)
}

// ExistsByAddressable delegates to the underlying repository implementation
func (r *PostgresAddressRepository) ExistsByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return r.repo.ExistsByAddressable(ctx, addressableID, addressableType)
}

// HasPrimaryAddress delegates to the underlying repository implementation
func (r *PostgresAddressRepository) HasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	return r.repo.HasPrimaryAddress(ctx, addressableID, addressableType)
}
