package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/helper/cache"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// postgresUserRepository implements UserRepository interface
type postgresUserRepository struct {
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.UserRepository {
	return &postgresUserRepository{
		db:          db,
		redisClient: redisClient,
	}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.UserName,
		user.Email,
		user.Phone,
		user.Password,
		user.EmailVerifiedAt,
		user.PhoneVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err == nil {
		// Invalidate user cache
		cache.InvalidatePattern(ctx, cache.PATTERN_USER_CACHE)
	}

	return err
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User

	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_USER_BY_ID, id.String())
	err := cache.GetJSON(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`

	err = r.db.QueryRow(ctx, query, id).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &user, cache.DefaultCacheDuration)

	return &user, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User

	// Try to get from cache first
	cacheKey := fmt.Sprintf(cache.KEY_USER_BY_EMAIL, email)
	err := cache.GetJSON(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`

	err = r.db.QueryRow(ctx, query, email).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// Cache the result
	cache.SetJSON(ctx, cacheKey, &user, cache.DefaultCacheDuration)

	return &user, nil
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE username = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.db.QueryRow(ctx, query, username).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepository) GetByPhone(ctx context.Context, phone string) (*entities.User, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE phone = $1 AND deleted_at IS NULL
	`

	var user entities.User
	err := r.db.QueryRow(ctx, query, phone).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	query := `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at, deleted_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
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
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// Old Search method removed - replaced with aggregate-based interface

func (r *postgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, phone = $4, password = $5, 
		    email_verified_at = $6, phone_verified_at = $7, updated_at = $8, deleted_at = $9
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.UserName,
		user.Email,
		user.Phone,
		user.Password,
		user.EmailVerifiedAt,
		user.PhoneVerifiedAt,
		user.UpdatedAt,
		user.DeletedAt,
	)

	if err == nil {
		// Invalidate user cache
		cache.InvalidatePattern(ctx, cache.PATTERN_USER_CACHE)
	}

	return err
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err == nil {
		// Invalidate user cache
		cache.InvalidatePattern(ctx, cache.PATTERN_USER_CACHE)
	}
	return err
}

func (r *postgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, phone).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *postgresUserRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
	searchQuery := `
		SELECT COUNT(*) FROM users 
		WHERE deleted_at IS NULL 
		AND (username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1)
	`
	var count int64
	err := r.db.QueryRow(ctx, searchQuery, fmt.Sprintf("%%%s%%", query)).Scan(&count)
	return count, err
}

// Stub methods to satisfy the interface - TODO: Implement properly
func (r *postgresUserRepository) FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	return nil, fmt.Errorf("FindAll not implemented")
}

func (r *postgresUserRepository) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		WHERE ur.role_id = $1 AND u.deleted_at IS NULL`, roleID).Scan(&count)
	return count, err
}

// Stub methods for aggregate-based interface - TODO: Implement properly
func (r *postgresUserRepository) Save(ctx context.Context, user *aggregates.UserAggregate) error {
	return fmt.Errorf("Save not implemented")
}

func (r *postgresUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	return nil, fmt.Errorf("FindByID not implemented")
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	return nil, fmt.Errorf("FindByEmail not implemented")
}

func (r *postgresUserRepository) FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error) {
	return nil, fmt.Errorf("FindByUsername not implemented")
}

func (r *postgresUserRepository) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	return nil, fmt.Errorf("FindByPhone not implemented")
}

func (r *postgresUserRepository) Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	return nil, fmt.Errorf("Search not implemented")
}
