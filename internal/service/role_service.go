package service

import (
	"context"
	"errors"
	"strings"

	"go-rest/internal/model"
	"go-rest/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrInvalidRoleID = errors.New("invalid role id")
	ErrInvalidRole   = errors.New("invalid role")
)

type RoleService struct {
	roles *repository.RoleRepository
}

func NewRoleService(roles *repository.RoleRepository) *RoleService {
	return &RoleService{roles: roles}
}

func (s *RoleService) List(ctx context.Context, limit int) ([]model.Role, error) {
	return s.roles.List(ctx, limit)
}

func (s *RoleService) Create(ctx context.Context, name string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidRole
	}
	role := &model.Role{Name: name}
	if err := s.roles.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidRoleID
	}
	_, err := s.roles.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return s.roles.DeleteByID(ctx, id)
}

