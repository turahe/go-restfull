package repository

import (
	"context"
	"testing"
	"time"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestCommentRepository_Create_ListByPostID_PostExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.User{}, &model.Category{}, &model.Post{}, &model.Comment{}, &model.Tag{}, &model.Media{})
	repo := NewCommentRepository(db)

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
	assert.NoError(t, db.WithContext(ctx).Create(p).Error)

	exists, err := repo.PostExists(ctx, p.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	cmt := &model.Comment{
		PostID:    p.ID,
		UserID:    u.ID,
		Content:   "hi",
		CreatedBy: u.ID,
		UpdatedBy: u.ID,
	}
	assert.NoError(t, repo.Create(ctx, cmt))
	assert.NotZero(t, cmt.ID)

	rows, err := repo.ListByPostID(ctx, p.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)
}

