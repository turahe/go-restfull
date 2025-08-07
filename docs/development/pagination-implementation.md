# Pagination Service Implementation

This document explains how pagination has been implemented as a service layer in the application, following clean architecture principles.

## Overview

Pagination has been moved from the controller layer to the service layer to maintain better separation of concerns and make the pagination logic reusable across different services.

## Architecture

### 1. Pagination Service Interface

Located in `internal/domain/services/pagination_service.go`:

```go
type PaginationService interface {
    // CreatePaginatedResponse creates a standardized paginated response
    CreatePaginatedResponse(ctx context.Context, data interface{}, total int64, paginationParams *pagination.PaginationRequest) *response.PaginatedResult
    
    // ValidatePaginationParams validates and normalizes pagination parameters
    ValidatePaginationParams(params *pagination.PaginationRequest) *pagination.PaginationRequest
}
```

### 2. Service Layer Pagination Methods

Each service that needs pagination implements these methods:

#### User Service (`internal/application/services/user_service.go`)

```go
// GetUsersWithPagination retrieves users with pagination and returns total count
func (s *userService) GetUsersWithPagination(ctx context.Context, page, perPage int, search string) ([]*entities.User, int64, error)

// GetUsersCount returns total count of users (for pagination)
func (s *userService) GetUsersCount(ctx context.Context, search string) (int64, error)
```

#### Post Service (`internal/application/services/post_service.go`)

```go
// GetPostsWithPagination retrieves posts with pagination and returns total count
func (s *postService) GetPostsWithPagination(ctx context.Context, page, perPage int, search, status string) ([]*entities.Post, int64, error)

// GetPostsCount returns total count of posts (for pagination)
func (s *postService) GetPostsCount(ctx context.Context, search, status string) (int64, error)
```

### 3. Repository Layer Enhancements

#### User Repository

Added methods to `internal/domain/repositories/user_repository.go`:
```go
// CountBySearch returns the total number of users matching the search query
CountBySearch(ctx context.Context, query string) (int64, error)
```

Implemented in `internal/infrastructure/adapters/user_repository.go`:
```go
func (r *postgresUserRepository) CountBySearch(ctx context.Context, query string) (int64, error) {
    searchQuery := `
        SELECT COUNT(*) FROM users 
        WHERE deleted_at IS NULL 
        AND (username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1)
    `
    var count int64
    err := r.db.QueryRow(ctx, searchQuery, fmt.Sprintf("%%%s%%", query)).Scan(&count)
    return count, err
}
```

#### Post Repository

Added methods to `internal/domain/repositories/post_repository.go`:
```go
// CountBySearch returns the total number of posts matching the search query
CountBySearch(ctx context.Context, query string) (int64, error)

// CountBySearchPublished returns the total number of published posts matching the search query
CountBySearchPublished(ctx context.Context, query string) (int64, error)

// SearchPublished searches published posts by query
SearchPublished(ctx context.Context, query string, limit, offset int) ([]*entities.Post, error)
```

### 4. Controller Layer Simplification

Controllers now use the service layer pagination methods instead of handling pagination logic directly:

#### User Controller

```go
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
    // Get pagination parameters from middleware
    pagination := middleware.GetPaginationParams(ctx)

    // Use the service layer pagination method
    users, total, err := c.userService.GetUsersWithPagination(ctx.Context(), pagination.Page, pagination.PerPage, pagination.Search)
    if err != nil {
        return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
            Status:  "error",
            Message: err.Error(),
        })
    }

    // Convert to response DTOs
    userResponses := make([]responses.UserResponse, len(users))
    for i, user := range users {
        userResponses[i] = *responses.NewUserResponse(user)
    }

    // Create paginated response using pagination service
    paginatedResult := c.paginationService.CreatePaginatedResponse(ctx.Context(), userResponses, total, nil)

    return ctx.JSON(responses.SuccessResponse{
        Status: "success",
        Data:   paginatedResult,
    })
}
```

#### Post Controller

```go
func (c *PostController) GetPosts(ctx *fiber.Ctx) error {
    // Get pagination parameters from middleware
    pagination := middleware.GetPaginationParams(ctx)

    // Get additional filters
    status := ctx.Query("status", "")

    // Use the service layer pagination method
    posts, total, err := c.postService.GetPostsWithPagination(ctx.Context(), pagination.Page, pagination.PerPage, pagination.Search, status)
    if err != nil {
        return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
            Status:  "error",
            Message: err.Error(),
        })
    }

    // Convert to response DTOs and return paginated result
    // ... rest of the implementation
}
```

## Benefits of This Approach

### 1. **Separation of Concerns**
- Controllers focus on HTTP request/response handling
- Services handle business logic and pagination
- Repositories handle data access and counting

### 2. **Reusability**
- Pagination logic is centralized in the service layer
- Can be easily applied to other entities (comments, tags, etc.)
- Consistent pagination behavior across the application

### 3. **Testability**
- Service layer pagination methods can be unit tested independently
- Repository count methods can be tested separately
- Controllers become simpler and easier to test

### 4. **Performance**
- Accurate total counts from database queries
- Efficient pagination with proper offset/limit
- Caching can be implemented at the repository level

### 5. **Maintainability**
- Changes to pagination logic only require service layer updates
- Consistent API responses across all paginated endpoints
- Easy to add new pagination features (sorting, filtering, etc.)

## Usage Examples

### Basic Pagination
```http
GET /api/v1/users?page=1&per_page=10
```

### Pagination with Search
```http
GET /api/v1/users?page=1&per_page=10&search=john
```

### Pagination with Status Filter (Posts)
```http
GET /api/v1/posts?page=1&per_page=10&status=published&search=technology
```

## Response Format

All paginated responses follow a consistent format:

```json
{
  "status": "success",
  "data": {
    "data": [...],
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total_items": 100,
      "total_pages": 10,
      "has_next_page": true,
      "has_prev_page": false,
      "next_page": 2,
      "previous_page": null,
      "from": 1,
      "to": 10
    }
  }
}
```

## Future Enhancements

1. **Advanced Filtering**: Add support for multiple filter criteria
2. **Sorting**: Implement dynamic sorting by different fields
3. **Caching**: Add Redis caching for frequently accessed paginated results
4. **Cursor-based Pagination**: Implement cursor-based pagination for better performance with large datasets
5. **Export**: Add support for exporting paginated data to CSV/Excel

## Implementation Checklist

- [x] Create PaginationService interface
- [x] Implement PaginationServiceImpl
- [x] Add pagination methods to UserService
- [x] Add pagination methods to PostService
- [x] Enhance UserRepository with CountBySearch
- [x] Enhance PostRepository with CountBySearch, CountBySearchPublished, SearchPublished
- [x] Update UserController to use service pagination
- [x] Update PostController to use service pagination
- [x] Add PaginationService to dependency injection container
- [x] Test build and ensure all interfaces are properly implemented 