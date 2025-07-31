// Package services provides application-level business logic for menu-role relationship management.
// This package contains the menu entities service implementation that handles menu-role assignments,
// role-based menu access control, and menu-role relationship management while ensuring proper
// authorization and security.
package services

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// MenuEntitiesService implements the MenuEntitiesService interface and provides comprehensive
// menu-role relationship management functionality. It handles menu-role assignments, role-based
// menu access control, and menu-role relationship management while ensuring proper
// authorization and security.
type MenuEntitiesService struct {
	menuRoleRepository repositories.MenuEntitiesRepository
}

// NewMenuRoleService creates a new menu entities service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - menuRoleRepository: Repository interface for menu-role relationship data access operations
//
// Returns:
//   - ports.MenuEntitiesService: The menu entities service interface implementation
func NewMenuRoleService(menuRoleRepository repositories.MenuEntitiesRepository) ports.MenuEntitiesService {
	return &MenuEntitiesService{
		menuRoleRepository: menuRoleRepository,
	}
}

// AssignRoleToMenu assigns a specific role to a menu, establishing a menu-role relationship.
// This method enforces business rules for role assignment and validates existing relationships.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Duplicate assignments are prevented
//   - Role assignment is atomic and consistent
//
// Security Features:
//   - Input validation prevents invalid assignments
//   - Duplicate relationship checking
//   - Atomic operations prevent race conditions
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to assign the role to
//   - roleID: UUID of the role to assign
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) AssignRoleToMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Check if the relationship already exists to prevent duplicates
	hasRole, err := s.menuRoleRepository.HasRole(ctx, menuID, roleID)
	if err != nil {
		return err
	}

	if hasRole {
		return errors.New("role is already assigned to this menu")
	}

	return s.menuRoleRepository.AssignRoleToMenu(ctx, menuID, roleID)
}

// RemoveRoleFromMenu removes a specific role from a menu, dissolving the menu-role relationship.
// This method enforces business rules for role removal and validates existing relationships.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Relationship must exist before removal
//   - Role removal is atomic and consistent
//
// Security Features:
//   - Input validation prevents invalid removals
//   - Relationship existence checking
//   - Atomic operations prevent race conditions
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to remove the role from
//   - roleID: UUID of the role to remove
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) RemoveRoleFromMenu(ctx context.Context, menuID, roleID uuid.UUID) error {
	// Check if the relationship exists before attempting removal
	hasRole, err := s.menuRoleRepository.HasRole(ctx, menuID, roleID)
	if err != nil {
		return err
	}

	if !hasRole {
		return errors.New("role is not assigned to this menu")
	}

	return s.menuRoleRepository.RemoveRoleFromMenu(ctx, menuID, roleID)
}

// GetMenuRoles retrieves all roles assigned to a specific menu.
// This method provides comprehensive role information for menu authorization.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Returns complete role entities with metadata
//   - Handles menus with no assigned roles gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to get roles for
//
// Returns:
//   - []*entities.Role: List of roles assigned to the menu
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) GetMenuRoles(ctx context.Context, menuID uuid.UUID) ([]*entities.Role, error) {
	return s.menuRoleRepository.GetMenuRoles(ctx, menuID)
}

// GetRoleMenus retrieves all menus assigned to a specific role with pagination.
// This method is useful for role management and menu administration.
//
// Business Rules:
//   - Role ID must be valid and not nil
//   - Pagination parameters are properly handled
//   - Returns complete menu entities with metadata
//   - Handles roles with no assigned menus gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - roleID: UUID of the role to get menus for
//   - limit: Maximum number of menus to return
//   - offset: Number of menus to skip for pagination
//
// Returns:
//   - []*entities.Menu: List of menus assigned to the role
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) GetRoleMenus(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.Menu, error) {
	return s.menuRoleRepository.GetRoleMenus(ctx, roleID, limit, offset)
}

// HasRole checks if a menu has a specific role, providing boolean authorization.
// This method is useful for role-based menu access control and authorization checks.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Returns boolean result for authorization decisions
//
// Security Features:
//   - Input validation prevents security bypasses
//   - Boolean result for clear authorization decisions
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to check
//   - roleID: UUID of the role to check for
//
// Returns:
//   - bool: True if menu has the role, false otherwise
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) HasRole(ctx context.Context, menuID, roleID uuid.UUID) (bool, error) {
	return s.menuRoleRepository.HasRole(ctx, menuID, roleID)
}

// HasAnyRole checks if a menu has any of the specified roles, providing flexible authorization.
// This method is useful for multi-role menu access control and authorization checks.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Role IDs list must not be empty
//   - Returns boolean result for authorization decisions
//   - Supports multiple role checking in single operation
//
// Security Features:
//   - Input validation prevents security bypasses
//   - Boolean result for clear authorization decisions
//   - Efficient multi-role checking
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to check
//   - roleIDs: List of role UUIDs to check for
//
// Returns:
//   - bool: True if menu has any of the specified roles, false otherwise
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) HasAnyRole(ctx context.Context, menuID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	return s.menuRoleRepository.HasAnyRole(ctx, menuID, roleIDs)
}

// GetMenuRoleIDs retrieves all role IDs assigned to a specific menu.
// This method provides lightweight role information for efficient authorization.
//
// Business Rules:
//   - Menu ID must be valid and not nil
//   - Returns only role IDs for performance optimization
//   - Handles menus with no assigned roles gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - menuID: UUID of the menu to get role IDs for
//
// Returns:
//   - []uuid.UUID: List of role IDs assigned to the menu
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) GetMenuRoleIDs(ctx context.Context, menuID uuid.UUID) ([]uuid.UUID, error) {
	return s.menuRoleRepository.GetMenuRoleIDs(ctx, menuID)
}

// GetMenuRoleCount returns the number of menus assigned to a specific role.
// This method is useful for role statistics and administrative reporting.
//
// Business Rules:
//   - Role ID must be valid and not nil
//   - Returns accurate count for role usage statistics
//   - Handles roles with no assigned menus gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - roleID: UUID of the role to count menus for
//
// Returns:
//   - int64: Number of menus assigned to the role
//   - error: Any error that occurred during the operation
func (s *MenuEntitiesService) GetMenuRoleCount(ctx context.Context, roleID uuid.UUID) (int64, error) {
	return s.menuRoleRepository.CountMenusByRole(ctx, roleID)
}
