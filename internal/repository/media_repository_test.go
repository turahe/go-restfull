package repository

import (
	"context"
	"testing"
	"time"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMediaRepository_Create_List_Find_Attach_And_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.User{}, &model.Category{}, &model.Post{}, &model.Comment{}, &model.Media{}, &model.Tag{})

	// Create join tables used by SoftDeleteByID raw SQL deletes (SQLite won't have them unless created).
	// Create tables minimally with media_id column.
	assert.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS post_media (post_id integer, media_id integer)").Error)
	assert.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS user_media (user_id integer, media_id integer)").Error)
	assert.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS category_media (category_id integer, media_id integer)").Error)
	assert.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS comment_media (comment_id integer, media_id integer)").Error)

	repo := NewMediaRepository(db, zap.NewNop())

	u := &model.User{Name: "A", Email: "a@b.com", Password: "x"}
	assert.NoError(t, db.WithContext(ctx).Create(u).Error)
	cat := &model.Category{Name: "Tech", Slug: "tech", CreatedBy: u.ID, UpdatedBy: u.ID}
	assert.NoError(t, db.WithContext(ctx).Create(cat).Error)
	p := &model.Post{Title: "T", Slug: "t", Content: "c", UserID: u.ID, CategoryID: cat.ID, CreatedBy: u.ID, UpdatedBy: u.ID, CreatedAt: time.Now()}
	assert.NoError(t, db.WithContext(ctx).Create(p).Error)
	cmt := &model.Comment{PostID: p.ID, UserID: u.ID, Content: "hi", CreatedBy: u.ID, UpdatedBy: u.ID}
	assert.NoError(t, db.WithContext(ctx).Create(cmt).Error)

	m := &model.Media{
		UserID:        u.ID,
		MediaType:     "image",
		OriginalName:  "a.png",
		MimeType:      "image/png",
		Size:          10,
		StoragePath:   "x",
		CreatedBy:     u.ID,
		UpdatedBy:     u.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	assert.NoError(t, repo.Create(ctx, m))
	assert.NotZero(t, m.ID)

	assert.NoError(t, repo.AttachMedia(ctx, m.ID, "Post", p.ID))

	rows, err := repo.ListByUserID(ctx, u.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	got, err := repo.FindByIDAndUserID(ctx, m.ID, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, m.ID, got.ID)

	assert.NoError(t, repo.SoftDeleteByID(ctx, m.ID, u.ID, u.ID))
}

