package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Menu represents a menu entity in the domain layer
type Menu struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description,omitempty"`
	URL            string     `json:"url,omitempty"`
	Icon           string     `json:"icon,omitempty"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	RecordLeft     *uint64    `json:"record_left,omitempty"`
	RecordRight    *uint64    `json:"record_right,omitempty"`
	RecordOrdering *uint64    `json:"record_ordering,omitempty"`
	RecordDepth    *uint64    `json:"record_depth,omitempty"`
	IsActive       bool       `json:"is_active"`
	IsVisible      bool       `json:"is_visible"`
	Target         string     `json:"target,omitempty"` // _blank, _self, etc.
	CreatedBy      uuid.UUID  `json:"created_by"`
	UpdatedBy      uuid.UUID  `json:"updated_by"`
	DeletedBy      *uuid.UUID `json:"deleted_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Relationships
	Parent   *Menu   `json:"parent,omitempty"`
	Children []*Menu `json:"children,omitempty"`
	Roles    []*Role `json:"roles,omitempty"`
}

// NewMenu creates a new menu instance
func NewMenu(name, slug, description, url, icon string, parentID *uuid.UUID) *Menu {
	return &Menu{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		URL:         url,
		Icon:        icon,
		ParentID:    parentID,
		IsActive:    true,
		IsVisible:   true,
		Target:      "_self",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Children:    []*Menu{},
		Roles:       []*Role{},
	}
}

// UpdateMenu updates menu information
func (m *Menu) UpdateMenu(name, slug, description, url, icon string, recordOrdering uint64, parentID *uuid.UUID) {
	m.Name = name
	m.Slug = slug
	m.Description = description
	m.URL = url
	m.Icon = icon
	m.RecordOrdering = &recordOrdering
	m.ParentID = parentID
	m.UpdatedAt = time.Now()
}

// Activate activates the menu
func (m *Menu) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate deactivates the menu
func (m *Menu) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// Show makes the menu visible
func (m *Menu) Show() {
	m.IsVisible = true
	m.UpdatedAt = time.Now()
}

// Hide makes the menu invisible
func (m *Menu) Hide() {
	m.IsVisible = false
	m.UpdatedAt = time.Now()
}

// SetTarget sets the target for the menu link
func (m *Menu) SetTarget(target string) {
	m.Target = target
	m.UpdatedAt = time.Now()
}

// SoftDelete marks the menu as deleted
func (m *Menu) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
}

// IsDeleted checks if the menu is soft deleted
func (m *Menu) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsActiveMenu checks if the menu is active
func (m *Menu) IsActiveMenu() bool {
	return m.IsActive && !m.IsDeleted()
}

// IsVisibleMenu checks if the menu is visible
func (m *Menu) IsVisibleMenu() bool {
	return m.IsVisible && m.IsActive && !m.IsDeleted()
}

// AddChild adds a child menu
func (m *Menu) AddChild(child *Menu) {
	if child != nil {
		m.Children = append(m.Children, child)
	}
}

// RemoveChild removes a child menu
func (m *Menu) RemoveChild(childID uuid.UUID) {
	for i, child := range m.Children {
		if child.ID == childID {
			m.Children = append(m.Children[:i], m.Children[i+1:]...)
			break
		}
	}
}

// AddRole adds a role to the menu
func (m *Menu) AddRole(role *Role) {
	if role != nil {
		m.Roles = append(m.Roles, role)
	}
}

// RemoveRole removes a role from the menu
func (m *Menu) RemoveRole(roleID uuid.UUID) {
	for i, role := range m.Roles {
		if role.ID == roleID {
			m.Roles = append(m.Roles[:i], m.Roles[i+1:]...)
			break
		}
	}
}

// HasRole checks if the menu has a specific role
func (m *Menu) HasRole(roleID uuid.UUID) bool {
	for _, role := range m.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

// Validate validates the menu data
func (m *Menu) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("menu name is required")
	}

	if strings.TrimSpace(m.Slug) == "" {
		return errors.New("menu slug is required")
	}

	return nil
}

// IsRoot checks if the menu is a root menu (no parent)
func (m *Menu) IsRoot() bool {
	return m.ParentID == nil
}

// IsLeaf checks if the menu is a leaf menu (no children)
func (m *Menu) IsLeaf() bool {
	return len(m.Children) == 0
}

// GetDepth calculates the depth of the menu in the hierarchy
func (m *Menu) GetDepth() int {
	if m.IsRoot() {
		return 0
	}

	depth := 1
	parent := m.Parent
	for parent != nil && !parent.IsRoot() {
		depth++
		parent = parent.Parent
	}

	return depth
}

// GetWidth returns the width of the node in the nested set
func (m *Menu) GetWidth() int64 {
	return int64(*m.RecordRight - *m.RecordLeft + 1)
}

// IsDescendantOf checks if this menu is a descendant of the given menu
func (m *Menu) IsDescendantOf(ancestor *Menu) bool {
	return *m.RecordLeft > *ancestor.RecordLeft && *m.RecordRight < *ancestor.RecordRight
}

// IsAncestorOf checks if this menu is an ancestor of the given menu
func (m *Menu) IsAncestorOf(descendant *Menu) bool {
	return *m.RecordLeft < *descendant.RecordLeft && *m.RecordRight > *descendant.RecordRight
}

