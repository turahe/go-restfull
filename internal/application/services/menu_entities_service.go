package services

import (
	"context"
	"errors"
	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// MenuEntitiesService implements the MenuEntitiesService interface
type MenuEntitiesService struct {
	menuRoleRepository repositories.MenuEntitiesRepository
}

// NewMenuRoleService creates a new menu-role service
func NewMenuRoleService(menuRoleRepository repositories.MenuEntitiesRepository) ports.MenuEntitiesService {
	return &MenuEntitiesService{
		menuRoleRepository: menuRoleRepository,
	}
}

// AssignRoleToMenu assigns a role to a menu
func (s *MenuEntitiesService) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Check if the relationship already exists
	hasRole, err := s.menuRoleRepository.HasRole(ctx, menuID, roleID)
	if err != nil {
		return err
	}

	if hasRole {
		return errors.New("role is already assigned to this menu")
	}

	return s.menuRoleRepository.AssignRoleToMenu(ctx, menuID, roleID)
}

// RemoveRoleFromMenu removes a role from a menu
func (s *MenuEntitiesService) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Check if the relationship exists
	hasRole, err := s.menuRoleRepository.HasRole(ctx, menuID, roleID)
	if err != nil {
		return err
	}

	if !hasRole {
		return errors.New("role is not assigned to this menu")
	}

	return s.menuRoleRepository.RemoveRoleFromMenu(ctx, menuID, roleID)
}

// GetMenuRoles retrieves all roles assigned to a menu
func (s *MenuEntitiesService) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
	return s.menuRoleRepository.GetMenuRoles(ctx, menuID)
}

// GetRoleMenus retrieves all menus assigned to a role
func (s *MenuEntitiesService) GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRoleRepository.GetRoleMenus(ctx, roleID, limit, offset)
}

// HasRole checks if a menu has a specific role
func (s *MenuEntitiesService) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	return s.menuRoleRepository.HasRole(ctx, menuID, roleID)
}

// HasAnyRole checks if a menu has any of the specified roles
func (s *MenuEntitiesService) HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	return s.menuRoleRepository.HasAnyRole(ctx, menuID, roleIDs)
}

// GetMenuRoleIDs retrieves all role IDs assigned to a menu
func (s *MenuEntitiesService) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	return s.menuRoleRepository.GetMenuRoleIDs(ctx, menuID)
}

// GetMenuRoleCount returns the number of menus assigned to a role
func (s *MenuEntitiesService) GetMenuRoleCount(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return s.menuRoleRepository.CountMenusByRole(ctx, roleID)
}
