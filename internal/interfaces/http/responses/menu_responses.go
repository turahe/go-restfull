package responses

import (
	"webapi/internal/domain/entities"
)

// MenuResponse represents a menu in API responses
type MenuResponse struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Slug           string         `json:"slug"`
	Description    string         `json:"description,omitempty"`
	URL            string         `json:"url,omitempty"`
	Icon           string         `json:"icon,omitempty"`
	ParentID       *string        `json:"parent_id,omitempty"`
	RecordLeft     int64          `json:"record_left"`
	RecordRight    int64          `json:"record_right"`
	RecordOrdering int64          `json:"record_ordering"`
	IsActive       bool           `json:"is_active"`
	IsVisible      bool           `json:"is_visible"`
	Target         string         `json:"target,omitempty"`
	Children       []MenuResponse `json:"children,omitempty"`
}

// MenuListResponse represents a list of menus with pagination
type MenuListResponse struct {
	Menus []MenuResponse `json:"menus"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Page  int            `json:"page"`
}

// NewMenuResponse creates a new MenuResponse from menu entity
func NewMenuResponse(menu *entities.Menu) *MenuResponse {
	response := &MenuResponse{
		ID:             menu.ID.String(),
		Name:           menu.Name,
		Slug:           menu.Slug,
		Description:    menu.Description,
		URL:            menu.URL,
		Icon:           menu.Icon,
		RecordLeft:     menu.RecordLeft,
		RecordRight:    menu.RecordRight,
		RecordOrdering: menu.RecordOrdering,
		IsActive:       menu.IsActive,
		IsVisible:      menu.IsVisible,
		Target:         menu.Target,
	}

	if menu.ParentID != nil {
		parentID := menu.ParentID.String()
		response.ParentID = &parentID
	}

	return response
}

// NewMenuListResponse creates a new MenuListResponse from menu entities
func NewMenuListResponse(menus []*entities.Menu, total int64, limit, page int) *MenuListResponse {
	menuResponses := make([]MenuResponse, len(menus))
	for i, menu := range menus {
		menuResponses[i] = *NewMenuResponse(menu)
	}

	return &MenuListResponse{
		Menus: menuResponses,
		Total: total,
		Limit: limit,
		Page:  page,
	}
}

// BuildMenuTree builds a hierarchical menu tree from flat menu list
func BuildMenuTree(menus []*entities.Menu) []MenuResponse {
	menuMap := make(map[string]*MenuResponse)
	var rootMenus []MenuResponse

	// First pass: create all menu responses
	for _, menu := range menus {
		menuResponse := NewMenuResponse(menu)
		menuMap[menu.ID.String()] = menuResponse
	}

	// Second pass: build hierarchy
	for _, menu := range menus {
		menuResponse := menuMap[menu.ID.String()]

		if menu.ParentID == nil {
			// Root menu
			rootMenus = append(rootMenus, *menuResponse)
		} else {
			// Child menu
			if parent, exists := menuMap[menu.ParentID.String()]; exists {
				parent.Children = append(parent.Children, *menuResponse)
			}
		}
	}

	return rootMenus
}
