// Package responses provides HTTP response structures and utilities for the Go RESTful API.
// It follows Laravel API Resource patterns for consistent formatting across all endpoints.
package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// MenuItemResource represents a comprehensive menu item in API responses.
// This struct provides a complete view of a menu item including its hierarchical structure,
// navigation properties, access control, and computed tree properties.
// It follows the Laravel API Resource pattern for consistent formatting.
type MenuItemResource struct {
	// ID is the unique identifier for the menu item
	ID string `json:"id"`
	// Name is the display name of the menu item
	Name string `json:"name"`
	// Slug is the URL-friendly version of the menu item name
	Slug string `json:"slug"`
	// Description is an optional description of the menu item
	Description string `json:"description,omitempty"`
	// URL is the optional destination URL for the menu item
	URL string `json:"url,omitempty"`
	// Icon is the optional icon identifier for the menu item
	Icon string `json:"icon,omitempty"`
	// ParentID is the optional ID of the parent menu item for hierarchical menus
	ParentID *string `json:"parent_id,omitempty"`
	// RecordLeft is used for nested set model operations (tree structure)
	RecordLeft *uint64 `json:"record_left,omitempty"`
	// RecordRight is used for nested set model operations (tree structure)
	RecordRight *uint64 `json:"record_right,omitempty"`
	// RecordOrdering determines the display order of menu items
	RecordOrdering *uint64 `json:"record_ordering,omitempty"`
	// RecordDepth indicates the nesting level of the menu item in the tree
	RecordDepth *uint64 `json:"record_depth,omitempty"`
	// IsActive indicates whether the menu item is currently active/enabled
	IsActive bool `json:"is_active"`
	// IsVisible indicates whether the menu item should be displayed to users
	IsVisible bool `json:"is_visible"`
	// Target specifies the target attribute for the menu link (e.g., "_blank", "_self")
	Target string `json:"target,omitempty"`
	// CreatedBy is the ID of the user who created the menu item
	CreatedBy string `json:"created_by"`
	// UpdatedBy is the ID of the user who last updated the menu item
	UpdatedBy string `json:"updated_by"`
	// DeletedBy is the optional ID of the user who deleted the menu item
	DeletedBy *string `json:"deleted_by,omitempty"`
	// CreatedAt is the timestamp when the menu item was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when the menu item was last updated
	UpdatedAt string `json:"updated_at"`
	// DeletedAt is the optional timestamp when the menu item was soft-deleted
	DeletedAt *string `json:"deleted_at,omitempty"`

	// Computed fields for tree operations and status checking
	// IsDeleted indicates whether the menu item has been soft-deleted
	IsDeleted bool `json:"is_deleted"`
	// IsRoot indicates whether this menu item is at the root level (no parent)
	IsRoot bool `json:"is_root"`
	// IsLeaf indicates whether this menu item has no children
	IsLeaf bool `json:"is_leaf"`
	// Depth indicates the nesting level of the menu item in the tree
	Depth int `json:"depth"`
	// Width indicates the number of descendants this menu item has
	Width int64 `json:"width"`

	// Nested resources for complete menu structure
	// Parent contains the parent menu item if this is a child item
	Parent *MenuItemResource `json:"parent,omitempty"`
	// Children contains the nested child menu items
	Children []MenuItemResource `json:"children,omitempty"`
	// Roles contains the roles that have access to this menu item
	Roles []RoleResource `json:"roles,omitempty"`
}

// MenuCollection represents a collection of menu items.
// This follows the Laravel API Resource Collection pattern for consistent pagination
// and metadata handling across all collection endpoints.
type MenuCollection struct {
	// Data contains the array of menu item resources
	Data []MenuItemResource `json:"data"`
	// Meta contains collection metadata (pagination, counts, etc.)
	Meta CollectionMeta `json:"meta"`
	// Links contains navigation links (first, last, prev, next)
	Links CollectionLinks `json:"links"`
}

// MenuResourceResponse represents a single menu item response.
// This wrapper provides a consistent response structure with response codes
// and messages, following the standard API response format.
type MenuResourceResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the menu item resource
	Data MenuItemResource `json:"data"`
}

// MenuCollectionResponse represents a collection of menu items response.
// This wrapper provides a consistent response structure for collections with
// response codes and messages.
type MenuCollectionResponse struct {
	// ResponseCode indicates the HTTP status code for the operation
	ResponseCode int `json:"response_code"`
	// ResponseMessage provides a human-readable description of the operation result
	ResponseMessage string `json:"response_message"`
	// Data contains the menu collection
	Data MenuCollection `json:"data"`
}

