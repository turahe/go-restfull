package adapters

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAddressRepository provides the concrete implementation of the AddressRepository interface
// using PostgreSQL as the underlying data store. This struct handles all address-related
// database operations including CRUD operations, search, and address management.
type PostgresAddressRepository struct {
	db *pgxpool.Pool // PostgreSQL connection pool for database operations
}

// NewPostgresAddressRepository creates a new instance of PostgresAddressRepository
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//
// Returns:
//   - repositories.AddressRepository: interface implementation for address management
func NewPostgresAddressRepository(db *pgxpool.Pool) repositories.AddressRepository {
	return &PostgresAddressRepository{
		db: db,
	}
}

// Create persists a new address to the database
// This method inserts a new address record with all required fields including
// addressable entity association, location details, and metadata.
//
// Parameters:
//   - ctx: context for the database operation
//   - address: pointer to the address entity to create
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) Create(ctx context.Context, address *entities.Address) error {
	query := `
		INSERT INTO addresses (
			id, addressable_id, addressable_type, address_line1, address_line2,
			city, state, postal_code, country, latitude, longitude,
			is_primary, address_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.db.Exec(ctx, query,
		address.ID,
		address.AddressableID,
		address.AddressableType,
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.Latitude,
		address.Longitude,
		address.IsPrimary,
		address.AddressType,
		address.CreatedAt,
		address.UpdatedAt,
	)

	return err
}

// GetByID retrieves an address by its unique identifier
// This method performs a soft-delete aware query, only returning addresses that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the address to retrieve
//
// Returns:
//   - *entities.Address: pointer to the found address entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE id = $1 AND deleted_at IS NULL
	`

	var address entities.Address
	err := r.db.QueryRow(ctx, query, id).Scan(
		&address.ID,
		&address.AddressableID,
		&address.AddressableType,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.Latitude,
		&address.Longitude,
		&address.IsPrimary,
		&address.AddressType,
		&address.CreatedAt,
		&address.UpdatedAt,
		&address.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}

	return &address, nil
}

// Update modifies an existing address in the database
// This method updates all address fields and sets the updated_at timestamp.
//
// Parameters:
//   - ctx: context for the database operation
//   - address: pointer to the address entity with updated values
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) Update(ctx context.Context, address *entities.Address) error {
	query := `
		UPDATE addresses SET
			addressable_id = $2, addressable_type = $3, address_line1 = $4,
			address_line2 = $5, city = $6, state = $7, postal_code = $8,
			country = $9, latitude = $10, longitude = $11, is_primary = $12,
			address_type = $13, updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		address.ID,
		address.AddressableID,
		address.AddressableType,
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.Latitude,
		address.Longitude,
		address.IsPrimary,
		address.AddressType,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found or already deleted")
	}

	return nil
}

// Delete performs a soft delete of an address by setting the deleted_at timestamp
// This method preserves the data while marking it as deleted for business logic purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the address to delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE addresses SET
			deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found or already deleted")
	}

	return nil
}

// GetByAddressable retrieves all addresses for a specific addressable entity
// This method returns addresses ordered by creation date, with primary addresses first.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - []*entities.Address: slice of address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) GetByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
		ORDER BY is_primary DESC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, addressableID, addressableType)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by addressable: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// GetPrimaryByAddressable retrieves the primary address for a specific addressable entity
// This method returns the first primary address found for the given entity.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - *entities.Address: pointer to the primary address entity, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) GetPrimaryByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND is_primary = true AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
	`

	var address entities.Address
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(
		&address.ID,
		&address.AddressableID,
		&address.AddressableType,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.Latitude,
		&address.Longitude,
		&address.IsPrimary,
		&address.AddressType,
		&address.CreatedAt,
		&address.UpdatedAt,
		&address.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get primary address by addressable: %w", err)
	}

	return &address, nil
}

// GetByAddressableAndType retrieves addresses for a specific addressable entity and address type
// This method filters addresses by both the addressable entity and the specific address type.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//   - addressType: type of address to filter by
//
// Returns:
//   - []*entities.Address: slice of address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) GetByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND address_type = $3 AND deleted_at IS NULL
		ORDER BY is_primary DESC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, addressableID, addressableType, addressType)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by addressable and type: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// SetPrimary sets an address as the primary address for its addressable entity