// IsSiblingOf checks if this menu is a sibling of the given menu
func (m *Menu) IsSiblingOf(other *Menu) bool {
	return m.ParentID != nil && other.ParentID != nil && *m.ParentID == *other.ParentID
}

// GetSiblingCount returns the number of siblings this menu has
func (m *Menu) GetSiblingCount() int {
	if m.IsRoot() {
		return 0
	}
	return len(m.Parent.Children) - 1
}

// GetPositionAmongSiblings returns the position of this menu among its siblings
func (m *Menu) GetPositionAmongSiblings() int {
	if m.IsRoot() {
		return 0
	}

	for i, sibling := range m.Parent.Children {
		if sibling.ID == m.ID {
			return i
		}
	}
	return -1
}

// GetNextSibling returns the next sibling menu
func (m *Menu) GetNextSibling() *Menu {
	if m.IsRoot() {
		return nil
	}

	position := m.GetPositionAmongSiblings()
	if position >= 0 && position < len(m.Parent.Children)-1 {
		return m.Parent.Children[position+1]
	}
	return nil
}

// GetPreviousSibling returns the previous sibling menu
func (m *Menu) GetPreviousSibling() *Menu {
	if m.IsRoot() {
		return nil
	}

	position := m.GetPositionAmongSiblings()
	if position > 0 {
		return m.Parent.Children[position-1]
	}
	return nil
}

// GetFirstChild returns the first child menu
func (m *Menu) GetFirstChild() *Menu {
	if len(m.Children) > 0 {
		return m.Children[0]
	}
	return nil
}

// GetLastChild returns the last child menu
func (m *Menu) GetLastChild() *Menu {
	if len(m.Children) > 0 {
		return m.Children[len(m.Children)-1]
	}
	return nil
}

// GetAncestors returns all ancestor menus (parent, grandparent, etc.)
func (m *Menu) GetAncestors() []*Menu {
	var ancestors []*Menu
	parent := m.Parent

	for parent != nil {
		ancestors = append([]*Menu{parent}, ancestors...)
		parent = parent.Parent
	}

	return ancestors
}

// GetDescendants returns all descendant menus (children, grandchildren, etc.)
func (m *Menu) GetDescendants() []*Menu {
	var descendants []*Menu

	for _, child := range m.Children {
		descendants = append(descendants, child)
		descendants = append(descendants, child.GetDescendants()...)
	}

	return descendants
}

// GetLeaves returns all leaf nodes (menus with no children)
func (m *Menu) GetLeaves() []*Menu {
	var leaves []*Menu

	if m.IsLeaf() {
		leaves = append(leaves, m)
	} else {
		for _, child := range m.Children {
			leaves = append(leaves, child.GetLeaves()...)
		}
	}

	return leaves
}

// GetPath returns the path from root to this menu
func (m *Menu) GetPath() []*Menu {
	path := []*Menu{m}
	ancestors := m.GetAncestors()
	path = append(ancestors, path...)
	return path
}

// GetPathString returns the path as a string representation
func (m *Menu) GetPathString() string {
	path := m.GetPath()
	var pathStrings []string

	for _, menu := range path {
		pathStrings = append(pathStrings, menu.Name)
	}

	return strings.Join(pathStrings, " > ")
}

// IsInPath checks if the given menu is in the path from root to this menu
func (m *Menu) IsInPath(target *Menu) bool {
	path := m.GetPath()

	for _, menu := range path {
		if menu.ID == target.ID {
			return true
		}
	}

	return false
}

// GetSubtreeSize returns the total number of nodes in the subtree rooted at this menu
func (m *Menu) GetSubtreeSize() int {
	size := 1 // Include self

	for _, child := range m.Children {
		size += child.GetSubtreeSize()
	}

	return size
}

// GetMaxDepth returns the maximum depth of the subtree rooted at this menu
func (m *Menu) GetMaxDepth() int {
	if m.IsLeaf() {
		return 0
	}

	maxDepth := 0
	for _, child := range m.Children {
		childDepth := child.GetMaxDepth()
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}

	return maxDepth + 1
}

// IsBalanced checks if the subtree rooted at this menu is balanced
func (m *Menu) IsBalanced() bool {
	if m.IsLeaf() {
		return true
	}

	depths := make(map[int]bool)
	for _, child := range m.Children {
		depths[child.GetMaxDepth()] = true
	}

	// All children should have the same depth for a balanced tree
	return len(depths) <= 1
}

// GetLevel returns the level of this menu in the tree (0 for root, 1 for first level, etc.)
func (m *Menu) GetLevel() int {
	return m.GetDepth()
}

// IsAtLevel checks if this menu is at the specified level
func (m *Menu) IsAtLevel(level int) bool {
	return m.GetLevel() == level
}

// GetNodesAtLevel returns all nodes at the specified level in the subtree
func (m *Menu) GetNodesAtLevel(level int) []*Menu {
	var nodes []*Menu

	if m.GetLevel() == level {
		nodes = append(nodes, m)
	} else if m.GetLevel() < level {
		for _, child := range m.Children {
			nodes = append(nodes, child.GetNodesAtLevel(level)...)
		}
	}

	return nodes
}
