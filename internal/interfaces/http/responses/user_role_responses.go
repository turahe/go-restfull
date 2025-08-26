package responses

import (
	"github.com/turahe/go-restfull/internal/domain/entities"
)

// UserRoleResource represents a user-role relationship in API responses
type UserRoleResource struct {
	UserID string        `json:"user_id"`
	RoleID string        `json:"role_id"`
	User   *UserResource `json:"user,omitempty"`
	Role   *RoleResource `json:"role,omitempty"`
}

// UserRoleCollection represents a collection of user-role relationships
type UserRoleCollection struct {
	Data  []UserRoleResource `json:"data"`
	Meta  CollectionMeta     `json:"meta"`
	Links CollectionLinks    `json:"links"`
}

// UserRoleResourceResponse represents a single user-role response
type UserRoleResourceResponse struct {
	ResponseCode    int              `json:"response_code"`
	ResponseMessage string           `json:"response_message"`
	Data            UserRoleResource `json:"data"`
}

// UserRoleCollectionResponse represents a collection of user-role relationships response
type UserRoleCollectionResponse struct {
	ResponseCode    int                `json:"response_code"`
	ResponseMessage string             `json:"response_message"`
	Data            UserRoleCollection `json:"data"`
}

// RoleUserResource represents a role-user relationship in API responses
type RoleUserResource struct {
	RoleID string        `json:"role_id"`
	UserID string        `json:"user_id"`
	Role   *RoleResource `json:"role,omitempty"`
	User   *UserResource `json:"user,omitempty"`
}

// RoleUserCollection represents a collection of role-user relationships
type RoleUserCollection struct {
	Data  []RoleUserResource `json:"data"`
	Meta  CollectionMeta     `json:"meta"`
	Links CollectionLinks    `json:"links"`
}

// RoleUserCollectionResponse represents a collection of role-user relationships response
type RoleUserCollectionResponse struct {
	ResponseCode    int                `json:"response_code"`
	ResponseMessage string             `json:"response_message"`
	Data            RoleUserCollection `json:"data"`
}

// NewUserRoleResource creates a new UserRoleResource from user and role entities
func NewUserRoleResource(userID, roleID string, user *entities.User, role *entities.Role) UserRoleResource {
	resource := UserRoleResource{
		UserID: userID,
		RoleID: roleID,
	}

	if user != nil {
		resource.User = &UserResource{
			ID:       user.ID.String(),
			Username: user.UserName,
			Email:    user.Email,
			Phone:    user.Phone,
		}
	}

	if role != nil {
		resource.Role = &RoleResource{
			ID:          role.ID.String(),
			Name:        role.Name,
			Slug:        role.Slug,
			Description: role.Description,
			IsActive:    role.IsActive,
		}
	}

	return resource
}

// NewUserRoleCollection creates a new UserRoleCollection
func NewUserRoleCollection(userRoles []UserRoleResource) UserRoleCollection {
	return UserRoleCollection{
		Data: userRoles,
	}
}

// NewUserRoleResourceResponse creates a new UserRoleResourceResponse
func NewUserRoleResourceResponse(userID, roleID string, user *entities.User, role *entities.Role) UserRoleResourceResponse {
	return UserRoleResourceResponse{
		ResponseCode:    200,
		ResponseMessage: "User role operation successful",
		Data:            NewUserRoleResource(userID, roleID, user, nil),
	}
}

// NewUserRoleCollectionResponse creates a new UserRoleCollectionResponse
func NewUserRoleCollectionResponse(userRoles []UserRoleResource) UserRoleCollectionResponse {
	return UserRoleCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "User roles retrieved successfully",
		Data:            NewUserRoleCollection(userRoles),
	}
}

// NewRoleUserResource creates a new RoleUserResource from role and user entities
func NewRoleUserResource(roleID, userID string, role *entities.Role, user *entities.User) RoleUserResource {
	resource := RoleUserResource{
		RoleID: roleID,
		UserID: userID,
	}

	if role != nil {
		resource.Role = &RoleResource{
			ID:          role.ID.String(),
			Name:        role.Name,
			Slug:        role.Slug,
			Description: role.Description,
			IsActive:    role.IsActive,
		}
	}

	if user != nil {
		resource.User = &UserResource{
			ID:       user.ID.String(),
			Username: user.UserName,
			Email:    user.Email,
			Phone:    user.Phone,
		}
	}

	return resource
}

// NewRoleUserCollection creates a new RoleUserCollection
func NewRoleUserCollection(roleUsers []RoleUserResource) RoleUserCollection {
	return RoleUserCollection{
		Data: roleUsers,
	}
}

// NewRoleUserCollectionResponse creates a new RoleUserCollectionResponse
func NewRoleUserCollectionResponse(roleUsers []RoleUserResource) RoleUserCollectionResponse {
	return RoleUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Role users retrieved successfully",
		Data:            NewRoleUserCollection(roleUsers),
	}
}

// NewPaginatedRoleUserCollection creates a new RoleUserCollection with pagination
func NewPaginatedRoleUserCollection(roleUsers []RoleUserResource, page, perPage, total int, baseURL string) RoleUserCollection {
	collection := NewRoleUserCollection(roleUsers)

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

// NewPaginatedRoleUserCollectionResponse creates a new RoleUserCollectionResponse with pagination
func NewPaginatedRoleUserCollectionResponse(roleUsers []RoleUserResource, page, perPage, total int, baseURL string) RoleUserCollectionResponse {
	return RoleUserCollectionResponse{
		ResponseCode:    200,
		ResponseMessage: "Role users retrieved successfully",
		Data:            NewPaginatedRoleUserCollection(roleUsers, page, perPage, total, baseURL),
	}
}
