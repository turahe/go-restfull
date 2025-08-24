package adapters

import (
	"context"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresRoleRepository struct {
	*BaseTransactionalRepository
	db          *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewPostgresRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.RoleRepository {
	return &PostgresRoleRepository{BaseTransactionalRepository: NewBaseTransactionalRepository(db), db: db, redisClient: redisClient}
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	query := `INSERT INTO roles (id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,NOW(),NOW(),$6,$7)`
	_, err := r.db.Exec(ctx, query, role.ID, role.Name, role.Slug, role.Description, role.IsActive, role.CreatedBy, role.UpdatedBy)
	return err
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE id = $1 AND deleted_at IS NULL`
	var role entities.Role
	if err := r.db.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE slug = $1 AND deleted_at IS NULL`
	var role entities.Role
	if err := r.db.QueryRow(ctx, query, slug).Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *PostgresRoleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	query := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE is_active = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	q := `SELECT id, name, slug, description, is_active, created_at, updated_at, created_by, updated_by FROM roles WHERE deleted_at IS NULL AND (name ILIKE $1 OR slug ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy, &role.UpdatedBy); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	query := `UPDATE roles SET name=$1, slug=$2, description=$3, updated_at=NOW() WHERE id=$4 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, role.Name, role.Slug, role.Description, role.ID)
	return err
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresRoleRepository) Activate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = true, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresRoleRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET is_active = false, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresRoleRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresRoleRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresRoleRepository) CountActive(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM roles WHERE is_active = true AND deleted_at IS NULL`
	var count int64
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
