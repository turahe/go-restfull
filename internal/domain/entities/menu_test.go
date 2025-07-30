package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMenu_Success(t *testing.T) {
	name := "Dashboard"
	slug := "dashboard"
	description := "Main dashboard menu"
	url := "/dashboard"
	icon := "dashboard-icon"
	recordOrdering := int64(1)
	parentID := uuid.New()

	menu := entities.NewMenu(name, slug, description, url, icon, recordOrdering, &parentID)

	assert.NotNil(t, menu)
	assert.Equal(t, name, menu.Name)
	assert.Equal(t, slug, menu.Slug)
	assert.Equal(t, description, menu.Description)
	assert.Equal(t, url, menu.URL)
	assert.Equal(t, icon, menu.Icon)
	assert.Equal(t, &parentID, menu.ParentID)
	assert.Equal(t, recordOrdering, menu.RecordOrdering)
	assert.NotEqual(t, uuid.Nil, menu.ID)
	assert.False(t, menu.CreatedAt.IsZero())
	assert.False(t, menu.UpdatedAt.IsZero())
	assert.Nil(t, menu.DeletedAt)
	assert.Equal(t, int64(0), menu.RecordLeft)
	assert.Equal(t, int64(0), menu.RecordRight)
	assert.True(t, menu.IsActive)
	assert.True(t, menu.IsVisible)
	assert.Equal(t, "_self", menu.Target)
	assert.Empty(t, menu.Children)
	assert.Empty(t, menu.Roles)
}

func TestNewMenu_WithoutParentID(t *testing.T) {
	name := "Dashboard"
	slug := "dashboard"
	description := "Main dashboard menu"
	url := "/dashboard"
	icon := "dashboard-icon"
	recordOrdering := int64(1)

	menu := entities.NewMenu(name, slug, description, url, icon, recordOrdering, nil)

	assert.NotNil(t, menu)
	assert.Equal(t, name, menu.Name)
	assert.Equal(t, slug, menu.Slug)
	assert.Equal(t, description, menu.Description)
	assert.Equal(t, url, menu.URL)
	assert.Equal(t, icon, menu.Icon)
	assert.Nil(t, menu.ParentID)
	assert.Equal(t, recordOrdering, menu.RecordOrdering)
	assert.NotEqual(t, uuid.Nil, menu.ID)
	assert.False(t, menu.CreatedAt.IsZero())
	assert.False(t, menu.UpdatedAt.IsZero())
	assert.Nil(t, menu.DeletedAt)
	assert.Equal(t, int64(0), menu.RecordLeft)
	assert.Equal(t, int64(0), menu.RecordRight)
	assert.True(t, menu.IsActive)
	assert.True(t, menu.IsVisible)
	assert.Equal(t, "_self", menu.Target)
	assert.Empty(t, menu.Children)
	assert.Empty(t, menu.Roles)
}

func TestMenu_UpdateMenu(t *testing.T) {
	menu := entities.NewMenu("Old Name", "old-slug", "Old description", "/old-url", "old-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	newParentID := uuid.New()
	menu.UpdateMenu("New Name", "new-slug", "New description", "/new-url", "new-icon", 2, &newParentID)

	assert.Equal(t, "New Name", menu.Name)
	assert.Equal(t, "new-slug", menu.Slug)
	assert.Equal(t, "New description", menu.Description)
	assert.Equal(t, "/new-url", menu.URL)
	assert.Equal(t, "new-icon", menu.Icon)
	assert.Equal(t, int64(2), menu.RecordOrdering)
	assert.Equal(t, &newParentID, menu.ParentID)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
}

func TestMenu_Activate(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	menu.IsActive = false
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Activate()

	assert.True(t, menu.IsActive)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, menu.IsActiveMenu())
}

func TestMenu_Deactivate(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Deactivate()

	assert.False(t, menu.IsActive)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, menu.IsActiveMenu())
}

func TestMenu_Show(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	menu.IsVisible = false
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Show()

	assert.True(t, menu.IsVisible)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, menu.IsVisibleMenu())
}

func TestMenu_Hide(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Hide()

	assert.False(t, menu.IsVisible)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, menu.IsVisibleMenu())
}

