//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"go-rest/internal/handler/request"
	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestUserRepository_Integration_Create_FindByEmail_FindByID_List(t *testing.T) {
	RunWithTx(t, func(tx *gorm.DB) {
		repo := NewUserRepository(tx, zap.NewNop())
		ctx := context.Background()

		u := &model.User{Name: "Integration User", Email: "int@example.com", Password: "secret"}
		err := repo.Create(ctx, u)
		require.NoError(t, err)
		assert.NotZero(t, u.ID)

		byEmail, err := repo.FindByEmail(ctx, "int@example.com")
		require.NoError(t, err)
		assert.Equal(t, u.ID, byEmail.ID)
		assert.Equal(t, "Integration User", byEmail.Name)

		byID, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		assert.Equal(t, u.ID, byID.ID)
		assert.Equal(t, "int@example.com", byID.Email)

		page, err := repo.List(ctx, request.UserListRequest{Limit: 10, Page: 1})
		require.NoError(t, err)
		items, ok := page.Items.([]model.User)
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(items), 1)
	})
}

func TestUserRepository_Integration_List_LimitClamp(t *testing.T) {
	RunWithTx(t, func(tx *gorm.DB) {
		repo := NewUserRepository(tx, zap.NewNop())
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			email := fmt.Sprintf("list%d@example.com", i)
			err := repo.Create(ctx, &model.User{Name: "U", Email: email, Password: "x"})
			require.NoError(t, err)
		}

		page, err := repo.List(ctx, request.UserListRequest{Limit: 2, Page: 1})
		require.NoError(t, err)
		items, ok := page.Items.([]model.User)
		require.True(t, ok)
		assert.Len(t, items, 2)

		pageDefault, err := repo.List(ctx, request.UserListRequest{Limit: -1, Page: 1})
		require.NoError(t, err)
		itemsDefault, ok := pageDefault.Items.([]model.User)
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(itemsDefault), 3)
	})
}
