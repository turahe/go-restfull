// Package services provides application-level business logic for menu management.
// This package contains the menu service implementation that handles menu creation,
// hierarchy management, visibility control, and user-specific menu access while
// ensuring proper navigation structure.
package services

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// MenuService implements the MenuService interface and provides comprehensive
// menu management functionality. It handles menu creation, hierarchy management,
// visibility control, user-specific access, and navigation structure while
// enforcing business rules and validation.
type MenuService struct {
	menuRepository repositories.MenuRepository
}

// NewMenuService creates a new menu service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - menuRepository: Repository interface for menu data access operations
//
// Returns:
//   - ports.MenuService: The menu service interface implementation
func NewMenuService(menuRepository repositories.MenuRepository) ports.MenuService {
	return &MenuService{
		menuRepository: menuRepository,
	}
}

// CreateMenu creates a new menu item with comprehensive validation and hierarchy support.
// This method enforces business rules for menu creation and supports hierarchical
// menu structures with proper ordering and slug uniqueness.
//
// Business Rules:
//   - Menu name and slug are required and validated
//   - Slug must be unique across the system
//   - Record ordering must be non-negative
//   - Parent ID is optional for root-level menus
//   - Menu validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - name: Display name of the menu item
//   - slug: Unique identifier for the menu item
//   - description: Optional description of the menu item
//   - url: URL or route for the menu item
//   - icon: Icon identifier for the menu item
//   - recordOrdering: Ordering position for menu display
//   - parentID: Optional parent menu ID for hierarchical structure
//
// Returns:
//   - *entities.Menu: The created menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) CreateMenu(ctx context.Context, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error) {
	// Validate menu name to ensure it's provided and meaningful
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("menu name is required")
	}

	// Validate menu slug to ensure it's provided and meaningful
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("menu slug is required")
	}

	// Validate record ordering to ensure it's non-negative
	if recordOrdering < 0 {
		return nil, errors.New("menu record ordering must be non-negative")
	}

	// Check if slug already exists to maintain uniqueness
	exists, err := s.menuRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("menu slug already exists")
	}

	// Create menu entity with the provided parameters
	menu := entities.NewMenu(name, slug, description, url, icon, parentID)
	// apply ordering if provided
	ro := uint64(recordOrdering)
	menu.RecordOrdering = &ro

	// Validate the menu entity to ensure proper structure
	if err := menu.Validate(); err != nil {
		return nil, err
	}

	// Persist the menu to the repository
	err = s.menuRepository.Create(ctx, menu)
	if err != nil {
		return nil, err
	}

	return menu, nil
}

// GetMenuByID retrieves a menu item by its unique identifier.
// This method provides access to individual menu details and metadata.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to retrieve
//
// Returns:
//   - *entities.Menu: The menu entity if found
//   - error: Error if menu not found or other issues occur
func (s *MenuService) GetMenuByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	return s.menuRepository.GetByID(ctx, id)
}

// GetMenuBySlug retrieves a menu item by its unique slug identifier.
// This method is useful for URL-based menu lookups and routing.
//
// Parameters:
//   - ctx: Context for the operation
//   - slug: Slug identifier of the menu to retrieve
//
// Returns:
//   - *entities.Menu: The menu entity if found
//   - error: Error if menu not found or other issues occur
func (s *MenuService) GetMenuBySlug(ctx context.Context, slug string) (*entities.Menu, error) {
	return s.menuRepository.GetBySlug(ctx, slug)
}

// GetAllMenus retrieves all menu items in the system with pagination.
// This method is useful for administrative purposes and menu management.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of menus to return
//   - offset: Number of menus to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of all menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetAllMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetAll(ctx, limit, offset)
}

// GetActiveMenus retrieves only active menu items with pagination.
// This method is useful for displaying menus that are currently enabled.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of active menus to return
//   - offset: Number of active menus to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of active menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetActiveMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetActive(ctx, limit, offset)
}

// GetVisibleMenus retrieves only visible menu items with pagination.
// This method is useful for displaying menus that are both active and visible.
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of visible menus to return
//   - offset: Number of visible menus to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of visible menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetVisibleMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetVisible(ctx, limit, offset)
}

// GetRootMenus retrieves all root-level menu items (no parent).
// This method is useful for building top-level navigation structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Menu: List of root-level menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	return s.menuRepository.GetRootMenus(ctx)
}

