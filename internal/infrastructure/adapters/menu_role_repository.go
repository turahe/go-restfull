package adapters

import (
	"context"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresMenuRoleRepository struct {
	repo repository.MenuRoleRepository
}

func NewPostgresMenuRoleRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuRoleRepository {
	return &PostgresMenuRoleRepository{
		repo: repository.NewMenuRoleRepository(db, redisClient),
	}
}

func (r *PostgresMenuRoleRepository) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	return r.repo.AssignRoleToMenu(ctx, menuID, roleID)
}

func (r *PostgresMenuRoleRepository) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	return r.repo.RemoveRoleFromMenu(ctx, menuID, roleID)
}

func (r *PostgresMenuRoleRepository) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
	// The repository should return entities directly, not models
	// If the repository returns models, we need to convert them
	// For now, assuming the repository returns entities
	return r.repo.GetMenuRoles(ctx, menuID)
}

func (r *PostgresMenuRoleRepository) GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error) {
	// The repository method doesn't take limit and offset parameters
	// We need to get all menus and then apply pagination
	allMenus, err := r.repo.GetRoleMenus(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Apply pagination manually
	start := offset
	end := offset + limit
	if start >= len(allMenus) {
		return []*entities.Menu{}, nil
	}
	if end > len(allMenus) {
		end = len(allMenus)
	}

	return allMenus[start:end], nil
}

func (r *PostgresMenuRoleRepository) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	return r.repo.HasRole(ctx, menuID, roleID)
}

func (r *PostgresMenuRoleRepository) HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	// This method is not available in the repository interface
	// We need to implement it by checking each role
	for _, roleID := range roleIDs {
		hasRole, err := r.repo.HasRole(ctx, menuID, roleID)
		if err != nil {
			return false, err
		}
		if hasRole {
			return true, nil
		}
	}
	return false, nil
}

func (r *PostgresMenuRoleRepository) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	return r.repo.GetMenuRoleIDs(ctx, menuID)
}

func (r *PostgresMenuRoleRepository) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return r.repo.CountMenusByRole(ctx, roleID)
}
