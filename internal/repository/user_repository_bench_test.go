package repository

import (
	"context"
	"fmt"
	"testing"

	"go-rest/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func BenchmarkUserRepository_Create(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := &model.User{Name: "B", Email: fmt.Sprintf("bench%d@b.com", i), Password: "x"}
		_ = repo.Create(ctx, u)
	}
}

func BenchmarkUserRepository_FindByID(b *testing.B) {
	ctx := context.Background()
	db := openBenchDB(b, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
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
	db := openBenchDB(b, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
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
	db := openBenchDB(b, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
	for i := 0; i < 50; i++ {
		_ = repo.Create(ctx, &model.User{Name: "B", Email: fmt.Sprintf("list%d@b.com", i), Password: "x"})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, 20)
	}
}

// openBenchDB opens an in-memory SQLite DB for benchmarks.
// Uses a unique name per call so parallel benchmark runs (-cpu N) don't share state.
func openBenchDB(b *testing.B, migrate ...any) *gorm.DB {
	b.Helper()
	dsn := "file:bench_" + uuid.New().String() + "?mode=memory&cache=private"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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
