package service

import (
	"context"
	"errors"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) List(ctx context.Context, limit int) ([]model.User, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]model.User), args.Error(1)
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
			svc := NewUserService(repo)

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
	repo.On("List", mock.Anything, 10).Return([]model.User{{ID: 1}}, nil).Once()

	svc := NewUserService(repo)
	rows, err := svc.List(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, uint(1), rows[0].ID)
	repo.AssertExpectations(t)
}

