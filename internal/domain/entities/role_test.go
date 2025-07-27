package entities_test

import (
	"testing"
	"time"

	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRole(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		slug        string
		description string
		wantErr     bool
	}{
		{
			name:        "Valid role creation",
			roleName:    "Admin",
			slug:        "admin",
			description: "Administrator role",
			wantErr:     false,
		},
		{
			name:        "Empty name",
			roleName:    "",
			slug:        "admin",
			description: "Administrator role",
			wantErr:     true,
		},
		{
			name:        "Empty slug",
			roleName:    "Admin",
			slug:        "",
			description: "Administrator role",
			wantErr:     true,
		},
		{
			name:        "Empty description",
			roleName:    "Admin",
			slug:        "admin",
			description: "",
			wantErr:     false, // Description is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := entities.NewRole(tt.roleName, tt.slug, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, tt.roleName, role.Name)
				assert.Equal(t, tt.slug, role.Slug)
				assert.Equal(t, tt.description, role.Description)
				assert.True(t, role.IsActive)
				assert.NotEqual(t, uuid.Nil, role.ID)
				assert.False(t, role.CreatedAt.IsZero())
				assert.False(t, role.UpdatedAt.IsZero())
			}
		})
	}
}

func TestRole_UpdateRole(t *testing.T) {
	role, err := entities.NewRole("OldRole", "old-role", "Old description")
	require.NoError(t, err)
	require.NotNil(t, role)

	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	tests := []struct {
		name        string
		roleName    string
		slug        string
		description string
	}{
		{
			name:        "Update all fields",
			roleName:    "NewRole",
			slug:        "new-role",
			description: "New description",
		},
		{
			name:        "Update only name",
			roleName:    "UpdatedRole",
			slug:        "",
			description: "",
		},
		{
			name:        "Update only slug",
			roleName:    "",
			slug:        "updated-slug",
			description: "",
		},
		{
			name:        "Update only description",
			roleName:    "",
			slug:        "",
			description: "Updated description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := role.UpdateRole(tt.roleName, tt.slug, tt.description)
			assert.NoError(t, err)

			if tt.roleName != "" {
				assert.Equal(t, tt.roleName, role.Name)
			}
			if tt.slug != "" {
				assert.Equal(t, tt.slug, role.Slug)
			}
			if tt.description != "" {
				assert.Equal(t, tt.description, role.Description)
			}

			assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
		})
	}
}

func TestRole_Activate(t *testing.T) {
	role, err := entities.NewRole("TestRole", "test-role", "Test description")
	require.NoError(t, err)
	require.NotNil(t, role)

	// Deactivate first
	role.Deactivate()
	assert.False(t, role.IsActive)

	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	role.Activate()

	assert.True(t, role.IsActive)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, role.IsActiveRole())
}

func TestRole_Deactivate(t *testing.T) {
	role, err := entities.NewRole("TestRole", "test-role", "Test description")
	require.NoError(t, err)
	require.NotNil(t, role)

	// Initially active
	assert.True(t, role.IsActive)

	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	role.Deactivate()

	assert.False(t, role.IsActive)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, role.IsActiveRole())
}

func TestRole_SoftDelete(t *testing.T) {
	role, err := entities.NewRole("TestRole", "test-role", "Test description")
	require.NoError(t, err)
	require.NotNil(t, role)

	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	role.SoftDelete()

	assert.NotNil(t, role.DeletedAt)
	assert.True(t, role.DeletedAt.After(originalUpdatedAt))
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, role.IsDeleted())
}

func TestRole_IsDeleted(t *testing.T) {
	role, err := entities.NewRole("TestRole", "test-role", "Test description")
	require.NoError(t, err)
	require.NotNil(t, role)

	// Initially not deleted
	assert.False(t, role.IsDeleted())

	// After soft delete
	role.SoftDelete()
	assert.True(t, role.IsDeleted())
}

func TestRole_IsActiveRole(t *testing.T) {
	role, err := entities.NewRole("TestRole", "test-role", "Test description")
	require.NoError(t, err)
	require.NotNil(t, role)

	// Initially active and not deleted
	assert.True(t, role.IsActiveRole())

	// After deactivation
	role.Deactivate()
	assert.False(t, role.IsActiveRole())

	// Reactivate
	role.Activate()
	assert.True(t, role.IsActiveRole())

	// After soft delete
	role.SoftDelete()
	assert.False(t, role.IsActiveRole())
}
