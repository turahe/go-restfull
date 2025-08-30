package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewOrganization_Success(t *testing.T) {
	name := "Test Organization"
	description := "Test organization description"
	code := "TEST-ORG"
	parentID := uuid.New()

	org, err := entities.NewOrganization(name, description, code, entities.OrganizationTypeCompany, &parentID)

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, name, org.Name)
	assert.Equal(t, &description, org.Description)
	assert.Equal(t, &code, org.Code)
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type)
	assert.Equal(t, entities.OrganizationStatusActive, org.Status)
	assert.Equal(t, &parentID, org.ParentID)
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.False(t, org.CreatedAt.IsZero())
	assert.False(t, org.UpdatedAt.IsZero())
	assert.Nil(t, org.DeletedAt)
	assert.Nil(t, org.RecordLeft)
	assert.Nil(t, org.RecordRight)
	assert.Nil(t, org.RecordDepth)
	assert.Nil(t, org.RecordOrdering)
	assert.Empty(t, org.Children)
}

func TestNewOrganization_WithoutParentID(t *testing.T) {
	name := "Test Organization"
	description := "Test organization description"
	code := "TEST-ORG"

	org, err := entities.NewOrganization(name, description, code, entities.OrganizationTypeCompany, nil)

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, name, org.Name)
	assert.Equal(t, &description, org.Description)
	assert.Equal(t, &code, org.Code)
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type)
	assert.Equal(t, entities.OrganizationStatusActive, org.Status)
	assert.Nil(t, org.ParentID)
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.False(t, org.CreatedAt.IsZero())
	assert.False(t, org.UpdatedAt.IsZero())
	assert.Nil(t, org.DeletedAt)
	assert.Nil(t, org.RecordLeft)
	assert.Nil(t, org.RecordRight)
	assert.Nil(t, org.RecordDepth)
	assert.Nil(t, org.RecordOrdering)
	assert.Empty(t, org.Children)
}

func TestNewOrganization_EmptyName(t *testing.T) {
	description := "Test organization description"
	code := "TEST-ORG"

	org, err := entities.NewOrganization("", description, code, entities.OrganizationTypeCompany, nil)

	assert.Error(t, err)
	assert.Nil(t, org)
	assert.Equal(t, "name is required", err.Error())
}

func TestNewOrganization_EmptyOptionalFields(t *testing.T) {
	name := "Test Organization"

	org, err := entities.NewOrganization(name, "", "", "", nil)

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, name, org.Name)
	assert.Nil(t, org.Description)
	assert.Nil(t, org.Code)
	assert.Nil(t, org.Type)
	assert.Equal(t, entities.OrganizationStatusActive, org.Status)
	assert.Nil(t, org.ParentID)
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.False(t, org.CreatedAt.IsZero())
	assert.False(t, org.UpdatedAt.IsZero())
	assert.Nil(t, org.DeletedAt)
}

func TestOrganization_UpdateOrganization(t *testing.T) {
	org, _ := entities.NewOrganization("Old Name", "Old description", "OLD-CODE", entities.OrganizationTypeCompany, nil)
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := org.UpdateOrganization("New Name", "New description", "NEW-CODE", entities.OrganizationTypeCompany)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)
	assert.Equal(t, "New description", *org.Description)
	assert.Equal(t, "NEW-CODE", *org.Code)
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))
}

func TestOrganization_UpdateOrganization_PartialUpdate(t *testing.T) {
	org, _ := entities.NewOrganization("Test Name", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalName := org.Name
	originalDescription := org.Description
	originalCode := org.Code

	err := org.UpdateOrganization("", "", "", entities.OrganizationTypeCompany)

	assert.NoError(t, err)
	assert.Equal(t, originalName, org.Name)                      // Should remain unchanged
	assert.Equal(t, originalDescription, org.Description)        // Should remain unchanged
	assert.Equal(t, originalCode, org.Code)                      // Should remain unchanged
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type) // Should be updated
	assert.Equal(t, originalCode, org.Code)                      // Should remain unchanged
}

func TestOrganization_SetStatus(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	err := org.SetStatus(entities.OrganizationStatusInactive)

	assert.NoError(t, err)
	assert.Equal(t, entities.OrganizationStatusInactive, org.Status)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))
}

func TestOrganization_SetStatus_InvalidStatus(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalStatus := org.Status

	err := org.SetStatus("invalid-status")

	assert.Error(t, err)
	assert.Equal(t, "invalid organization status", err.Error())
	assert.Equal(t, originalStatus, org.Status) // Should remain unchanged
}

func TestOrganization_SetParent(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	newParentID := uuid.New()
	org.SetParent(&newParentID)

	assert.Equal(t, &newParentID, org.ParentID)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))
}

func TestOrganization_SetParent_Nil(t *testing.T) {
	parentID := uuid.New()
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, &parentID)
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	org.SetParent(nil)

	assert.Nil(t, org.ParentID)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))
}

