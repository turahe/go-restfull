// Package services provides application-level business logic for user role management.
// This package contains the user role service implementation that handles role assignments,
// user-role relationships, and role-based access control while ensuring proper
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

// UserRoleService implements the UserRoleService interface and provides comprehensive
// user role management functionality. It handles role assignments, user-role relationships,
// and role-based access control while ensuring proper authorization and security.
type UserRoleService struct {
	userRoleRepository repositories.UserRoleRepository
}

// NewUserRoleService creates a new user role service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - userRoleRepository: Repository interface for user role data access operations
//
// Returns:
//   - ports.UserRoleService: The user role service interface implementation
func NewUserRoleService(userRoleRepository repositories.UserRoleRepository) ports.UserRoleService {
	return &UserRoleService{
		userRoleRepository: userRoleRepository,
	}
}

// AssignRoleToUser assigns a specific role to a user, establishing a user-role relationship.
// This method enforces business rules for role assignment and validates input parameters.
//
// Business Rules:
//   - User ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Duplicate assignments are handled by the repository layer
//   - Role assignment is atomic and consistent
//
// Security Features:
//   - Input validation prevents invalid assignments
//   - UUID validation ensures data integrity
//   - Atomic operations prevent race conditions
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to assign the role to
//   - roleID: UUID of the role to assign
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *UserRoleService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	// Validate role ID to ensure it's not nil
	if roleID == uuid.Nil {
		return errors.New("invalid role ID")
	}

	return s.userRoleRepository.AssignRoleToUser(ctx, userID, roleID)
}

// RemoveRoleFromUser removes a specific role from a user, dissolving the user-role relationship.
// This method enforces business rules for role removal and validates input parameters.
//
// Business Rules:
//   - User ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Non-existent assignments are handled gracefully
//   - Role removal is atomic and consistent
//
// Security Features:
//   - Input validation prevents invalid removals
//   - UUID validation ensures data integrity
//   - Atomic operations prevent race conditions
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to remove the role from
//   - roleID: UUID of the role to remove
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *UserRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	// Validate role ID to ensure it's not nil
	if roleID == uuid.Nil {
		return errors.New("invalid role ID")
	}

	return s.userRoleRepository.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles retrieves all roles assigned to a specific user.
// This method provides comprehensive role information for user authorization.
//
// Business Rules:
//   - User ID must be valid and not nil
//   - Returns complete role entities with metadata
//   - Handles users with no assigned roles gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get roles for
//
// Returns:
//   - []*entities.Role: List of roles assigned to the user
//   - error: Any error that occurred during the operation
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRoleRepository.GetUserRoles(ctx, userID)
}

// GetRoleUsers retrieves all users assigned to a specific role with pagination.
// This method is useful for role management and user administration.
//
// Business Rules:
//   - Role ID must be valid and not nil
//   - Default pagination values are applied for better UX
//   - Returns complete user entities with metadata
//   - Handles roles with no assigned users gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - roleID: UUID of the role to get users for
//   - limit: Maximum number of users to return (defaults to 10)
//   - offset: Number of users to skip for pagination
//
// Returns:
//   - []*entities.User: List of users assigned to the role
//   - error: Any error that occurred during the operation
func (s *UserRoleService) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	// Validate role ID to ensure it's not nil
	if roleID == uuid.Nil {
		return nil, errors.New("invalid role ID")
	}

	// Set default pagination values for better user experience
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.userRoleRepository.GetRoleUsers(ctx, roleID, limit, offset)
}

// HasRole checks if a user has a specific role, providing boolean authorization.
// This method is useful for role-based access control and authorization checks.
//
// Business Rules:
//   - User ID must be valid and not nil
//   - Role ID must be valid and not nil
//   - Returns boolean result for authorization decisions
//
// Security Features:
//   - Input validation prevents security bypasses
//   - Boolean result for clear authorization decisions
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to check
//   - roleID: UUID of the role to check for
//
// Returns:
//   - bool: True if user has the role, false otherwise
//   - error: Any error that occurred during the operation
func (s *UserRoleService) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return false, errors.New("invalid user ID")
	}
	// Validate role ID to ensure it's not nil
	if roleID == uuid.Nil {
		return false, errors.New("invalid role ID")
	}

	return s.userRoleRepository.HasRole(ctx, userID, roleID)
}

// HasAnyRole checks if a user has any of the specified roles, providing flexible authorization.
// This method is useful for multi-role authorization checks and access control.
//
// Business Rules:
//   - User ID must be valid and not nil
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
//   - userID: UUID of the user to check
//   - roleIDs: List of role UUIDs to check for
//
// Returns:
//   - bool: True if user has any of the specified roles, false otherwise
//   - error: Any error that occurred during the operation
func (s *UserRoleService) HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return false, errors.New("invalid user ID")
	}
	// Validate role IDs list to ensure it's not empty
	if len(roleIDs) == 0 {
		return false, errors.New("role IDs list cannot be empty")
	}

	return s.userRoleRepository.HasAnyRole(ctx, userID, roleIDs)
}

// GetUserRoleIDs retrieves all role IDs assigned to a specific user.
// This method provides lightweight role information for efficient authorization.
//
// Business Rules:
//   - User ID must be valid and not nil
//   - Returns only role IDs for performance optimization
//   - Handles users with no assigned roles gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: UUID of the user to get role IDs for
//
// Returns:
//   - []uuid.UUID: List of role IDs assigned to the user
//   - error: Any error that occurred during the operation
func (s *UserRoleService) GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Validate user ID to ensure it's not nil
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRoleRepository.GetUserRoleIDs(ctx, userID)
}

// GetUserRoleCount returns the number of users assigned to a specific role.
// This method is useful for role statistics and administrative reporting.
//
// Business Rules:
//   - Role ID must be valid and not nil
//   - Returns accurate count for role usage statistics
//   - Handles roles with no assigned users gracefully
//
// Parameters:
//   - ctx: Context for the operation
//   - roleID: UUID of the role to count users for
//
// Returns:
//   - int64: Number of users assigned to the role
//   - error: Any error that occurred during the operation
func (s *UserRoleService) GetUserRoleCount(ctx context.Context, roleID uuid.UUID) (int64, error) {
	// Validate role ID to ensure it's not nil
	if roleID == uuid.Nil {
		return 0, errors.New("invalid role ID")
	}

	return s.userRoleRepository.CountUsersByRole(ctx, roleID)
}
