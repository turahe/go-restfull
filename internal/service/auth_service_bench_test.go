package service

import (
	"context"
	"testing"
	"time"

	"go-rest/internal/model"
	"go-rest/internal/service/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

func BenchmarkAuthService_Register(b *testing.B) {
	ctx := context.Background()
	users := &mockAuthUserRepo{}
	users.On("FindByEmail", mock.Anything, "bench@example.com").Return((*model.User)(nil), gorm.ErrRecordNotFound)
	users.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(1).(*model.User)
		u.ID = 1
	})
	// nil rbac so we don't benchmark AssignRole
	svc := NewAuthService(users, &mockAuthRepo{}, nil, nil, &mockJWT{}, nil, 10, 30, 5, "pepper", zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Register(ctx, "Bench", "bench@example.com", "password123")
	}
}

func BenchmarkAuthService_Login(b *testing.B) {
	ctx := context.Background()
	hash, err := bcryptHash("password")
	if err != nil {
		b.Fatal(err)
	}
	users := &mockAuthUserRepo{}
	users.On("FindByEmail", mock.Anything, "login@example.com").Return(&model.User{ID: 1, Email: "login@example.com", Password: hash, Name: "U"}, nil)
	authRepo := &mockAuthRepo{}
	authRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.AuthSession")).Return(nil)
	authRepo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(nil)
	rbac := &mockRBAC{}
	rbac.On("RolesForUser", mock.Anything, uint(1)).Return([]string{"user"}, nil)
	rbac.On("PermissionsForUser", mock.Anything, uint(1)).Return([]string{}, nil)
	j := &mockJWT{}
	j.On("DefaultRegistered", "1", 10*time.Minute).Return(jwt.RegisteredClaims{})
	j.On("IssueAccessToken", mock.AnythingOfType("dto.AccessClaims")).Return("token", nil)
	svc := NewAuthService(users, authRepo, nil, rbac, j, nil, 10, 30, 5, "pepper", zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Login(ctx, "login@example.com", "password", dto.LoginMeta{DeviceID: "dev1"})
	}
}
