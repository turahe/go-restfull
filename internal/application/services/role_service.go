package services

import (
	"context"
	"errors"
	"strings"
	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
)

// RoleService implements the RoleService interface
type RoleService struct {
	roleRepository repositories.RoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepository repositories.RoleRepository) ports.RoleService {
	return &RoleService{
		roleRepository: roleRepository,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, name, slug, description string) (*entities.Role, error) {
	// Validate inputs
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("role name is required")
	}

	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("role slug is required")
	}

	// Check if slug already exists
	exists, err := s.roleRepository.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("role with this slug already exists")
	}

	// Create new role entity
	role, err := entities.NewRole(name, slug, description)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.roleRepository.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetRoleByID retrieves role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	return role, nil
}

// GetRoleBySlug retrieves role by slug
func (s *RoleService) GetRoleBySlug(ctx context.Context, slug string) (*entities.Role, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("role slug is required")
	}

	role, err := s.roleRepository.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	return role, nil
}

// GetAllRoles retrieves all roles with pagination
func (s *RoleService) GetAllRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
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

	// Filter out deleted roles
	var activeRoles []*entities.Role
	for _, role := range roles {
		if !role.IsDeleted() {
			activeRoles = append(activeRoles, role)
		}
	}

	return activeRoles, nil
}

// GetActiveRoles retrieves only active roles with pagination
func (s *RoleService) GetActiveRoles(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.roleRepository.GetActive(ctx, limit, offset)
}

// SearchRoles searches roles by query
func (s *RoleService) SearchRoles(ctx context.Context, query string, limit, offset int) ([]*entities.Role, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query is required")
	}

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

	// Filter out deleted roles
	var activeRoles []*entities.Role
	for _, role := range roles {
		if !role.IsDeleted() {
			activeRoles = append(activeRoles, role)
		}
	}

	return activeRoles, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, name, slug, description string) (*entities.Role, error) {
	// Get existing role
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

	// Update role
	err = role.UpdateRole(name, slug, description)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.roleRepository.Update(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// DeleteRole soft deletes role by ID
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	// Get existing role
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsDeleted() {
		return errors.New("role not found")
	}

	// Soft delete role
	return s.roleRepository.Delete(ctx, id)
}

// ActivateRole activates a role
func (s *RoleService) ActivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// Get existing role
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	// Activate role
	err = s.roleRepository.Activate(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get updated role
	return s.roleRepository.GetByID(ctx, id)
}

// DeactivateRole deactivates a role
func (s *RoleService) DeactivateRole(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	// Get existing role
	role, err := s.roleRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsDeleted() {
		return nil, errors.New("role not found")
	}

	// Deactivate role
	err = s.roleRepository.Deactivate(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get updated role
	return s.roleRepository.GetByID(ctx, id)
}

// GetRoleCount returns the total number of roles
func (s *RoleService) GetRoleCount(ctx context.Context) (int64, error) {
	return s.roleRepository.Count(ctx)
}

// GetActiveRoleCount returns the total number of active roles
func (s *RoleService) GetActiveRoleCount(ctx context.Context) (int64, error) {
	return s.roleRepository.CountActive(ctx)
} 