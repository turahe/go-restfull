package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/testutil"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BenchmarkUserRepository_Create(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{}, &model.UserMedia{})
	repo := NewUserRepository(db, zap.NewNop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := &model.User{Name: "B", Email: fmt.Sprintf("bench%d@b.com", i), Password: "x"}
		_ = repo.Create(ctx, u)
	}
}

func BenchmarkUserRepository_FindByID(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{}, &model.UserMedia{})
	repo := NewUserRepository(db, zap.NewNop())
	u := &model.User{Name: "B", Email: "findbyid@b.com", Password: "x"}
	if err := repo.Create(ctx, u); err != nil {
		b.Fatal(err)
	}
	id := u.ID
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, id)
	}
}

func BenchmarkUserRepository_FindByEmail(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{}, &model.UserMedia{})
	repo := NewUserRepository(db, zap.NewNop())
	u := &model.User{Name: "B", Email: "findbyemail@b.com", Password: "x"}
	if err := repo.Create(ctx, u); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByEmail(ctx, "findbyemail@b.com")
	}
}

func BenchmarkUserRepository_List(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{}, &model.UserMedia{})
	repo := NewUserRepository(db, zap.NewNop())
	for i := 0; i < 50; i++ {
		_ = repo.Create(ctx, &model.User{Name: "B", Email: fmt.Sprintf("list%d@b.com", i), Password: "x"})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, request.UserListRequest{
			PageRequest: request.PageRequest{Page: 1, Limit: 20},
		})
	}
}

// openBenchDB opens an in-memory SQLite DB for benchmarks.
// Uses a unique name per call so parallel benchmark runs (-cpu N) don't share state.
func openBenchDB(b *testing.B, migrate ...any) *gorm.DB {
	b.Helper()
	dsn := "file:bench_" + uuid.New().String() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(testutil.GormLogLevelFromEnv()),
	})
	if err != nil {
		b.Fatalf("open db: %v", err)
	}
	if len(migrate) > 0 {
		if err := db.AutoMigrate(migrate...); err != nil {
			b.Fatalf("migrate: %v", err)
		}
	}
	return db
}
