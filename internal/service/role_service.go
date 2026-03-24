package service

import (
	"context"
	"errors"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrInvalidRoleID = errors.New("invalid role id")
	ErrInvalidRole   = errors.New("invalid role")
)

type RoleService struct {
	roles *repository.RoleRepository
	log   *zap.Logger
}

func NewRoleService(roles *repository.RoleRepository, log *zap.Logger) *RoleService {
	return &RoleService{roles: roles, log: log}
}

func (s *RoleService) List(ctx context.Context, req request.RoleListRequest) (repository.CursorPage, error) {
	page, err := s.roles.List(ctx, req)
	if err != nil {
		s.log.Error("failed to list roles", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func (s *RoleService) Create(ctx context.Context, req request.CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{Name: req.Name}
	if err := s.roles.Create(ctx, role); err != nil {
		s.log.Error("failed to create role", zap.Error(err))
		return nil, err
	}
	return role, nil
}

func (s *RoleService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		s.log.Error("invalid role id")
		return ErrInvalidRoleID
	}
	_, err := s.roles.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("failed to find role by id", zap.Error(err))
			return ErrRoleNotFound
		}
		s.log.Error("failed to find role by id", zap.Error(err))
		return err
	}
	if err := s.roles.DeleteByID(ctx, id); err != nil {
		s.log.Error("failed to delete role by id", zap.Error(err))
		return err
	}
	return nil
}
