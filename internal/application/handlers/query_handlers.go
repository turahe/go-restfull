package handlers

import (
	"context"

	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/domain/aggregates"
	"github.com/turahe/go-restfull/internal/domain/valueobjects"
)

// QueryHandler represents a generic query handler interface
type QueryHandler[TQuery any, TResult any] interface {
	Handle(ctx context.Context, query TQuery) (TResult, error)
}

// PaginatedResult represents a paginated query result
type PaginatedResult[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// UserQueryHandlers defines all user query handlers
type UserQueryHandlers struct {
	GetUserByID       QueryHandler[queries.GetUserByIDQuery, *aggregates.UserAggregate]
	GetUserByEmail    QueryHandler[queries.GetUserByEmailQuery, *aggregates.UserAggregate]
	GetUserByUsername QueryHandler[queries.GetUserByUsernameQuery, *aggregates.UserAggregate]
	GetUserByPhone    QueryHandler[queries.GetUserByPhoneQuery, *aggregates.UserAggregate]
	ListUsers         QueryHandler[queries.ListUsersQuery, *PaginatedResult[*aggregates.UserAggregate]]
	GetUserRoles      QueryHandler[queries.GetUserRolesQuery, []valueobjects.Role]
	GetUserProfile    QueryHandler[queries.GetUserProfileQuery, *valueobjects.UserProfile]
	SearchUsers       QueryHandler[queries.SearchUsersQuery, *PaginatedResult[*aggregates.UserAggregate]]
}

// GetUserByIDQueryHandler handles get user by ID queries
type GetUserByIDQueryHandler interface {
	QueryHandler[queries.GetUserByIDQuery, *aggregates.UserAggregate]
}

// GetUserByEmailQueryHandler handles get user by email queries
type GetUserByEmailQueryHandler interface {
	QueryHandler[queries.GetUserByEmailQuery, *aggregates.UserAggregate]
}

// GetUserByUsernameQueryHandler handles get user by username queries
type GetUserByUsernameQueryHandler interface {
	QueryHandler[queries.GetUserByUsernameQuery, *aggregates.UserAggregate]
}

// GetUserByPhoneQueryHandler handles get user by phone queries
type GetUserByPhoneQueryHandler interface {
	QueryHandler[queries.GetUserByPhoneQuery, *aggregates.UserAggregate]
}

// ListUsersQueryHandler handles list users queries
type ListUsersQueryHandler interface {
	QueryHandler[queries.ListUsersQuery, *PaginatedResult[*aggregates.UserAggregate]]
}

// GetUserRolesQueryHandler handles get user roles queries
type GetUserRolesQueryHandler interface {
	QueryHandler[queries.GetUserRolesQuery, []valueobjects.Role]
}

// GetUserProfileQueryHandler handles get user profile queries
type GetUserProfileQueryHandler interface {
	QueryHandler[queries.GetUserProfileQuery, *valueobjects.UserProfile]
}

// SearchUsersQueryHandler handles search users queries
type SearchUsersQueryHandler interface {
	QueryHandler[queries.SearchUsersQuery, *PaginatedResult[*aggregates.UserAggregate]]
}