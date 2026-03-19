package service

import (
	"context"
	"errors"

	"go-rest/internal/model"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidUserID = errors.New("invalid user id")
)

type UserRepo interface {
	List(ctx context.Context, limit int) ([]model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

type UserService struct {
	users UserRepo
}

func NewUserService(users UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List(ctx context.Context, limit int) ([]model.User, error) {
	return s.users.List(ctx, limit)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	if id == 0 {
		return nil, ErrInvalidUserID
	}
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

