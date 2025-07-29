package adapters

import (
	"context"
	"strings"
	"time"
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
	return r.repo.Create(ctx, menu)
}

func (r *PostgresMenuRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *PostgresMenuRepository) GetBySlug(ctx context.Context, slug string) (*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allMenus, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to find by slug
	if err != nil {
		return nil, err
	}

	for _, menu := range allMenus {
		if menu.Slug == slug {
			return menu, nil
		}
	}

	return nil, nil // Not found
}

func (r *PostgresMenuRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return r.repo.GetAll(ctx, limit, offset)
}

func (r *PostgresMenuRepository) GetActive(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allMenus, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var activeMenus []*entities.Menu
	for _, menu := range allMenus {
		if menu.IsActive {
			activeMenus = append(activeMenus, menu)
		}
	}

	return activeMenus, nil
}

func (r *PostgresMenuRepository) GetVisible(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering the results
	allMenus, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var visibleMenus []*entities.Menu
	for _, menu := range allMenus {
		if menu.IsVisible {
			visibleMenus = append(visibleMenus, menu)
		}
	}

	return visibleMenus, nil
}

func (r *PostgresMenuRepository) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	return r.repo.GetRootMenus(ctx)
}

func (r *PostgresMenuRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entities.Menu, error) {
	return r.repo.GetByParentID(ctx, parentID)
}

func (r *PostgresMenuRepository) GetHierarchy(ctx context.Context) ([]*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by building the hierarchy
	allMenus, err := r.repo.GetAll(ctx, 1000, 0) // Get all menus to build hierarchy
	if err != nil {
		return nil, err
	}

	// Build hierarchy by organizing menus into a tree structure
	menuMap := make(map[uuid.UUID]*entities.Menu)
	var rootMenus []*entities.Menu

	// First pass: create a map of all menus
	for _, menu := range allMenus {
		menuMap[menu.ID] = menu
		menu.Children = []*entities.Menu{} // Initialize children slice
	}

	// Second pass: build the hierarchy
	for _, menu := range allMenus {
		if menu.ParentID == nil {
			rootMenus = append(rootMenus, menu)
		} else {
			if parent, exists := menuMap[*menu.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return rootMenus, nil
}

func (r *PostgresMenuRepository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by filtering based on user roles
	// For now, return all visible menus
	return r.GetVisible(ctx, 1000, 0)
}

func (r *PostgresMenuRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	// This method is not available in the repository interface
	// We need to implement it by searching through all menus
	allMenus, err := r.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var searchResults []*entities.Menu
	queryLower := strings.ToLower(query)
	for _, menu := range allMenus {
		if strings.Contains(strings.ToLower(menu.Name), queryLower) ||
			strings.Contains(strings.ToLower(menu.Description), queryLower) ||
			strings.Contains(strings.ToLower(menu.Slug), queryLower) {
			searchResults = append(searchResults, menu)
		}
	}

	return searchResults, nil
}

func (r *PostgresMenuRepository) Update(ctx context.Context, menu *entities.Menu) error {
	return r.repo.Update(ctx, menu)
}

func (r *PostgresMenuRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

func (r *PostgresMenuRepository) Activate(ctx context.Context, id uuid.UUID) error {
	// This method is not available in the repository interface
	// We need to implement it by getting the menu and updating it
	menu, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	menu.IsActive = true
	menu.UpdatedAt = time.Now()
	return r.repo.Update(ctx, menu)
}

func (r *PostgresMenuRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	// This method is not available in the repository interface
	// We need to implement it by getting the menu and updating it
	menu, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	menu.IsActive = false
	menu.UpdatedAt = time.Now()
	return r.repo.Update(ctx, menu)
}

func (r *PostgresMenuRepository) Show(ctx context.Context, id uuid.UUID) error {
	// This method is not available in the repository interface
	// We need to implement it by getting the menu and updating it
	menu, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	menu.IsVisible = true
	menu.UpdatedAt = time.Now()
	return r.repo.Update(ctx, menu)
}

func (r *PostgresMenuRepository) Hide(ctx context.Context, id uuid.UUID) error {
	// This method is not available in the repository interface
	// We need to implement it by getting the menu and updating it
	menu, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	menu.IsVisible = false
	menu.UpdatedAt = time.Now()
	return r.repo.Update(ctx, menu)
}

func (r *PostgresMenuRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.repo.ExistsBySlug(ctx, slug)
}

func (r *PostgresMenuRepository) Count(ctx context.Context) (int64, error) {
	return r.repo.Count(ctx)
}

func (r *PostgresMenuRepository) CountActive(ctx context.Context) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting filtered results
	allMenus, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to count
	if err != nil {
		return 0, err
	}

	var count int64
	for _, menu := range allMenus {
		if menu.IsActive {
			count++
		}
	}

	return count, nil
}

func (r *PostgresMenuRepository) CountVisible(ctx context.Context) (int64, error) {
	// This method is not available in the repository interface
	// We need to implement it by counting filtered results
	allMenus, err := r.repo.GetAll(ctx, 1000, 0) // Get a large number to count
	if err != nil {
		return 0, err
	}

	var count int64
	for _, menu := range allMenus {
		if menu.IsVisible {
			count++
		}
	}

	return count, nil
}
