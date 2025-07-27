package services

import (
	"context"
	"errors"
	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
)

// UserRoleService implements the UserRoleService interface
type UserRoleService struct {
	userRoleRepository repositories.UserRoleRepository
}

// NewUserRoleService creates a new user-role service
func NewUserRoleService(userRoleRepository repositories.UserRoleRepository) ports.UserRoleService {
	return &UserRoleService{
		userRoleRepository: userRoleRepository,
	}
}

// AssignRoleToUser assigns a role to a user
func (s *UserRoleService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	if roleID == uuid.Nil {
		return errors.New("invalid role ID")
	}

	return s.userRoleRepository.AssignRoleToUser(ctx, userID, roleID)
}

// RemoveRoleFromUser removes a role from a user
func (s *UserRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("invalid user ID")
	}
	if roleID == uuid.Nil {
		return errors.New("invalid role ID")
	}

	return s.userRoleRepository.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles retrieves all roles for a user
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRoleRepository.GetUserRoles(ctx, userID)
}

// GetRoleUsers retrieves all users for a role with pagination
func (s *UserRoleService) GetRoleUsers(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	if roleID == uuid.Nil {
		return nil, errors.New("invalid role ID")
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.userRoleRepository.GetRoleUsers(ctx, roleID, limit, offset)
}

// HasRole checks if a user has a specific role
func (s *UserRoleService) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, errors.New("invalid user ID")
	}
	if roleID == uuid.Nil {
		return false, errors.New("invalid role ID")
	}

	return s.userRoleRepository.HasRole(ctx, userID, roleID)
}

// HasAnyRole checks if a user has any of the specified roles
func (s *UserRoleService) HasAnyRole(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, errors.New("invalid user ID")
	}
	if len(roleIDs) == 0 {
		return false, errors.New("role IDs list cannot be empty")
	}

	return s.userRoleRepository.HasAnyRole(ctx, userID, roleIDs)
}

// GetUserRoleIDs retrieves all role IDs for a user
func (s *UserRoleService) GetUserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	return s.userRoleRepository.GetUserRoleIDs(ctx, userID)
}

// GetUserRoleCount returns the number of users assigned to a role
func (s *UserRoleService) GetUserRoleCount(ctx context.Context, roleID uuid.UUID) (int64, error) {
	if roleID == uuid.Nil {
		return 0, errors.New("invalid role ID")
	}

	return s.userRoleRepository.CountUsersByRole(ctx, roleID)
} 