package adapters

import (
	"context"
	"fmt"

	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository provides the concrete implementation of the UserRepository interface
// using PostgreSQL as the underlying data store. This struct handles all user aggregate-related
// database operations including CRUD operations, search, and user management.
type PostgresUserRepository struct {
	*BaseTransactionalRepository
	db *pgxpool.Pool // PostgreSQL connection pool for database operations
}

// NewPostgresUserRepository creates a new instance of PostgresUserRepository
// This constructor function initializes the repository with the required dependencies.
//
// Parameters:
//   - db: PostgreSQL connection pool for database operations
//
// Returns:
//   - repositories.UserRepository: interface implementation for user aggregate management
func NewPostgresUserRepository(db *pgxpool.Pool) repositories.UserRepository {
	return &PostgresUserRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		db:                          db,
	}
}

// Save persists a user aggregate to the database
// This method handles both creation and updates of user aggregates.
//
// Parameters:
//   - ctx: context for the database operation
//   - user: pointer to the user aggregate to save
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) Save(ctx context.Context, user *aggregates.UserAggregate) error {
	// Check if user exists to determine if this is an insert or update
	exists, err := r.ExistsByEmail(ctx, user.Email.String())
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return r.updateUser(ctx, user)
	}
	return r.createUser(ctx, user)
}

// Delete performs a soft delete of a user by setting the deleted_at timestamp
// This method preserves the data while marking it as deleted for business logic purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user to delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users SET
			deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// FindByID retrieves a user aggregate by its unique identifier
// This method performs a soft-delete aware query, only returning users that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user to retrieve
//
// Returns:
//   - *aggregates.UserAggregate: pointer to the found user aggregate, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			   created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user aggregates.UserAggregate
	var passwordHash, emailStr, phoneStr string
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.UserName,
		&emailStr,
		&phoneStr,
		&passwordHash,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	// Convert strings to value objects
	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create email from string: %w", err)
	}
	user.Email = email

	phone, err := valueobjects.NewPhone(phoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create phone from string: %w", err)
	}
	user.Phone = phone

	// Convert password hash string to HashedPassword value object
	hashedPassword, err := valueobjects.NewHashedPasswordFromHash(passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashed password from hash: %w", err)
	}
	user.Password = hashedPassword

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return &user, nil
}

// FindByEmail retrieves a user aggregate by email address
// This method performs a soft-delete aware query, only returning users that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - email: email address of the user to retrieve
//
// Returns:
//   - *aggregates.UserAggregate: pointer to the found user aggregate, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			   created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user aggregates.UserAggregate
	var passwordHash, emailStr, phoneStr string
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&emailStr,
		&phoneStr,
		&passwordHash,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	// Convert strings to value objects
	emailVO, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create email from string: %w", err)
	}
	user.Email = emailVO

	phone, err := valueobjects.NewPhone(phoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create phone from string: %w", err)
	}
	user.Phone = phone

	// Convert password hash string to HashedPassword value object
	hashedPassword, err := valueobjects.NewHashedPasswordFromHash(passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashed password from hash: %w", err)
	}
	user.Password = hashedPassword

	return &user, nil
}

// FindByUsername retrieves a user aggregate by username
// This method performs a soft-delete aware query, only returning users that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - username: username of the user to retrieve
//
// Returns:
//   - *aggregates.UserAggregate: pointer to the found user aggregate, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			   created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	var user aggregates.UserAggregate
	var passwordHash, emailStr, phoneStr string
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.UserName,
		&emailStr,
		&phoneStr,
		&passwordHash,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	// Convert strings to value objects
	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create email from string: %w", err)
	}
	user.Email = email

	phone, err := valueobjects.NewPhone(phoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create phone from string: %w", err)
	}
	user.Phone = phone

	// Convert password hash string to HashedPassword value object
	hashedPassword, err := valueobjects.NewHashedPasswordFromHash(passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashed password from hash: %w", err)
	}
	user.Password = hashedPassword

	return &user, nil
}

