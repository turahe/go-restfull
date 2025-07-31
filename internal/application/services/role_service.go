// Package services provides application-level business logic for role management.
// This package contains the role service implementation that handles role creation,
// role lifecycle management, role assignments, and authorization while ensuring proper
// data integrity and business rules.
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

// RoleService implements the RoleService interface and provides comprehensive
// role management functionality. It handles role creation, role lifecycle management,
// role assignments, and authorization while ensuring proper data integrity
// and business rules.
type RoleService struct {
	roleRepository repositories.RoleRepository
}

// NewRoleService creates a new role service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - roleRepository: Repository interface for role data access operations
//
// Returns:
//   - ports.RoleService: The role service interface implementation
func NewRoleService(roleRepository repositories.RoleRepository) ports.RoleService {
	return &RoleService{
		roleRepository: roleRepository,
	}
}

// CreateRole creates a new role with comprehensive validation and slug uniqueness.
// This method enforces business rules for role creation and supports role lifecycle.
//
// Business Rules:
//   - Role name is required and cannot be empty
//   - Role slug is required and must be unique
//   - Role validation ensures proper structure
//   - Slug uniqueness prevents conflicts
//
// Parameters:
//   - ctx: Context for the operation
//   - name: Display name of the role
//   - slug: Unique slug identifier for the role
//   - description: Optional description of the role
//
// Returns:
//   - *entities.Role: The created role entity
//   - error: Any error that occurred during the operation
func (s *RoleService) CreateRole(ctx context.Context, name, slug, description string) (*entities.Role, error) {
	// Validate inputs to ensure required fields are provided
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("role name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("role slug is required")
	}

	// Check if slug already exists to prevent duplicates
	exists, err := s.roleRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("role with this slug already exists")
	}

	// Create new role entity with validated parameters
	role, err := entities.NewRole(name, slug, description)
	if err != nil {
		return nil, err
	}

	// Persist the role to the repository
	err = s.roleRepository.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetRoleByID retrieves a role by its unique identifier.
// This method includes soft delete checking to ensure deleted roles
// are not returned to the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the role to retrieve
//
// Returns:
//   - *entities.Role: The role entity if found
//   - error: Error if role not found or other issues occur
func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the role has been soft deleted
	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	return role, nil
}

// GetRoleBySlug retrieves a role by its unique slug identifier.
// This method is useful for slug-based role lookups and routing.
//
// Business Rules:
//   - Slug must be provided and not empty
//   - Role must exist and not be deleted
//
// Parameters:
//   - ctx: Context for the operation
//   - slug: Slug identifier of the role to retrieve
//
// Returns:
//   - *entities.Role: The role entity if found
//   - error: Error if role not found or other issues occur
func (s *RoleService) GetRoleBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	// Validate slug is provided
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("role slug is required")
	}

	role, err := s.roleRepository.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check if the role has been soft deleted
	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	return role, nil
}

// GetAllRoles retrieves all roles in the system with pagination.
// This method is useful for administrative purposes and role management.
//
// Business Rules:
//   - Default limit of 10 if not specified
//   - Offset must be non-negative
//   - Deleted roles are filtered out from results
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of roles to return (defaults to 10)
//   - offset: Number of roles to skip for pagination
//
// Returns:
//   - []*entities.Role: List of all active roles
//   - error: Any error that occurred during the operation
func (s *RoleService) GetAllRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// Set default values for pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	roles, err := s.roleRepository.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter out deleted roles to ensure only active roles are returned
	var activeRoles []*entities.Role
	for _, role := range roles {
		if !role.IsDeleted() {
			activeRoles = append(activeRoles, role)
		}
	}

	return activeRoles, nil
}

// GetActiveRoles retrieves only active roles with pagination.
// This method is useful for role assignment and authorization contexts.
//
// Business Rules:
//   - Default limit of 10 if not specified
//   - Offset must be non-negative
//   - Only active roles are returned
//
// Parameters:
//   - ctx: Context for the operation
//   - limit: Maximum number of active roles to return (defaults to 10)
//   - offset: Number of active roles to skip for pagination
//
// Returns:
//   - []*entities.Role: List of active roles
//   - error: Any error that occurred during the operation
func (s *RoleService) GetActiveRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	// Set default values for pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.roleRepository.GetActive(ctx, limit, offset)
}

