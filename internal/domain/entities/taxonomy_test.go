package entities_test

import (
	"testing"
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTaxonomy_Success(t *testing.T) {
	name := "Technology"
	slug := "technology"
	code := "tech"
	description := "Technology related content"
	parentID := uuid.New()

	taxonomy := entities.NewTaxonomy(name, slug, code, description, &parentID)

	assert.NotNil(t, taxonomy)
	assert.Equal(t, name, taxonomy.Name)
	assert.Equal(t, slug, taxonomy.Slug)
	assert.Equal(t, code, taxonomy.Code)
	assert.Equal(t, description, taxonomy.Description)
	assert.Equal(t, &parentID, taxonomy.ParentID)
	assert.NotEqual(t, uuid.Nil, taxonomy.ID)
	assert.False(t, taxonomy.CreatedAt.IsZero())
	assert.False(t, taxonomy.UpdatedAt.IsZero())
	assert.Nil(t, taxonomy.DeletedAt)
	assert.Equal(t, int64(0), taxonomy.RecordLeft)
	assert.Equal(t, int64(0), taxonomy.RecordRight)
	assert.Equal(t, int64(0), taxonomy.RecordDepth)
	assert.Empty(t, taxonomy.Children)
}

func TestNewTaxonomy_WithoutParentID(t *testing.T) {
	name := "Technology"
	slug := "technology"
	code := "tech"
	description := "Technology related content"

	taxonomy := entities.NewTaxonomy(name, slug, code, description, nil)

	assert.NotNil(t, taxonomy)
	assert.Equal(t, name, taxonomy.Name)
	assert.Equal(t, slug, taxonomy.Slug)
	assert.Equal(t, code, taxonomy.Code)
	assert.Equal(t, description, taxonomy.Description)
	assert.Nil(t, taxonomy.ParentID)
	assert.NotEqual(t, uuid.Nil, taxonomy.ID)
	assert.False(t, taxonomy.CreatedAt.IsZero())
	assert.False(t, taxonomy.UpdatedAt.IsZero())
	assert.Nil(t, taxonomy.DeletedAt)
	assert.Equal(t, int64(0), taxonomy.RecordLeft)
	assert.Equal(t, int64(0), taxonomy.RecordRight)
	assert.Equal(t, int64(0), taxonomy.RecordDepth)
	assert.Empty(t, taxonomy.Children)
}

func TestTaxonomy_UpdateTaxonomy(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Old Name", "old-slug", "old-code", "Old description", nil)
	originalUpdatedAt := taxonomy.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	newParentID := uuid.New()
	taxonomy.UpdateTaxonomy("New Name", "new-slug", "new-code", "New description", &newParentID)

	assert.Equal(t, "New Name", taxonomy.Name)
	assert.Equal(t, "new-slug", taxonomy.Slug)
	assert.Equal(t, "new-code", taxonomy.Code)
	assert.Equal(t, "New description", taxonomy.Description)
	assert.Equal(t, &newParentID, taxonomy.ParentID)
	assert.True(t, taxonomy.UpdatedAt.After(originalUpdatedAt))
}

func TestTaxonomy_SoftDelete(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Test Taxonomy", "test-taxonomy", "test", "Test description", nil)
	originalUpdatedAt := taxonomy.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	taxonomy.SoftDelete()

	assert.NotNil(t, taxonomy.DeletedAt)
	assert.True(t, taxonomy.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, taxonomy.IsDeleted())
}

func TestTaxonomy_IsDeleted(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Test Taxonomy", "test-taxonomy", "test", "Test description", nil)

	// Initially not deleted
	assert.False(t, taxonomy.IsDeleted())

	// After soft delete
	taxonomy.SoftDelete()
	assert.True(t, taxonomy.IsDeleted())
}

func TestTaxonomy_AddChild(t *testing.T) {
	parent := entities.NewTaxonomy("Parent", "parent", "parent", "Parent description", nil)
	child := entities.NewTaxonomy("Child", "child", "child", "Child description", &parent.ID)

	parent.AddChild(child)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child, parent.Children[0])
}

func TestTaxonomy_AddChild_NilChild(t *testing.T) {
	parent := entities.NewTaxonomy("Parent", "parent", "parent", "Parent description", nil)

	parent.AddChild(nil)

	assert.Len(t, parent.Children, 0)
}

