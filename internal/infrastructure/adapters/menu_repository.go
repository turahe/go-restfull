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

type PostgresMenuRepository struct {
	repo repository.MenuRepository
}

func NewPostgresMenuRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuRepository {
	return &PostgresMenuRepository{
		repo: repository.NewMenuRepository(db, redisClient),
	}
}

func (r *PostgresMenuRepository) Create(ctx context.Context, menu *entities.Menu) error {
	menuModel := &model.Menu{
		ID:             menu.ID.String(),
		Name:           menu.Name,
		Slug:           menu.Slug,
		Description:    menu.Description,
		URL:            menu.URL,
		Icon:           menu.Icon,
		ParentID:       nil, // Will be set below
		RecordLeft:     menu.RecordLeft,
		RecordRight:    menu.RecordRight,
		RecordOrdering: menu.RecordOrdering,
		IsActive:       menu.IsActive,
		IsVisible:      menu.IsVisible,
		Target:         menu.Target,
		CreatedAt:      menu.CreatedAt,
		UpdatedAt:      menu.UpdatedAt,
		CreatedBy:      "",
		UpdatedBy:      "",
	}

	// Handle parent ID
	if menu.ParentID != nil {
		parentIDStr := menu.ParentID.String()
		menuModel.ParentID = &parentIDStr
	}

	// Handle deleted at
	if menu.DeletedAt != nil {
		menuModel.DeletedAt = menu.DeletedAt
	}

	return r.repo.Create(ctx, menuModel)
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	menuModel, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertModelToEntity(menuModel), nil
}

func (r *PostgresMenuRepository) GetBySlug(ctx context.Context, slug string) (*entities.Menu, error) {
	menuModel, err := r.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return r.convertModelToEntity(menuModel), nil
}

func (r *PostgresMenuRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetActive(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetVisible(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetVisible(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetRootMenus(ctx)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetHierarchy(ctx context.Context) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error) {
	menuModels, err := r.repo.GetUserMenus(ctx, userID)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	menuModels, err := r.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return r.convertModelsToEntities(menuModels), nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, menu *entities.Menu) error {
	menuModel := &model.Menu{
		ID:             menu.ID.String(),
		Name:           menu.Name,
		Slug:           menu.Slug,
		Description:    menu.Description,
		URL:            menu.URL,
		Icon:           menu.Icon,
		ParentID:       nil, // Will be set below
		RecordLeft:     menu.RecordLeft,
		RecordRight:    menu.RecordRight,
		RecordOrdering: menu.RecordOrdering,
		IsActive:       menu.IsActive,
		IsVisible:      menu.IsVisible,
		Target:         menu.Target,
		CreatedAt:      menu.CreatedAt,
		UpdatedAt:      menu.UpdatedAt,
		CreatedBy:      "",
		UpdatedBy:      "",
	}

	// Handle parent ID
	if menu.ParentID != nil {
		parentIDStr := menu.ParentID.String()
		menuModel.ParentID = &parentIDStr
	}

	// Handle deleted at
	if menu.DeletedAt != nil {
		menuModel.DeletedAt = menu.DeletedAt
	}

	return r.repo.Update(ctx, menuModel)
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresMenuRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return r.repo.Activate(ctx, id)
}

func (r *PostgresMenuRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.repo.Deactivate(ctx, id)
}

func (r *PostgresMenuRepository) Show(ctx context.Context, id uuid.UUID) error {
	return r.repo.Show(ctx, id)
}

func (r *PostgresMenuRepository) Hide(ctx context.Context, id uuid.UUID) error {
	return r.repo.Hide(ctx, id)
}

func (r *PostgresMenuRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.repo.ExistsBySlug(ctx, slug)
}

func (r *PostgresMenuRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

func (r *PostgresMenuRepository) CountActive(ctx context.Context) (int64, error) {
	return r.repo.CountActive(ctx)
}

func (r *PostgresMenuRepository) CountVisible(ctx context.Context) (int64, error) {
	return r.repo.CountVisible(ctx)
}

// convertModelToEntity converts a menu model to entity
func (r *PostgresMenuRepository) convertModelToEntity(menuModel *model.Menu) *entities.Menu {
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

	return menu
}

// convertModelsToEntities converts menu models to entities
func (r *PostgresMenuRepository) convertModelsToEntities(menuModels []*model.Menu) []*entities.Menu {
	var result []*entities.Menu

	for _, menuModel := range menuModels {
		result = append(result, r.convertModelToEntity(menuModel))
	}

	return result
}
