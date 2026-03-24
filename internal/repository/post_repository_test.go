package repository

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPostRepository_CRUD_SlugExists_ListCursor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.User{}, &model.Category{}, &model.Post{}, &model.Media{})
	repo := NewPostRepository(db, zap.NewNop())

	u := &model.User{Name: "A", Email: "a@b.com", Password: "x"}
	assert.NoError(t, db.WithContext(ctx).Create(u).Error)
	cat := &model.Category{Name: "Tech", Slug: "tech", CreatedBy: u.ID, UpdatedBy: u.ID}
	assert.NoError(t, db.WithContext(ctx).Create(cat).Error)

	p := &model.Post{
		Title:      "T",
		Slug:       "t",
		Content:    "c",
		UserID:     u.ID,
		CategoryID: cat.ID,
		CreatedBy:  u.ID,
		UpdatedBy:  u.ID,
		CreatedAt:  time.Now(),
	}
	assert.NoError(t, repo.Create(ctx, p))
	assert.NotZero(t, p.ID)

	ok, err := repo.SlugExists(ctx, "t")
	assert.NoError(t, err)
	assert.True(t, ok)

	got, err := repo.FindBySlug(ctx, "t")
	assert.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)

	page, err := repo.ListCursor(ctx, request.PostListRequest{Limit: 10, Page: 1})
	assert.NoError(t, err)
	items, ok := page.Items.([]model.Post)
	assert.True(t, ok)
	assert.Len(t, items, 1)
}