func TestMenu_SetTarget(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.SetTarget("_blank")

	assert.Equal(t, "_blank", menu.Target)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
}

func TestMenu_SoftDelete(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.SoftDelete()

	assert.NotNil(t, menu.DeletedAt)
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, menu.IsDeleted())
}

func TestMenu_IsDeleted(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)

	// Initially not deleted
	assert.False(t, menu.IsDeleted())

	// After soft delete
	menu.SoftDelete()
	assert.True(t, menu.IsDeleted())
}

func TestMenu_IsActiveMenu(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)

	// Initially active and not deleted
	assert.True(t, menu.IsActiveMenu())

	// After deactivation
	menu.Deactivate()
	assert.False(t, menu.IsActiveMenu())

	// After reactivation
	menu.Activate()
	assert.True(t, menu.IsActiveMenu())

	// After soft delete
	menu.SoftDelete()
	assert.False(t, menu.IsActiveMenu())
}

func TestMenu_IsVisibleMenu(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)

	// Initially visible, active and not deleted
	assert.True(t, menu.IsVisibleMenu())

	// After hiding
	menu.Hide()
	assert.False(t, menu.IsVisibleMenu())

	// After showing
	menu.Show()
	assert.True(t, menu.IsVisibleMenu())

	// After deactivation
	menu.Deactivate()
	assert.False(t, menu.IsVisibleMenu())

	// After reactivation
	menu.Activate()
	assert.True(t, menu.IsVisibleMenu())

	// After soft delete
	menu.SoftDelete()
	assert.False(t, menu.IsVisibleMenu())
}

func TestMenu_AddChild(t *testing.T) {
	parent := entities.NewMenu("Parent", "parent", "Parent description", "/parent", "parent-icon", 1, nil)
	child := entities.NewMenu("Child", "child", "Child description", "/child", "child-icon", 2, &parent.ID)

	parent.AddChild(child)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child, parent.Children[0])
}

func TestMenu_AddChild_NilChild(t *testing.T) {
	parent := entities.NewMenu("Parent", "parent", "Parent description", "/parent", "parent-icon", 1, nil)

	parent.AddChild(nil)

	assert.Len(t, parent.Children, 0)
}

func TestMenu_RemoveChild(t *testing.T) {
	parent := entities.NewMenu("Parent", "parent", "Parent description", "/parent", "parent-icon", 1, nil)
	child1 := entities.NewMenu("Child1", "child1", "Child1 description", "/child1", "child1-icon", 2, &parent.ID)
	child2 := entities.NewMenu("Child2", "child2", "Child2 description", "/child2", "child2-icon", 3, &parent.ID)

	parent.AddChild(child1)
	parent.AddChild(child2)

	assert.Len(t, parent.Children, 2)

	// Remove child1
	parent.RemoveChild(child1.ID)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child2, parent.Children[0])
}

func TestMenu_RemoveChild_NonExistent(t *testing.T) {
	parent := entities.NewMenu("Parent", "parent", "Parent description", "/parent", "parent-icon", 1, nil)
	child := entities.NewMenu("Child", "child", "Child description", "/child", "child-icon", 2, &parent.ID)

	parent.AddChild(child)
	assert.Len(t, parent.Children, 1)

	// Remove non-existent child
	parent.RemoveChild(uuid.New())

	assert.Len(t, parent.Children, 1) // Should remain unchanged
}

func TestMenu_AddRole(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	role := &entities.Role{ID: uuid.New(), Name: "Admin", Slug: "admin"}

	menu.AddRole(role)

	assert.Len(t, menu.Roles, 1)
	assert.Equal(t, role, menu.Roles[0])
}

func TestMenu_AddRole_NilRole(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)

	menu.AddRole(nil)

	assert.Len(t, menu.Roles, 0)
}

func TestMenu_RemoveRole(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	role1 := &entities.Role{ID: uuid.New(), Name: "Admin", Slug: "admin"}
	role2 := &entities.Role{ID: uuid.New(), Name: "User", Slug: "user"}

	menu.AddRole(role1)
	menu.AddRole(role2)

	assert.Len(t, menu.Roles, 2)

	// Remove role1
	menu.RemoveRole(role1.ID)

	assert.Len(t, menu.Roles, 1)
	assert.Equal(t, role2, menu.Roles[0])
}

