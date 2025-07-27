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
	roleModels, err := r.repo.GetMenuRoles(ctx, menuID)
	if err != nil {
		return nil, err
	}

	return r.convertRoleModelsToEntities(roleModels), nil
}

func (r *PostgresMenuRoleRepository) GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetRoleMenus(ctx, roleID, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertMenuModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRoleRepository) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	return r.repo.HasRole(ctx, menuID, roleID)
}

func (r *PostgresMenuRoleRepository) HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	return r.repo.HasAnyRole(ctx, menuID, roleIDs)
}

func (r *PostgresMenuRoleRepository) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	return r.repo.GetMenuRoleIDs(ctx, menuID)
}

func (r *PostgresMenuRoleRepository) CountMenusByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return r.repo.CountMenusByRole(ctx, roleID)
}

// convertRoleModelsToEntities converts role models to entities
func (r *PostgresMenuRoleRepository) convertRoleModelsToEntities(roleModels []*model.Role) []*entities.Role {
	var result []*entities.Role

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

		// Handle deleted at
		if roleModel.DeletedAt != nil {
			role.DeletedAt = roleModel.DeletedAt
		}

		result = append(result, role)
	}

	return result
}

// convertMenuModelsToEntities converts menu models to entities
func (r *PostgresMenuRoleRepository) convertMenuModelsToEntities(menuModels []*model.Menu) []*entities.Menu {
	var result []*entities.Menu

	for _, menuModel := range menuModels {
		menuID, _ := uuid.Parse(menuModel.ID)

		menu := &entities.Menu{
			ID:             menuID,
			Name:           menuModel.Name,
			Slug:           menuModel.Slug,
			Description:    menuModel.Description,
			URL:            menuModel.URL,
			Icon:           menuModel.Icon,
			ParentID:       nil, // Will be set below
			RecordLeft:     menuModel.RecordLeft,
			RecordRight:    menuModel.RecordRight,
			RecordOrdering: menuModel.RecordOrdering,
			IsActive:       menuModel.IsActive,
			IsVisible:      menuModel.IsVisible,
			Target:         menuModel.Target,
			CreatedAt:      menuModel.CreatedAt,
			UpdatedAt:      menuModel.UpdatedAt,
			Children:       []*entities.Menu{},
			Roles:          []*entities.Role{},
		}

		// Handle parent ID
		if menuModel.ParentID != nil {
			if parentID, err := uuid.Parse(*menuModel.ParentID); err == nil {
				menu.ParentID = &parentID
			}
		}

		// Handle deleted at
		if menuModel.DeletedAt != nil {
			menu.DeletedAt = menuModel.DeletedAt
		}

		result = append(result, menu)
	}

	return result
}
