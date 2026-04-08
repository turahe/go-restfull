package repository

import (
	"context"
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCommentRepository_Create_ListByPostID_PostExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.User{}, &model.CategoryModel{}, &model.Post{}, &model.Comment{}, &model.Tag{}, &model.Media{})
	repo := NewCommentRepository(db, zap.NewNop())

	u := &model.User{Name: "A", Email: "a@b.com", Password: "x"}
	assert.NoError(t, db.WithContext(ctx).Create(u).Error)
	cat := &model.CategoryModel{Name: "Tech", Slug: "tech", Lft: 1, Rgt: 2, Depth: 0, CreatedBy: u.ID, UpdatedBy: u.ID}
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
	assert.NoError(t, db.WithContext(ctx).Create(p).Error)

	exists, err := repo.PostExists(ctx, p.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	cmt, err := repo.CreateRoot(ctx, p.ID, u.ID, "hi", u.ID)
	assert.NoError(t, err)
	assert.NotZero(t, cmt.ID)

	rows, err := repo.ListByPostID(ctx, p.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
}
