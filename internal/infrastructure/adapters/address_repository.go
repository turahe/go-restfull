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

type addressRepository struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) repositories.AddressRepository {
	return &addressRepository{
		db: db,
	}
}

func (r *addressRepository) Create(ctx context.Context, address *entities.Address) error {
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

func (r *addressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
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
			return nil, fmt.Errorf("address not found")
		}
		return nil, err
	}

	return &address, nil
}

func (r *addressRepository) Update(ctx context.Context, address *entities.Address) error {
	query := `
		UPDATE addresses SET
			address_line1 = $2, address_line2 = $3, city = $4, state = $5,
			postal_code = $6, country = $7, latitude = $8, longitude = $9,
			is_primary = $10, address_type = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		address.ID,
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
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found")
	}

	return nil
}

func (r *addressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE addresses SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found")
	}

	return nil
}

func (r *addressRepository) GetByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) ([]*entities.Address, error) {
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
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) GetPrimaryByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND is_primary = true AND deleted_at IS NULL
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
			return nil, fmt.Errorf("primary address not found")
		}
		return nil, err
	}

	return &address, nil
}

func (r *addressRepository) GetByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) ([]*entities.Address, error) {
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
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) SetPrimary(ctx context.Context, id uuid.UUID, addressableID uuid.UUID, addressableType entities.AddressableType) error {
	query := `
		UPDATE addresses SET is_primary = true, updated_at = NOW()
		WHERE id = $1 AND addressable_id = $2 AND addressable_type = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, addressableID, addressableType)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("address not found")
	}

	return nil
}

func (r *addressRepository) UnsetOtherPrimaries(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, excludeID uuid.UUID) error {
	var query string
	var args []interface{}

	if excludeID == uuid.Nil {
		query = `
			UPDATE addresses SET is_primary = false, updated_at = NOW()
			WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
		`
		args = []interface{}{addressableID, addressableType}
	} else {
		query = `
			UPDATE addresses SET is_primary = false, updated_at = NOW()
			WHERE addressable_id = $1 AND addressable_type = $2 AND id != $3 AND deleted_at IS NULL
		`
		args = []interface{}{addressableID, addressableType, excludeID}
	}

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *addressRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE city ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, "%"+city+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) SearchByState(ctx context.Context, state string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE state ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, "%"+state+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE country ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, "%"+country+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) SearchByPostalCode(ctx context.Context, postalCode string, limit, offset int) ([]*entities.Address, error) {
	query := `
		SELECT id, addressable_id, addressable_type, address_line1, address_line2,
			   city, state, postal_code, country, latitude, longitude,
			   is_primary, address_type, created_at, updated_at, deleted_at
		FROM addresses
		WHERE postal_code ILIKE $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, "%"+postalCode+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAddresses(rows)
}

func (r *addressRepository) CountByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (int64, error) {
	query := `
		SELECT COUNT(*) FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&count)
	return count, err
}

func (r *addressRepository) CountByType(ctx context.Context, addressType entities.AddressType) (int64, error) {
	query := `
		SELECT COUNT(*) FROM addresses
		WHERE address_type = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressType).Scan(&count)
	return count, err
}

func (r *addressRepository) CountByAddressableAndType(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType, addressType entities.AddressType) (int64, error) {
	query := `
		SELECT COUNT(*) FROM addresses
		WHERE addressable_id = $1 AND addressable_type = $2 AND address_type = $3 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, addressableID, addressableType, addressType).Scan(&count)
	return count, err
}

func (r *addressRepository) ExistsByAddressable(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM addresses
			WHERE addressable_id = $1 AND addressable_type = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&exists)
	return exists, err
}

func (r *addressRepository) HasPrimaryAddress(ctx context.Context, addressableID uuid.UUID, addressableType entities.AddressableType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM addresses
			WHERE addressable_id = $1 AND addressable_type = $2 AND is_primary = true AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, addressableID, addressableType).Scan(&exists)
	return exists, err
}

func (r *addressRepository) scanAddresses(rows pgx.Rows) ([]*entities.Address, error) {
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
			return nil, err
		}
		addresses = append(addresses, &address)
	}

	return addresses, nil
}