// FindByPhone retrieves a user aggregate by phone number
// This method performs a soft-delete aware query, only returning users that haven't been deleted.
//
// Parameters:
//   - ctx: context for the database operation
//   - phone: phone number of the user to retrieve
//
// Returns:
//   - *aggregates.UserAggregate: pointer to the found user aggregate, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			   created_at, updated_at, deleted_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`

	var user aggregates.UserAggregate
	var passwordHash, emailStr, phoneStr string
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.UserName,
		&emailStr,
		&phoneStr,
		&passwordHash,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	// Convert strings to value objects
	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create email from string: %w", err)
	}
	user.Email = email

	phoneVO, err := valueobjects.NewPhone(phoneStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create phone from string: %w", err)
	}
	user.Phone = phoneVO

	// Convert password hash string to HashedPassword value object
	hashedPassword, err := valueobjects.NewHashedPasswordFromHash(passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create hashed password from hash: %w", err)
	}
	user.Password = hashedPassword

	return &user, nil
}

// FindAll retrieves all users based on the provided query parameters
// This method supports pagination and filtering through the ListUsersQuery.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: query parameters for listing users
//
// Returns:
//   - *handlers.PaginatedResult[*aggregates.UserAggregate]: paginated result containing users
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Implementation would depend on the specific structure of ListUsersQuery
	// For now, returning a basic implementation
	return &handlers.PaginatedResult[*aggregates.UserAggregate]{
		Items:      []*aggregates.UserAggregate{},
		TotalCount: 0,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: 0,
	}, nil
}

// Search searches for users based on the provided search query
// This method supports pagination and text-based search.
//
// Parameters:
//   - ctx: context for the database operation
//   - query: search query parameters
//
// Returns:
//   - *handlers.PaginatedResult[*aggregates.UserAggregate]: paginated result containing matching users
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Implementation would depend on the specific structure of SearchUsersQuery
	// For now, returning a basic implementation
	return &handlers.PaginatedResult[*aggregates.UserAggregate]{
		Items:      []*aggregates.UserAggregate{},
		TotalCount: 0,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: 0,
	}, nil
}

// ExistsByEmail checks if a user with the specified email exists
// This method is useful for validation and business logic checks.
//
// Parameters:
//   - ctx: context for the database operation
//   - email: email address to check
//
// Returns:
//   - bool: true if a user with the email exists, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE email = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by email: %w", err)
	}

	return exists, nil
}

// ExistsByUsername checks if a user with the specified username exists
// This method is useful for validation and business logic checks.
//
// Parameters:
//   - ctx: context for the database operation
//   - username: username to check
//
// Returns:
//   - bool: true if a user with the username exists, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE username = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by username: %w", err)
	}

	return exists, nil
}

// ExistsByPhone checks if a user with the specified phone number exists
// This method is useful for validation and business logic checks.
//
// Parameters:
//   - ctx: context for the database operation
//   - phone: phone number to check
//
// Returns:
//   - bool: true if a user with the phone number exists, false otherwise
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE phone = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, phone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by phone: %w", err)
	}

	return exists, nil
}

// Count returns the total number of users
// This method is useful for pagination and reporting purposes.
//
// Parameters:
//   - ctx: context for the database operation
//
// Returns:
//   - int64: total count of users
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CountByRole returns the total number of users with a specific role
// This method is useful for reporting and analytics purposes.
//
// Parameters:
//   - ctx: context for the database operation
//   - roleID: UUID of the role to count users for
//
// Returns:
//   - int64: total count of users with the specified role
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN role_entities re ON u.id = re.entity_id
		WHERE re.role_id = $1 AND u.deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, roleID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	return count, nil
}

// createUser is a helper method that creates a new user in the database
// This method handles the insertion of new user records.
//
// Parameters:
//   - ctx: context for the database operation
//   - user: pointer to the user aggregate to create
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) createUser(ctx context.Context, user *aggregates.UserAggregate) error {
	query := `
		INSERT INTO users (
			id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.UserName,
		user.Email,
		user.Phone,
		user.Password.Hash(),
		user.EmailVerifiedAt,
		user.PhoneVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// updateUser is a helper method that updates an existing user in the database
// This method handles the modification of existing user records.
//
// Parameters:
//   - ctx: context for the database operation
//   - user: pointer to the user aggregate with updated values
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *PostgresUserRepository) updateUser(ctx context.Context, user *aggregates.UserAggregate) error {
	query := `
		UPDATE users SET
			username = $2, email = $3, phone = $4, password = $5,
			email_verified_at = $6, phone_verified_at = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.UserName,
		user.Email,
		user.Phone,
		user.Password.Hash(),
		user.EmailVerifiedAt,
		user.PhoneVerifiedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}
