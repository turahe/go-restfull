package repository

import (
	"context"
	"testing"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/model"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestRoleRepository_CRUD_FindByName_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := openTestDB(t, &model.Role{})
	repo := NewRoleRepository(db, zap.NewNop())

	r := &model.Role{Name: "admin"}
	assert.NoError(t, repo.Create(ctx, r))
	assert.NotZero(t, r.ID)

	got, err := repo.FindByName(ctx, "admin")
	assert.NoError(t, err)
	assert.Equal(t, r.ID, got.ID)

	page, err := repo.List(ctx, request.RoleListRequest{
		PageRequest: request.PageRequest{Page: 1, Limit: 10},
	})
	assert.NoError(t, err)
	items, ok := page.Items.([]model.Role)
	assert.True(t, ok)
	assert.Len(t, items, 1)

	got2, err := repo.FindByID(ctx, r.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", got2.Name)

	assert.NoError(t, repo.DeleteByID(ctx, r.ID))
	_, err = repo.FindByID(ctx, r.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
