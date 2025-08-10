package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostgresMenuRepository struct {
	*BaseTransactionalRepository
	repo repository.MenuRepository
}

func NewPostgresMenuRepository(db *pgxpool.Pool, redisClient redis.Cmdable) repositories.MenuRepository {
	return &PostgresMenuRepository{
		BaseTransactionalRepository: NewBaseTransactionalRepository(db),
		repo:                        repository.NewMenuRepository(db, redisClient),
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

// Nested Set Tree Traversal methods
func (r *PostgresMenuRepository) GetDescendants(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	// Cast to the concrete repository type to access nested set methods
	if concreteRepo, ok := r.repo.(interface {
		GetDescendants(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetDescendants(ctx, menuID)
	}
	return nil, fmt.Errorf("GetDescendants method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetAncestors(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetAncestors(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetAncestors(ctx, menuID)
	}
	return nil, fmt.Errorf("GetAncestors method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetSiblings(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetSiblings(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetSiblings(ctx, menuID)
	}
	return nil, fmt.Errorf("GetSiblings method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetPath(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetPath(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetPath(ctx, menuID)
	}
	return nil, fmt.Errorf("GetPath method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetTree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetTree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetTree(ctx, menuID)
	}
	return nil, fmt.Errorf("GetTree method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetSubtree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetSubtree(ctx context.Context, menuID uuid.UUID) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetSubtree(ctx, menuID)
	}
	return nil, fmt.Errorf("GetSubtree method not available in underlying repository")
}

// Advanced Nested Set Operations
func (r *PostgresMenuRepository) AddChild(ctx context.Context, parentID, childID uuid.UUID) error {
	if concreteRepo, ok := r.repo.(interface {
		AddChild(ctx context.Context, parentID, childID uuid.UUID) error
	}); ok {
		return concreteRepo.AddChild(ctx, parentID, childID)
	}
	return fmt.Errorf("AddChild method not available in underlying repository")
}

func (r *PostgresMenuRepository) MoveSubtree(ctx context.Context, menuID, newParentID uuid.UUID) error {
	if concreteRepo, ok := r.repo.(interface {
		MoveSubtree(ctx context.Context, menuID, newParentID uuid.UUID) error
	}); ok {
		return concreteRepo.MoveSubtree(ctx, menuID, newParentID)
	}
	return fmt.Errorf("MoveSubtree method not available in underlying repository")
}

func (r *PostgresMenuRepository) DeleteSubtree(ctx context.Context, menuID uuid.UUID) error {
	if concreteRepo, ok := r.repo.(interface {
		DeleteSubtree(ctx context.Context, menuID uuid.UUID) error
	}); ok {
		return concreteRepo.DeleteSubtree(ctx, menuID)
	}
	return fmt.Errorf("DeleteSubtree method not available in underlying repository")
}

func (r *PostgresMenuRepository) InsertBetween(ctx context.Context, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error {
	if concreteRepo, ok := r.repo.(interface {
		InsertBetween(ctx context.Context, menu *entities.Menu, leftSiblingID, rightSiblingID *uuid.UUID) error
	}); ok {
		return concreteRepo.InsertBetween(ctx, menu, leftSiblingID, rightSiblingID)
	}
	return fmt.Errorf("InsertBetween method not available in underlying repository")
}

func (r *PostgresMenuRepository) SwapPositions(ctx context.Context, menu1ID, menu2ID uuid.UUID) error {
	if concreteRepo, ok := r.repo.(interface {
		SwapPositions(ctx context.Context, menu1ID, menu2ID uuid.UUID) error
	}); ok {
		return concreteRepo.SwapPositions(ctx, menu1ID, menu2ID)
	}
	return fmt.Errorf("SwapPositions method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetLeafNodes(ctx context.Context) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetLeafNodes(ctx context.Context) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetLeafNodes(ctx)
	}
	return nil, fmt.Errorf("GetLeafNodes method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetInternalNodes(ctx context.Context) ([]*entities.Menu, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetInternalNodes(ctx context.Context) ([]*entities.Menu, error)
	}); ok {
		return concreteRepo.GetInternalNodes(ctx)
	}
	return nil, fmt.Errorf("GetInternalNodes method not available in underlying repository")
}

// Batch Operations
func (r *PostgresMenuRepository) BatchMoveSubtrees(ctx context.Context, moves []struct {
	MenuID      uuid.UUID
	NewParentID uuid.UUID
}) error {
	if concreteRepo, ok := r.repo.(interface {
		BatchMoveSubtrees(ctx context.Context, moves []struct {
			MenuID      uuid.UUID
			NewParentID uuid.UUID
		}) error
	}); ok {
		return concreteRepo.BatchMoveSubtrees(ctx, moves)
	}
	return fmt.Errorf("BatchMoveSubtrees method not available in underlying repository")
}

func (r *PostgresMenuRepository) BatchInsertBetween(ctx context.Context, insertions []struct {
	Menu           *entities.Menu
	LeftSiblingID  *uuid.UUID
	RightSiblingID *uuid.UUID
}) error {
	if concreteRepo, ok := r.repo.(interface {
		BatchInsertBetween(ctx context.Context, insertions []struct {
			Menu           *entities.Menu
			LeftSiblingID  *uuid.UUID
			RightSiblingID *uuid.UUID
		}) error
	}); ok {
		return concreteRepo.BatchInsertBetween(ctx, insertions)
	}
	return fmt.Errorf("BatchInsertBetween method not available in underlying repository")
}

// Tree Maintenance and Optimization
func (r *PostgresMenuRepository) ValidateTree(ctx context.Context) ([]string, error) {
	if concreteRepo, ok := r.repo.(interface {
		ValidateTree(ctx context.Context) ([]string, error)
	}); ok {
		return concreteRepo.ValidateTree(ctx)
	}
	return nil, fmt.Errorf("ValidateTree method not available in underlying repository")
}

func (r *PostgresMenuRepository) RebuildTree(ctx context.Context) error {
	if concreteRepo, ok := r.repo.(interface {
		RebuildTree(ctx context.Context) error
	}); ok {
		return concreteRepo.RebuildTree(ctx)
	}
	return fmt.Errorf("RebuildTree method not available in underlying repository")
}

func (r *PostgresMenuRepository) OptimizeTree(ctx context.Context) error {
	if concreteRepo, ok := r.repo.(interface {
		OptimizeTree(ctx context.Context) error
	}); ok {
		return concreteRepo.OptimizeTree(ctx)
	}
	return fmt.Errorf("OptimizeTree method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetTreeStatistics(ctx context.Context) (map[string]interface{}, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetTreeStatistics(ctx context.Context) (map[string]interface{}, error)
	}); ok {
		return concreteRepo.GetTreeStatistics(ctx)
	}
	return nil, fmt.Errorf("GetTreeStatistics method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetTreeHeight(ctx context.Context) (uint64, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetTreeHeight(ctx context.Context) (uint64, error)
	}); ok {
		return concreteRepo.GetTreeHeight(ctx)
	}
	return 0, fmt.Errorf("GetTreeHeight method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetLevelWidth(ctx context.Context, level uint64) (int64, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetLevelWidth(ctx context.Context, level uint64) (int64, error)
	}); ok {
		return concreteRepo.GetLevelWidth(ctx, level)
	}
	return 0, fmt.Errorf("GetLevelWidth method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetSubtreeSize(ctx context.Context, menuID uuid.UUID) (int64, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetSubtreeSize(ctx context.Context, menuID uuid.UUID) (int64, error)
	}); ok {
		return concreteRepo.GetSubtreeSize(ctx, menuID)
	}
	return 0, fmt.Errorf("GetSubtreeSize method not available in underlying repository")
}

func (r *PostgresMenuRepository) GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	if concreteRepo, ok := r.repo.(interface {
		GetTreePerformanceMetrics(ctx context.Context) (map[string]interface{}, error)
	}); ok {
		return concreteRepo.GetTreePerformanceMetrics(ctx)
	}
	return nil, fmt.Errorf("GetTreePerformanceMetrics method not available in underlying repository")
}

func (r *PostgresMenuRepository) ValidateTreeIntegrity(ctx context.Context) ([]string, error) {
	if concreteRepo, ok := r.repo.(interface {
		ValidateTreeIntegrity(ctx context.Context) ([]string, error)
	}); ok {
		return concreteRepo.ValidateTreeIntegrity(ctx)
	}
	return nil, fmt.Errorf("ValidateTreeIntegrity method not available in underlying repository")
}