func TestTaxonomy_RemoveChild(t *testing.T) {
	parent := entities.NewTaxonomy("Parent", "parent", "parent", "Parent description", nil)
	child1 := entities.NewTaxonomy("Child1", "child1", "child1", "Child1 description", &parent.ID)
	child2 := entities.NewTaxonomy("Child2", "child2", "child2", "Child2 description", &parent.ID)

	parent.AddChild(child1)
	parent.AddChild(child2)

	assert.Len(t, parent.Children, 2)

	// Remove child1
	parent.RemoveChild(child1.ID)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child2, parent.Children[0])
}

func TestTaxonomy_RemoveChild_NonExistent(t *testing.T) {
	parent := entities.NewTaxonomy("Parent", "parent", "parent", "Parent description", nil)
	child := entities.NewTaxonomy("Child", "child", "child", "Child description", &parent.ID)

	parent.AddChild(child)
	assert.Len(t, parent.Children, 1)

	// Remove non-existent child
	parent.RemoveChild(uuid.New())

	assert.Len(t, parent.Children, 1) // Should remain unchanged
}

func TestTaxonomy_Validate_Success(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Valid Name", "valid-slug", "valid", "Valid description", nil)

	err := taxonomy.Validate()

	assert.NoError(t, err)
}

func TestTaxonomy_Validate_EmptyName(t *testing.T) {
	taxonomy := entities.NewTaxonomy("", "valid-slug", "valid", "Valid description", nil)

	err := taxonomy.Validate()

	assert.Error(t, err)
	assert.Equal(t, "taxonomy name is required", err.Error())
}

func TestTaxonomy_Validate_WhitespaceName(t *testing.T) {
	taxonomy := entities.NewTaxonomy("   ", "valid-slug", "valid", "Valid description", nil)

	err := taxonomy.Validate()

	assert.Error(t, err)
	assert.Equal(t, "taxonomy name is required", err.Error())
}

func TestTaxonomy_Validate_EmptySlug(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Valid Name", "", "valid", "Valid description", nil)

	err := taxonomy.Validate()

	assert.Error(t, err)
	assert.Equal(t, "taxonomy slug is required", err.Error())
}

func TestTaxonomy_Validate_WhitespaceSlug(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Valid Name", "   ", "valid", "Valid description", nil)

	err := taxonomy.Validate()

	assert.Error(t, err)
	assert.Equal(t, "taxonomy slug is required", err.Error())
}

func TestTaxonomy_IsRoot(t *testing.T) {
	// Root taxonomy (no parent)
	root := entities.NewTaxonomy("Root", "root", "root", "Root description", nil)
	assert.True(t, root.IsRoot())

	// Child taxonomy (has parent)
	parentID := uuid.New()
	child := entities.NewTaxonomy("Child", "child", "child", "Child description", &parentID)
	assert.False(t, child.IsRoot())
}

func TestTaxonomy_IsLeaf(t *testing.T) {
	// Leaf taxonomy (no children)
	leaf := entities.NewTaxonomy("Leaf", "leaf", "leaf", "Leaf description", nil)
	assert.True(t, leaf.IsLeaf())

	// Parent taxonomy (has children)
	parent := entities.NewTaxonomy("Parent", "parent", "parent", "Parent description", nil)
	child := entities.NewTaxonomy("Child", "child", "child", "Child description", &parent.ID)
	parent.AddChild(child)
	assert.False(t, parent.IsLeaf())
}

func TestTaxonomy_GetDepth(t *testing.T) {
	// Root taxonomy
	root := entities.NewTaxonomy("Root", "root", "root", "Root description", nil)
	assert.Equal(t, 0, root.GetDepth())

	// Child taxonomy
	child := entities.NewTaxonomy("Child", "child", "child", "Child description", &root.ID)
	child.Parent = root
	assert.Equal(t, 1, child.GetDepth())

	// Grandchild taxonomy
	grandchild := entities.NewTaxonomy("Grandchild", "grandchild", "grandchild", "Grandchild description", &child.ID)
	grandchild.Parent = child
	assert.Equal(t, 2, grandchild.GetDepth())
}

