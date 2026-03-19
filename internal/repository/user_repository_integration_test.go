//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"go-rest/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_Integration_Create_FindByEmail_FindByID_List(t *testing.T) {
	RunWithTx(t, func(tx *gorm.DB) {
		repo := NewUserRepository(tx)
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

		list, err := repo.List(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)
	})
}

func TestUserRepository_Integration_List_LimitClamp(t *testing.T) {
	RunWithTx(t, func(tx *gorm.DB) {
		repo := NewUserRepository(tx)
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			email := fmt.Sprintf("list%d@example.com", i)
			err := repo.Create(ctx, &model.User{Name: "U", Email: email, Password: "x"})
			require.NoError(t, err)
		}

		rows, err := repo.List(ctx, 2)
		require.NoError(t, err)
		assert.Len(t, rows, 2)

		rowsDefault, err := repo.List(ctx, -1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rowsDefault), 3)
	})
}
