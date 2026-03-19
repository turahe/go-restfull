package repository

import (
	"context"
	"sync"
	"testing"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserRepository_Create_FindByEmail_FindByID(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.User{}, &model.Media{}, &model.UserMedia{}, &model.Role{}, &model.UserRole{})
	repo := NewUserRepository(db, zap.NewNop())
	ctx := context.Background()

	u := &model.User{Name: "A", Email: "a@b.com", Password: "x"}
	assert.NoError(t, repo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	gotByEmail, err := repo.FindByEmail(ctx, "a@b.com")
	assert.NoError(t, err)
	assert.Equal(t, u.ID, gotByEmail.ID)

	gotByID, err := repo.FindByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, u.ID, gotByID.ID)
	assert.Equal(t, "a@b.com", gotByID.Email)
}

func TestUserRepository_List_LimitClamp(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.User{}, &model.Media{}, &model.UserMedia{}, &model.Role{}, &model.UserRole{})
	repo := NewUserRepository(db, zap.NewNop())
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		assert.NoError(t, repo.Create(ctx, &model.User{Name: "N", Email: string(rune('a'+i)) + "@b.com", Password: "x"}))
	}

	page, err := repo.List(ctx, request.UserListRequest{Limit: 2, Page: 1})
	assert.NoError(t, err)
	items, ok := page.Items.([]model.User)
	assert.True(t, ok)
	assert.Len(t, items, 2)

	page2, err := repo.List(ctx, request.UserListRequest{Limit: -1, Page: 1})
	assert.NoError(t, err)
	items2, ok := page2.Items.([]model.User)
	assert.True(t, ok)
	assert.Len(t, items2, 3)
}

// TestUserRepository_ConcurrentCreate_SameEmail runs concurrent Creates with the same email.
// Expect exactly one success; others must get a duplicate/unique constraint error.
// Run with: go test -race -run TestUserRepository_ConcurrentCreate_SameEmail ./internal/repository/...
func TestUserRepository_ConcurrentCreate_SameEmail(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.User{}, &model.Media{}, &model.UserMedia{}, &model.Role{}, &model.UserRole{})
	repo := NewUserRepository(db, zap.NewNop())
	ctx := context.Background()
	const concurrency = 20
	email := "concurrent-same@example.com"

	done := make(chan struct{}, concurrency)
	var successCount int
	var successMu sync.Mutex
	for i := 0; i < concurrency; i++ {
		go func() {
			u := &model.User{Name: "Concurrent", Email: email, Password: "x"}
			err := repo.Create(ctx, u)
			if err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < concurrency; i++ {
		<-done
	}
	assert.Equal(t, 1, successCount, "expected exactly one successful Create with same email")
}