func TestMenu_RemoveRole_NonExistent(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	role := &entities.Role{ID: uuid.New(), Name: "Admin", Slug: "admin"}

	menu.AddRole(role)
	assert.Len(t, menu.Roles, 1)

	// Remove non-existent role
	menu.RemoveRole(uuid.New())

	assert.Len(t, menu.Roles, 1) // Should remain unchanged
}

func TestMenu_HasRole(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	role := &entities.Role{ID: uuid.New(), Name: "Admin", Slug: "admin"}

	// Initially no roles
	assert.False(t, menu.HasRole(role.ID))

	// After adding role
	menu.AddRole(role)
	assert.True(t, menu.HasRole(role.ID))

	// After removing role
	menu.RemoveRole(role.ID)
	assert.False(t, menu.HasRole(role.ID))
}

func TestMenu_Validate_Success(t *testing.T) {
	menu := entities.NewMenu("Valid Name", "valid-slug", "Valid description", "/valid", "valid-icon", 1, nil)

	err := menu.Validate()

	assert.NoError(t, err)
}

func TestMenu_Validate_EmptyName(t *testing.T) {
	menu := entities.NewMenu("", "valid-slug", "Valid description", "/valid", "valid-icon", 1, nil)

	err := menu.Validate()

	assert.Error(t, err)
	assert.Equal(t, "menu name is required", err.Error())
}

func TestMenu_Validate_WhitespaceName(t *testing.T) {
	menu := entities.NewMenu("   ", "valid-slug", "Valid description", "/valid", "valid-icon", 1, nil)

	err := menu.Validate()

	assert.Error(t, err)
	assert.Equal(t, "menu name is required", err.Error())
}

func TestMenu_Validate_EmptySlug(t *testing.T) {
	menu := entities.NewMenu("Valid Name", "", "Valid description", "/valid", "valid-icon", 1, nil)

	err := menu.Validate()

	assert.Error(t, err)
	assert.Equal(t, "menu slug is required", err.Error())
}

func TestMenu_Validate_WhitespaceSlug(t *testing.T) {
	menu := entities.NewMenu("Valid Name", "   ", "Valid description", "/valid", "valid-icon", 1, nil)

	err := menu.Validate()

	assert.Error(t, err)
	assert.Equal(t, "menu slug is required", err.Error())
}

func TestMenu_Validate_NegativeRecordOrdering(t *testing.T) {
	menu := entities.NewMenu("Valid Name", "valid-slug", "Valid description", "/valid", "valid-icon", -1, nil)

	err := menu.Validate()

	assert.Error(t, err)
	assert.Equal(t, "menu record ordering must be non-negative", err.Error())
}

func TestMenu_IsRoot(t *testing.T) {
	// Root menu (no parent)
	root := entities.NewMenu("Root", "root", "Root description", "/root", "root-icon", 1, nil)
	assert.True(t, root.IsRoot())

	// Child menu (has parent)
	parentID := uuid.New()
	child := entities.NewMenu("Child", "child", "Child description", "/child", "child-icon", 2, &parentID)
	assert.False(t, child.IsRoot())
}

func TestMenu_IsLeaf(t *testing.T) {
	// Leaf menu (no children)
	leaf := entities.NewMenu("Leaf", "leaf", "Leaf description", "/leaf", "leaf-icon", 1, nil)
	assert.True(t, leaf.IsLeaf())

	// Parent menu (has children)
	parent := entities.NewMenu("Parent", "parent", "Parent description", "/parent", "parent-icon", 1, nil)
	child := entities.NewMenu("Child", "child", "Child description", "/child", "child-icon", 2, &parent.ID)
	parent.AddChild(child)
	assert.False(t, parent.IsLeaf())
}

