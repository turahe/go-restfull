package repository

import (
	"context"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTagRepository_CRUD_SlugExists_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.Tag{})
	repo := NewTagRepository(db, zap.NewNop())

	tag := &model.Tag{Name: "Go", Slug: "go"}
	assert.NoError(t, repo.Create(ctx, tag))
	assert.NotZero(t, tag.ID)

	ok, err := repo.SlugExists(ctx, "go")
	assert.NoError(t, err)
	assert.True(t, ok)

	got, err := repo.FindBySlug(ctx, "go")
	assert.NoError(t, err)
	assert.Equal(t, tag.ID, got.ID)

	tag.Name = "Golang"
	assert.NoError(t, repo.Update(ctx, tag))

	got2, err := repo.FindByID(ctx, tag.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Golang", got2.Name)

	page, err := repo.List(ctx, request.TagListRequest{Limit: 10, Name: "", Page: 1})
	assert.NoError(t, err)
	items, ok := page.Items.([]model.Tag)
	assert.True(t, ok)
	assert.Len(t, items, 1)

	idsRows, err := repo.FindByIDs(ctx, []uint{tag.ID})
	assert.NoError(t, err)
	assert.Len(t, idsRows, 1)

	assert.NoError(t, repo.DeleteByID(ctx, tag.ID))
}

