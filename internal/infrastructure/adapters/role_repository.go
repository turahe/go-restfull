package adapters

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresRoleRepository struct {
	repo repository.RoleRepository
}

func NewPostgresRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.RoleRepository {
	return &PostgresRoleRepository{
		repo: repository.NewRoleRepository(db, redisClient),
	}
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	roleModel := &model.Role{
		ID:          role.ID.String(),
		Name:        role.Name,
		Slug:        role.Slug,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		CreatedBy:   "",
		UpdatedBy:   "",
	}

	return r.repo.Create(ctx, roleModel)
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	roleModel, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	roleID, _ := uuid.Parse(roleModel.ID)
	role := &entities.Role{
		ID:          roleID,
		Name:        roleModel.Name,
		Slug:        roleModel.Slug,
		Description: roleModel.Description,
		IsActive:    roleModel.IsActive,
		CreatedAt:   roleModel.CreatedAt,
		UpdatedAt:   roleModel.UpdatedAt,
	}

	return role, nil
}

func (r *PostgresRoleRepository) GetBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	roleModel, err := r.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	roleID, _ := uuid.Parse(roleModel.ID)
	role := &entities.Role{
		ID:          roleID,
		Name:        roleModel.Name,
		Slug:        roleModel.Slug,
		Description: roleModel.Description,
		IsActive:    roleModel.IsActive,
		CreatedAt:   roleModel.CreatedAt,
		UpdatedAt:   roleModel.UpdatedAt,
	}

	return role, nil
}

func (r *PostgresRoleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	roleModels, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var roles []*entities.Role
	for _, roleModel := range roleModels {
		roleID, _ := uuid.Parse(roleModel.ID)
		role := &entities.Role{
			ID:          roleID,
			Name:        roleModel.Name,
			Slug:        roleModel.Slug,
			Description: roleModel.Description,
			IsActive:    roleModel.IsActive,
			CreatedAt:   roleModel.CreatedAt,
			UpdatedAt:   roleModel.UpdatedAt,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *PostgresRoleRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	roleModels, err := r.repo.GetActive(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var roles []*entities.Role
	for _, roleModel := range roleModels {
		roleID, _ := uuid.Parse(roleModel.ID)
		role := &entities.Role{
			ID:          roleID,
			Name:        roleModel.Name,
			Slug:        roleModel.Slug,
			Description: roleModel.Description,
			IsActive:    roleModel.IsActive,
			CreatedAt:   roleModel.CreatedAt,
			UpdatedAt:   roleModel.UpdatedAt,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *PostgresRoleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	roleModels, err := r.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	var roles []*entities.Role
	for _, roleModel := range roleModels {
		roleID, _ := uuid.Parse(roleModel.ID)
		role := &entities.Role{
			ID:          roleID,
			Name:        roleModel.Name,
			Slug:        roleModel.Slug,
			Description: roleModel.Description,
			IsActive:    roleModel.IsActive,
			CreatedAt:   roleModel.CreatedAt,
			UpdatedAt:   roleModel.UpdatedAt,
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	roleModel := &model.Role{
		ID:          role.ID.String(),
		Name:        role.Name,
		Slug:        role.Slug,
		Description: role.Description,
		IsActive:    role.IsActive,
		UpdatedAt:   role.UpdatedAt,
		UpdatedBy:   "",
	}

	return r.repo.Update(ctx, roleModel)
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresRoleRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return r.repo.Activate(ctx, id)
}

func (r *PostgresRoleRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.repo.Deactivate(ctx, id)
}

func (r *PostgresRoleRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.repo.ExistsBySlug(ctx, slug)
}

func (r *PostgresRoleRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

func (r *PostgresRoleRepository) CountActive(ctx context.Context) (int64, error) {
	return r.repo.CountActive(ctx)
}
