package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// MenuItemResource represents a comprehensive menu item in API responses
type MenuItemResource struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description,omitempty"`
	URL            string  `json:"url,omitempty"`
	Icon           string  `json:"icon,omitempty"`
	ParentID       *string `json:"parent_id,omitempty"`
	RecordLeft     *uint64 `json:"record_left,omitempty"`
	RecordRight    *uint64 `json:"record_right,omitempty"`
	RecordOrdering *uint64 `json:"record_ordering,omitempty"`
	RecordDepth    *uint64 `json:"record_depth,omitempty"`
	IsActive       bool    `json:"is_active"`
	IsVisible      bool    `json:"is_visible"`
	Target         string  `json:"target,omitempty"`
	CreatedBy      string  `json:"created_by"`
	UpdatedBy      string  `json:"updated_by"`
	DeletedBy      *string `json:"deleted_by,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	DeletedAt      *string `json:"deleted_at,omitempty"`

	// Computed fields
	IsDeleted bool  `json:"is_deleted"`
	IsRoot    bool  `json:"is_root"`
	IsLeaf    bool  `json:"is_leaf"`
	Depth     int   `json:"depth"`
	Width     int64 `json:"width"`

	// Nested resources
	Parent   *MenuItemResource  `json:"parent,omitempty"`
	Children []MenuItemResource `json:"children,omitempty"`
	Roles    []RoleResource     `json:"roles,omitempty"`
}

// MenuCollection represents a collection of menu items
type MenuCollection struct {
	Data  []MenuItemResource `json:"data"`
	Meta  CollectionMeta     `json:"meta"`
	Links CollectionLinks    `json:"links"`
}

// MenuResourceResponse represents a single menu item response
type MenuResourceResponse struct {
	ResponseCode    int              `json:"response_code"`
	ResponseMessage string           `json:"response_message"`
	Data            MenuItemResource `json:"data"`
}

// MenuCollectionResponse represents a collection of menu items response
type MenuCollectionResponse struct {
	ResponseCode    int            `json:"response_code"`
	ResponseMessage string         `json:"response_message"`
	Data            MenuCollection `json:"data"`
}

// NewMenuResource creates a new MenuItemResource from menu entity
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

		// Computed fields
		IsDeleted: menu.IsDeleted(),
		IsRoot:    menu.IsRoot(),
		IsLeaf:    menu.IsLeaf(),
		Depth:     menu.GetDepth(),
	}

	// Set optional fields
	if menu.ParentID != nil {
		parentID := menu.ParentID.String()
		resource.ParentID = &parentID
	}

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

	if menu.DeletedBy != nil {
		deletedBy := menu.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if menu.DeletedAt != nil {
		deletedAt := menu.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	// Calculate width if record boundaries are available
	if menu.RecordLeft != nil && menu.RecordRight != nil {
		resource.Width = menu.GetWidth()
	}

	// Set nested resources
	if menu.Parent != nil {
		parentResource := NewMenuResource(menu.Parent)
		resource.Parent = &parentResource
	}

	if len(menu.Children) > 0 {
		childrenResources := make([]MenuItemResource, len(menu.Children))
		for i, child := range menu.Children {
			childrenResources[i] = NewMenuResource(child)
		}
		resource.Children = childrenResources
	}

	if len(menu.Roles) > 0 {
		roleResources := make([]RoleResource, len(menu.Roles))
		for i, role := range menu.Roles {
			roleResources[i] = NewRoleResource(role)
		}
		resource.Roles = roleResources
	}

	return resource
}

// NewMenuResourceResponse creates a new MenuResourceResponse
func NewMenuResourceResponse(menu *entities.Menu) MenuResourceResponse {
	return MenuResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Menu retrieved successfully",
		Data:            NewMenuResource(menu),
	}
}

// NewMenuCollection creates a new MenuCollection
func NewMenuCollection(menus []*entities.Menu) MenuCollection {
	menuResources := make([]MenuItemResource, len(menus))
	for i, menu := range menus {
		menuResources[i] = NewMenuResource(menu)
	}

	return MenuCollection{
		Data: menuResources,
	}
}

// NewPaginatedMenuCollection creates a new MenuCollection with pagination
func NewPaginatedMenuCollection(menus []*entities.Menu, page, perPage, total int, baseURL string) MenuCollection {
	collection := NewMenuCollection(menus)

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	from := (page-1)*perPage + 1
	to := page * perPage
	if to > total {
		to = total
	}

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

	// Build pagination links
	collection.Links = CollectionLinks{
		First: buildPaginationLink(baseURL, 1, perPage),
		Last:  buildPaginationLink(baseURL, totalPages, perPage),
	}

	if page > 1 {
		collection.Links.Prev = buildPaginationLink(baseURL, page-1, perPage)
	}

	if page < totalPages {
		collection.Links.Next = buildPaginationLink(baseURL, page+1, perPage)
	}

	return collection
}

// NewMenuCollectionResponse creates a new MenuCollectionResponse
func NewMenuCollectionResponse(menus []*entities.Menu) MenuCollectionResponse {
	return MenuCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Menus retrieved successfully",
		Data:            NewMenuCollection(menus),
	}
}

// NewPaginatedMenuCollectionResponse creates a new MenuCollectionResponse with pagination
func NewPaginatedMenuCollectionResponse(menus []*entities.Menu, page, perPage, total int, baseURL string) MenuCollectionResponse {
	return MenuCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Menus retrieved successfully",
		Data:            NewPaginatedMenuCollection(menus, page, perPage, total, baseURL),
	}
}