func TestTaxonomy_GetWidth(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Test", "test", "test", "Test description", nil)
	taxonomy.RecordLeft = 1
	taxonomy.RecordRight = 5

	width := taxonomy.GetWidth()

	assert.Equal(t, int64(5), width) // 5 - 1 + 1 = 5
}

func TestTaxonomy_IsDescendantOf(t *testing.T) {
	ancestor := entities.NewTaxonomy("Ancestor", "ancestor", "ancestor", "Ancestor description", nil)
	ancestor.RecordLeft = 1
	ancestor.RecordRight = 10

	descendant := entities.NewTaxonomy("Descendant", "descendant", "descendant", "Descendant description", &ancestor.ID)
	descendant.RecordLeft = 3
	descendant.RecordRight = 7

	// Should be descendant
	assert.True(t, descendant.IsDescendantOf(ancestor))

	// Should not be descendant of itself
	assert.False(t, descendant.IsDescendantOf(descendant))

	// Should not be descendant of unrelated taxonomy
	unrelated := entities.NewTaxonomy("Unrelated", "unrelated", "unrelated", "Unrelated description", nil)
	unrelated.RecordLeft = 20
	unrelated.RecordRight = 25
	assert.False(t, descendant.IsDescendantOf(unrelated))
}

func TestTaxonomy_IsAncestorOf(t *testing.T) {
	ancestor := entities.NewTaxonomy("Ancestor", "ancestor", "ancestor", "Ancestor description", nil)
	ancestor.RecordLeft = 1
	ancestor.RecordRight = 10

	descendant := entities.NewTaxonomy("Descendant", "descendant", "descendant", "Descendant description", &ancestor.ID)
	descendant.RecordLeft = 3
	descendant.RecordRight = 7

	// Should be ancestor
	assert.True(t, ancestor.IsAncestorOf(descendant))

	// Should not be ancestor of itself
	assert.False(t, ancestor.IsAncestorOf(ancestor))

	// Should not be ancestor of unrelated taxonomy
	unrelated := entities.NewTaxonomy("Unrelated", "unrelated", "unrelated", "Unrelated description", nil)
	unrelated.RecordLeft = 20
	unrelated.RecordRight = 25
	assert.False(t, ancestor.IsAncestorOf(unrelated))
}

func TestTaxonomy_SoftDelete_MultipleCalls(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Test Taxonomy", "test-taxonomy", "test", "Test description", nil)

	// First soft delete
	taxonomy.SoftDelete()
	firstDeletedAt := taxonomy.DeletedAt

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	// Second soft delete
	taxonomy.SoftDelete()
	secondDeletedAt := taxonomy.DeletedAt

	// Should update the deleted timestamp
	assert.True(t, secondDeletedAt.After(*firstDeletedAt))
	assert.True(t, taxonomy.IsDeleted())
}

func TestTaxonomy_UpdateTaxonomy_AllFields(t *testing.T) {
	taxonomy := entities.NewTaxonomy("Old Name", "old-slug", "old-code", "Old description", nil)
	originalUpdatedAt := taxonomy.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	newParentID := uuid.New()
	taxonomy.UpdateTaxonomy("New Name", "new-slug", "new-code", "New description", &newParentID)

	assert.Equal(t, "New Name", taxonomy.Name)
	assert.Equal(t, "new-slug", taxonomy.Slug)
	assert.Equal(t, "new-code", taxonomy.Code)
	assert.Equal(t, "New description", taxonomy.Description)
	assert.Equal(t, &newParentID, taxonomy.ParentID)
	assert.True(t, taxonomy.UpdatedAt.After(originalUpdatedAt))
}

func TestTaxonomy_UpdateTaxonomy_RemoveParent(t *testing.T) {
	parentID := uuid.New()
	taxonomy := entities.NewTaxonomy("Test", "test", "test", "Test description", &parentID)
	originalUpdatedAt := taxonomy.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	taxonomy.UpdateTaxonomy("New Name", "new-slug", "new-code", "New description", nil)

	assert.Equal(t, "New Name", taxonomy.Name)
	assert.Equal(t, "new-slug", taxonomy.Slug)
	assert.Equal(t, "new-code", taxonomy.Code)
	assert.Equal(t, "New description", taxonomy.Description)
	assert.Nil(t, taxonomy.ParentID)
	assert.True(t, taxonomy.UpdatedAt.After(originalUpdatedAt))
}
