package service

import (
	"context"
	"errors"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"
	"go-rest/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidUserID = errors.New("invalid user id")
)

type UserRepo interface {
	List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

type UserService struct {
	users UserRepo
	log   *zap.Logger
}

func NewUserService(users UserRepo, log *zap.Logger) *UserService {
	return &UserService{users: users, log: log}
}

func (s *UserService) List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error) {
	page, err := s.users.List(ctx, req)
	if err != nil {
		s.log.Error("failed to list users", zap.Error(err))
		return repository.CursorPage{}, err
	}
	return page, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	if id == 0 {
		s.log.Error("invalid user id")
		return nil, ErrInvalidUserID
	}
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error("user not found")
			return nil, ErrUserNotFound
		}
		s.log.Error("failed to find user by id", zap.Error(err))
		return nil, err
	}
	return u, nil
}
