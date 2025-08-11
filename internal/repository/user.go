package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/cache"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
)

// UserRepositoryImpl provides the PostgreSQL implementation of UserRepository.
// This struct handles all user-related database operations including CRUD operations,
// user authentication, role management, and Redis caching for performance.
type UserRepositoryImpl struct {
	pgxPool     *pgxpool.Pool // PostgreSQL connection pool for database operations
	redisClient redis.Cmdable // Redis client for caching operations
}

// NewUserRepository creates a new instance of UserRepositoryImpl
// This constructor function initializes the repository with the required dependencies
// including PostgreSQL connection pool and Redis client for caching.
//
// Parameters:
//   - pgxPool: PostgreSQL connection pool for database operations
//   - redisClient: Redis client for caching operations
//
// Returns:
//   - repositories.UserRepository: interface implementation for user management
func NewUserRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) repositories.UserRepository {
	return &UserRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

// Save persists a user aggregate to the database
// This method handles both insert and update operations based on whether the user exists.
// It first attempts to insert a new user, and if a unique constraint violation occurs,
// it updates the existing user record instead.
//
// Parameters:
//   - ctx: context for the database operation
//   - user: pointer to the user aggregate to save
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *UserRepositoryImpl) Save(ctx context.Context, user *aggregates.UserAggregate) error {
	// Try to insert first (for new users)
	insertQuery := `
		INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pgxPool.Exec(ctx, insertQuery,
		user.ID,
		user.UserName,
		user.Email.String(),
		user.Phone.String(),
		user.Password.Hash(),
		user.EmailVerifiedAt,
		user.PhoneVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Check if it's a unique constraint violation (user already exists)
		if isUniqueConstraintViolation(err) {
			// User exists, update it
			updateQuery := `
				UPDATE users 
				SET username = $2, email = $3, phone = $4, password = $5, 
					email_verified_at = $6, phone_verified_at = $7, updated_at = $8
				WHERE id = $1
			`

			_, err = r.pgxPool.Exec(ctx, updateQuery,
				user.ID,
				user.UserName,
				user.Email.String(),
				user.Phone.String(),
				user.Password.Hash(),
				user.EmailVerifiedAt,
				user.PhoneVerifiedAt,
				user.UpdatedAt,
			)
		}
	}

	if err == nil {
		// Invalidate user cache
		cache.InvalidatePattern(ctx, cache.PATTERN_USER_CACHE)
	}

	return err
}

// Delete performs a soft delete by setting the deleted_at timestamp
// This method marks a user as deleted without physically removing the record
// from the database. The record remains but is excluded from normal queries.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user to delete
//
// Returns:
//   - error: nil if successful, or database error if the operation fails
func (r *UserRepositoryImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	_, err := r.pgxPool.Exec(ctx, query, time.Now(), userID)
	if err == nil {
		// Invalidate user cache
		cache.InvalidatePattern(ctx, cache.PATTERN_USER_CACHE)
	}
	return err
}

// FindByID retrieves a user aggregate by its unique identifier.
// Returns nil and error if the user is not found or has been deleted.
// FindByID retrieves a user aggregate by its UUID from the database
// This method queries the database for a specific user and converts the result
// to a UserAggregate with all associated entities and value objects.
//
// Parameters:
//   - ctx: context for the database operation
//   - userID: UUID of the user to retrieve
//
// Returns:
//   - *aggregates.UserAggregate: pointer to the found user aggregate, or nil if not found
//   - error: nil if successful, or database error if the operation fails
func (r *UserRepositoryImpl) FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.pgxPool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Convert to aggregate
	return r.convertToAggregate(&user)
}

// FindByEmail retrieves a user aggregate by email address.
// Returns nil and error if the user is not found or has been deleted.
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.pgxPool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Convert to aggregate
	return r.convertToAggregate(&user)
}

// FindByUsername retrieves a user aggregate by username.
// Returns nil and error if the user is not found or has been deleted.
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE username = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.pgxPool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Convert to aggregate
	return r.convertToAggregate(&user)
}

// FindByPhone retrieves a user aggregate by phone number.
// Returns nil and error if the user is not found or has been deleted.
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE phone = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.pgxPool.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.PhoneVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Convert to aggregate
	return r.convertToAggregate(&user)
}

// FindAll retrieves a paginated list of user aggregates.
// Results are ordered by creation date (newest first).
func (r *UserRepositoryImpl) FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	// Base WHERE
	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1
	if query.Search != nil && *query.Search != "" {
		where += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, fmt.Sprintf("%%%s%%", *query.Search))
		argIdx++
	}

	// Count
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	var total int
	if err := r.pgxPool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Pagination
	offset := (query.Page - 1) * query.PageSize
	listSQL := fmt.Sprintf(`SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, query.PageSize, offset)

	rows, err := r.pgxPool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*aggregates.UserAggregate
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.ID, &u.UserName, &u.Email, &u.Phone, &u.Password, &u.EmailVerifiedAt, &u.PhoneVerifiedAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, err
		}
		agg, err := r.convertToAggregate(&u)
		if err != nil {
			return nil, err
		}
		items = append(items, agg)
	}

	totalPages := 0
	if query.PageSize > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}

	return &handlers.PaginatedResult[*aggregates.UserAggregate]{
		Items:      items,
		TotalCount: total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Search performs a search operation on users with pagination.
// Results are ordered by creation date (newest first).
func (r *UserRepositoryImpl) Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Build search query
	searchQuery := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users 
		WHERE deleted_at IS NULL
		  AND (username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	// Count total results for pagination
	countQuery := `
		SELECT COUNT(*)
		FROM users 
		WHERE deleted_at IS NULL
		  AND (username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1)
	`

	searchTerm := "%" + query.Query + "%"
	offset := (query.Page - 1) * query.PageSize

	// Get total count
	var total int64
	err := r.pgxPool.QueryRow(ctx, countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get paginated results
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*aggregates.UserAggregate
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.UserName,
			&user.Email,
			&user.Phone,
			&user.Password,
			&user.EmailVerifiedAt,
			&user.PhoneVerifiedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Convert to aggregate
		userAgg, err := r.convertToAggregate(&user)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user to aggregate: %w", err)
		}
		users = append(users, userAgg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	// Calculate pagination info
	totalPages := 0
	if query.PageSize > 0 {
		totalPages = (int(total) + query.PageSize - 1) / query.PageSize
	}

	return &handlers.PaginatedResult[*aggregates.UserAggregate]{
		Items:      users,
		TotalCount: int(total),
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ExistsByEmail checks if a user with the given email exists and is not deleted.
// Returns true if the user exists, false otherwise.
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

// ExistsByUsername checks if a user with the given username exists and is not deleted.
// Returns true if the user exists, false otherwise.
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

// ExistsByPhone checks if a user with the given phone number exists and is not deleted.
// Returns true if the user exists, false otherwise.
func (r *UserRepositoryImpl) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, phone).Scan(&exists)
	return exists, err
}

// Count returns the total number of non-deleted users in the database.
func (r *UserRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountByRole returns the number of users assigned to a specific role.
func (r *UserRepositoryImpl) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.pgxPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		WHERE ur.role_id = $1 AND u.deleted_at IS NULL`, roleID).Scan(&count)
	return count, err
}

// convertToAggregate converts an entities.User to aggregates.UserAggregate.
// This method handles the conversion of primitive types to value objects.
func (r *UserRepositoryImpl) convertToAggregate(user *entities.User) (*aggregates.UserAggregate, error) {
	// Create email value object
	email, err := valueobjects.NewEmail(user.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Create phone value object
	phone, err := valueobjects.NewPhone(user.Phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone: %w", err)
	}

	// Create password value object
	password, err := valueobjects.NewHashedPasswordFromHash(user.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid password hash: %w", err)
	}

	// Create user aggregate
	userAggregate := &aggregates.UserAggregate{
		ID:              user.ID,
		UserName:        user.UserName,
		Email:           email,
		Phone:           phone,
		Password:        password,
		EmailVerifiedAt: user.EmailVerifiedAt,
		PhoneVerifiedAt: user.PhoneVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		DeletedAt:       user.DeletedAt,
		Version:         1, // TODO: Implement versioning
	}

	return userAggregate, nil
}

// isUniqueConstraintViolation checks if the error is a unique constraint violation.
// This helper method is used to determine whether to insert or update a user.
func isUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check for PostgreSQL unique constraint violation
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key value violates unique constraint") ||
		strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "duplicate key")
}

// BeginTransaction starts a new transaction for this repository
func (r *UserRepositoryImpl) BeginTransaction(ctx context.Context) (repositories.Transaction, error) {
	// This is a placeholder implementation since the actual transaction management
	// is handled by the adapter layer through BaseTransactionalRepository
	// The concrete repository doesn't need to implement transaction logic
	return nil, fmt.Errorf("transactions should be handled through the adapter layer")
}

// WithTransaction executes repository operations within a transaction
func (r *UserRepositoryImpl) WithTransaction(ctx context.Context, fn func(repositories.Transaction) error) error {
	// This is a placeholder implementation since the actual transaction management
	// is handled by the adapter layer through BaseTransactionalRepository
	// The concrete repository doesn't need to implement transaction logic
	return fmt.Errorf("transactions should be handled through the adapter layer")
}
