package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewRole_Success(t *testing.T) {
	name := "Admin"
	slug := "admin"
	description := "Administrator role"

	role, err := entities.NewRole(name, slug, description)

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, name, role.Name)
	assert.Equal(t, slug, role.Slug)
	assert.Equal(t, description, role.Description)
	assert.True(t, role.IsActive)
	assert.NotEqual(t, uuid.Nil, role.ID)
	assert.False(t, role.CreatedAt.IsZero())
	assert.False(t, role.UpdatedAt.IsZero())
	assert.Nil(t, role.DeletedAt)
}

func TestNewRole_EmptyName(t *testing.T) {
	slug := "admin"
	description := "Administrator role"

	role, err := entities.NewRole("", slug, description)

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role name is required", err.Error())
}

func TestNewRole_EmptySlug(t *testing.T) {
	name := "Admin"
	description := "Administrator role"

	role, err := entities.NewRole(name, "", description)

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Equal(t, "role slug is required", err.Error())
}

func TestNewRole_EmptyDescription(t *testing.T) {
	name := "Admin"
	slug := "admin"

	role, err := entities.NewRole(name, slug, "")

	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, name, role.Name)
	assert.Equal(t, slug, role.Slug)
	assert.Equal(t, "", role.Description)
	assert.True(t, role.IsActive)
}

func TestRole_UpdateRole(t *testing.T) {
	role, _ := entities.NewRole("Old Name", "old-slug", "Old description")
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := role.UpdateRole("New Name", "new-slug", "New description")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", role.Name)
	assert.Equal(t, "new-slug", role.Slug)
	assert.Equal(t, "New description", role.Description)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
}

func TestRole_UpdateRole_PartialUpdate(t *testing.T) {
	role, _ := entities.NewRole("Old Name", "old-slug", "Old description")
	originalName := role.Name
	originalSlug := role.Slug

	err := role.UpdateRole("", "", "New description")

	assert.NoError(t, err)
	assert.Equal(t, originalName, role.Name) // Should remain unchanged
	assert.Equal(t, originalSlug, role.Slug) // Should remain unchanged
	assert.Equal(t, "New description", role.Description)
}

func TestRole_UpdateRole_OnlyName(t *testing.T) {
	role, _ := entities.NewRole("Old Name", "old-slug", "Old description")
	originalSlug := role.Slug

	err := role.UpdateRole("New Name", "", "")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", role.Name)
	assert.Equal(t, originalSlug, role.Slug) // Should remain unchanged
	assert.Equal(t, "", role.Description)    // Description is always set, even if empty
}

func TestRole_UpdateRole_OnlySlug(t *testing.T) {
	role, _ := entities.NewRole("Old Name", "old-slug", "Old description")
	originalName := role.Name

	err := role.UpdateRole("", "new-slug", "")

	assert.NoError(t, err)
	assert.Equal(t, originalName, role.Name) // Should remain unchanged
	assert.Equal(t, "new-slug", role.Slug)
	assert.Equal(t, "", role.Description) // Description is always set, even if empty
}

func TestRole_UpdateRole_EmptyStrings(t *testing.T) {
	role, _ := entities.NewRole("Old Name", "old-slug", "Old description")
	originalName := role.Name
	originalSlug := role.Slug

	err := role.UpdateRole("", "", "")

	assert.NoError(t, err)
	assert.Equal(t, originalName, role.Name)
	assert.Equal(t, originalSlug, role.Slug)
	assert.Equal(t, "", role.Description) // Description is always set, even if empty
}

func TestRole_Activate(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")
	role.IsActive = false
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	role.Activate()

	assert.True(t, role.IsActive)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, role.IsActiveRole())
}

func TestRole_Deactivate(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	role.Deactivate()

	assert.False(t, role.IsActive)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, role.IsActiveRole())
}

func TestRole_SoftDelete(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	role.SoftDelete()

	assert.NotNil(t, role.DeletedAt)
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, role.IsDeleted())
}

func TestRole_IsDeleted(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")

	// Initially not deleted
	assert.False(t, role.IsDeleted())

	// After soft delete
	role.SoftDelete()
	assert.True(t, role.IsDeleted())
}

func TestRole_IsActiveRole(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")

	// Initially active and not deleted
	assert.True(t, role.IsActiveRole())

	// After deactivation
	role.Deactivate()
	assert.False(t, role.IsActiveRole())

	// After reactivation
	role.Activate()
	assert.True(t, role.IsActiveRole())

	// After soft delete
	role.SoftDelete()
	assert.False(t, role.IsActiveRole())
}

func TestRole_SoftDelete_MultipleCalls(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")

	// First soft delete
	role.SoftDelete()
	firstDeletedAt := role.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	role.SoftDelete()
	secondDeletedAt := role.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, role.IsDeleted())
}

func TestRole_Activate_AlreadyActive(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	role.Activate()

	// Should update timestamp even if already active
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, role.IsActive)
}

func TestRole_Deactivate_AlreadyInactive(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")
	role.IsActive = false
	originalUpdatedAt := role.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	role.Deactivate()

	// Should update timestamp even if already inactive
	assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, role.IsActive)
}

func TestRole_StateTransitions(t *testing.T) {
	role, _ := entities.NewRole("Test Role", "test-role", "Test description")

	// Initially active and not deleted
	assert.True(t, role.IsActive)
	assert.False(t, role.IsDeleted())
	assert.True(t, role.IsActiveRole())

	// Deactivate
	role.Deactivate()
	assert.False(t, role.IsActive)
	assert.False(t, role.IsDeleted())
	assert.False(t, role.IsActiveRole())

	// Activate
	role.Activate()
	assert.True(t, role.IsActive)
	assert.False(t, role.IsDeleted())
	assert.True(t, role.IsActiveRole())

	// Soft delete
	role.SoftDelete()
	assert.True(t, role.IsActive) // IsActive remains true
	assert.True(t, role.IsDeleted())
	assert.False(t, role.IsActiveRole()) // But IsActiveRole returns false
}
