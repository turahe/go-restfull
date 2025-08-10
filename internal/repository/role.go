package repository

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entities.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Role, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error)
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

type RoleRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewRoleRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) RoleRepository {
	return &RoleRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *RoleRepositoryImpl) Create(ctx context.Context, role *entities.Role) error {
	query := `INSERT INTO roles (id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), $6, $7)`

	_, err := r.pgxPool.Exec(ctx, query,
		role.ID, role.Name, role.Slug, role.Description, role.IsActive, role.CreatedBy, role.UpdatedBy)
	return err
}

func (r *RoleRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE id = $1 AND deleted_at IS NULL`

	var role entities.Role
	var createdBy, updatedBy *string

	err := r.pgxPool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE slug = $1 AND deleted_at IS NULL`

	var role entities.Role
	var createdBy, updatedBy *string

	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
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

// scanRoleRow is a helper function to scan a role row from database
func (r *RoleRepositoryImpl) scanRoleRow(rows pgx.Rows) (*entities.Role, error) {
	var role entities.Role
	var createdBy, updatedBy *string

	err := rows.Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &createdBy, &updatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepositoryImpl) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
			  FROM roles WHERE is_active = true AND deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
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

func (r *RoleRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	searchQuery := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by
					FROM roles WHERE deleted_at IS NULL AND
					(name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1)
					ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	searchTerm := "%" + query + "%"
	rows, err := r.pgxPool.Query(ctx, searchQuery, searchTerm, limit, offset)
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

func (r *RoleRepositoryImpl) Update(ctx context.Context, role *entities.Role) error {
	query := `UPDATE roles SET name = $1, slug = $2, description = $3, updated_at = NOW()
			  WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, role.Name, role.Slug, role.Description, role.ID)
	return err
}

func (r *RoleRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

func (r *RoleRepositoryImpl) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

func (r *RoleRepositoryImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.pgxPool.Exec(ctx, query, id)
	return err
}

func (r *RoleRepositoryImpl) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE slug = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.pgxPool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

func (r *RoleRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *RoleRepositoryImpl) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE is_active = true AND deleted_at IS NULL`

	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
