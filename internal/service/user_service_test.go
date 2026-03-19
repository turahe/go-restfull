package service

import (
	"context"
	"errors"
	"testing"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"
	"go-rest/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"go.uber.org/zap"
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
			svc := NewUserService(repo, nil, zap.NewNop())

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
	repo.On("List", mock.Anything, request.UserListRequest{Limit: 10}).Return(repository.CursorPage{
		Items: []model.User{{ID: 1}},
	}, nil).Once()

	svc := NewUserService(repo, nil, zap.NewNop())
	page, err := svc.List(ctx, request.UserListRequest{Limit: 10})

	assert.NoError(t, err)
	items, ok := page.Items.([]model.User)
	assert.True(t, ok)
	assert.Len(t, items, 1)
	assert.Equal(t, uint(1), items[0].ID)
	repo.AssertExpectations(t)
}

