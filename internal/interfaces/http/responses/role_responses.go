package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// RoleResource represents a role in API responses
type RoleResource struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedBy   string  `json:"created_by"`
	UpdatedBy   string  `json:"updated_by"`
	DeletedBy   *string `json:"deleted_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`

	// Computed fields
	IsDeleted    bool `json:"is_deleted"`
	IsActiveRole bool `json:"is_active_role"`
}

// RoleCollection represents a collection of roles
type RoleCollection struct {
	Data  []RoleResource  `json:"data"`
	Meta  CollectionMeta  `json:"meta"`
	Links CollectionLinks `json:"links"`
}

// RoleResourceResponse represents a single role response
type RoleResourceResponse struct {
	ResponseCode    int          `json:"response_code"`
	ResponseMessage string       `json:"response_message"`
	Data            RoleResource `json:"data"`
}

// RoleCollectionResponse represents a collection of roles response
type RoleCollectionResponse struct {
	ResponseCode    int            `json:"response_code"`
	ResponseMessage string         `json:"response_message"`
	Data            RoleCollection `json:"data"`
}

// NewRoleResource creates a new RoleResource from role entity
func NewRoleResource(role *entities.Role) RoleResource {
	resource := RoleResource{
		ID:          role.ID.String(),
		Name:        role.Name,
		Slug:        role.Slug,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedBy:   role.CreatedBy.String(),
		UpdatedBy:   role.UpdatedBy.String(),
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),

		// Computed fields
		IsDeleted:    role.IsDeleted(),
		IsActiveRole: role.IsActiveRole(),
	}

	// Set optional fields
	if role.DeletedBy != nil {
		deletedBy := role.DeletedBy.String()
		resource.DeletedBy = &deletedBy
	}

	if role.DeletedAt != nil {
		deletedAt := role.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resource.DeletedAt = &deletedAt
	}

	return resource
}

// NewRoleResourceResponse creates a new RoleResourceResponse
func NewRoleResourceResponse(role *entities.Role) RoleResourceResponse {
	return RoleResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "Role retrieved successfully",
		Data:            NewRoleResource(role),
	}
}

// NewRoleCollection creates a new RoleCollection
func NewRoleCollection(roles []*entities.Role) RoleCollection {
	roleResources := make([]RoleResource, len(roles))
	for i, role := range roles {
		roleResources[i] = NewRoleResource(role)
	}

	return RoleCollection{
		Data: roleResources,
	}
}

// NewPaginatedRoleCollection creates a new RoleCollection with pagination
func NewPaginatedRoleCollection(roles []*entities.Role, page, perPage, total int, baseURL string) RoleCollection {
	collection := NewRoleCollection(roles)

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

// NewRoleCollectionResponse creates a new RoleCollectionResponse
func NewRoleCollectionResponse(roles []*entities.Role) RoleCollectionResponse {
	return RoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewRoleCollection(roles),
	}
}

// NewPaginatedRoleCollectionResponse creates a new RoleCollectionResponse with pagination
func NewPaginatedRoleCollectionResponse(roles []*entities.Role, page, perPage, total int, baseURL string) RoleCollectionResponse {
	return RoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Roles retrieved successfully",
		Data:            NewPaginatedRoleCollection(roles, page, perPage, total, baseURL),
	}
}
