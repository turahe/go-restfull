package repository

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BenchmarkAuthRepository_CreateSession(b *testing.B) {
	ctx := context.Background()
	db := openAuthBenchDB(b)
	repo := NewAuthRepository(db, zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := &model.AuthSession{
			ID:         uuid.New().String(),
			UserID:     1,
			DeviceID:   "dev",
			IPAddress:  "127.0.0.1",
			UserAgent:  "bench",
			LastSeenAt: time.Now(),
		}
		_ = repo.CreateSession(ctx, s)
	}
}

func BenchmarkAuthRepository_SessionActive(b *testing.B) {
	ctx := context.Background()
	db := openAuthBenchDB(b)
	repo := NewAuthRepository(db, zap.NewNop())
	s := &model.AuthSession{
		ID:         "sess-active",
		UserID:     1,
		DeviceID:   "dev",
		IPAddress:  "127.0.0.1",
		UserAgent:  "bench",
		LastSeenAt: time.Now(),
	}
	if err := repo.CreateSession(ctx, s); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.SessionActive(ctx, "sess-active")
	}
}

func BenchmarkAuthRepository_FindRefreshTokenByHash(b *testing.B) {
	ctx := context.Background()
	db := openAuthBenchDB(b)
	repo := NewAuthRepository(db, zap.NewNop())
	hash := "hash-" + uuid.New().String()
	rt := &model.RefreshToken{
		SessionID:   uuid.New().String(),
		UserID:      1,
		TokenHash:   hash,
		TokenFamily: uuid.New().String(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	if err := repo.CreateRefreshToken(ctx, rt); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindRefreshTokenByHash(ctx, hash)
	}
}

func openAuthBenchDB(b *testing.B) *gorm.DB {
	b.Helper()
	name := "file:authbench_" + uuid.New().String() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	if err != nil {
		b.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuthSession{}, &model.RefreshToken{}, &model.RevokedJTI{}); err != nil {
		b.Fatalf("migrate: %v", err)
	}
	return db
}