// SearchRoles searches for roles based on a query string.
// This method supports full-text search capabilities for finding roles
// by name, description, or other attributes.
//
// Business Rules:
//   - Search query is required and cannot be empty
//   - Default limit of 10 if not specified
//   - Offset must be non-negative
//   - Deleted roles are filtered out from search results
//
// Parameters:
//   - ctx: Context for the operation
//   - query: Search query string (required)
//   - limit: Maximum number of search results to return (defaults to 10)
//   - offset: Number of search results to skip for pagination
//
// Returns:
//   - []*entities.Role: List of matching active roles
//   - error: Any error that occurred during the operation
func (s *RoleService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	// Validate search query is provided
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query is required")
	}

	// Set default values for pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	roles, err := s.roleRepository.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter out deleted roles to ensure only active roles are returned
	var activeRoles []*entities.Role
	for _, role := range roles {
		if !role.IsDeleted() {
			activeRoles = append(activeRoles, role)
		}
	}

	return activeRoles, nil
}

// UpdateRole updates an existing role's information and metadata.
// This method enforces business rules and maintains data integrity during updates.
//
// Business Rules:
//   - Role must exist and not be deleted
//   - Updated slug must be unique if changed
//   - Role validation ensures proper structure
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the role to update
//   - name: Updated display name of the role
//   - slug: Updated unique slug identifier for the role
//   - description: Updated description of the role
//
// Returns:
//   - *entities.Role: The updated role entity
//   - error: Any error that occurred during the operation
func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, name, slug, description string) (*entities.Role, error) {
	// Retrieve existing role to ensure it exists and is not deleted
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	// Check if new slug already exists (if slug is being changed)
	if slug != "" && slug != role.Slug {
		exists, err := s.roleRepository.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("role with this slug already exists")
		}
	}

	// Update the role entity with new information
	err = role.UpdateRole(name, slug, description)
	if err != nil {
		return nil, err
	}

	// Persist the updated role to the repository
	err = s.roleRepository.Update(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// DeleteRole performs a soft delete of a role by marking it as deleted
// rather than physically removing it from the database. This preserves data
// integrity and allows for potential recovery.
//
// Business Rules:
//   - Role must exist before deletion
//   - Soft delete preserves role data
//   - Deleted roles are not returned in queries
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the role to delete
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	// Retrieve existing role to ensure it exists and is not deleted
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsDeleted() {
		return errors.New("role not found")
	}

	// Perform soft delete by marking the role as deleted
	return s.roleRepository.Delete(ctx, id)
}

// ActivateRole activates a role, making it available for assignment.
// This method is part of the role lifecycle management.
//
// Business Rules:
//   - Role must exist and not be deleted
//   - Activation status is updated atomically
//   - Activated roles become available for assignment
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the role to activate
//
// Returns:
//   - *entities.Role: The activated role entity
//   - error: Any error that occurred during the operation
func (s *RoleService) ActivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// Retrieve existing role to ensure it exists and is not deleted
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	// Activate the role
	err = s.roleRepository.Activate(ctx, id)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated role to return current state
	return s.roleRepository.GetByID(ctx, id)
}

// DeactivateRole deactivates a role, making it unavailable for assignment.
// This method is part of the role lifecycle management.
//
// Business Rules:
//   - Role must exist and not be deleted
//   - Deactivation status is updated atomically
//   - Deactivated roles become unavailable for assignment
//
// Parameters:
//   - ctx: Context for the operation
//   - id: UUID of the role to deactivate
//
// Returns:
//   - *entities.Role: The deactivated role entity
//   - error: Any error that occurred during the operation
func (s *RoleService) DeactivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// Retrieve existing role to ensure it exists and is not deleted
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	// Deactivate the role
	err = s.roleRepository.Deactivate(ctx, id)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated role to return current state
	return s.roleRepository.GetByID(ctx, id)
}

// GetRoleCount returns the total number of roles in the system.
// This method is useful for administrative reporting and statistics.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of roles
//   - error: Any error that occurred during the operation
func (s *RoleService) GetRoleCount(ctx context.Context) (int64, error) {
	return s.roleRepository.Count(ctx)
}

// GetActiveRoleCount returns the total number of active roles in the system.
// This method is useful for role assignment statistics and reporting.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - int64: Total count of active roles
//   - error: Any error that occurred during the operation
func (s *RoleService) GetActiveRoleCount(ctx context.Context) (int64, error) {
	return s.roleRepository.CountActive(ctx)
}
