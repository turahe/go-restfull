package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestTagRepository_CRUD_SlugExists_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.Tag{})
	repo := NewTagRepository(db)

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

	rows, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	idsRows, err := repo.FindByIDs(ctx, []uint{tag.ID})
	assert.NoError(t, err)
	assert.Len(t, idsRows, 1)

	assert.NoError(t, repo.DeleteByID(ctx, tag.ID))
}

