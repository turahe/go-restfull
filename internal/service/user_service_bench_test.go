package service

import (
	"context"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func BenchmarkUserService_GetByID(b *testing.B) {
	ctx := context.Background()
	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, uint(123)).Return(&model.User{ID: 123, Email: "a@b.com", Name: "A"}, nil)
	svc := NewUserService(repo, nil, nil, nil, zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetByID(ctx, 123)
	}
}

func BenchmarkUserService_List(b *testing.B) {
	ctx := context.Background()
	repo := &mockUserRepo{}
	users := make([]model.User, 20)
	for i := range users {
		users[i] = model.User{ID: uint(i + 1), Email: "a@b.com", Name: "A"}
	}
	listReq := request.UserListRequest{
		PageRequest: request.PageRequest{Limit: 20},
	}
	repo.On("List", mock.Anything, listReq).Return(repository.CursorPage{
		Items: users,
	}, nil)
	svc := NewUserService(repo, nil, nil, nil, zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.List(ctx, listReq)
	}
}
