package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create_FindByEmail_FindByID(t *testing.T) {
	t.Parallel()
	db := openTestDB(t, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
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
	db := openTestDB(t, &model.User{}, &model.Media{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		assert.NoError(t, repo.Create(ctx, &model.User{Name: "N", Email: string(rune('a'+i)) + "@b.com", Password: "x"}))
	}

	rows, err := repo.List(ctx, 2)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)

	rows2, err := repo.List(ctx, -1)
	assert.NoError(t, err)
	assert.Len(t, rows2, 3)
}

