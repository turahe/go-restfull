# Laravel-Style Resource Responses

This document explains the implementation of Laravel-style API resource responses in the Go RESTful API project.

## Overview

The address controller has been updated to use Laravel-style resource responses, providing consistent and well-structured API responses that follow Laravel's API resource patterns.

## What Are Laravel API Resources?

Laravel API resources are a way to transform models into JSON responses with consistent formatting. They provide:

- **Consistent structure**: All responses follow the same format
- **Data transformation**: Domain entities are transformed into API-friendly formats
- **Pagination support**: Built-in pagination metadata and links
- **Flexibility**: Easy to customize response structure

## Implementation Details

### 1. Address Resource Structure

```go
type AddressResource struct {
    ID              string     `json:"id"`
    AddressableID   string     `json:"addressable_id"`
    AddressableType string     `json:"addressable_type"`
    AddressLine1    string     `json:"address_line1"`
    AddressLine2    *string    `json:"address_line2,omitempty"`
    City            string     `json:"city"`
    State           string     `json:"state"`
    PostalCode      string     `json:"postal_code"`
    Country         string     `json:"country"`
    Latitude        *float64   `json:"latitude,omitempty"`
    Longitude       *float64   `json:"longitude,omitempty"`
    IsPrimary       bool       `json:"is_primary"`
    AddressType     string     `json:"address_type"`
    FullAddress     string     `json:"full_address"`        // Computed field
    HasCoordinates  bool       `json:"has_coordinates"`    // Computed field
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
```

### 2. Collection Structure

```go
type AddressCollection struct {
    Data  []AddressResource `json:"data"`
    Meta  *CollectionMeta   `json:"meta,omitempty"`
    Links *CollectionLinks  `json:"links,omitempty"`
}
```

### 3. Pagination Metadata

```go
type CollectionMeta struct {
    CurrentPage  int   `json:"current_page"`
    PerPage      int   `json:"per_page"`
    TotalItems   int64 `json:"total_items"`
    TotalPages   int   `json:"total_pages"`
    HasNextPage  bool  `json:"has_next_page"`
    HasPrevPage  bool  `json:"has_prev_page"`
    NextPage     int   `json:"next_page,omitempty"`
    PreviousPage int   `json:"previous_page,omitempty"`
    From         int   `json:"from"`
    To           int   `json:"to"`
}
```

### 4. Pagination Links

```go
type CollectionLinks struct {
    First string `json:"first,omitempty"`
    Last  string `json:"last,omitempty"`
    Prev  string `json:"prev,omitempty"`
    Next  string `json:"next,omitempty"`
}
```

## Response Examples

### Single Address Response

```json
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "addressable_id": "550e8400-e29b-41d4-a716-446655440001",
    "addressable_type": "user",
    "address_line1": "123 Main St",
    "address_line2": "Apt 4B",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "USA",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "is_primary": true,
    "address_type": "home",
    "full_address": "123 Main St, Apt 4B, New York, NY 10001, USA",
    "has_coordinates": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Collection Response

```json
{
  "status": "success",
  "data": {
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "addressable_id": "550e8400-e29b-41d4-a716-446655440001",
        "addressable_type": "user",
        "address_line1": "123 Main St",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA",
        "is_primary": true,
        "address_type": "home",
        "full_address": "123 Main St, New York, NY 10001, USA",
        "has_coordinates": false,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ]
  }
}
```

### Paginated Collection Response

```json
{
  "status": "success",
  "data": {
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "addressable_id": "550e8400-e29b-41d4-a716-446655440001",
        "addressable_type": "user",
        "address_line1": "123 Main St",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA",
        "is_primary": true,
        "address_type": "home",
        "full_address": "123 Main St, New York, NY 10001, USA",
        "has_coordinates": false,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "meta": {
      "current_page": 1,
      "per_page": 10,
      "total_items": 25,
      "total_pages": 3,
      "has_next_page": true,
      "has_prev_page": false,
      "next_page": 2,
      "previous_page": 0,
      "from": 1,
      "to": 10
    },
    "links": {
      "first": "/api/v1/addresses/search/city?city=New%20York",
      "last": "/api/v1/addresses/search/city?city=New%20York&page=3",
      "prev": "",
      "next": "/api/v1/addresses/search/city?city=New%20York&page=2"
    }
  }
}
```

## Usage in Controllers

### Single Resource Response

```go
// For single address operations (create, read, update)
return ctx.Status(http.StatusOK).JSON(responses.NewAddressResourceResponse(address))
```

### Collection Response

```go
// For multiple addresses without pagination
return ctx.Status(http.StatusOK).JSON(responses.NewAddressCollectionResponse(addresses))
```

### Paginated Collection Response

```go
// For search operations with pagination
return ctx.Status(http.StatusOK).JSON(responses.NewPaginatedAddressCollectionResponse(
    addresses, page, limit, total, baseURL,
))
```

## Benefits

1. **Consistency**: All address endpoints now return responses in the same format
2. **Computed Fields**: Additional fields like `full_address` and `has_coordinates` are automatically included
3. **Pagination**: Built-in pagination support with metadata and navigation links
4. **Maintainability**: Centralized response formatting makes it easier to modify response structure
5. **Developer Experience**: Frontend developers get predictable response structures
6. **Laravel Familiarity**: Developers familiar with Laravel will recognize the response patterns

## Migration from Old Response Format

The old response format used `SuccessResponse` with generic data:

```go
// Old format
return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
    Status: "success",
    Data:   address,
})

// New format
return ctx.Status(http.StatusOK).JSON(responses.NewAddressResourceResponse(address))
```

## Future Enhancements

1. **Conditional Fields**: Add support for including/excluding fields based on user permissions
2. **Resource Transformers**: Create reusable transformers for common entity types
3. **API Versioning**: Support different response formats for different API versions
4. **Caching**: Cache transformed resources for better performance
5. **Validation**: Add response validation to ensure consistency

## Files Modified

- `internal/interfaces/http/responses/address_responses.go` - New resource response structures
- `internal/interfaces/http/controllers/address_controller.go` - Updated to use new responses
- `docs/development/laravel-resource-responses.md` - This documentation file

## Testing

To test the new resource responses:

1. Run the application
2. Make requests to address endpoints
3. Verify that responses follow the new structure
4. Check that pagination works correctly for search endpoints
5. Ensure computed fields are present in responses

## Conclusion

The implementation of Laravel-style resource responses provides a more professional and consistent API experience. The structured responses make it easier for frontend developers to work with the API and provide better documentation through the response format itself.
