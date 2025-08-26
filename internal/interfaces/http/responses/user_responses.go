package responses

import (
	"time"

	"github.com/turahe/go-restfull/internal/domain/entities"
)

// UserResource represents a single user in API responses
// Following Laravel API Resource pattern for consistent formatting
type UserResource struct {
	ID              string         `json:"id"`
	Username        string         `json:"username"`
	Email           string         `json:"email"`
	Phone           string         `json:"phone"`
	Avatar          *string        `json:"avatar,omitempty"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time     `json:"phone_verified_at,omitempty"`
	IsEmailVerified bool           `json:"is_email_verified"`
	IsPhoneVerified bool           `json:"is_phone_verified"`
	HasAvatar       bool           `json:"has_avatar"`
	Roles           []RoleResource `json:"roles,omitempty"`
	Menus           []MenuResource `json:"menus,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}

// RoleResource is defined in role_responses.go to avoid duplication

// MenuResource represents a menu in user responses
type MenuResource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Order       *uint64   `json:"order,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserCollection represents a collection of users
// Following Laravel API Resource Collection pattern
type UserCollection struct {
	Data  []UserResource   `json:"data"`
	Meta  *CollectionMeta  `json:"meta,omitempty"`
	Links *CollectionLinks `json:"links,omitempty"`
}

// UserResourceResponse represents a single user response with Laravel-style formatting
type UserResourceResponse struct {
	Status string       `json:"status"`
	Data   UserResource `json:"data"`
}

// UserCollectionResponse represents a collection response with Laravel-style formatting
type UserCollectionResponse struct {
	Status string         `json:"status"`
	Data   UserCollection `json:"data"`
}

// NewUserResource creates a new UserResource from a User entity
// This transforms the domain entity into a consistent API response format
func NewUserResource(user *entities.User) *UserResource {
	var avatar *string
	if user.Avatar != "" {
		avatar = &user.Avatar
	}

	// Transform roles
	var roles []RoleResource
	if user.Roles != nil {
		roles = make([]RoleResource, len(user.Roles))
		for i, role := range user.Roles {
			roles[i] = NewRoleResource(role)
		}
	}

	// Transform menus
	var menus []MenuResource
	if user.Menus != nil {
		menus = make([]MenuResource, len(user.Menus))
		for i, menu := range user.Menus {
			var parentID *string
			if menu.ParentID != nil {
				parentIDStr := menu.ParentID.String()
				parentID = &parentIDStr
			}

			menus[i] = MenuResource{
				ID:          menu.ID.String(),
				Name:        menu.Name,
				Slug:        menu.Slug,
				Description: menu.Description,
				ParentID:    parentID,
				Order:       menu.RecordOrdering,
				CreatedAt:   menu.CreatedAt,
				UpdatedAt:   menu.UpdatedAt,
			}
		}
	}

	return &UserResource{
		ID:              user.ID.String(),
		Username:        user.UserName,
		Email:           user.Email,
		Phone:           user.Phone,
		Avatar:          avatar,
		EmailVerifiedAt: user.EmailVerifiedAt,
		PhoneVerifiedAt: user.PhoneVerifiedAt,
		IsEmailVerified: user.EmailVerifiedAt != nil,
		IsPhoneVerified: user.PhoneVerifiedAt != nil,
		HasAvatar:       user.Avatar != "",
		Roles:           roles,
		Menus:           menus,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		DeletedAt:       user.DeletedAt,
	}
}

// NewUserCollection creates a new UserCollection from a slice of User entities
// This transforms multiple domain entities into a consistent API response format
func NewUserCollection(users []*entities.User) *UserCollection {
	userResources := make([]UserResource, len(users))
	for i, user := range users {
		userResources[i] = *NewUserResource(user)
	}

	return &UserCollection{
		Data: userResources,
	}
}

// NewPaginatedUserCollection creates a new UserCollection with pagination metadata
// This follows Laravel's paginated resource collection pattern
func NewPaginatedUserCollection(
	users []*entities.User,
	page, perPage int,
	total int64,
	baseURL string,
) *UserCollection {
	collection := NewUserCollection(users)

	totalPages := (int(total) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	from := (page-1)*perPage + 1
	to := page * perPage
	if to > int(total) {
		to = int(total)
	}

	collection.Meta = &CollectionMeta{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   total,
		TotalPages:   totalPages,
		HasNextPage:  page < totalPages,
		HasPrevPage:  page > 1,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		From:         from,
		To:           to,
	}

	// Generate pagination links
	collection.Links = &CollectionLinks{
		First: generatePageURL(baseURL, 1),
		Last:  generatePageURL(baseURL, totalPages),
		Prev:  generatePageURL(baseURL, page-1),
		Next:  generatePageURL(baseURL, page+1),
	}

	return collection
}

// NewUserResourceResponse creates a new single user response
func NewUserResourceResponse(user *entities.User) *UserResourceResponse {
	return &UserResourceResponse{
		Status: "success",
		Data:   *NewUserResource(user),
	}
}

// NewUserCollectionResponse creates a new user collection response
func NewUserCollectionResponse(users []*entities.User) *UserCollectionResponse {
	return &UserCollectionResponse{
		Status: "success",
		Data:   *NewUserCollection(users),
	}
}

// NewPaginatedUserCollectionResponse creates a new paginated user collection response
func NewPaginatedUserCollectionResponse(
	users []*entities.User,
	page, perPage int,
	total int64,
	baseURL string,
) *UserCollectionResponse {
	return &UserCollectionResponse{
		Status: "success",
		Data:   *NewPaginatedUserCollection(users, page, perPage, total, baseURL),
	}
}