// This method first unsets other primary addresses for the same entity, then sets the specified address as primary.
//
// Parameters:
//   - ctx: context for the database operation
//   - id: UUID of the address to set as primary
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) SetPrimary(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error {
	// First, unset other primary addresses
	if err := r.UnsetOtherPrimaries(ctx, addressableID, addressableType, id); err != nil {
		return fmt.Errorf("failed to unset other primary addresses: %w", err)
	}

	// Then, set this address as primary
	query := `
		UPDATE addresses SET
			is_primary = true, updated_at = NOW()
		WHERE id = $1 AND addressable_id = $2 AND addressable_type = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, addressableID, addressableType)
	if err != nil {
		return fmt.Errorf("failed to set address as primary: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found or already deleted")
	}

	return nil
}

// UnsetOtherPrimaries removes the primary flag from all other addresses of the same addressable entity
// This method is used when setting a new primary address to ensure only one primary address exists.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//   - excludeID: UUID of the address to exclude from unsetting (usually the new primary)
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) UnsetOtherPrimaries(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, excludeID uuid.UUID) error {
	query := `
		UPDATE addresses SET
			is_primary = false, updated_at = NOW()
		WHERE addressable_id = $1 AND addressable_type = $2 AND id != $3 AND is_primary = true AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, addressableID, addressableType, excludeID)
	if err != nil {
		return fmt.Errorf("failed to unset other primary addresses: %w", err)
	}

	return nil
}

// Search performs a general search across address fields
// This method searches across city, state, country, and postal code fields.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query string
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: slice of matching address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Address, error) {
	searchQuery := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE deleted_at IS NULL AND (
			city ILIKE $1 OR state ILIKE $1 OR country ILIKE $1 OR postal_code ILIKE $1
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// SearchByCity searches for addresses in a specific city
// This method performs a case-insensitive search for addresses in the specified city.
//
// Parameters:
//   - ctx: context for the database operation
//   - city: city name to search for
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: slice of matching address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE city ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + city + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by city: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// SearchByState searches for addresses in a specific state
// This method performs a case-insensitive search for addresses in the specified state.
//
// Parameters:
//   - ctx: context for the database operation
//   - state: state name to search for
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: slice of matching address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) SearchByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE state ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + state + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by state: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// SearchByCountry searches for addresses in a specific country
// This method performs a case-insensitive search for addresses in the specified country.
//
// Parameters:
//   - ctx: context for the database operation
//   - country: country name to search for
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: slice of matching address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE country ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + country + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by country: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// SearchByPostalCode searches for addresses with a specific postal code
// This method performs a case-insensitive search for addresses with the specified postal code.
//
// Parameters:
//   - ctx: context for the database operation
//   - postalCode: postal code to search for
//   - limit: maximum number of results to return
//   - offset: number of results to skip for pagination
//
// Returns:
//   - []*entities.Address: slice of matching address entities, or empty slice if none found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) SearchByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE postal_code ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + postalCode + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by postal code: %w", err)
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

// CountByAddressable returns the total count of addresses for a specific addressable entity
// This method is useful for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - int64: total count of addresses for the entity
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) CountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count addresses by addressable: %w", err)
	}

	return count, nil
}

// CountByType returns the total count of addresses of a specific type
// This method is useful for reporting and analytics purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressType: type of address to count
//
// Returns:
//   - int64: total count of addresses of the specified type
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) CountByType(ctx context.Context, addressType entities.AddressType) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM addresses
		WHERE address_type = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count addresses by type: %w", err)
	}

	return count, nil
}

// CountByAddressableAndType returns the total count of addresses for a specific addressable entity and type
// This method is useful for pagination and reporting purposes when filtering by both entity and type.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//   - addressType: type of address to count
//
// Returns:
//   - int64: total count of addresses matching the criteria
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) CountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND address_type = $3 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressableID, addressableType, addressType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count addresses by addressable and type: %w", err)
	}

	return count, nil
}

// ExistsByAddressable checks if an addressable entity has any addresses
// This method is useful for validation and business logic checks.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - bool: true if the entity has addresses, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) ExistsByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM addresses
			WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if addressable has addresses: %w", err)
	}

	return exists, nil
}

// HasPrimaryAddress checks if an addressable entity has a primary address
// This method is useful for validation and business logic checks.
//
// Parameters:
//   - ctx: context for the database operation
//   - addressableID: UUID of the addressable entity
//   - addressableType: type of the addressable entity
//
// Returns:
//   - bool: true if the entity has a primary address, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) HasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM addresses
			WHERE addressable_id = $1 AND addressable_type = $2 AND is_primary = true AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if addressable has primary address: %w", err)
	}

	return exists, nil
}

// scanAddresses is a helper method that scans database rows into address entities
// This method handles the repetitive task of scanning address data from database rows.
//
// Parameters:
//   - rows: database rows containing address data
//
// Returns:
//   - []*entities.Address: slice of scanned address entities
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresAddressRepository) scanAddresses(rows pgx.Rows) ([]*entities.Address, error) {
	var addresses []*entities.Address

	for rows.Next() {
		var address entities.Address
		err := rows.Scan(
			&address.ID,
			&address.AddressableID,
			&address.AddressableType,
			&address.AddressLine1,
			&address.AddressLine2,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.Latitude,
			&address.Longitude,
			&address.IsPrimary,
			&address.AddressType,
			&address.CreatedAt,
			&address.UpdatedAt,
			&address.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		addresses = append(addresses, &address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over address rows: %w", err)
	}

	return addresses, nil
}
