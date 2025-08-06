package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
	"github.com/turahe/go-restfull/internal/shared/errors"
)

// PostgreSQLUserRepository implements the UserRepository interface using PostgreSQL
type PostgreSQLUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgreSQLUserRepository creates a new PostgreSQL user repository
func NewPostgreSQLUserRepository(db *pgxpool.Pool) repositories.UserRepository {
	return &PostgreSQLUserRepository{
		db: db,
	}
}

// Save saves a user aggregate to the database
func (r *PostgreSQLUserRepository) Save(ctx context.Context, user *aggregates.UserAggregate) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if user exists
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		// Update existing user
		_, err = tx.Exec(ctx, `
			UPDATE users 
			SET username = $2, email = $3, phone = $4, password = $5, 
				email_verified_at = $6, phone_verified_at = $7, 
				updated_at = $8, deleted_at = $9, version = $10
			WHERE id = $1 AND version = $11`,
			user.ID, user.UserName, user.Email.String(), user.Phone.String(), user.Password.Hash(),
			user.EmailVerifiedAt, user.PhoneVerifiedAt, user.UpdatedAt, user.DeletedAt, 
			user.Version, user.Version-1)
		
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// Insert new user
		_, err = tx.Exec(ctx, `
			INSERT INTO users (id, username, email, phone, password, email_verified_at, 
				phone_verified_at, created_at, updated_at, deleted_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			user.ID, user.UserName, user.Email.String(), user.Phone.String(), user.Password.Hash(),
			user.EmailVerifiedAt, user.PhoneVerifiedAt, user.CreatedAt, user.UpdatedAt, 
			user.DeletedAt, user.Version)
		
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
	}

	// Save user profile if exists
	if user.Profile != nil {
		err = r.saveUserProfile(ctx, tx, user.ID, *user.Profile)
		if err != nil {
			return fmt.Errorf("failed to save user profile: %w", err)
		}
	}

	// Save user roles
	err = r.saveUserRoles(ctx, tx, user.ID, user.Roles)
	if err != nil {
		return fmt.Errorf("failed to save user roles: %w", err)
	}

	return tx.Commit(ctx)
}

// Delete deletes a user from the database
func (r *PostgreSQLUserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// FindByID finds a user by ID
func (r *PostgreSQLUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at, deleted_at, version
		FROM users WHERE id = $1 AND deleted_at IS NULL`, userID)
	
	return r.scanUserAggregate(ctx, row, userID)
}

// FindByEmail finds a user by email
func (r *PostgreSQLUserRepository) FindByEmail(ctx context.Context, email string) (*aggregates.UserAggregate, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at, deleted_at, version
		FROM users WHERE email = $1 AND deleted_at IS NULL`, email)
	
	var userID uuid.UUID
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound(uuid.Nil)
		}
		return nil, fmt.Errorf("failed to scan user ID: %w", err)
	}
	
	return r.scanUserAggregate(ctx, row, userID)
}

// FindByUsername finds a user by username
func (r *PostgreSQLUserRepository) FindByUsername(ctx context.Context, username string) (*aggregates.UserAggregate, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at, deleted_at, version
		FROM users WHERE username = $1 AND deleted_at IS NULL`, username)
	
	var userID uuid.UUID
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound(uuid.Nil)
		}
		return nil, fmt.Errorf("failed to scan user ID: %w", err)
	}
	
	return r.scanUserAggregate(ctx, row, userID)
}

// FindByPhone finds a user by phone
func (r *PostgreSQLUserRepository) FindByPhone(ctx context.Context, phone string) (*aggregates.UserAggregate, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at, deleted_at, version
		FROM users WHERE phone = $1 AND deleted_at IS NULL`, phone)
	
	var userID uuid.UUID
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound(uuid.Nil)
		}
		return nil, fmt.Errorf("failed to scan user ID: %w", err)
	}
	
	return r.scanUserAggregate(ctx, row, userID)
}

// FindAll finds all users with pagination and filters
func (r *PostgreSQLUserRepository) FindAll(ctx context.Context, query queries.ListUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*query.Search+"%")
		argIndex++
	}

	if query.RoleID != nil {
		conditions = append(conditions, fmt.Sprintf("id IN (SELECT user_id FROM user_roles WHERE role_id = $%d)", argIndex))
		args = append(args, *query.RoleID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if query.SortBy != nil && *query.SortBy != "" {
		direction := "ASC"
		if query.SortDir != nil && *query.SortDir == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", *query.SortBy, direction)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Calculate pagination
	offset := (query.Page - 1) * query.PageSize
	totalPages := (totalCount + query.PageSize - 1) / query.PageSize

	// Fetch users
	fetchQuery := fmt.Sprintf(`
		SELECT id, username, email, phone, password, email_verified_at, phone_verified_at,
			created_at, updated_at, deleted_at, version
		FROM users %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)
	
	args = append(args, query.PageSize, offset)
	
	rows, err := r.db.Query(ctx, fetchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*aggregates.UserAggregate
	for rows.Next() {
		user, err := r.scanUserAggregateFromRow(ctx, rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return &handlers.PaginatedResult[*aggregates.UserAggregate]{
		Items:      users,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Search searches users
func (r *PostgreSQLUserRepository) Search(ctx context.Context, query queries.SearchUsersQuery) (*handlers.PaginatedResult[*aggregates.UserAggregate], error) {
	// Convert to ListUsersQuery
	listQuery := queries.ListUsersQuery{
		Page:     query.Page,
		PageSize: query.PageSize,
		Search:   &query.Query,
	}
	
	return r.FindAll(ctx, listQuery)
}

// ExistsByEmail checks if a user exists by email
func (r *PostgreSQLUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)", email).Scan(&exists)
	return exists, err
}

// ExistsByUsername checks if a user exists by username
func (r *PostgreSQLUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)", username).Scan(&exists)
	return exists, err
}

// ExistsByPhone checks if a user exists by phone
func (r *PostgreSQLUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND deleted_at IS NULL)", phone).Scan(&exists)
	return exists, err
}

// Count returns the total number of users
func (r *PostgreSQLUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count)
	return count, err
}

// CountByRole returns the number of users with a specific role
func (r *PostgreSQLUserRepository) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		WHERE ur.role_id = $1 AND u.deleted_at IS NULL`, roleID).Scan(&count)
	return count, err
}

// Helper methods

func (r *PostgreSQLUserRepository) scanUserAggregate(ctx context.Context, row interface{}, userID uuid.UUID) (*aggregates.UserAggregate, error) {
	// This is a simplified implementation - in reality you'd need proper scanning logic
	// and loading of related data (profile, roles)
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLUserRepository) scanUserAggregateFromRow(ctx context.Context, row interface{}) (*aggregates.UserAggregate, error) {
	// This is a simplified implementation - in reality you'd need proper scanning logic
	// and loading of related data (profile, roles)
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLUserRepository) saveUserProfile(ctx context.Context, tx interface{}, userID uuid.UUID, profile valueobjects.UserProfile) error {
	// Implementation for saving user profile
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLUserRepository) saveUserRoles(ctx context.Context, tx interface{}, userID uuid.UUID, roles []valueobjects.Role) error {
	// Implementation for saving user roles
	return fmt.Errorf("not implemented")
}