// NewMenuResource creates a new MenuItemResource from menu entity.
// This function transforms the domain entity into a comprehensive API response format,
// handling all optional fields, computed properties, and nested resources.
//
// Parameters:
//   - menu: The menu domain entity to convert
//
// Returns:
//   - A new MenuItemResource with all fields properly formatted
func NewMenuResource(menu *entities.Menu) MenuItemResource {
	resource := MenuItemResource{
		ID:          menu.ID.String(),
		Name:        menu.Name,
		Slug:        menu.Slug,
		Description: menu.Description,
		URL:         menu.URL,
		Icon:        menu.Icon,
		IsActive:    menu.IsActive,
		IsVisible:   menu.IsVisible,
		Target:      menu.Target,
		CreatedBy:   menu.CreatedBy.String(),
		UpdatedBy:   menu.UpdatedBy.String(),
		CreatedAt:   menu.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   menu.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Set computed fields based on menu state
		IsDeleted: menu.IsDeleted(),
		IsRoot:    menu.IsRoot(),
		IsLeaf:    menu.IsLeaf(),
		Depth:     menu.GetDepth(),
	}

	// Handle optional parent ID for hierarchical menus
	if menu.ParentID != nil {
		parentID := menu.ParentID.String()
		resource.ParentID = &parentID
	}

	// Set nested set model fields if available
	if menu.RecordLeft != nil {
		resource.RecordLeft = menu.RecordLeft
	}

	if menu.RecordRight != nil {
		resource.RecordRight = menu.RecordRight
	}

	if menu.RecordOrdering != nil {
		resource.RecordOrdering = menu.RecordOrdering
	}

	if menu.RecordDepth != nil {
		resource.RecordDepth = menu.RecordDepth
	}

	// Handle soft deletion information
	if menu.DeletedBy != nil {
		deletedBy := menu.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if menu.DeletedAt != nil {
		deletedAt := menu.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	// Calculate width if record boundaries are available for tree operations
	if menu.RecordLeft != nil && menu.RecordRight != nil {
		resource.Width = menu.GetWidth()
	}

	// Set nested parent resource if available
	if menu.Parent != nil {
		parentResource := NewMenuResource(menu.Parent)
		resource.Parent = &parentResource
	}

	// Set nested children resources if available
	if len(menu.Children) > 0 {
		childrenResources := make([]MenuItemResource, len(menu.Children))
		for i, child := range menu.Children {
			childrenResources[i] = NewMenuResource(child)
		}
		resource.Children = childrenResources
	}

	// Set associated role resources if available
	if len(menu.Roles) > 0 {
		roleResources := make([]RoleResource, len(menu.Roles))
		for i, role := range menu.Roles {
			roleResources[i] = NewRoleResource(role)
		}
		resource.Roles = roleResources
	}

	return resource
}

// NewMenuResourceResponse creates a new MenuResourceResponse.
// This function wraps a MenuItemResource in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - menu: The menu domain entity to convert and wrap
//
// Returns:
//   - A new MenuResourceResponse with success status and menu data
func NewMenuResourceResponse(menu *entities.Menu) MenuResourceResponse {
	return MenuResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Menu retrieved successfully",
		Data:            NewMenuResource(menu),
	}
}

// NewMenuCollection creates a new MenuCollection.
// This function transforms multiple menu domain entities into a consistent
// API response format, creating a collection that can be easily serialized to JSON.
//
// Parameters:
//   - menus: Slice of menu domain entities to convert
//
// Returns:
//   - A new MenuCollection with all menu items properly formatted
func NewMenuCollection(menus []*entities.Menu) MenuCollection {
	menuResources := make([]MenuItemResource, len(menus))
	for i, menu := range menus {
		menuResources[i] = NewMenuResource(menu)
	}

	return MenuCollection{
		Data: menuResources,
	}
}

// NewPaginatedMenuCollection creates a new MenuCollection with pagination.
// This function follows Laravel's paginated resource collection pattern and provides
// comprehensive pagination information including current page, total pages, and navigation links.
//
// Parameters:
//   - menus: Slice of menu domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated MenuCollection with metadata and navigation links
func NewPaginatedMenuCollection(menus []*entities.Menu, page, perPage, total int, baseURL string) MenuCollection {
	collection := NewMenuCollection(menus)

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

	// Set pagination metadata
	collection.Meta = CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   int64(total),
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Build pagination navigation links
	collection.Links = CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	// Add previous page link if not on first page
	if page > 1 {
		collection.Links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	// Add next page link if not on last page
	if page < totalPages {
		collection.Links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return collection
}

// NewMenuCollectionResponse creates a new MenuCollectionResponse.
// This function wraps a MenuCollection in a standard API response format
// with appropriate response codes and success messages.
//
// Parameters:
//   - menus: Slice of menu domain entities to convert and wrap
//
// Returns:
//   - A new MenuCollectionResponse with success status and menu collection data
func NewMenuCollectionResponse(menus []*entities.Menu) MenuCollectionResponse {
	return MenuCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Menus retrieved successfully",
		Data:            NewMenuCollection(menus),
	}
}

// NewPaginatedMenuCollectionResponse creates a new MenuCollectionResponse with pagination.
// This function wraps a paginated MenuCollection in a standard API response format
// with appropriate response codes and success messages, including all pagination metadata.
//
// Parameters:
//   - menus: Slice of menu domain entities for the current page
//   - page: Current page number (1-based)
//   - perPage: Number of items per page
//   - total: Total number of items across all pages
//   - baseURL: Base URL for generating pagination links
//
// Returns:
//   - A new paginated MenuCollectionResponse with success status and pagination data
func NewPaginatedMenuCollectionResponse(menus []*entities.Menu, page, perPage, total int, baseURL string) MenuCollectionResponse {
	return MenuCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Menus retrieved successfully",
		Data:            NewPaginatedMenuCollection(menus, page, perPage, total, baseURL),
	}
}