func TestOrganization_SoftDelete(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	org.SoftDelete()

	assert.NotNil(t, org.DeletedAt)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, org.IsDeleted())
}

func TestOrganization_IsDeleted(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)

	// Initially not deleted
	assert.False(t, org.IsDeleted())

	// After soft delete
	org.SoftDelete()
	assert.True(t, org.IsDeleted())
}

func TestOrganization_IsActive(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)

	// Initially active
	assert.True(t, org.IsActive())

	// Set to inactive
	err := org.SetStatus(entities.OrganizationStatusInactive)
	assert.NoError(t, err)
	assert.False(t, org.IsActive())

	// Set to suspended
	err = org.SetStatus(entities.OrganizationStatusSuspended)
	assert.NoError(t, err)
	assert.False(t, org.IsActive())

	// Set back to active
	err = org.SetStatus(entities.OrganizationStatusActive)
	assert.NoError(t, err)
	assert.True(t, org.IsActive())
}

func TestOrganization_IsRoot(t *testing.T) {
	// Root organization (no parent)
	root, _ := entities.NewOrganization("Root Organization", "Root description", "ROOT-CODE", entities.OrganizationTypeCompany, nil)
	assert.True(t, root.IsRoot())

	// Child organization (has parent)
	parentID := uuid.New()
	child, _ := entities.NewOrganization("Child Organization", "Child description", "CHILD-CODE", entities.OrganizationTypeCompany, &parentID)
	assert.False(t, child.IsRoot())
}

func TestOrganization_SoftDelete_MultipleCalls(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)

	// First soft delete
	org.SoftDelete()
	firstDeletedAt := org.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	org.SoftDelete()
	secondDeletedAt := org.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, org.IsDeleted())
}

func TestOrganization_UpdateOrganization_EmptyFields(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalName := org.Name
	originalDescription := org.Description
	originalCode := org.Code

	err := org.UpdateOrganization("", "", "", entities.OrganizationTypeCompany)

	assert.NoError(t, err)
	assert.Equal(t, originalName, org.Name)               // Should remain unchanged
	assert.Equal(t, originalDescription, org.Description) // Should remain unchanged
	assert.Equal(t, originalCode, org.Code)               // Should remain unchanged
}

func TestOrganization_UpdateOrganization_OnlyName(t *testing.T) {
	org, _ := entities.NewOrganization("Old Name", "Old description", "OLD-CODE", entities.OrganizationTypeCompany, nil)
	originalDescription := org.Description
	originalCode := org.Code

	err := org.UpdateOrganization("New Name", "", "", entities.OrganizationTypeCompany)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)
	assert.Equal(t, originalDescription, org.Description)        // Should remain unchanged
	assert.Equal(t, originalCode, org.Code)                      // Should remain unchanged
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type) // Should remain unchanged
}

func TestOrganization_UpdateOrganization_OnlyEmail(t *testing.T) {
	org, _ := entities.NewOrganization("Test Name", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)
	originalName := org.Name
	originalDescription := org.Description
	originalCode := org.Code

	err := org.UpdateOrganization("", "", "", entities.OrganizationTypeCompany)

	assert.NoError(t, err)
	assert.Equal(t, originalName, org.Name)               // Should remain unchanged
	assert.Equal(t, originalDescription, org.Description) // Should remain unchanged
	assert.Equal(t, originalCode, org.Code)               // Should remain unchanged
	assert.Equal(t, entities.OrganizationTypeCompany, *org.Type)
}

func TestOrganization_SetStatus_AllStatuses(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)

	// Test all valid statuses
	statuses := []entities.OrganizationStatus{
		entities.OrganizationStatusActive,
		entities.OrganizationStatusInactive,
		entities.OrganizationStatusSuspended,
	}

	for _, status := range statuses {
		err := org.SetStatus(status)
		assert.NoError(t, err)
		assert.Equal(t, status, org.Status)
	}
}

func TestOrganization_StatusTransitions(t *testing.T) {
	org, _ := entities.NewOrganization("Test Organization", "Test description", "TEST-CODE", entities.OrganizationTypeCompany, nil)

	// Initially active
	assert.True(t, org.IsActive())
	assert.Equal(t, entities.OrganizationStatusActive, org.Status)

	// Set to inactive
	err := org.SetStatus(entities.OrganizationStatusInactive)
	assert.NoError(t, err)
	assert.False(t, org.IsActive())
	assert.Equal(t, entities.OrganizationStatusInactive, org.Status)

	// Set to suspended
	err = org.SetStatus(entities.OrganizationStatusSuspended)
	assert.NoError(t, err)
	assert.False(t, org.IsActive())
	assert.Equal(t, entities.OrganizationStatusSuspended, org.Status)

	// Set back to active
	err = org.SetStatus(entities.OrganizationStatusActive)
	assert.NoError(t, err)
	assert.True(t, org.IsActive())
	assert.Equal(t, entities.OrganizationStatusActive, org.Status)
}
