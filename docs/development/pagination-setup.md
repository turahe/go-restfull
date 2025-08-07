# Pagination Setup Guide

This document explains how to use the pagination system in the Go RESTful API.

## Overview

The pagination system provides a standardized way to handle paginated responses across all API endpoints. It includes:

- **Pagination Middleware**: Automatically parses pagination parameters from query strings
- **Pagination Helpers**: Helper functions to create paginated responses
- **Standardized Response Format**: Consistent pagination metadata in API responses

## Features

- ✅ **Page-based pagination** (page, per_page)
- ✅ **Search functionality** (search parameter)
- ✅ **Sorting** (sort_by, sort_desc)
- ✅ **Comprehensive metadata** (total_items, total_pages, has_next_page, etc.)
- ✅ **Reasonable limits** (max 100 items per page)
- ✅ **Backward compatibility** with existing pagination

## Usage

### 1. Query Parameters

All paginated endpoints accept these query parameters:

```
GET /api/v1/users?page=1&per_page=10&search=john&sort_by=created_at&sort_desc=true
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number (1-based) |
| `per_page` | int | 10 | Items per page (max 100) |
| `search` | string | "" | Search term |
| `sort_by` | string | "created_at" | Field to sort by |
| `sort_desc` | bool | true | Sort descending (true/false, 1/0) |

### 2. Response Format

Paginated responses follow this structure:

```json
{
  "status": "success",
  "data": {
    "data": [
      // Array of items
    ],
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

### 3. Controller Implementation

Here's how to implement pagination in a controller:

```go
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
    // Parse pagination parameters
    pageStr := ctx.Query("page", "1")
    perPageStr := ctx.Query("per_page", "10")
    search := ctx.Query("search", "")

    // Convert to integers with defaults
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }

    perPage, err := strconv.Atoi(perPageStr)
    if err != nil || perPage < 1 {
        perPage = 10
    }

    // Ensure reasonable limits
    if perPage > 100 {
        perPage = 100
    }

    offset := (page - 1) * perPage

    var users []*entities.User
    var total int64
    var err2 error

    if search != "" {
        users, err2 = c.userService.SearchUsers(ctx.Context(), search, perPage, offset)
        total = int64(len(users)) // In real implementation, get total count
    } else {
        users, err2 = c.userService.GetAllUsers(ctx.Context(), perPage, offset)
        total = int64(len(users)) // In real implementation, get total count
    }

    if err2 != nil {
        return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
            Status:  "error",
            Message: err2.Error(),
        })
    }

    // Convert to response DTOs
    userResponses := make([]responses.UserResponse, len(users))
    for i, user := range users {
        userResponses[i] = *responses.NewUserResponse(user)
    }

    // Create paginated response using helper
    paginatedResult := response.CreatePaginatedResult(userResponses, page, perPage, total)

    return ctx.JSON(responses.SuccessResponse{
        Status: "success",
        Data:   paginatedResult,
    })
}
```

### 4. Using Pagination Middleware

For automatic pagination parameter parsing, you can use the pagination middleware:

```go
// In routes
users := rbacProtected.Group("/users", middleware.PaginationMiddleware())
users.Get("/", userController.GetUsers)

// In controller
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
    pagination := middleware.GetPaginationParams(ctx)
    offset := middleware.GetOffset(ctx)
    
    // Use pagination.Page, pagination.PerPage, pagination.Search, etc.
    // Use offset for database queries
}
```

## Helper Functions

### CreatePaginatedResult

Creates a complete paginated response:

```go
paginatedResult := response.CreatePaginatedResult(data, page, perPage, totalItems)
```

### CreatePaginationResponse

Creates only the pagination metadata:

```go
pagination := response.CreatePaginationResponse(page, perPage, totalItems)
```

## Database Integration

For proper pagination, your repository methods should return both data and total count:

```go
// In repository interface
func (r *UserRepository) GetAllUsersWithPagination(ctx context.Context, limit, offset int) ([]entities.User, int64, error)

// In repository implementation
func (r *UserRepository) GetAllUsersWithPagination(ctx context.Context, limit, offset int) ([]entities.User, int64, error) {
    var users []entities.User
    var total int64
    
    // Get total count
    err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    // Get paginated data
    rows, err := r.db.Query(ctx, "SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    // Parse results...
    
    return users, total, nil
}
```

## Examples

### Basic Pagination
```
GET /api/v1/users?page=1&per_page=10
```

### Search with Pagination
```
GET /api/v1/users?page=1&per_page=20&search=john&sort_by=username&sort_desc=false
```

### Next Page
```
GET /api/v1/users?page=2&per_page=10
```

## Best Practices

1. **Always validate pagination parameters** - Ensure page and per_page are positive integers
2. **Set reasonable limits** - Maximum 100 items per page
3. **Use consistent sorting** - Default to created_at DESC for most endpoints
4. **Include total count** - Always return total_items for proper pagination
5. **Handle edge cases** - Empty results, invalid page numbers, etc.
6. **Use indexes** - Ensure database indexes on sort_by fields for performance

## Migration from Legacy Pagination

The system maintains backward compatibility with the existing pagination structure. You can gradually migrate endpoints to use the new pagination helpers while keeping the old ones working. 