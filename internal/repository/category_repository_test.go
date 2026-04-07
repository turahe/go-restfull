package repository

import (
	"context"
	"testing"

	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const testActor = uint(42)

func TestCategoryRepository_CreateRoot_FirstAndSecond(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()

	a, err := repo.CreateRoot(ctx, "A", testActor)
	require.NoError(t, err)
	assert.Equal(t, 1, a.Lft)
	assert.Equal(t, 2, a.Rgt)
	assert.Equal(t, 0, a.Depth)
	assert.Nil(t, a.ParentID)
	assert.Equal(t, testActor, a.CreatedBy)

	b, err := repo.CreateRoot(ctx, "B", testActor)
	require.NoError(t, err)
	assert.Equal(t, 3, b.Lft)
	assert.Equal(t, 4, b.Rgt)
}

func TestCategoryRepository_CreateRoot_DuplicateName(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()
	_, err := repo.CreateRoot(ctx, "Only", testActor)
	require.NoError(t, err)
	_, err = repo.CreateRoot(ctx, "Only", testActor)
	assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

func TestCategoryRepository_CreateChild_ShiftAndNested(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()

	root, err := repo.CreateRoot(ctx, "Root", testActor)
	require.NoError(t, err)

	child, err := repo.CreateChild(ctx, root.ID, "Child", testActor)
	require.NoError(t, err)
	assert.Equal(t, root.ID, *child.ParentID)
	assert.Equal(t, 2, child.Lft)
	assert.Equal(t, 3, child.Rgt)
	assert.Equal(t, 1, child.Depth)

	var r model.CategoryModel
	require.NoError(t, db.First(&r, root.ID).Error)
	assert.Equal(t, 1, r.Lft)
	assert.Equal(t, 4, r.Rgt)
}

func TestCategoryRepository_GetTree_GetSubtree_FindByIDs(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()

	root, _ := repo.CreateRoot(ctx, "R", testActor)
	_, _ = repo.CreateChild(ctx, root.ID, "C1", testActor)
	c2, _ := repo.CreateChild(ctx, root.ID, "C2", testActor)

	all, err := repo.GetTree(ctx)
	require.NoError(t, err)
	require.Len(t, all, 3)

	sub, err := repo.GetSubtree(ctx, root.ID)
	require.NoError(t, err)
	require.Len(t, sub, 3)

	sub2, err := repo.GetSubtree(ctx, c2.ID)
	require.NoError(t, err)
	require.Len(t, sub2, 1)
	assert.Equal(t, "C2", sub2[0].Name)

	by, err := repo.FindByIDs(ctx, []uint{root.ID})
	require.NoError(t, err)
	require.Len(t, by, 1)
}

func TestCategoryRepository_FindByIDs_Empty(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	_, err := repo.FindByIDs(context.Background(), nil)
	assert.Error(t, err)
}

func TestCategoryRepository_UpdateName(t *testing.T) {
	db := openTestDB(t, &model.CategoryModel{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()
	r, err := repo.CreateRoot(ctx, "A", testActor)
	require.NoError(t, err)
	up, err := repo.UpdateName(ctx, r.ID, "Alpha", testActor)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", up.Name)
	assert.Equal(t, testActor, up.UpdatedBy)
}

func TestCategoryRepository_DeleteSubtree_EmptyTree(t *testing.T) {
	db := openTestDB(t, &model.User{}, &model.CategoryModel{}, &model.Post{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()
	r, err := repo.CreateRoot(ctx, "Root", testActor)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteSubtree(ctx, r.ID, testActor))
	var n int64
	require.NoError(t, db.Model(&model.CategoryModel{}).Count(&n).Error)
	assert.Equal(t, int64(0), n)
	var unscoped int64
	require.NoError(t, db.Unscoped().Model(&model.CategoryModel{}).Count(&unscoped).Error)
	assert.Equal(t, int64(1), unscoped)
}

func TestCategoryRepository_DeleteSubtree_BlockedByPost(t *testing.T) {
	db := openTestDB(t, &model.User{}, &model.CategoryModel{}, &model.Post{})
	repo := NewCategoryRepository(db, zap.NewNop())
	ctx := context.Background()

	u := &model.User{Name: "U", Email: "u@x.com", Password: "x"}
	require.NoError(t, db.WithContext(ctx).Create(u).Error)
	cat, err := repo.CreateRoot(ctx, "Cat", testActor)
	require.NoError(t, err)
	p := &model.Post{
		Title: "T", Slug: "s", Content: "c",
		UserID: u.ID, CategoryID: cat.ID, CreatedBy: u.ID, UpdatedBy: u.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)

	err = repo.DeleteSubtree(ctx, cat.ID, testActor)
	assert.ErrorIs(t, err, ErrCategorySubtreeHasPosts)
}
