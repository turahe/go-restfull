package repository

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error)
	HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
	HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error)
	GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	CountUsersByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
}

type UserRoleRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewUserRoleRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) UserRoleRepository {
	return &UserRoleRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *UserRoleRepositoryImpl) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (id, user_id, role_id, created_at, updated_at)
			  VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
			  ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := r.pgxPool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *UserRoleRepositoryImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	_, err := r.pgxPool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *UserRoleRepositoryImpl) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	query := `SELECT r.id, r.name, r.slug, r.description, r.is_active, r.created_at, r.updated_at, r.created_by, r.updated_by
			  FROM roles r
			  INNER JOIN user_roles ur ON r.id = ur.role_id
			  WHERE ur.user_id = $1 AND r.deleted_at IS NULL AND r.is_active = true
			  ORDER BY r.created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *UserRoleRepositoryImpl) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	query := `SELECT u.id, u.username, u.email, u.phone, u.password, u.email_verified_at, u.phone_verified_at, 
			  u.created_at, u.updated_at, u.deleted_at, u.created_by, u.updated_by, u.deleted_by
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  WHERE ur.role_id = $1 AND u.deleted_at IS NULL
			  ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pgxPool.Query(ctx, query, roleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user, err := r.scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// scanRoleRow is a helper function to scan a role row from database
func (r *UserRoleRepositoryImpl) scanRoleRow(rows pgx.Rows) (*entities.Role, error) {
	var role entities.Role
	var createdBy, updatedBy string

	err := rows.Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// scanUserRow is a helper function to scan a user row from database
func (r *UserRoleRepositoryImpl) scanUserRow(rows pgx.Rows) (*entities.User, error) {
	var user entities.User
	var createdBy, updatedBy, deletedBy string

	err := rows.Scan(
		&user.ID, &user.UserName, &user.Email, &user.Phone, &user.Password,
		&user.EmailVerifiedAt, &user.PhoneVerifiedAt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		&createdBy, &updatedBy, &deletedBy,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRoleRepositoryImpl) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				INNER JOIN roles r ON ur.role_id = r.id
				WHERE ur.user_id = $1 AND ur.role_id = $2 AND r.deleted_at IS NULL AND r.is_active = true
			  )`

	var hasRole bool
	err := r.pgxPool.QueryRow(ctx, query, userID, roleID).Scan(&hasRole)
	return hasRole, err
}

func (r *UserRoleRepositoryImpl) HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}

	query := `SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				INNER JOIN roles r ON ur.role_id = r.id
				WHERE ur.user_id = $1 AND ur.role_id = ANY($2) AND r.deleted_at IS NULL AND r.is_active = true
			  )`

	var hasAnyRole bool
	err := r.pgxPool.QueryRow(ctx, query, userID, roleIDs).Scan(&hasAnyRole)
	return hasAnyRole, err
}

func (r *UserRoleRepositoryImpl) GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT ur.role_id
			  FROM user_roles ur
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE ur.user_id = $1 AND r.deleted_at IS NULL AND r.is_active = true
			  ORDER BY r.created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []uuid.UUID
	for rows.Next() {
		var roleID uuid.UUID
		err := rows.Scan(&roleID)
		if err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}

func (r *UserRoleRepositoryImpl) CountUsersByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*)
			  FROM user_roles ur
			  INNER JOIN users u ON ur.user_id = u.id
			  WHERE ur.role_id = $1 AND u.deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query, roleID).Scan(&count)
	return count, err
}
