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

// MenuService implements the MenuService interface
type MenuService struct {
	menuRepository repositories.MenuRepository
}

// NewMenuService creates a new menu service
func NewMenuService(menuRepository repositories.MenuRepository) ports.MenuService {
	return &MenuService{
		menuRepository: menuRepository,
	}
}

// CreateMenu creates a new menu
func (s *MenuService) CreateMenu(ctx context.Context, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error) {
	// Validate input
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("menu name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("menu slug is required")
	}

	if recordOrdering < 0 {
		return nil, errors.New("menu record ordering must be non-negative")
	}

	// Check if slug already exists
	exists, err := s.menuRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("menu slug already exists")
	}

	// Create menu entity
	menu := entities.NewMenu(name, slug, description, url, icon, recordOrdering, parentID)

	// Validate menu
	if err := menu.Validate(); err != nil {
		return nil, err
	}

	// Save to repository
	err = s.menuRepository.Create(ctx, menu)
	if err != nil {
		return nil, err
	}

	return menu, nil
}

// GetMenuByID retrieves a menu by ID
func (s *MenuService) GetMenuByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	return s.menuRepository.GetByID(ctx, id)
}

// GetMenuBySlug retrieves a menu by slug
func (s *MenuService) GetMenuBySlug(ctx context.Context, slug string) (*entities.Menu, error) {
	return s.menuRepository.GetBySlug(ctx, slug)
}

// GetAllMenus retrieves all menus with pagination
func (s *MenuService) GetAllMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetAll(ctx, limit, offset)
}

// GetActiveMenus retrieves active menus with pagination
func (s *MenuService) GetActiveMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetActive(ctx, limit, offset)
}

// GetVisibleMenus retrieves visible menus with pagination
func (s *MenuService) GetVisibleMenus(ctx context.Context, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRepository.GetVisible(ctx, limit, offset)
}

// GetRootMenus retrieves root menus (no parent)
func (s *MenuService) GetRootMenus(ctx context.Context) ([]*entities.Menu, error) {
	return s.menuRepository.GetRootMenus(ctx)
}

// GetMenuHierarchy retrieves the complete menu hierarchy
func (s *MenuService) GetMenuHierarchy(ctx context.Context) ([]*entities.Menu, error) {
	return s.menuRepository.GetHierarchy(ctx)
}

// GetUserMenus retrieves menus accessible to a specific user
func (s *MenuService) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]*entities.Menu, error) {
	return s.menuRepository.GetUserMenus(ctx, userID)
}

// SearchMenus searches menus by query
func (s *MenuService) SearchMenus(ctx context.Context, query string, limit, offset int) ([]*entities.Menu, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query is required")
	}

	return s.menuRepository.Search(ctx, query, limit, offset)
}

// UpdateMenu updates a menu
func (s *MenuService) UpdateMenu(ctx context.Context, id uuid.UUID, name, slug, description, url, icon string, recordOrdering int64, parentID *uuid.UUID) (*entities.Menu, error) {
	// Get existing menu
	existingMenu, err := s.menuRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate input
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("menu name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("menu slug is required")
	}

	if recordOrdering < 0 {
		return nil, errors.New("menu record ordering must be non-negative")
	}

	// Check if slug already exists (excluding current menu)
	if slug != existingMenu.Slug {
		exists, err := s.menuRepository.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("menu slug already exists")
		}
	}

	// Update menu
	existingMenu.UpdateMenu(name, slug, description, url, icon, recordOrdering, parentID)

	// Save to repository
	err = s.menuRepository.Update(ctx, existingMenu)
	if err != nil {
		return nil, err
	}

	// Return updated menu entity
	return s.GetMenuByID(ctx, id)
}

// DeleteMenu deletes a menu
func (s *MenuService) DeleteMenu(ctx context.Context, id uuid.UUID) error {
	// Check if menu exists
	_, err := s.menuRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.menuRepository.Delete(ctx, id)
}

// ActivateMenu activates a menu
func (s *MenuService) ActivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Activate(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// DeactivateMenu deactivates a menu
func (s *MenuService) DeactivateMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Deactivate(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// ShowMenu makes a menu visible
func (s *MenuService) ShowMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Show(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// HideMenu makes a menu invisible
func (s *MenuService) HideMenu(ctx context.Context, id uuid.UUID) (*entities.Menu, error) {
	err := s.menuRepository.Hide(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetMenuByID(ctx, id)
}

// GetMenuCount returns the total number of menus
func (s *MenuService) GetMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.Count(ctx)
}

// GetActiveMenuCount returns the number of active menus
func (s *MenuService) GetActiveMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.CountActive(ctx)
}

// GetVisibleMenuCount returns the number of visible menus
func (s *MenuService) GetVisibleMenuCount(ctx context.Context) (int64, error) {
	return s.menuRepository.CountVisible(ctx)
}