// GetMenuHierarchy retrieves the complete menu hierarchy with parent-child relationships.
// This method is useful for building complete navigation trees and menu structures.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Menu: Complete menu hierarchy with relationships
//   - error: Any error that occurred during the operation
func (s *MenuService) GetMenuHierarchy(ctx context.Context) ([]*entities.Menu, error) {
	return s.menuRepository.GetHierarchy(ctx)
}

// GetUserMenus retrieves menu items accessible to a specific user.
// This method supports role-based menu access and user-specific navigation.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get menus for
//
// Returns:
//   - []*entities.Menu: List of menus accessible to the user
//   - error: Any error that occurred during the operation
func (s *MenuService) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error) {
	return s.menuRepository.GetUserMenus(ctx, userID)
}

// SearchMenus searches for menu items based on a query string.
// This method supports full-text search capabilities for finding menus
// by name, description, or other attributes.
//
// Business Rules:
//   - Search query must be provided and validated
//   - Empty queries are not allowed
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string
//   - limit: Maximum number of search results to return
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of matching menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) SearchMenus(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	// Validate search query to ensure meaningful search
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query is required")
	}

	return s.menuRepository.Search(ctx, query, limit, offset)
}

// UpdateMenu updates an existing menu item with new information.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Menu must exist and be accessible
//   - Menu name and slug are required and validated
//   - Slug must be unique (excluding current menu)
//   - Record ordering must be non-negative
//   - Menu validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to update
//   - name: Updated display name of the menu item
//   - slug: Updated unique identifier for the menu item
//   - description: Updated description of the menu item
//   - url: Updated URL or route for the menu item
//   - icon: Updated icon identifier for the menu item
//   - recordOrdering: Updated ordering position for menu display
//   - parentID: Updated parent menu ID for hierarchical structure
//
// Returns:
//   - *entities.Menu: The updated menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) UpdateMenu(ctx context.Context, id uuid.UUID, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error) {
	// Retrieve existing menu to ensure it exists and is accessible
	existingMenu, err := s.menuRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate updated menu name to ensure it's provided and meaningful
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("menu name is required")
	}

	// Validate updated menu slug to ensure it's provided and meaningful
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("menu slug is required")
	}

	// Validate updated record ordering to ensure it's non-negative
	if recordOrdering < 0 {
		return nil, errors.New("menu record ordering must be non-negative")
	}

	// Check if updated slug already exists (excluding current menu)
	if slug != existingMenu.Slug {
		exists, err := s.menuRepository.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("menu slug already exists")
		}
	}

	// Update the menu entity with new information
	existingMenu.UpdateMenu(name, slug, description, url, icon, uint64(recordOrdering), parentID)

	// Persist the updated menu to the repository
	err = s.menuRepository.Update(ctx, existingMenu)
	if err != nil {
		return nil, err
	}

	// Return the updated menu entity
	return s.GetMenuByID(ctx, id)
}

// DeleteMenu performs a soft delete of a menu item by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - Menu must exist before deletion
//   - Soft delete preserves menu structure
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *MenuService) DeleteMenu(ctx context.Context, id uuid.UUID) error {
	// Check if menu exists before attempting deletion
	_, err := s.menuRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.menuRepository.Delete(ctx, id)
}

// ActivateMenu activates a menu item, making it available for use.
// This method is part of the menu lifecycle management.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to activate
//
// Returns:
//   - *entities.Menu: The activated menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) ActivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Activate(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// DeactivateMenu deactivates a menu item, making it unavailable for use.
// This method is part of the menu lifecycle management.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to deactivate
//
// Returns:
//   - *entities.Menu: The deactivated menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) DeactivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Deactivate(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// ShowMenu makes a menu item visible to users.
// This method controls menu visibility in the user interface.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to show
//
// Returns:
//   - *entities.Menu: The visible menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) ShowMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Show(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// HideMenu makes a menu item invisible to users.
// This method controls menu visibility in the user interface.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the menu to hide
//
// Returns:
//   - *entities.Menu: The hidden menu entity
//   - error: Any error that occurred during the operation
func (s *MenuService) HideMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Hide(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// GetMenuCount returns the total number of menu items in the system.
// This method is useful for statistics and administrative reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.Count(ctx)
}

// GetActiveMenuCount returns the total number of active menu items.
// This method is useful for monitoring active menu usage.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of active menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetActiveMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.CountActive(ctx)
}

// GetVisibleMenuCount returns the total number of visible menu items.
// This method is useful for monitoring visible menu usage.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of visible menu items
//   - error: Any error that occurred during the operation
func (s *MenuService) GetVisibleMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.CountVisible(ctx)
}
