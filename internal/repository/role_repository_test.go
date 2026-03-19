package repository

import (
	"context"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRoleRepository_CRUD_FindByName_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.Role{})
	repo := NewRoleRepository(db)

	r := &model.Role{Name: "admin"}
	assert.NoError(t, repo.Create(ctx, r))
	assert.NotZero(t, r.ID)

	got, err := repo.FindByName(ctx, "admin")
	assert.NoError(t, err)
	assert.Equal(t, r.ID, got.ID)

	rows, err := repo.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	got2, err := repo.FindByID(ctx, r.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", got2.Name)

	assert.NoError(t, repo.DeleteByID(ctx, r.ID))
	_, err = repo.FindByID(ctx, r.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

