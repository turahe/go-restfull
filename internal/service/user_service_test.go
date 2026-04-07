package service

import (
	"context"
	"errors"
	"testing"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(repository.CursorPage), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, u *model.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*model.User)
	return u, args.Error(1)
}

type mockUserRBAC struct {
	mock.Mock
}

func (m *mockUserRBAC) AssignRoleByID(ctx context.Context, userID uint, roleID uint) (bool, error) {
	args := m.Called(ctx, userID, roleID)
	return args.Bool(0), args.Error(1)
}

type mockRoleLookup struct {
	mock.Mock
}

func (m *mockRoleLookup) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	args := m.Called(ctx, id)
	r, _ := args.Get(0).(*model.Role)
	return r, args.Error(1)
}

func (m *mockRoleLookup) FindByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	r, _ := args.Get(0).(*model.Role)
	return r, args.Error(1)
}

func TestUserService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		id        uint
		mockSetup func(r *mockUserRepo)
		wantErr   error
		wantNil   bool
	}{
		{
			name:    "invalid id",
			id:      0,
			wantErr: ErrInvalidUserID,
			wantNil: true,
		},
		{
			name: "not found maps to service error",
			id:   123,
			mockSetup: func(r *mockUserRepo) {
				r.On("FindByID", mock.Anything, uint(123)).Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
			},
			wantErr: ErrUserNotFound,
			wantNil: true,
		},
		{
			name: "db error passthrough",
			id:   123,
			mockSetup: func(r *mockUserRepo) {
				r.On("FindByID", mock.Anything, uint(123)).Return((*model.User)(nil), errors.New("db down")).Once()
			},
			wantErr: errors.New("db down"),
			wantNil: true,
		},
		{
			name: "success",
			id:   123,
			mockSetup: func(r *mockUserRepo) {
				r.On("FindByID", mock.Anything, uint(123)).Return(&model.User{ID: 123, Email: "a@b.com"}, nil).Once()
			},
			wantErr: nil,
			wantNil: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockUserRepo{}
			if tc.mockSetup != nil {
				tc.mockSetup(repo)
			}
			svc := NewUserService(repo, nil, nil, nil, zap.NewNop())

			u, err := svc.GetByID(ctx, tc.id)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.wantNil {
				assert.Nil(t, u)
			} else {
				assert.NotNil(t, u)
				assert.Equal(t, tc.id, u.ID)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestUserService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &mockUserRepo{}
	listReq := request.UserListRequest{
		PageRequest: request.PageRequest{Limit: 10},
	}
	repo.On("List", mock.Anything, listReq).Return(repository.CursorPage{
		Items: []model.User{{ID: 1}},
	}, nil).Once()

	svc := NewUserService(repo, nil, nil, nil, zap.NewNop())
	page, err := svc.List(ctx, listReq)

	assert.NoError(t, err)
	items, ok := page.Items.([]model.User)
	assert.True(t, ok)
	assert.Len(t, items, 1)
	assert.Equal(t, uint(1), items[0].ID)
	repo.AssertExpectations(t)
}

func TestUserService_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("email taken", func(t *testing.T) {
		t.Parallel()
		repo := &mockUserRepo{}
		roles := &mockRoleLookup{}
		roles.On("FindByName", mock.Anything, entities.RoleUser).Return(&model.Role{ID: 1, Name: entities.RoleUser}, nil).Once()
		repo.On("FindByEmail", mock.Anything, "a@b.com").Return(&model.User{ID: 1}, nil).Once()
		svc := NewUserService(repo, roles, nil, nil, zap.NewNop())
		_, err := svc.Create(ctx, request.CreateUserRequest{Name: "N", Email: "a@b.com", Password: "password1", ConfirmPassword: "password1"})
		assert.ErrorIs(t, err, ErrEmailTaken)
		repo.AssertExpectations(t)
		roles.AssertExpectations(t)
	})

	t.Run("success assigns role", func(t *testing.T) {
		t.Parallel()
		repo := &mockUserRepo{}
		roles := &mockRoleLookup{}
		rbac := &mockUserRBAC{}
		roles.On("FindByName", mock.Anything, entities.RoleUser).Return(&model.Role{ID: 10, Name: entities.RoleUser}, nil).Once()
		repo.On("FindByEmail", mock.Anything, "a@b.com").Return((*model.User)(nil), gorm.ErrRecordNotFound).Once()
		repo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = 42
		}).Once()
		rbac.On("AssignRoleByID", mock.Anything, uint(42), uint(10)).Return(true, nil).Once()

		svc := NewUserService(repo, roles, rbac, nil, zap.NewNop())
		out, err := svc.Create(ctx, request.CreateUserRequest{Name: "N", Email: "A@B.com", Password: "password1", ConfirmPassword: "password1"})
		assert.NoError(t, err)
		assert.Equal(t, uint(42), out.User.ID)
		assert.Equal(t, "a@b.com", out.User.Email)
		assert.Equal(t, uint(10), out.RoleID)
		repo.AssertExpectations(t)
		roles.AssertExpectations(t)
		rbac.AssertExpectations(t)
	})

	t.Run("role id not found", func(t *testing.T) {
		t.Parallel()
		repo := &mockUserRepo{}
		roles := &mockRoleLookup{}
		rid := uint(999)
		roles.On("FindByID", mock.Anything, uint(999)).Return((*model.Role)(nil), gorm.ErrRecordNotFound).Once()
		svc := NewUserService(repo, roles, nil, nil, zap.NewNop())
		_, err := svc.Create(ctx, request.CreateUserRequest{Name: "N", Email: "x@y.com", Password: "password1", ConfirmPassword: "password1", RoleID: &rid})
		assert.ErrorIs(t, err, ErrRoleNotFound)
		roles.AssertExpectations(t)
	})
}