func TestMenu_GetDepth(t *testing.T) {
	// Root menu
	root := entities.NewMenu("Root", "root", "Root description", "/root", "root-icon", 1, nil)
	assert.Equal(t, 0, root.GetDepth())

	// Child menu
	child := entities.NewMenu("Child", "child", "Child description", "/child", "child-icon", 2, &root.ID)
	child.Parent = root
	assert.Equal(t, 1, child.GetDepth())

	// Grandchild menu
	grandchild := entities.NewMenu("Grandchild", "grandchild", "Grandchild description", "/grandchild", "grandchild-icon", 3, &child.ID)
	grandchild.Parent = child
	assert.Equal(t, 2, grandchild.GetDepth())
}

func TestMenu_GetWidth(t *testing.T) {
	menu := entities.NewMenu("Test", "test", "Test description", "/test", "test-icon", 1, nil)
	menu.RecordLeft = 1
	menu.RecordRight = 5

	width := menu.GetWidth()

	assert.Equal(t, int64(5), width) // 5 - 1 + 1 = 5
}

func TestMenu_IsDescendantOf(t *testing.T) {
	ancestor := entities.NewMenu("Ancestor", "ancestor", "Ancestor description", "/ancestor", "ancestor-icon", 1, nil)
	ancestor.RecordLeft = 1
	ancestor.RecordRight = 10

	descendant := entities.NewMenu("Descendant", "descendant", "Descendant description", "/descendant", "descendant-icon", 2, &ancestor.ID)
	descendant.RecordLeft = 3
	descendant.RecordRight = 7

	// Should be descendant
	assert.True(t, descendant.IsDescendantOf(ancestor))

	// Should not be descendant of itself
	assert.False(t, descendant.IsDescendantOf(descendant))

	// Should not be descendant of unrelated menu
	unrelated := entities.NewMenu("Unrelated", "unrelated", "Unrelated description", "/unrelated", "unrelated-icon", 3, nil)
	unrelated.RecordLeft = 20
	unrelated.RecordRight = 25
	assert.False(t, descendant.IsDescendantOf(unrelated))
}

func TestMenu_IsAncestorOf(t *testing.T) {
	ancestor := entities.NewMenu("Ancestor", "ancestor", "Ancestor description", "/ancestor", "ancestor-icon", 1, nil)
	ancestor.RecordLeft = 1
	ancestor.RecordRight = 10

	descendant := entities.NewMenu("Descendant", "descendant", "Descendant description", "/descendant", "descendant-icon", 2, &ancestor.ID)
	descendant.RecordLeft = 3
	descendant.RecordRight = 7

	// Should be ancestor
	assert.True(t, ancestor.IsAncestorOf(descendant))

	// Should not be ancestor of itself
	assert.False(t, ancestor.IsAncestorOf(ancestor))

	// Should not be ancestor of unrelated menu
	unrelated := entities.NewMenu("Unrelated", "unrelated", "Unrelated description", "/unrelated", "unrelated-icon", 3, nil)
	unrelated.RecordLeft = 20
	unrelated.RecordRight = 25
	assert.False(t, ancestor.IsAncestorOf(unrelated))
}

func TestMenu_SoftDelete_MultipleCalls(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)

	// First soft delete
	menu.SoftDelete()
	firstDeletedAt := menu.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	menu.SoftDelete()
	secondDeletedAt := menu.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, menu.IsDeleted())
}

func TestMenu_Activate_AlreadyActive(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Activate()

	// Should update timestamp even if already active
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, menu.IsActive)
}

func TestMenu_Deactivate_AlreadyInactive(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	menu.IsActive = false
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Deactivate()

	// Should update timestamp even if already inactive
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, menu.IsActive)
}

func TestMenu_Show_AlreadyVisible(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Show()

	// Should update timestamp even if already visible
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, menu.IsVisible)
}

func TestMenu_Hide_AlreadyHidden(t *testing.T) {
	menu := entities.NewMenu("Test Menu", "test-menu", "Test description", "/test", "test-icon", 1, nil)
	menu.IsVisible = false
	originalUpdatedAt := menu.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	menu.Hide()

	// Should update timestamp even if already hidden
	assert.True(t, menu.UpdatedAt.After(originalUpdatedAt))
	assert.False(t, menu.IsVisible)
}
