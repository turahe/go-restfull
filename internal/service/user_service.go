package service

import (
	"context"
	"errors"
	"strings"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/usecase"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidUserID = errors.New("invalid user id")
)

// UserCreateOutcome is returned by Create so handlers can expose roles.id.
type UserCreateOutcome struct {
	User   *model.User
	RoleID uint
}

type UserRepo interface {
	List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

// userRoleAssigner assigns RBAC roles by roles.id after user creation.
type userRoleAssigner interface {
	AssignRoleByID(ctx context.Context, userID uint, roleID uint) (bool, error)
}

// roleLookup loads model.Role by id or name for roleId / response metadata.
type roleLookup interface {
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
}

type UserService struct {
	users UserRepo
	roles roleLookup
	rbac  userRoleAssigner
	media *usecase.MediaUsecase
	log   *zap.Logger
}

func NewUserService(users UserRepo, roles roleLookup, rbac userRoleAssigner, media *usecase.MediaUsecase, log *zap.Logger) *UserService {
	return &UserService{users: users, roles: roles, rbac: rbac, media: media, log: log}
}

// Create provisions a new user (admin-only at HTTP layer). Mirrors Register + default role assignment.
func (s *UserService) Create(ctx context.Context, req request.CreateUserRequest) (*UserCreateOutcome, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	name := strings.TrimSpace(req.Name)
	if email == "" || name == "" || req.Password == "" {
		return nil, errors.New("name, email, password are required")
	}

	var assignRoleID uint

	if req.RoleID != nil && *req.RoleID > 0 {
		if s.roles == nil {
			return nil, errors.New("role lookup is not configured")
		}
		rrow, err := s.roles.FindByID(ctx, *req.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrRoleNotFound
			}
			return nil, err
		}
		assignRoleID = rrow.ID
	} else {
		if s.roles == nil {
			return nil, errors.New("role lookup is not configured")
		}
		rrow, err := s.roles.FindByName(ctx, entities.RoleUser)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrRoleNotFound
			}
			return nil, err
		}
		assignRoleID = rrow.ID
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		s.log.Error("email already registered", zap.String("email", email))
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to generate password hash", zap.Error(err))
		return nil, err
	}

	u := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		s.log.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	if s.rbac != nil && assignRoleID > 0 {
		if _, err := s.rbac.AssignRoleByID(ctx, u.ID, assignRoleID); err != nil {
			s.log.Error("failed to assign role", zap.Error(err))
			return nil, err
		}
	}
	return &UserCreateOutcome{User: u, RoleID: assignRoleID}, nil
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

	// Avatar is already loaded by the repository (with fallback).
	// If media service wiring is present, refresh it to keep behavior consistent.
	if s.media != nil {
		avatar, err := s.media.UserAvatar(ctx, u)
		if err != nil {
			s.log.Error("failed to get user avatar", zap.Error(err))
			return nil, err
		}
		u.Avatar = avatar
	}

	return u, nil
}
