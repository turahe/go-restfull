package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestCategoryRepository_CRUD_SlugExists_FindByIDs_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.Category{}, &model.Media{})
	repo := NewCategoryRepository(db)

	cat := &model.Category{Name: "Tech", Slug: "tech", CreatedBy: 1, UpdatedBy: 1}
	assert.NoError(t, repo.Create(ctx, cat))
	assert.NotZero(t, cat.ID)

	ok, err := repo.SlugExists(ctx, "tech")
	assert.NoError(t, err)
	assert.True(t, ok)

	got, err := repo.FindBySlug(ctx, "tech")
	assert.NoError(t, err)
	assert.Equal(t, cat.ID, got.ID)

	cat.Name = "Technology"
	assert.NoError(t, repo.Update(ctx, cat))

	got2, err := repo.FindByID(ctx, cat.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Technology", got2.Name)

	rows, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	idsRows, err := repo.FindByIDs(ctx, []uint{cat.ID})
	assert.NoError(t, err)
	assert.Len(t, idsRows, 1)

	assert.NoError(t, repo.SoftDeleteByID(ctx, cat.ID, 1))
}

